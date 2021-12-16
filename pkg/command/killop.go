package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	//"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func init() {
	Register("killOp", func() Command {
		return &KillOp{}
	})
}

// the struct for the 'find' command.
type KillOp struct {
	KillOp  int    `bson:"killOp"`
	OpID    int    `bson:"op"`
	Comment string `bson:"comment,omitempty"`

	Common `bson:",inline"`
}

func (m *KillOp) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
