package mongoproxy

import (
	"context"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/config"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

func TestProxy(t *testing.T) {
	cfg := &config.Config{}

	if err := cfg.Load(); err != nil {
		t.Fatal(err)
	}

	proxy, err := NewProxy(nil, cfg)
	if err != nil {
		t.Fatal(err)
	}

	p := plugins.BuildPipeline([]plugins.Plugin{}, func(context.Context, *plugins.Request) (bson.D, error) {
		return bson.D{
			{"cursor", bson.D{{"id", 1}}},
		}, nil
	})

	// Test the proxy

	// Start with cursor handling
	t.Run("cursor", func(t *testing.T) {
		// Send a command to get a cursor response
		r := plugins.Request{
			CursorCache: proxy,
			CC:          plugins.NewClientConnection(),
			CommandName: "find",
			Command:     &command.Find{Collection: "foo"},
		}

		p(context.TODO(), &r)

	})
}
