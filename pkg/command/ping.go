package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("ping", func() Command {
		return &Ping{}
	})
}

// the struct for the 'update' command.
type Ping struct {
	Ping int `bson:"ping"`

	Common `bson:",inline"`
}

func (m *Ping) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
