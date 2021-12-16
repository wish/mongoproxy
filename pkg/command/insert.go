package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("insert", func() Command {
		return &Insert{}
	})
}

// the struct for the 'update' command.
type Insert struct {
	Collection string   `bson:"insert"`
	Documents  []bson.D `bson:"documents"`
	Ordered    *bool    `bson:"ordered,omitempty"`
	//selector                 description.ServerSelector
	WriteConcern             *WriteConcern `bson:"writeConcern,omitempty"`
	BypassDocumentValidation *bool         `bson:"bypassDocumentValidation,omitempty"`

	Common `bson:",inline"`
}

func (m *Insert) GetCollection() string { return m.Collection }

func (m *Insert) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
