package filtercommand

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/mongoerror"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

// Name of the plugin
const Name = "filtercommand"

func init() {
	plugins.Register(func() plugins.Plugin {
		return &FilterCommandPlugin{
			conf: FilterCommandPluginConfig{},
		}
	})
}

type FilterCommandPluginConfig struct {
	// FilterCommands defines a list of commands to return `CommandNotFound` for
	FilterCommands []string `bson:"filterCommands"`
	filterCommands map[string]struct{}
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type FilterCommandPlugin struct {
	conf FilterCommandPluginConfig
}

func (p *FilterCommandPlugin) Name() string { return Name }

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *FilterCommandPlugin) Configure(d bson.D) error {
	// Load config
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&p.conf); err != nil {
		return err
	}

	p.conf.filterCommands = make(map[string]struct{})
	for _, f := range p.conf.FilterCommands {
		p.conf.filterCommands[f] = struct{}{}
	}

	return nil
}

// Process is the function executed when a message is called in the pipeline.
func (p *FilterCommandPlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	if _, ok := p.conf.filterCommands[r.CommandName]; ok {
		return mongoerror.CommandNotFound.ErrMessage("no such command: '" + r.CommandName + "'"), nil
	}
	return next(ctx, r)
}
