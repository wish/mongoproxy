package command

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("create", func() Command {
		return &Create{}
	})
}

// Create mongo command
type Create struct {
	Collection          string        `bson:"create"`
	Capped              *bool         `bson:"capped,omitempty"`
	AutoIndexID         *bool         `bson:"autoIndexId,omitempty"`
	Size                *int          `bson:"size,omitempty"`
	Max                 *int          `bson:"max,omitempty"`
	StorageEngine       bson.D        `bson:"storageEngine,omitempty"`
	Validator           bson.D        `bson:"validator,omitempty"`
	ValidationLevel     string        `bson:"validationLevel,omitempty"`
	ValidationAction    string        `bson:"validationAction,omitempty"`
	IndexOptionDefaults bson.D        `bson:"indexOptionDefaults,omitempty"`
	ViewOn              string        `bson:"viewOn,omitempty"`
	Pipeline            primitive.A   `bson:"pipeline,omitempty"`
	Collation           *Collation    `bson:"collation,omitempty"`
	WriteConcern        *WriteConcern `bson:"writeConcern,omitempty"`
	// New in 4.4
	// Comment interface{} `bson"comment,omitempty"`

	Common `bson:",inline"`
}

// GetCollection returns the collection name for this Command
func (m *Create) GetCollection() string { return m.Collection }

// From BSOND loads a command from a bson.D
func (m *Create) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
