package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("buildinfo", func() Command {
		return &BuildInfo{}
	})

	Register("buildInfo", func() Command {
		return &BuildInfo{}
	})
}

// the struct for the 'update' command.
type BuildInfo struct {
	V  int `bson:"buildinfo"`
	V2 int `bson:"buildInfo"`

	Common `bson:",inline"`
}

func (m *BuildInfo) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
