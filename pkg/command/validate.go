package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("validate", func() Command {
		return &Validate{}
	})
}

// the struct for the 'update' command.
type Validate struct {
	Collection string `bson:"validate"`
	Full       *bool  `bson:"full,omitempty"`

	Common `bson:",inline"`
}

func (m *Validate) GetCollection() string { return m.Collection }

func (m *Validate) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(m); err != nil {
		return err
	}

	return nil
}
