package schema

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
)

type SchemaLoader interface {
	// Load returns a bool (whether something was loaded) + error
	Load(ctx context.Context, glob string) error
	// Querier returns a querier of the current loaded config. The
	// returned querier must provide a consistent view over the data
	// (this way we get consistent schema enforcement throughout
	// all calls in a single request)
	Querier() SchemaQuerier
}

type SchemaValidator interface {
	// ValidateInsert will validate the schema of the passed in object.
	ValidateInsert(ctx context.Context, database, collection string, obj bson.D) error
	// ValidateUpdate will validate the schema of the passed in object.
	ValidateUpdate(ctx context.Context, database, collection string, obj bson.D) error
}

type SchemaQuerier interface {
	SchemaValidator

	// TODO: introspection interface
}
