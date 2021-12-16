package defaults

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
	"github.com/wish/mongoproxy/pkg/mongowire"
)

var (
	slowlogTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_slowlog_logged_total",
		Help: "The total number of slow queries logged",
	}, []string{"db", "collection", "command", "readpref"})
)

const (
	Name = "slowlog"
)

func init() {
	plugins.Register(func() plugins.Plugin {
		return &SlowlogPlugin{
			conf: SlowlogPluginConfig{
				RequestLengthLimit: 1024,
			},
		}
	})
}

type SlowlogPluginConfig struct {
	SlowlogThreshold   string `bson:"slowlogThreshold"`
	thresholdDuration  time.Duration
	RequestLengthLimit int `bson:"requestLengthLimit"`
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type SlowlogPlugin struct {
	conf SlowlogPluginConfig
}

func (p *SlowlogPlugin) Name() string { return Name }

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *SlowlogPlugin) Configure(d bson.D) error {
	// Load config
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&p.conf); err != nil {
		return err
	}

	dur, err := time.ParseDuration(p.conf.SlowlogThreshold)
	if err != nil {
		return err
	}
	p.conf.thresholdDuration = dur

	return nil
}

// Process is the function executed when a message is called in the pipeline.
func (p *SlowlogPlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	start := time.Now()
	result, err := next(ctx, r)
	if took := time.Since(start); took > p.conf.thresholdDuration {
		slowlogTotal.WithLabelValues(
			command.GetCommandDatabase(r.Command),
			command.GetCommandCollection(r.Command),
			r.CommandName,
			command.GetCommandReadPreferenceMode(r.Command),
		).Inc()
		logrus.Infof("Slowlog: took=%s request=%s", took, mongowire.ToJson(r.Command, p.conf.RequestLengthLimit))
	}
	return result, err
}
