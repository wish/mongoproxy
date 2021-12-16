package integrationtest

import (
	"context"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/wish/mongoproxy/pkg/mongoproxy"
	"github.com/wish/mongoproxy/pkg/mongoproxy/config"
)

var (
	ctx       = context.Background()
	proxyPort = 27016
	proxyURI  = fmt.Sprintf("mongodb://localhost:%d/test", proxyPort)
	logLevel  = "debug"
	mongoAddr = "mongodb://localhost:27017"
)

func init() {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Fatalf("Unknown log level %s: %v", logLevel, err)
	}
	logrus.SetLevel(level)

	if v, ok := os.LookupEnv("MONGO_ADDR"); ok {
		fmt.Println("Overriding MONGOADDR in tests")
		mongoAddr = v
	}
	fmt.Println("addr", mongoAddr)
}

func SetupProxy(t *testing.T, cfg *config.Config) (*mongoproxy.Proxy, func()) {
	l, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", proxyPort))
	if err != nil {
		t.Fatal(err)
	}

	if cfg == nil {
		c := config.DefaultConfig
		cfg = &c
	}
	cfg.BindAddr = l.Addr().String() // Use port from listener
	cfg.Plugins = append(cfg.Plugins, config.PluginConfig{
		Name: "mongo",
		Config: bson.D{
			{"connectTimeout", "100ms"},
			{"mongoAddr", mongoAddr},
		},
	})

	proxy, err := mongoproxy.NewProxy(l, cfg)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		if err := proxy.Serve(); err != nil && err != mongoproxy.ErrServerClosed {
			logrus.Error(err)
		}
	}()
	time.Sleep(time.Millisecond * 10) // TODO: wait?

	return proxy, func() {
		proxy.Shutdown(context.TODO())
		l.Close()
	}
}

func SetupClient(t *testing.T, clientOpts ...*options.ClientOptions) *mongo.Client {
	// Base options should only use ApplyURI. The full set should have the user-supplied options after uriOpts so they
	// will win out in the case of conflicts.
	uriOpts := options.Client().ApplyURI(proxyURI)
	allClientOpts := append([]*options.ClientOptions{uriOpts}, clientOpts...)

	client, err := mongo.Connect(ctx, allClientOpts...)
	assert.Nil(t, err)

	// Call Ping with a low timeout to ensure the cluster is running and fail-fast if not.
	pingCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	err = client.Ping(pingCtx, nil)
	if err != nil {
		// Clean up in failure cases.
		_ = client.Disconnect(ctx)

		// Use t.Fatalf instead of assert because we want to fail fast if the cluster is down.
		t.Fatalf("error pinging cluster: %v", err)
	}

	return client
}
