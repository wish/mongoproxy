package command

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("endSessions", func() Command {
		return &EndSessions{}
	})
}

// the struct for the 'update' command.
type EndSessions struct {
	SessionIDs []bsoncore.Document `bson:"endSessions"`

	Common `bson:",inline"`
}

func (m *EndSessions) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
