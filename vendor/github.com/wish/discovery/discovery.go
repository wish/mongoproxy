package discovery

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
	"github.com/sirupsen/logrus"
	"github.com/wish/discovery/resolver"
)

// NewDiscoveryFromEnv is a convenience method for creating a client from environment config
func NewDiscoveryFromEnv() (Discovery, error) {
	c, err := ConfigFromEnv()
	if err != nil {
		return nil, err
	}
	return NewDiscovery(*c)
}

// NewDiscovery returns a new discovery client based on the given config
func NewDiscovery(c DiscoveryConfig) (Discovery, error) {
	r, err := resolver.NewResolver(c.ResolvConf)
	if err != nil {
		return nil, err
	}
	return &discovery{c: c, r: r}, nil
}

type discovery struct {
	c DiscoveryConfig
	r *resolver.DNSResolver
}

// GetServiceAddresses will do the lookup and return the results from DNS
func (d *discovery) GetServiceAddresses(ctx context.Context, query string) (ServiceAddresses, error) {
	// Unfortunately there isn't a good mechanism to parse just the host/port section (which is the query we are expecting)
	// so we are going to add it and get a result
	u, err := url.Parse("notreal://" + query)
	if err != nil {
		return nil, err
	}

	// check if query is an IP already, is so return it
	if ip := net.ParseIP(u.Hostname()); ip != nil {
		portNum := uint16(0)
		if u.Port() != "" {
			portNumTmp, err := strconv.ParseUint(u.Port(), 10, 16)
			if err != nil {
				return nil, err
			}
			portNum = uint16(portNumTmp)
		}
		return ServiceAddresses{
			ServiceAddress{
				Name:     u.Hostname(),
				IP:       ip,
				Port:     portNum,
				isStatic: true,
			},
		}, nil
	}

	// At this point we then need to do the resolution in order as defined by config
	for _, resolutionType := range d.c.ResolutionPriority {
		var addrs ServiceAddresses
		var err error
		switch resolutionType {
		case SRV:
			addrs, err = d.getServiceAddressesSRV(ctx, u)
		case AAAA:
			addrs, err = d.getServiceAddressesAAAA(ctx, u)
		case A:
			addrs, err = d.getServiceAddressesA(ctx, u)
		default:
			return nil, fmt.Errorf("Unknown resolution type: %v", resolutionType)
		}

		// If any of the layers had an error, we'll return that error instead of continuing on.
		// Errors from these layers mean that we have tried all resolvers and we just get errors
		// from them, so asking the same resolvers for different things should get the same errors
		if err != nil {
			return nil, err
		}

		// If we got no addresses from the resolver, we will continue trying down the types
		if len(addrs) > 0 {
			return ShuffleServiceAddresses(addrs), nil
		}
	}

	// If we tried all resolutionTypes and got nothing we'll simply return nil, nil
	// since there was no error and we didn't find anything
	return nil, nil
}

func (d *discovery) getServiceAddressesSRV(ctx context.Context, u *url.URL) (ServiceAddresses, error) {
	m := &dns.Msg{}
	m.SetQuestion(dns.Fqdn(u.Hostname()), dns.TypeSRV)

	resp, err := d.r.ExchangeContext(ctx, m)
	if err != nil {
		return nil, err
	}
	if resp.MsgHdr.Rcode != dns.RcodeSuccess {
		return nil, errors.New(dns.RcodeToString[resp.MsgHdr.Rcode])
	}

	// map for name -> address mapping in case we need to get IPs for specific hosts
	nameToAddr := make(map[string]*ServiceAddress)
	addrs := make(ServiceAddresses, len(resp.Answer))

	now := time.Now()
	// Go over answers and make base entry
	for i, answer := range resp.Answer {
		srvAnswer, ok := answer.(*dns.SRV)
		if !ok {
			return nil, fmt.Errorf("Invalid DNS response!")
		}
		addrs[i] = ServiceAddress{
			Name:      strings.TrimSuffix(srvAnswer.Target, "."),
			Port:      srvAnswer.Port,
			Priority:  srvAnswer.Priority,
			Weight:    srvAnswer.Weight,
			expiresAt: now.Add(time.Second * time.Duration(answer.Header().Ttl)),
		}
		nameToAddr[srvAnswer.Target] = &addrs[i]
	}

	// Now we want to fill in the IPs for all those service addresses
	// we'll start by looking at the additional section
	for _, e := range resp.Extra {
		switch extra := e.(type) {
		case *dns.A:
			if addrEntry, ok := nameToAddr[extra.Hdr.Name]; ok {
				addrEntry.IP = extra.A
				delete(nameToAddr, extra.Hdr.Name)
			}
		case *dns.AAAA:
			if addrEntry, ok := nameToAddr[extra.Hdr.Name]; ok {
				addrEntry.IP = extra.AAAA
				delete(nameToAddr, extra.Hdr.Name)
			}
		}
	}

	// Now we loop over any that still need IPs
	for name, addrEntry := range nameToAddr {
	RESOLUTION_LOOP:
		for _, resolutionType := range d.c.ResolutionPriority {
			entryUrl := &url.URL{Host: name}
			var addrs []ServiceAddress
			var err error
			switch resolutionType {
			case AAAA:
				addrs, err = d.getServiceAddressesAAAA(ctx, entryUrl)
			case A:
				addrs, err = d.getServiceAddressesA(ctx, entryUrl)
			}
			if err != nil {
				return nil, err
			}
			if len(addrs) > 0 {
				addrEntry.IP = addrs[0].IP
				break RESOLUTION_LOOP
			}
		}
		if addrEntry.IP == nil {
			return nil, fmt.Errorf("No resolution for %s", name)
		}
	}

	return addrs, nil
}

func (d *discovery) getServiceAddressesAAAA(ctx context.Context, u *url.URL) (ServiceAddresses, error) {
	m := &dns.Msg{}
	m.SetQuestion(dns.Fqdn(u.Hostname()), dns.TypeAAAA)

	resp, err := d.r.ExchangeContext(ctx, m)
	if err != nil {
		return nil, err
	}

	if resp != nil && resp.Rcode != dns.RcodeSuccess {
		return nil, errors.New(dns.RcodeToString[resp.Rcode])
	}

	now := time.Now()
	addrs := make(ServiceAddresses, len(resp.Answer))
	for i, record := range resp.Answer {
		if t, ok := record.(*dns.AAAA); ok {
			addrs[i] = ServiceAddress{
				Name:      t.AAAA.String(),
				IP:        t.AAAA,
				expiresAt: now.Add(time.Second * time.Duration(record.Header().Ttl)),
			}
		} else {
			// TODO: this would mean the DNS resolver is returning invalid results, we might want to change this to either
			// (1) return an error which would stop resolution or (2) return nil, nil to force the caller to fall through
			// to the next query type
			logrus.Warningf("Got an unexpected/invalid response to DNS query; expected=%v actual=%v", dns.TypeAAAA, record.Header().Rrtype)
		}
	}
	return addrs, nil
}

func (d *discovery) getServiceAddressesA(ctx context.Context, u *url.URL) (ServiceAddresses, error) {
	m := &dns.Msg{}
	m.SetQuestion(dns.Fqdn(u.Hostname()), dns.TypeA)

	resp, err := d.r.ExchangeContext(ctx, m)
	if err != nil {
		return nil, err
	}

	if resp != nil && resp.Rcode != dns.RcodeSuccess {
		return nil, errors.New(dns.RcodeToString[resp.Rcode])
	}

	now := time.Now()
	addrs := make(ServiceAddresses, len(resp.Answer))
	for i, record := range resp.Answer {
		if t, ok := record.(*dns.A); ok {
			addrs[i] = ServiceAddress{
				Name:      t.A.String(),
				IP:        t.A,
				expiresAt: now.Add(time.Second * time.Duration(record.Header().Ttl)),
			}
		} else {
			// TODO: this would mean the DNS resolver is returning invalid results, we might want to change this to either
			// (1) return an error which would stop resolution or (2) return nil, nil to force the caller to fall through
			// to the next query type
			logrus.Warningf("Got an unexpected/invalid response to DNS query; expected=%v actual=%v", dns.TypeA, record.Header().Rrtype)
		}
	}
	return addrs, nil
}

// SubscribeServiceAddresses will do the lookup and give the results to Callback
func (d *discovery) SubscribeServiceAddresses(ctx context.Context, q string, cb Callback) error {
	addrs, err := d.GetServiceAddresses(ctx, q)
	if err != nil {
		return err
	}

	if err := cb(ctx, addrs); err != nil {
		return err
	}

	go d.backgroundCallback(ctx, q, cb, addrs)

	return nil
}

func (d *discovery) backgroundCallback(ctx context.Context, q string, cb Callback, results ServiceAddresses) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Get the closest expiry time (if one record will expire we want to re-query)
	getTimerDur := func(results ServiceAddresses) time.Duration {
		var t time.Time
		for _, addr := range results {
			if t.IsZero() {
				t = addr.expiresAt
			} else {
				if addr.expiresAt.Before(t) {
					t = addr.expiresAt
				}
			}
		}
		dur := t.Sub(time.Now())
		if dur < d.c.MinRetryInterval {
			dur = d.c.MinRetryInterval
		}
		return dur
	}

	timer := time.NewTimer(getTimerDur(results))
	for {
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return
		case <-timer.C:
			newResults, err := d.GetServiceAddresses(ctx, q)
			if err != nil {
				timer.Reset(d.c.SubscribeRetryInterval)
			} else {
				// If there was a change, we execute the callback
				if !results.Equal(newResults) {
					// If we get an error applying the callback, we'll continue retrying
					if err := cb(ctx, newResults); err != nil {
						logrus.Errorf("Error in ServiceAddress callback, retrying: %v", err)
						timer.Reset(d.c.SubscribeRetryInterval)
						continue
					} else {
						results = newResults
					}
				}

				// Regardless of callback, we update the TTL timer etc.
				timer.Reset(getTimerDur(results))
			}
		}
	}
}
