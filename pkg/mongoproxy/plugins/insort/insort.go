package insort

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

const Name = "insort"

func init() {
	plugins.Register(func() plugins.Plugin {
		return &InSortPlugin{}
	})
}

type InSortPluginConfig struct {
	InLimit int `bson:"inlimit"`
}

// This is a plugin that handles sending the request to the acutual downstream mongo
type InSortPlugin struct {
	conf InSortPluginConfig
}

func (p *InSortPlugin) Name() string { return Name }

// Configure configures this plugin with the given configuration object. Returns
// an error if the configuration is invalid for the plugin.
func (p *InSortPlugin) Configure(d bson.D) error {
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
func (p *InSortPlugin) Process(ctx context.Context, r *plugins.Request, next plugins.PipelineFunc) (bson.D, error) {
	switch cmd := r.Command.(type) {
	case *command.Find:
		if cmd.Filter != nil {
			if err := PreprocessFilter(cmd.Filter, p.conf.InLimit); err != nil {
				if bsonErr, ok := err.(*InLenError); ok {
					return bsonErr.BSONError(), nil
				}
				return nil, err
			}
		}
	case *command.FindAndModify:
		if err := PreprocessFilter(cmd.Query, p.conf.InLimit); err != nil {
			if bsonErr, ok := err.(*InLenError); ok {
				return bsonErr.BSONError(), nil
			}
			return nil, err
		}
	case *command.Update:
		for i := range cmd.Updates {
			if err := PreprocessFilter(cmd.Updates[i].Query, p.conf.InLimit); err != nil {
				if bsonErr, ok := err.(*InLenError); ok {
					return bsonErr.BSONError(), nil
				}
				return nil, err
			}
		}
	}

	return next(ctx, r)
}
