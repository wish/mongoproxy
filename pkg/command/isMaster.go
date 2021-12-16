package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	//"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
)

func init() {
	Register("isMaster", func() Command {
		return &IsMaster{}
	})
	Register("ismaster", func() Command {
		return &IsMaster{}
	})
}

// IsMaster mongo command
type IsMaster struct {
	IsMaster       int      `bson:"isMaster"`
	IsMasterLegacy int      `bson:"ismaster"`
	Client         bson.D   `bson:"client"` // TODO parse out
	Compression    []string `bson:"compression"`
	HostInfo       string   `bson:"hostInfo"`

	Common `bson:",inline"`
}

// From BSOND loads a command from a bson.D
func (m *IsMaster) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
