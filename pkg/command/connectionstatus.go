package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("connectionStatus", func() Command {
		return &ConnectionStatus{}
	})
}

// the struct for the 'find' command.
type ConnectionStatus struct {
	ConnectionStatus int   `bson:"connectionStatus"`
	ShowPrivileges   *bool `bson:"showPrivileges,omitempty"`

	Common `bson:",inline"`
}

func (m *ConnectionStatus) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
