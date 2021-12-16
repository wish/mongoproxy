package authzlib

import (
	"context"
	"sync/atomic"
)

// Authz implements Authorization.
type Authz struct {
	querier atomic.Value
	urls    []string
}

// TODO - remove this when SchemaQuerier is actually implemented
type SchemaQuerier interface{}

// TODO: maintain a version of the config loaded
// LoadConfig will load the given URLs of rules and
// expand the rules using the SchemaQuerier (for annotation
// based rules) returning an error if there was an issue
func (a *Authz) LoadConfig(ctx context.Context, paths []string, q *SchemaQuerier) error {
	var querier AuthzSchema
	for _, path := range paths {
		if err := querier.getSchema(path); err != nil {
			return err
		}
	}
	// TODO - use SchemaQuerier to generate authz conditions for annotations

	a.querier.Store(&querier)
	a.urls = paths

	return nil
}

// Querier returns a querier of the current loaded config
// the returned querier must provide a consistent view
// over the data (this way we get consistent authz throughout
// a single request)
func (a *Authz) Querier() AuthorizationQuerier {
	return a.GetSchema()
}

func (a *Authz) GetSchema() *AuthzSchema {
	tmp := a.querier.Load()
	if tmp == nil {
		return nil
	}

	return tmp.(*AuthzSchema)
}
