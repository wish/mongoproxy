package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("drop", func() Command {
		return &Drop{}
	})
}

// the struct for the 'update' command.
type Drop struct {
	Collection   string        `bson:"drop"`
	WriteConcern *WriteConcern `bson:"writeConcern,omitempty"`
	Comment      string        `bson:"comment,omitempty"`

	Common `bson:",inline"`
}

func (m *Drop) GetCollection() string { return m.Collection }

func (m *Drop) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
