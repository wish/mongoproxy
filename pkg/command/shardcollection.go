package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("shardCollection", func() Command {
		return &ShardCollection{}
	})
}

// the struct for the 'update' command.
type ShardCollection struct {
	Collection string `bson:"shardCollection"`
	Key        bson.D `bson:"key"`
	Unique     *bool  `bson:"unique,omitempty"`

	// New in 4.4
	// Comment interface{} `bson"comment,omitempty"`

	Common `bson:",inline"`
}

func (m *ShardCollection) GetCollection() string { return m.Collection }

func (m *ShardCollection) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
