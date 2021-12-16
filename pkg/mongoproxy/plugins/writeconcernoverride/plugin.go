package writeconcernoverride

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

const Name = "writeconcernoverride"

func init() {
	plugins.Register(func() plugins.Plugin {
		return &WriteconcernOverridePlugin{
			conf: WriteconcernOverridePluginConfig{},
		}
	})
}

type WriteconcernOverridePluginConfig struct {
	UpdateOverrides map[string]interface{} `bson:"updateOverride"`
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type WriteconcernOverridePlugin struct {
	conf WriteconcernOverridePluginConfig
}

func (p *WriteconcernOverridePlugin) Name() string { return Name }

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *WriteconcernOverridePlugin) Configure(d bson.D) error {
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
func (p *WriteconcernOverridePlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {

	switch cmd := r.Command.(type) {
	case *command.Update:
		if cmd.WriteConcern != nil {
			if writeConcern, ok := cmd.WriteConcern.W.(string); ok {
				override, ok := p.conf.UpdateOverrides[writeConcern]
				if ok {
					cmd.WriteConcern.W = override
				}
			}
		}
	}
	return next(ctx, r)
}
