package opentracing

import (
	"context"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/sirupsen/logrus"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

const (
	Name            = "opentracing"
	TRACE_ID_PREFIX = "open-trace-id"
)

func init() {
	plugins.Register(func() plugins.Plugin {
		return &OpentracingPlugin{
			conf: OpentracingPluginConfig{},
		}
	})
}

type OpentracingPluginConfig struct {
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type OpentracingPlugin struct {
	conf OpentracingPluginConfig

	tracer opentracing.Tracer
}

func (p *OpentracingPlugin) Name() string { return Name }

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *OpentracingPlugin) Configure(d bson.D) error {
	// Load config
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&p.conf); err != nil {
		return err
	}

	// TODO: config
	cfg, err := jaegercfg.FromEnv()
	if err != nil {
		return err
	}
	cfg.ServiceName = "mongoproxy"
	cfg.Headers = &jaeger.HeadersConfig{TraceContextHeaderName: TRACE_ID_PREFIX}

	tracer, _, err := cfg.NewTracer()
	if err != nil {
		return err
	}
	opentracing.SetGlobalTracer(tracer)

	p.tracer = tracer

	return nil
}

func (p *OpentracingPlugin) extractSpanFromComment(traceIDs map[string]struct{}, comment string) opentracing.SpanContext {
	var spanCtx opentracing.SpanContext

	// Extract span from comment
	if strings.HasPrefix(comment, TRACE_ID_PREFIX) {
		// Value is the remaining (until a space or end of string)
		idx := strings.Index(comment, " ")
		if idx == -1 {
			idx = len(comment)
		}

		id := comment[len(TRACE_ID_PREFIX)+1 : idx]
		if _, ok := traceIDs[id]; ok {
			return nil
		}
		traceIDs[id] = struct{}{}
		var err error
		spanCtx, err = p.tracer.Extract(opentracing.TextMap, opentracing.TextMapCarrier(map[string]string{
			TRACE_ID_PREFIX: id,
		}))
		if err != nil {
			logrus.Errorf("Error extracting trace info: %v", err)
		}
	}

	return spanCtx
}

func getString(d bson.D, keys ...string) string {
	vRaw, ok := bsonutil.Lookup(d, keys...)
	if !ok {
		return ""
	}

	v, ok := vRaw.(string)
	if !ok {
		return ""
	}
	return v
}

// Process is the function executed when a message is called in the pipeline.
func (p *OpentracingPlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	var (
		// As comments can be in a variety of places (filters, commands, etc.)
		// we want to check all comments for traceIDs; we'll start a span per
		// unique traceID that we find
		comments []string
		traceIDs = make(map[string]struct{})
	)
	switch cmd := r.Command.(type) {
	case *command.Aggregate:
		comments = []string{cmd.Comment}
	case *command.Distinct:
		comments = []string{getString(cmd.Query, "$comment")}
	case *command.Drop:
		comments = []string{cmd.Comment}
	case *command.DropDatabase:
		comments = []string{cmd.Comment}
	case *command.Explain:
		comments = []string{cmd.Comment}
	case *command.Find:
		comments = []string{getString(cmd.Filter, "$comment"), cmd.Comment}
	case *command.KillOp:
		comments = []string{cmd.Comment}
	case *command.MapReduce:
		comments = []string{cmd.Comment}
	case *command.Update:
		comments = make([]string, 0, len(cmd.Updates))
		for _, update := range cmd.Updates {
			comments = append(comments, getString(update.Query, "$comment"))
		}
	}

	var span opentracing.Span

	for _, comment := range comments {
		spanCtx := p.extractSpanFromComment(traceIDs, comment)
		if spanCtx != nil {
			s := p.tracer.StartSpan(r.CommandName, ext.RPCServerOption(spanCtx))
			s = s.SetTag("mongoproxy.database", command.GetCommandDatabase(r.Command))
			s = s.SetTag("mongoproxy.collection", command.GetCommandCollection(r.Command))

			defer s.Finish()
			if span == nil {
				span = s
				ctx = opentracing.ContextWithSpan(ctx, s)
			}
		}
	}

	return next(ctx, r)
}
