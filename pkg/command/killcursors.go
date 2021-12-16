package command

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	//"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func init() {
	Register("killCursors", func() Command {
		return &KillCursors{}
	})
}

// the struct for the 'find' command.
type KillCursors struct {
	Collection string      `bson:"killCursors"`
	Cursors    primitive.A `bson:"cursors"`
	// New in 4.4
	//Comment interface{} `bson:"comment,omitempty"`

	Common `bson:",inline"`
}

func (m *KillCursors) GetCollection() string { return m.Collection }

func (m *KillCursors) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
