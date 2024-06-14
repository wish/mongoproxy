package nohint

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

const (
	Name = "nohint"
)

func init() {
	plugins.Register(func() plugins.Plugin {
		return &NohintPlugin{}
	})
}

// This is a plugin that simply strips hints out
type NohintPlugin struct {
}

func (p *NohintPlugin) Name() string { return Name }

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *NohintPlugin) Configure(d bson.D) error {
	return nil
}

// Process is the function executed when a message is called in the pipeline.
func (p *NohintPlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	switch cmd := r.Command.(type) {

	case *command.Aggregate:
		cmd.Hint = nil

	case *command.Count:
		cmd.Hint = nil

	case *command.Delete:
		cmd.Hint = nil

	case *command.Find:
		cmd.Hint = nil

	case *command.Update:
		for i, u := range cmd.Updates {
			if u.Hint != nil {
				u.Hint = nil
				cmd.Updates[i] = u
			}
		}
	}
	return next(ctx, r)
}
