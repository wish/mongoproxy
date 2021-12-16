package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("dropIndexes", func() Command {
		return &DropIndexes{}
	})
}

// the struct for the 'update' command.
type DropIndexes struct {
	Collection   string        `bson:"dropIndexes"`
	Index        interface{}   `bson:"index"`
	WriteConcern *WriteConcern `bson:"writeConcern,omitempty"`
	// New in 4.4
	//Comment      interface{}        `bson:"comment,omitempty"`

	Common `bson:",inline"`
}

func (m *DropIndexes) GetCollection() string { return m.Collection }

func (m *DropIndexes) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
