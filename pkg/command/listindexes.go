package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("listIndexes", func() Command {
		return &ListIndexes{}
	})
}

// the struct for the 'update' command.
type ListIndexes struct {
	Collection string `bson:"listIndexes"`
	// New in 4.4
	//Comment             interface{}      `bson:"comment,omitempty"`

	Common  `bson:",inline"`
	*Cursor `bson:"cursor,omitempty"`
}

func (m *ListIndexes) GetCollection() string { return m.Collection }

func (m *ListIndexes) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
