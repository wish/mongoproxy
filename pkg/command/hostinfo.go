package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("hostInfo", func() Command {
		return &HostInfo{}
	})
}

// the struct for the 'update' command.
type HostInfo struct {
	HostInfo interface{} `bson:"hostInfo"`

	Common `bson:",inline"`
}

func (m *HostInfo) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(m); err != nil {
		return err
	}

	return nil
}
