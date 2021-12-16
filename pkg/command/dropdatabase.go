package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("dropDatabase", func() Command {
		return &DropDatabase{}
	})
}

// the struct for the 'update' command.
type DropDatabase struct {
	DropDatabase int           `bson:"dropDatabase"`
	WriteConcern *WriteConcern `bson:"writeConcern,omitempty"`
	Comment      string        `bson:"comment,omitempty"`

	Common `bson:",inline"`
}

func (m *DropDatabase) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
