package schema

import (
	"context"
	"encoding/json"
	"github.com/cespare/xxhash/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"gopkg.in/fsnotify.v1"
	"io/ioutil"
	"log"
	"path"
	"sync/atomic"
	"time"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoerror"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

var (
	schemaUpdates = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_schema_updates_total",
		Help: "The total schema updates completed",
	}, []string{"success"})
	schemaVersion = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "mongoproxy_plugins_schema_config_hash",
		Help: "The current hash of the schema config file loaded",
	})

	schemaDeny = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_schema_deny_total",
		Help: "The total deny returns of a command",
	}, []string{"db", "collection", "command"})
)

const (
	Name = "schema"
)

func init() {
	plugins.Register(func() plugins.Plugin {
		return &SchemaPlugin{
			conf: SchemaPluginConfig{},
		}
	})
}

type SchemaPluginConfig struct {
	// SchemaPath is the path on disk to the schema file to load + watch for changes
	SchemaPath string `bson:"schemaPath"`
	// Log EnforceSchema errors
	EnforceSchemaLogOnly bool `bson:"enforceSchemaLogOnly"`
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type SchemaPlugin struct {
	conf SchemaPluginConfig

	s atomic.Value
}

func (p *SchemaPlugin) Name() string { return Name }

func (p *SchemaPlugin) GetSchema() *ClusterSchema {
	tmp := p.s.Load()
	if ret, ok := tmp.(*ClusterSchema); ok {
		return ret
	}
	return nil
}

func (p *SchemaPlugin) LoadSchema() (err error) {
	defer func() {
		if err != nil {
			schemaUpdates.WithLabelValues("false").Add(1)
		} else {
			schemaUpdates.WithLabelValues("true").Add(1)
			logrus.Infof("Schema update succeed")
		}
	}()
	b, err := ioutil.ReadFile(p.conf.SchemaPath)
	if err != nil {
		return err
	}

	var schema ClusterSchema
	if err := json.Unmarshal(b, &schema); err != nil {
		return err
	}

	p.s.Store(&schema)
	schemaVersion.Set(float64(xxhash.Sum64(b)))

	return nil
}

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *SchemaPlugin) Configure(d bson.D) error {
	// Load config
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&p.conf); err != nil {
		return err
	}

	// load schema
	if err := p.LoadSchema(); err != nil {
		return err
	}

	// start watch
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			p.LoadSchema()
			time.Sleep(5 * time.Minute)
		}
	}()

	if err := watcher.Add(path.Dir(p.conf.SchemaPath)); err != nil {
		return err
	}
	logrus.Infof("Schema watcher added path: %s", path.Dir(p.conf.SchemaPath))
	return nil
}

// Process is the function executed when a message is called in the pipeline.
func (p *SchemaPlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	switch cmd := r.Command.(type) {
	case *command.Insert:
		schema := p.GetSchema()
		for _, document := range cmd.Documents {
			if err := schema.ValidateInsert(ctx, cmd.Database, cmd.Collection, document); err != nil {
				schemaDeny.WithLabelValues(cmd.Database, cmd.Collection, r.CommandName).Inc()
				if !p.conf.EnforceSchemaLogOnly {
					return mongoerror.DocumentValidationFailure.ErrMessage(err.Error()), nil
				}
				logrus.Warningf("ENFORCE SCHEMA LOGONLY: %s, in db: %s, collection: %s, with cmd: %s",
					err.Error(), cmd.Database, cmd.Collection, r.CommandName)
			}
		}

	case *command.FindAndModify:
		if len(cmd.Update) > 0 {
			schema := p.GetSchema()
			logrus.Debugf("command findAndModify: %s", cmd.Update)
			if err := schema.ValidateUpdate(ctx, cmd.Database, cmd.Collection, cmd.Update, bsonutil.GetBoolDefault(cmd.Upsert, false)); err != nil {
				schemaDeny.WithLabelValues(cmd.Database, cmd.Collection, r.CommandName).Inc()
				if !p.conf.EnforceSchemaLogOnly {
					return mongoerror.DocumentValidationFailure.ErrMessage(err.Error()), nil
				}
				logrus.Warningf("ENFORCE SCHEMA LOGONLY: %s, in db: %s, collection: %s, with cmd: %s",
					err.Error(), cmd.Database, cmd.Collection, r.CommandName)
			}
		}

	case *command.Update:
		schema := p.GetSchema()
		for _, updateDoc := range cmd.Updates {
			logrus.Debugf("command Update wiht doc: %v", updateDoc)
			if err := schema.ValidateUpdate(ctx, cmd.Database, cmd.Collection, updateDoc.U, bsonutil.GetBoolDefault(updateDoc.Upsert, false)); err != nil {
				schemaDeny.WithLabelValues(cmd.Database, cmd.Collection, r.CommandName).Inc()
				if !p.conf.EnforceSchemaLogOnly {
					return mongoerror.DocumentValidationFailure.ErrMessage(err.Error()), nil
				}
				logrus.Warningf("ENFORCE SCHEMA LOGONLY: %s, in db: %s, collection: %s, with cmd: %s",
					err.Error(), cmd.Database, cmd.Collection, r.CommandName)
			}
		}
	}
	return next(ctx, r)
}