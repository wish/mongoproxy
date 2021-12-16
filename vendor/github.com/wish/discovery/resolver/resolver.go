package resolver

import (
	"context"
	"fmt"
	"net"

	"github.com/miekg/dns"
)

func NewResolver(path string) (*DNSResolver, error) {
	cfg, err := dns.ClientConfigFromFile(path)
	if err != nil {
		return nil, err
	}

	return &DNSResolver{
		cfg: cfg,
	}, nil
}

func NewResolverFromConfig(cfg *dns.ClientConfig) *DNSResolver {
	return &DNSResolver{cfg}
}

type DNSResolver struct {
	cfg *dns.ClientConfig
}

func (r *DNSResolver) ExchangeContext(ctx context.Context, msg *dns.Msg) (*dns.Msg, error) {
	return r.exchangeContext(ctx, msg, 0)
}

func (r *DNSResolver) exchangeContext(ctx context.Context, msg *dns.Msg, attempt int) (*dns.Msg, error) {
	if attempt >= len(r.cfg.Servers) {
		return nil, fmt.Errorf("No more available servers to attempt")
	}

	udpClient := &dns.Client{}
	resp, _, err := udpClient.ExchangeContext(ctx, msg, r.cfg.Servers[attempt]+":"+r.cfg.Port)
	// if it truncated, then fallback to TCP
	if err == nil && resp.MsgHdr.Truncated {
		tcpClient := &dns.Client{Net: "tcp"}
		resp, _, err = tcpClient.ExchangeContext(ctx, msg, r.cfg.Servers[attempt]+":"+r.cfg.Port)
	}

	if err != nil {
		if _, ok := err.(*net.OpError); ok {
			// if got an error, fallback to the next one
			return r.exchangeContext(ctx, msg, attempt+1)
		} else {
			return nil, err
		}
	}
	// TODO: remove the dns.ErrTruncated? Truncated responses from TCP requests aren't valid
	// but some resolvers (systemd's) seems to do so-- so we'll let it fall through
	if resp.MsgHdr.Truncated {
		// if got an error, fallback to the next one
		return r.exchangeContext(ctx, msg, attempt+1)
	}

	// Depending on the response code from the resolver we may want to fall through
	switch resp.MsgHdr.Rcode {
	// If the resolver had an error resolving, we should fall through to the next
	case dns.RcodeServerFailure:
		// if got an error, fallback to the next one
		return r.exchangeContext(ctx, msg, attempt+1)

	default:
		return resp, nil
	}
}
