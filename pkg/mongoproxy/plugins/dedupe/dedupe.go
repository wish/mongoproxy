package dedupe

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/sync/singleflight"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

var (
	commandDedupeCounterVec = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_dedupe_command_total",
		Help: "The total number of deduplicated commands",
	}, []string{"db", "collection", "command", "readpref"})
)

const Name = "dedupe"

func init() {
	plugins.Register(func() plugins.Plugin {
		return &DedupePlugin{
			conf: DedupePluginConfig{},
		}
	})
}

type DedupePluginConfig struct {
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type DedupePlugin struct {
	conf DedupePluginConfig
	g    *singleflight.Group
}

func (p *DedupePlugin) Name() string { return Name }

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *DedupePlugin) Configure(d bson.D) error {
	p.g = &singleflight.Group{}

	return nil
}

// Process is the function executed when a message is called in the pipeline.
func (p *DedupePlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	switch cmd := r.Command.(type) {
	case *command.Find:
		readPref := command.GetCommandReadPreferenceMode(r.Command)
		useInflight := cmd.SingleBatch != nil && *cmd.SingleBatch && (readPref == readpref.SecondaryMode.String() || readPref == readpref.SecondaryPreferredMode.String() || readPref == readpref.NearestMode.String())

		if useInflight {
			deduped := true
			ch := p.g.DoChan(DedupeKey(cmd), func() (interface{}, error) {
				deduped = false
				return next(ctx, r)
			})

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case ret := <-ch:
				if deduped {
					commandDedupeCounterVec.WithLabelValues(cmd.Database, cmd.Collection, r.CommandName, command.GetCommandReadPreferenceMode(r.Command)).Inc()
				}
				if ret.Err != nil {
					return nil, ret.Err
				}
				return ret.Val.(bson.D), nil
			}
		}
	}

	return next(ctx, r)
}

// DedupeKey returns a string to use as the key for deduplication
func DedupeKey(f *command.Find) string {
	key := bson.D{{"k", bson.A{
		f.Common.Database,
		f.Collection,
		f.Filter,
		f.Limit,
		f.Skip,
		f.Sort,
		f.MaxTimeMS,
		f.Projection,
		f.Hint,
		f.Common.ReadPreference,
		f.Collation,
		f.AllowPartialResults,
		f.ReadConcern,
	}}}

	b, _ := bson.Marshal(key)
	return string(b)
}
