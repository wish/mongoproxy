package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("update", func() Command {
		return &Update{}
	})
}

// the struct for the 'update' command.
type Update struct {
	Collection               string            `bson:"update"`
	Updates                  []UpdateStatement `bson:"updates"`
	WriteConcern             *WriteConcern     `bson:"writeConcern,omitempty"`
	Ordered                  *bool             `bson:"ordered,omitempty"`
	BypassDocumentValidation *bool             `bson:"bypassDocumentValidation,omitempty"`

	Common `bson:",inline"`
}

func (m *Update) GetCollection() string { return m.Collection }

func (m *Update) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}

type UpdateStatement struct {
	Query        bson.D      `bson:"q"`
	U            bson.D      `bson:"u"`
	Upsert       *bool       `bson:"upsert,omitempty"`
	Multi        *bool       `bson:"multi,omitempty"`
	Collation    *Collation  `bson:"collation,omitempty"`
	ArrayFilters interface{} `bson:"arrayFilters,omitempty"` // TODO
	Hint         interface{} `bson:"hint,omitempty"`
}
