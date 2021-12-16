package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	//"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func init() {
	Register("isdbgrid", func() Command {
		return &IsDBGrid{}
	})
}

// the struct for the 'find' command.
type IsDBGrid struct {
	IsDBGrid int `bson:"isdbgrid"`

	Common `bson:",inline"`
}

func (m *IsDBGrid) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
