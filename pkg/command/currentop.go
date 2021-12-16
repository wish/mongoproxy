package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("currentOp", func() Command {
		return &CurrentOp{}
	})
}

// the struct for the 'update' command.
type CurrentOp struct {
	CurrentOp interface{} `bson:"currentOp"`
	OwnOps    *bool       `bson:"$ownOps,omitempty"`
	All       *bool       `bson:"$all,omitempty"`
	Filter    bson.D      `bson:"filter,omitempty"`

	Common `bson:",inline"`
}

func (m *CurrentOp) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(m); err != nil {
		return err
	}

	return nil
}
