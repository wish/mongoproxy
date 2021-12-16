package discovery

import (
	"context"
	"net"
	"time"
)

// Callback is a function which will be called with updated ServiceAddresses
type Callback func(context.Context, ServiceAddresses) error

// Discovery defines the interface for this package; this way if there are additional
// implementation in the future the interface will remain the same
type Discovery interface {
	// GetServiceAddresses will do the lookup and return the list in-order from DNS
	GetServiceAddresses(context.Context, string) (ServiceAddresses, error)
	// SubscribeServiceAddresses will do the lookup and give the results to Callback
	SubscribeServiceAddresses(context.Context, string, Callback) error
}

// ServiceAddress is a struct which contains the SD information about a given target
type ServiceAddress struct {
	Name     string
	IP       net.IP
	Port     uint16
	Priority uint16
	Weight   uint16

	expiresAt time.Time
	isStatic  bool
}

// IsExpired returns whether the ServiceAddress is stale.
// Note: if this was statically defined (hard-coded IP in the query) it will
// never expire
func (s ServiceAddress) IsExpired() bool {
	if s.isStatic {
		return false
	}
	return time.Now().After(s.expiresAt)
}

// Equal returns whether this address matches another one
func (s ServiceAddress) Equal(o ServiceAddress) bool {
	if s.Name != o.Name {
		return false
	}

	if s.Port != o.Port {
		return false
	}

	if s.Priority != o.Priority {
		return false
	}

	if s.Weight != o.Weight {
		return false
	}

	if !s.IP.Equal(o.IP) {
		return false
	}

	if s.isStatic != o.isStatic {
		return false
	}

	if !s.isStatic {
		if !s.expiresAt.Equal(o.expiresAt) {
			return false
		}
	}

	return true
}

// ServiceAddresses is a list of the information about targets, for the same service name
type ServiceAddresses []ServiceAddress

// Equal checks if the lists have the same addresses in them (regardless of sort order)
func (s ServiceAddresses) Equal(o ServiceAddresses) bool {
	if len(s) != len(o) {
		return false
	}

	us := make(map[string]ServiceAddress)
	for _, sa := range s {
		us[sa.Name] = sa
	}

	for _, sa := range o {
		ourSA, ok := us[sa.Name]
		if !ok {
			return false
		}

		if !ourSA.Equal(sa) {
			return false
		}
	}
	return true
}
