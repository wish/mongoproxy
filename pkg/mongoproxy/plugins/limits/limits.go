package defaults

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/time/rate"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

var (
	streamDelayTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_limits_stream_delay_seconds_total",
		Help: "The total stream delay added in seconds",
	}, []string{"db", "collection", "command", "readpref"})
)

const (
	Name                = "limits"
	GetMoreRatelimitKey = "limit.getmore"
)

func init() {
	plugins.Register(func() plugins.Plugin {
		return &LimitsPlugin{
			conf: LimitsPluginConfig{
				BatchSizeLimit:         10000,
				GetMoreStreamRatelimit: 20000,
			},
		}
	})
}

type LimitsPluginConfig struct {
	// BatchSizeLimit defines a limit for a query's batch size
	BatchSizeLimit int32 `bson:"batchSizeLimit"`

	// GetMoreStreamRatelimit defines a ratelimit for any given getMore stream
	GetMoreStreamRatelimit int `bson:"getMoreStreamRatelimit"`
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type LimitsPlugin struct {
	conf LimitsPluginConfig
}

func (p *LimitsPlugin) Name() string { return Name }

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *LimitsPlugin) Configure(d bson.D) error {
	// Load config
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&p.conf); err != nil {
		return err
	}

	return nil
}

// Process is the function executed when a message is called in the pipeline.
func (p *LimitsPlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	switch cmd := r.Command.(type) {

	case *command.Find:
		if cmd.BatchSize != nil {
			if *cmd.BatchSize > p.conf.BatchSizeLimit {
				return nil, fmt.Errorf("limit too high") // TODO: better bson error
			}
		} else {
			l := p.conf.BatchSizeLimit
			cmd.BatchSize = &l
		}

	case *command.GetMore:
		if cmd.BatchSize != nil {
			if *cmd.BatchSize > p.conf.BatchSizeLimit {
				return nil, fmt.Errorf("limit too high") // TODO: better bson error
			}
		} else {
			l := p.conf.BatchSizeLimit
			cmd.BatchSize = &l
		}

		// Now we want to ratelimit based on the batch size we could get instead of what we actually get.
		// This means that for the last batch (that isn't full) we may over-ratelimit, but we now enforce
		// the ratelimit before the work is done by the downstream mongo cluster
		l, ok := r.CC.Map[GetMoreRatelimitKey]
		if !ok {
			l = rate.NewLimiter(rate.Limit(p.conf.GetMoreStreamRatelimit), int(p.conf.GetMoreStreamRatelimit))
			r.CC.Map[GetMoreRatelimitKey] = l
		}

		waitStart := time.Now()
		if err := l.(*rate.Limiter).WaitN(ctx, int(*cmd.BatchSize)); err != nil {
			return nil, err
		}
		waitDuration := time.Since(waitStart)
		if waitDuration > time.Millisecond {
			streamDelayTotal.WithLabelValues(cmd.Database, cmd.Collection, r.CommandName, command.GetCommandReadPreferenceMode(r.Command)).Add(float64(waitDuration.Seconds()))
		}

	}
	return next(ctx, r)
}
