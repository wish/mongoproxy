package config

import (
	"fmt"
	"io/ioutil"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

// DefaultConfig is a base default config
var DefaultConfig = Config{
	RequestLengthLimit: 1024,
}

// ConfigFromFile loads a config (based on DefaultConfig) from the given path
func ConfigFromFile(path string) (*Config, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := DefaultConfig

	if err := bson.UnmarshalExtJSON(b, true, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Load(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Config is the configuration struct for mongoproxy
type Config struct {
	BindAddr    string         `bson:"bindAddr"`
	Plugins     []PluginConfig `bson:"plugins"`
	Compressors []string       `bson:"compressors"`
	// IdleCursorTimeoutMillis
	IdleCursorTimeoutMillis *string `bson:"idleCursorTimeoutMillis"`
	IdleCursorTimeout       time.Duration

	InternalIdentity *plugins.StaticIdentity `bson:"internalIdentity"`

	RequestLengthLimit int `bson:"requestLengthLimit"`
}

// Load will load all configuration
func (c *Config) Load() error {
	if c.IdleCursorTimeoutMillis != nil {
		d, err := time.ParseDuration(*c.IdleCursorTimeoutMillis)
		if err != nil {
			return err
		}
		c.IdleCursorTimeout = d
	} else {
		c.IdleCursorTimeout = time.Minute * 30 // Default timeout
	}

	return nil
}

// GetPlugins returns a list of plugin instances for the given config
func (c *Config) GetPlugins() ([]plugins.Plugin, error) {
	ps := make([]plugins.Plugin, len(c.Plugins))
	for i, config := range c.Plugins {
		p, ok := plugins.GetPlugin(config.Name)
		if !ok {
			return nil, fmt.Errorf("unknown plugin %s", config.Name)
		}
		if err := p.Configure(config.Config); err != nil {
			return nil, err
		}
		ps[i] = p
	}

	return ps, nil
}

type PluginConfig struct {
	Name   string `bson:"name"`
	Config bson.D `bson:"config"`
}
