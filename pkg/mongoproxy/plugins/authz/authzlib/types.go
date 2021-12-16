package authzlib

import (
	"context"
)

// AuthorizationMethod is an int8 to indicate CRUD permissions
type AuthorizationMethod int8

// CRUD permissions
const (
	_ = iota // ignore first value

	Create AuthorizationMethod = iota
	Read
	Update
	Delete

	create string = "Create"
	read   string = "Read"
	update string = "Update"
	delete string = "Delete"
)

func (m AuthorizationMethod) String() string {
	switch m {
	case Create:
		return create
	case Read:
		return read
	case Update:
		return update
	case Delete:
		return delete
	default:
		return "Unknown"
	}
}

// Authorization is an interface that handles loading in the config
// and creating the querier
type Authorization interface {
	// LoadConfig will load the given URLs of rules and
	// expand the rules using the SchemaQuerier (for annotation
	// based rules) returning an error if there was an issue
	LoadConfig(ctx context.Context, paths []string, data SchemaQuerier) error
	// Querier returns a querier of the current loaded config
	// the returned querier must provide a consistent view
	// over the data (this way we get consistent authz throughout
	// a single request)
	Querier() AuthorizationQuerier
}

// AuthorizationQuerier is an interface that handles authorizing
// mongo requests.
type AuthorizationQuerier interface {
	// Authorize will authorize the given request based on the URI
	// passed in. The URI might be a subset (e.g. DB/Collection) in
	// cases where we want to pre-check permissions (e.g. if no
	// permissions on anything, just fail to avoid the subsequent
	// lookups
	Authorize(ctx context.Context, identities []string, method AuthorizationMethod, resource Resource) AuthorizeResult
}

type AuthorizeResult struct {
	AuthorizationMethod
	Resource
	IdentityName string
	Rule         *Rule

	LogOnlyRules []Rule
}
