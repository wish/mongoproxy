package discovery

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// ResolutionType is a DNS query type (SRV, A, AAAA)
type ResolutionType int

const (
	SRV ResolutionType = iota
	AAAA
	A
)

var (
	// DefaultConfig is the default discovery config
	DefaultConfig = DiscoveryConfig{
		ResolvConf:             "/etc/resolv.conf",
		ResolutionPriority:     []ResolutionType{SRV, AAAA, A},
		MinRetryInterval:       time.Second,
		SubscribeRetryInterval: time.Second,
	}
)

// DiscoveryConfig is the config for the discovery client
type DiscoveryConfig struct {
	// ResolvConf is the path to resolv.conf
	ResolvConf             string
	ResolutionPriority     []ResolutionType
	MinRetryInterval       time.Duration
	SubscribeRetryInterval time.Duration
}

const (
	// Environment Variable Names
	envResolvConfg            = "DISCOVERY_RESOLV_CONF"
	envResolutionPriority     = "DISCOVERY_RESOLUTION_PRIORITY"
	envMinRetryInterval       = "DISCOVERY_MIN_RETRY_INTERVAL"
	envSubscribeRetryInterval = "DISCOVERY_SUBSCRIBE_RETRY_INTERVAL"
)

func ConfigFromEnv() (*DiscoveryConfig, error) {
	c := DefaultConfig

	if v, ok := os.LookupEnv(envResolvConfg); ok {
		c.ResolvConf = v
	}

	if v, ok := os.LookupEnv(envResolutionPriority); ok {
		items := strings.Split(v, ",")
		types := make([]ResolutionType, len(items))
		for i, item := range items {
			var t ResolutionType
			switch item {
			case "SRV":
				t = SRV
			case "AAAA":
				t = AAAA
			case "A":
				t = A
			default:
				return nil, fmt.Errorf("Unknown ResolutionType: %s", item)
			}
			types[i] = t
		}
		c.ResolutionPriority = types
	}

	if v, ok := os.LookupEnv(envMinRetryInterval); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, err
		}
		c.MinRetryInterval = d
	}

	if v, ok := os.LookupEnv(envSubscribeRetryInterval); ok {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, err
		}
		c.SubscribeRetryInterval = d
	}

	return &c, nil
}
