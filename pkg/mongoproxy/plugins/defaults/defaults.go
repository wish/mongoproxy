package defaults

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

const Name = "defaults"

func init() {
	plugins.Register(func() plugins.Plugin {
		return &DefaultPlugin{
			conf: DefaultPluginConfig{},
		}
	})
}

type DefaultPluginConfig struct {
	DefaultReadConcern *command.ReadConcern `bson:"defaultReadConcern"`
	DefaultMaxTimeMS   *int64               `bson:"defaultMaxTimeMS"`
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type DefaultPlugin struct {
	conf DefaultPluginConfig
}

func (p *DefaultPlugin) Name() string { return Name }

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *DefaultPlugin) Configure(d bson.D) error {
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
func (p *DefaultPlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	switch cmd := r.Command.(type) {
	case *command.Aggregate:
		if p.conf.DefaultReadConcern != nil && cmd.ReadConcern == nil {
			tmp := *p.conf.DefaultReadConcern
			cmd.ReadConcern = &tmp
		}
		if p.conf.DefaultMaxTimeMS != nil && cmd.MaxTimeMS == nil {
			tmp := *p.conf.DefaultMaxTimeMS
			cmd.MaxTimeMS = &tmp
		}
	case *command.Count:
		if p.conf.DefaultReadConcern != nil && cmd.ReadConcern == nil {
			tmp := *p.conf.DefaultReadConcern
			cmd.ReadConcern = &tmp
		}
		if p.conf.DefaultMaxTimeMS != nil && cmd.MaxTimeMS == nil {
			tmp := *p.conf.DefaultMaxTimeMS
			cmd.MaxTimeMS = &tmp
		}
	case *command.Distinct:
		if p.conf.DefaultReadConcern != nil && cmd.ReadConcern == nil {
			tmp := *p.conf.DefaultReadConcern
			cmd.ReadConcern = &tmp
		}
		if p.conf.DefaultMaxTimeMS != nil && cmd.MaxTimeMS == nil {
			tmp := *p.conf.DefaultMaxTimeMS
			cmd.MaxTimeMS = &tmp
		}
	case *command.Find:
		if p.conf.DefaultReadConcern != nil && cmd.ReadConcern == nil {
			tmp := *p.conf.DefaultReadConcern
			cmd.ReadConcern = &tmp
		}
		if p.conf.DefaultMaxTimeMS != nil && cmd.MaxTimeMS == nil {
			tmp := *p.conf.DefaultMaxTimeMS
			cmd.MaxTimeMS = &tmp
		}

	}
	return next(ctx, r)
}
