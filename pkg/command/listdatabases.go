package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("listDatabases", func() Command {
		return &ListDatabases{}
	})
}

// the struct for the 'update' command.
type ListDatabases struct {
	ListDatabases       int    `bson:"listDatabases"`
	Filter              bson.D `bson:"filter,omitempty"`
	NameOnly            *bool  `bson:"nameOnly,omitempty"`
	AuthorizedDatabases *bool  `bson:"authorizedDatabases,omitempty"`
	// New in 4.4
	//Comment             interface{}      `bson:"comment,omitempty"`

	Common `bson:",inline"`
}

func (m *ListDatabases) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
