package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("saslStart", func() Command {
		return &SaslStart{}
	})
}

// the struct for the 'find' command.
type SaslStart struct {
	SaslStart     int    `bson:"saslStart"`
	Mechanism     string `bson:"mechanism"`
	Payload       []byte `bson:"payload"`
	AutoAuthorize int    `bson:"autoAuthorize,omitempty"`
	Options       bson.D `bson:"options,omitempty"` // TODO: expand

	Common `bson:",inline"`
}

func (m *SaslStart) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
