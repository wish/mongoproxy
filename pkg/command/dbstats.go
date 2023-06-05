package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("dbstats", func() Command {
		return &DbStats{}
	})
	Register("dbStats", func() Command {
		return &DbStats{}
	})
}

// the struct for the 'find' command.
type DbStats struct {
	DbStats int `bson:"dbstats"`
	Scale   int `bson:"scale,omitempty"`

	Common `bson:",inline"`
}

func (m *DbStats) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
