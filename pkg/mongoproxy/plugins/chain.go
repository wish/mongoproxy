package plugins

import (
	"context"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	pluginSummary = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "mongoproxy_plugins_duration_seconds",
		Help:       "Summary of HandleMongo calls",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 1.0: 0.0},
		MaxAge:     time.Minute,
	}, []string{"i", "plugin", "status"})
)

type ChainFunc func(PipelineFunc) PipelineFunc

// BuildPipeline takes a plugin chain and creates a pipeline, returning
// a PipelineFunc that starts the pipeline when called.
func BuildPipeline(m []Plugin, base PipelineFunc) PipelineFunc {

	if len(m) == 0 {
		return base
	}
	pipeline := wrapPlugin(len(m)-1, m[len(m)-1])(base)
	for i := len(m) - 2; i >= 0; i-- {
		pipeline = wrapPlugin(i, m[i])(pipeline)
	}

	return pipeline
}

// wrapPlugin returns a closure ChainFunc that wraps over the plugin p, which
// can input and output PipelineFuncs to help with chaining.
func wrapPlugin(i int, p Plugin) ChainFunc {
	return ChainFunc(func(next PipelineFunc) PipelineFunc {
		return PipelineFunc(func(ctx context.Context, req *Request) (bson.D, error) {
			start := time.Now()
			d, err := p.Process(ctx, req, next)
			pluginSummary.WithLabelValues(strconv.Itoa(i), p.Name(), statusForErr(err)).Observe(time.Since(start).Seconds())
			return d, err
		})
	})
}

func statusForErr(err error) string {
	if err == nil {
		return "success"
	}
	return "error"
}
