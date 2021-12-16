package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("killAllSessions", func() Command {
		return &KillAllSessions{}
	})
}

// the struct for the 'find' command.
type KillAllSessions struct {
	KillAllSessions []KillAllSessionsFilter `bson:"killAllSessions"`

	Common `bson:",inline"`
}

type KillAllSessionsFilter struct {
	User string `bson:"user,omitempty"`
	DB   string `bson:"db,omitempty"`
}

func (m *KillAllSessions) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
