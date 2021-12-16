package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("deleteIndexes", func() Command {
		return &DeleteIndexes{}
	})
}

// the struct for the 'update' command.
type DeleteIndexes struct {
	Collection string      `bson:"deleteIndexes"`
	Index      interface{} `bson:"index"`

	Common `bson:",inline"`
}

func (m *DeleteIndexes) GetCollection() string { return m.Collection }

func (m *DeleteIndexes) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
