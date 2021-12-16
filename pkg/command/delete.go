package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("delete", func() Command {
		return &Delete{}
	})
}

// the struct for the 'update' command.
type Delete struct {
	Collection   string        `bson:"delete"`
	Deletes      []bson.D      `bson:"deletes"`
	Ordered      *bool         `bson:"ordered,omitempty"`
	WriteConcern *WriteConcern `bson:"writeConcern,omitempty"`
	Hint         interface{}   `bson:"hint,omitempty"`

	Common `bson:",inline"`
}

func (m *Delete) GetCollection() string { return m.Collection }

func (m *Delete) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
