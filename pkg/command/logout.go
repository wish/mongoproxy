package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("logout", func() Command {
		return &Logout{}
	})
}

// the struct for the 'logout' command.
type Logout struct {
	Logout int `bson:"logout"`

	Common `bson:",inline"`
}

func (m *Logout) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
