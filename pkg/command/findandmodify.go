package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("findAndModify", func() Command {
		return &FindAndModify{}
	})

	Register("findandmodify", func() Command {
		return &FindAndModifyLegacy{}
	})
}

// the struct for the 'update' command.
type FindAndModify struct {
	Collection               string        `bson:"findAndModify"`
	Query                    bson.D        `bson:"query,omitempty"`
	Sort                     bson.D        `bson:"sort,omitempty"`
	Remove                   *bool         `bson:"remove,omitempty"`
	Update                   bson.D        `bson:"update,omitempty"`
	New                      *bool         `bson:"new,omitempty"`
	Fields                   bson.D        `bson:"fields,omitempty"`
	Upsert                   *bool         `bson:"upsert,omitempty"`
	BypassDocumentValidation *bool         `bson:"bypassDocumentValidation,omitempty"`
	WriteConcern             *WriteConcern `bson:"writeConcern,omitempty"`
	Collation                *Collation    `bson:"collation,omitempty"`
	ArrayFilters             interface{}   `bson:"arrayFilters,omitempty"` // TODO

	Common `bson:",inline"`
}

func (m *FindAndModify) GetCollection() string { return m.Collection }

func (m *FindAndModify) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}

// TODO: find a way to combine? The only thing that changes is the tag for `collection`
// the struct for the 'update' command.
type FindAndModifyLegacy struct {
	Collection               string        `bson:"findandmodify"`
	Query                    bson.D        `bson:"query,omitempty"`
	Sort                     bson.D        `bson:"sort,omitempty"`
	Remove                   *bool         `bson:"remove,omitempty"`
	Update                   interface{}   `bson:"update,omitempty"`
	New                      *bool         `bson:"new,omitempty"`
	Fields                   bson.D        `bson:"fields,omitempty"`
	Upsert                   *bool         `bson:"upsert,omitempty"`
	BypassDocumentValidation *bool         `bson:"bypassDocumentValidation,omitempty"`
	WriteConcern             *WriteConcern `bson:"writeConcern,omitempty"`
	Collation                *Collation    `bson:"collation,omitempty"`
	ArrayFilters             interface{}   `bson:"arrayFilters,omitempty"` // TODO

	Common `bson:",inline"`
}

func (m *FindAndModifyLegacy) GetCollection() string { return m.Collection }

func (m *FindAndModifyLegacy) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
