package command

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("aggregate", func() Command {
		return &Aggregate{}
	})
}

// the struct for the 'find' command.
type Aggregate struct {
	Aggregate                bson.RawValue `bson:"aggregate"` // This is either a collection name or a 1 (int) to indicate a pipeline
	Pipeline                 primitive.A   `bson:"pipeline"`
	Cursor                   pipeCmdCursor `bson:"cursor"`
	Explain                  *bool         `bson:"explain,omitempty"`
	AllowDisk                *bool         `bson:"allowDiskUse,omitempty"` // TODO: disallow (or have option for it)
	MaxTimeMS                *int64        `bson:"maxTimeMS,omitempty"`
	BypassDocumentValidation *bool         `bson:"bypassDocumentValidation,omitempty"`
	ReadConcern              *ReadConcern  `bson:"readConcern,omitempty"`
	Collation                *Collation    `bson:"collation,omitempty"`
	Hint                     interface{}   `bson:"hint,omitempty"`
	Comment                  string        `bson:"comment,omitempty"`
	WriteConcern             *WriteConcern `bson:"writeConcern,omitempty"`

	Common `bson:",inline"`
}

func (a *Aggregate) GetCollection() string {
	if collection, ok := a.Aggregate.StringValueOK(); ok {
		return collection
	}
	return ""
}

type pipeCmdCursor struct {
	BatchSize *int `bson:"batchSize,omitempty"`
}

func (m *Aggregate) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
