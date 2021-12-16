package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("find", func() Command {
		return &Find{}
	})
}

// the struct for the 'find' command.
type Find struct {
	Collection          string      `bson:"find"`
	Filter              bson.D      `bson:"filter,omitempty"`
	Sort                bson.D      `bson:"sort,omitempty"`
	Projection          bson.D      `bson:"projection,omitempty"`
	Min                 bson.D      `bson:"min,omitempty"`
	Max                 bson.D      `bson:"max,omitempty"`
	Skip                *int64      `bson:"skip,omitempty"`
	Limit               *int64      `bson:"limit,omitempty"`
	Tailable            *bool       `bson:"tailable,omitempty"`
	OplogReplay         *bool       `bson:"oplogReplay,omitempty"` // TODO Check
	SingleBatch         *bool       `bson:"singleBatch,omitempty"`
	NoCursorTimeout     *bool       `bson:"noCursorTimeout,omitempty"` // TODO: check
	AllowDiskUse        *bool       `bson:"allowDiskUse,omitempty"`    // TODO: restrict usage of
	AllowPartialResults *bool       `bson:"allowPartialResults,omitempty"`
	AwaitData           *bool       `bson:"awaitData,omitempty"`
	BatchSize           *int32      `bson:"batchSize,omitempty"`
	Collation           *Collation  `bson:"collation,omitempty"`
	Comment             string      `bson:"comment,omitempty"`
	Hint                interface{} `bson:"hint,omitempty"`

	MaxTimeMS    *int64       `bson:"maxTimeMS,omitempty"`
	ShowRecordId *bool        `bson:"showRecordId,omitempty"`
	ReturnKey    *bool        `bson:"returnKey,omitempty"`
	ShowRecordID *bool        `bson:"showRecordID,omitempty"`
	ReadConcern  *ReadConcern `bson:"readConcern,omitempty"`

	Common `bson:",inline"`
}

func (m *Find) GetCollection() string { return m.Collection }

func (m *Find) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
