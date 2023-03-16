package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/description"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/operation"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoerror"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"

	"github.com/wish/discovery"
)

var (
	commandSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "mongoproxy_plugins_mongo_command_duration_seconds",
		Help:       "The duration of mongo commands",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 1.0: 0.0},
		MaxAge:     time.Minute,
	}, []string{"db", "collection", "command", "readpref"})
	commandInflight = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mongoproxy_plugins_mongo_command_inflight",
		Help: "The duration of mongo commands",
	}, []string{"db", "collection", "command", "readpref"})
	commandReceiveBytes = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_mongo_command_receive_bytes_total",
		Help: "The total number of bytes received from downstream",
	}, []string{"db", "collection", "command", "readpref"})
	mongoDiscoveryUpdate = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_mongo_discovery_update",
		Help: "The total number of updates from discovery",
	}, []string{"status"})
)

type contextKey string

func (c contextKey) String() string {
	return "mongo context key " + string(c)
}

var (
	contextKeyServer = contextKey("mongo.server")
)

const Name = "mongo"

func init() {
	plugins.Register(func() plugins.Plugin {
		return &MongoPlugin{}
	})
}

type MongoPluginConfig struct {
	MongoAddr string `bson:"mongoAddr"`
	// Default 30s
	ConnectTimeout *string `bson:"connectTimeout"`
	// zstd,zlib,snappy
	Compressors []string `bson:"compressors"`
	// Default 10s
	HeartbeatInterval *string `bson:"heartbeatInterval"`
	// Default is 0 (indefinite)
	MaxConnIdleTime *string `bson:"maxConnIdleTime"`
	// Default is 100
	MaxPoolSize *uint64 `bson:"maxPoolSize"`
	// Default is 0
	MinPoolSize *uint64 `bson:"minPoolSize"`
	// Default 30s (waiting for available connection)
	ServerSelectionTimeout *string `bson:"serverSelectionTimeout"`
	// How long to wait for socket operations. Default is 0 (infinite)
	SocketTimeout *string `bson:"socketTimeout"`
	// EnableDNSDiscovery enables background resolution of the DNS results to set the host list of the mongo driver
	EnableDNSDiscovery bool `bson:"enableDNSDiscovery"`
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type MongoPlugin struct {
	conf MongoPluginConfig
	c    *mongo.Client
	t    *topology.Topology
}

func (p *MongoPlugin) Name() string { return Name }

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *MongoPlugin) Configure(d bson.D) error {
	// Load config
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&p.conf); err != nil {
		return err
	}

	// Do setup
	opts := &options.ClientOptions{
		PoolMonitor: &PoolMonitor,
		Monitor:     &CommandMonitor,
	}

	if p.conf.ConnectTimeout != nil {
		d, err := time.ParseDuration(*p.conf.ConnectTimeout)
		if err != nil {
			return err
		}
		opts.ConnectTimeout = &d
	}

	if p.conf.Compressors != nil {
		opts.Compressors = p.conf.Compressors
	}

	if p.conf.HeartbeatInterval != nil {
		d, err := time.ParseDuration(*p.conf.HeartbeatInterval)
		if err != nil {
			return err
		}
		opts.HeartbeatInterval = &d
	}

	if p.conf.MaxConnIdleTime != nil {
		d, err := time.ParseDuration(*p.conf.MaxConnIdleTime)
		if err != nil {
			return err
		}
		opts.MaxConnIdleTime = &d
	}

	if p.conf.MaxPoolSize != nil {
		opts.MaxPoolSize = p.conf.MaxPoolSize
	}

	if p.conf.MinPoolSize != nil {
		opts.MinPoolSize = p.conf.MinPoolSize
	}

	if p.conf.ServerSelectionTimeout != nil {
		d, err := time.ParseDuration(*p.conf.ServerSelectionTimeout)
		if err != nil {
			return err
		}
		opts.ServerSelectionTimeout = &d
	}

	if p.conf.SocketTimeout != nil {
		d, err := time.ParseDuration(*p.conf.SocketTimeout)
		if err != nil {
			return err
		}
		opts.SocketTimeout = &d
	}

	opts = opts.ApplyURI(p.conf.MongoAddr)
	// If we have EnableDNSDiscovery we will be overriding the IPs etc. but we want to continue
	// asking for the same ServerName
	if p.conf.EnableDNSDiscovery && opts.TLSConfig != nil {
		opts.TLSConfig.ServerName = strings.Split(opts.Hosts[0], ":")[0]
	}

	client, err := mongo.NewClient(opts)
	if err != nil {
		return err
	}

	if err := client.Connect(context.TODO()); err != nil {
		return err
	}

	p.c = client
	p.t = extractTopology(client)

	if p.conf.EnableDNSDiscovery {
		discoveryClient, err := discovery.NewDiscoveryFromEnv()
		if err != nil {
			return err
		}

		discoveryTarget := strings.TrimPrefix(p.conf.MongoAddr, "mongodb://")
		logrus.Debugf("discover target: %s", discoveryTarget)

		if err := discoveryClient.SubscribeServiceAddresses(context.TODO(), discoveryTarget, func(ctx context.Context, addrs discovery.ServiceAddresses) (err error) {
			start := time.Now()
			defer func() {
				logrus.Debugf("UpdateSessions completed in %s", time.Since(start))
				if err != nil {
					mongoDiscoveryUpdate.WithLabelValues("success").Inc()
				} else {
					mongoDiscoveryUpdate.WithLabelValues("failure").Inc()
				}
			}()

			// If we didn't get any addresses, don't change anything
			if len(addrs) <= 0 {
				return fmt.Errorf("no addresses found")
			}

			ips := make([]string, len(addrs))
			for i, addr := range addrs {
				ips[i] = fmt.Sprintf("%s:%d", addr.IP.String(), addr.Port)
			}

			if !p.t.ProcessSRVResults(ips) {
				return fmt.Errorf("error updating addresses")
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (p *MongoPlugin) runCommand(ctx context.Context, db string, cmd command.Command, server driver.Server) (bsoncore.Document, driver.Server, error) {
	runCmdDoc, err := bson.Marshal(cmd)
	if err != nil {
		return nil, nil, err
	}

	op := operation.NewCommand(runCmdDoc).
		Database(db).
		CommandMonitor(&CommandMonitor)

	if server != nil {
		op = op.Deployment(driver.SingleServerDeployment{Server: server})
	} else {
		// TODO:?
		readSelect := description.CompositeSelector([]description.ServerSelector{
			//description.ReadPrefSelector(ro.ReadPreference),
			//description.LatencySelector(db.client.localThreshold),
		})

		op = op.ServerSelector(readSelect).Deployment(p.t)
	}

	err = op.Execute(ctx)

	return op.Result(), extractServer(op), err
}

// Process is the function executed when a message is called in the pipeline.
func (p *MongoPlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	start := time.Now()

	labels := []string{
		command.GetCommandDatabase(r.Command),
		command.GetCommandCollection(r.Command),
		r.CommandName,
		command.GetCommandReadPreferenceMode(r.Command),
	}

	commandInflight.WithLabelValues(labels...).Inc()
	defer func() {
		commandInflight.WithLabelValues(labels...).Dec()
		commandSummary.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
	}()

	// Wrap handleCommand to output b/w metrics
	runCommand := func(ctx context.Context, db string, cmd command.Command, server driver.Server) (bson.D, error) {
		d, cmdServer, err := p.runCommand(ctx, db, cmd, server)
		commandReceiveBytes.WithLabelValues(labels...).Add(float64(len(d)))

		var result bson.D
		if unmarshalErr := bson.Unmarshal(d, &result); unmarshalErr != nil {
			return result, unmarshalErr
		}

		if err != nil {
			errDoc, err := ErrorToDoc(err)
			if err != nil {
				return result, err
			}
			if len(result) == 0 {
				result = append(result, bson.E{"ok", 0})
			}
			return append(result, errDoc...), err
		}

		// If we weren't passed in a server and we got a server in the response
		if server == nil && cmdServer != nil {
			// If we have a cursor in the response; store the mapping of ID -> server
			if cursorIDRaw, ok := bsonutil.Lookup(result, "cursor", "id"); ok {
				if cursorID, ok := cursorIDRaw.(int64); ok && cursorID > 0 {
					logrus.Tracef("Store cursor: %v %v", cursorID, cmdServer)
					// TODO: TTL from cmd
					r.CursorCache.CreateCursor(cursorID).Map[contextKeyServer] = cmdServer
				}
			}
		}

		return result, nil
	}

	switch cmd := r.Command.(type) {
	case *command.Aggregate:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.SaslStart:
		// Always return error; we don't want authn to happen at this layer as we don't keep
		// connections for various clients separated.
		return mongoerror.AuthenticationFailed.ErrMessage("Authentication failed."), nil

	case *command.Count:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.Create:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.CreateIndexes:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.CurrentOp:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.Delete:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.DeleteIndexes:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.Distinct:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.Explain:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.FindAndModify:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.FindAndModifyLegacy:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.Find:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		result, err := runCommand(ctx, dbName, cmd, nil)
		if err != nil {
			return nil, err
		}

		return result, nil

	case *command.CollStats:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.Drop:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.DropDatabase:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.DropIndexes:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.EndSessions:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.ListDatabases:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.ListCollections:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.ListIndexes:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.GetMore:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		// TODO: move into runCommand?
		v, ok := r.CursorCache.GetCursor(cmd.CursorID).Map[contextKeyServer]
		if !ok {
			return mongoerror.CursorNotFound.ErrMessage("Cursor not found."), nil
		}

		result, err := runCommand(ctx, dbName, cmd, v.(driver.Server))

		if cursorIDRaw, ok := bsonutil.Lookup(result, "cursor", "id"); ok {
			if cursorID, ok := cursorIDRaw.(int64); ok && cursorID == 0 {
				r.CursorCache.CloseCursor(cmd.CursorID)
			}
		}

		return result, err

	case *command.KillAllSessions:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.GetNonce:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	// TODO: clear cursor entry from cache?
	case *command.KillCursors:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		var (
			cursorsKilled   primitive.A
			cursorsNotFound primitive.A
			cursorsAlive    primitive.A
			cursorsUnknown  primitive.A
		)

		// TODO: optional based on the number of downstreams!
		// Right now we use the same cursorIDs as we got from mongos; we need to map those through
		for _, cursorIDRaw := range cmd.Cursors {
			cursorID, ok := cursorIDRaw.(int64)
			if !ok {
				return nil, fmt.Errorf("invalid cursorID")
			}
			v, ok := r.CursorCache.GetCursor(cursorID).Map[contextKeyServer]
			if !ok {
				return mongoerror.CursorNotFound.ErrMessage("Cursor not found."), nil
			}

			result, err := runCommand(ctx, dbName, cmd, v.(driver.Server))
			if err != nil || !bsonutil.Ok(result) {
				cursorsUnknown = append(cursorsUnknown, cursorID)
				continue
			}

			v, ok = bsonutil.Lookup(result, "cursorsKilled")
			if ok {
				cursorsKilled = append(cursorsKilled, v.(primitive.A)...)
				for _, cursorIDRaw := range cursorsKilled {
					fmt.Println("kill a cursor", cursorIDRaw)
					cursorID, ok := cursorIDRaw.(int64)
					if ok {
						r.CursorCache.CloseCursor(cursorID)
					}
				}
			}
			v, ok = bsonutil.Lookup(result, "cursorsNotFound")
			if ok {
				cursorsNotFound = append(cursorsNotFound, v.(primitive.A)...)
			}
			v, ok = bsonutil.Lookup(result, "cursorsAlive")
			if ok {
				cursorsAlive = append(cursorsAlive, v.(primitive.A)...)
			}
			v, ok = bsonutil.Lookup(result, "cursorsUnknown")
			if ok {
				cursorsUnknown = append(cursorsUnknown, v.(primitive.A)...)
			}
		}

		return bson.D{
			{"cursorsKilled", cursorsKilled},
			{"cursorsNotFound", cursorsNotFound},
			{"cursorsAlive", cursorsAlive},
			{"cursorsUnknown", cursorsUnknown},
			{"ok", 1},
			// TODO: operationTime and $clusterTime
		}, nil

	case *command.KillOp:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.MapReduce:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.Validate:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.Update:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		resp, err := runCommand(ctx, dbName, cmd, nil)

		if err != nil {
			resp, err = ErrorToDoc(err)
			if err != nil {
				return nil, err
			}
		}
		// TODO: better hack around the error condition; we need to always have nModified set
		// https://jira.mongodb.org/browse/SERVER-13210
		if resp != nil {
			found := false
			for _, e := range resp {
				if e.Key == "nModified" {
					found = true
					break
				}
			}
			if !found {
				resp = append(resp, primitive.E{Key: "nModified", Value: 0})
			}
		}

		return resp, err

	case *command.Insert:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	case *command.ShardCollection:
		// TODO: some other way to not double-send the DB
		dbName := cmd.Database
		cmd.Database = ""

		return runCommand(ctx, dbName, cmd, nil)

	}

	return next(ctx, r)

}
