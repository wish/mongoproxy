package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("listCollections", func() Command {
		return &ListCollections{}
	})
}

// the struct for the 'update' command.
type ListCollections struct {
	ListCollections       int    `bson:"listCollections"`
	Filter                bson.D `bson:"filter,omitempty"`
	NameOnly              *bool  `bson:"nameOnly,omitempty"`
	AuthorizedDatabases   *bool  `bson:"authorizedDatabases,omitempty"`
	AuthorizedCollections *bool  `bson:"authorizedCollections,omitempty"`
	// New in 4.4
	//Comment             interface{}      `bson:"comment,omitempty"`

	MaxTimeMS *int64 `bson:"maxTimeMS,omitempty"`
	Common    `bson:",inline"`
	*Cursor   `bson:"cursor,omitempty"`
}

func (m *ListCollections) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
