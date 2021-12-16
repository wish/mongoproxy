package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("collStats", func() Command {
		return &CollStats{}
	})
}

// the struct for the 'find' command.
type CollStats struct {
	Collection string `bson:"collStats"`
	Scale      *int64 `bson:"scale,omitempty"`
	Verbose    *int64 `bson:"verbose,omitempty"`

	Common `bson:",inline"`
}

func (m *CollStats) GetCollection() string { return m.Collection }

func (m *CollStats) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
