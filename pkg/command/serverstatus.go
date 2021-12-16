package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("serverStatus", func() Command {
		return &ServerStatus{}
	})
}

// the struct for the 'update' command.
type ServerStatus struct {
	ServerStatus int `bson:"serverStatus"`

	Common `bson:",inline"`
}

func (m *ServerStatus) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
