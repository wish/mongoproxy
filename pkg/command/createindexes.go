package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("createIndexes", func() Command {
		return &CreateIndexes{}
	})
}

// the struct for the 'update' command.
type CreateIndexes struct {
	Collection   string               `bson:"createIndexes"`
	Indexes      []CreateIndexesIndex `bson:"indexes"`
	WriteConcern *WriteConcern        `bson:"writeConcern,omitempty"`
	CommitQuorum interface{}          `bson:"commitQuorum,omitempty"`

	// New in 4.4
	// Comment interface{} `bson"comment,omitempty"`

	Common `bson:",inline"`
}

func (m *CreateIndexes) GetCollection() string { return m.Collection }

func (m *CreateIndexes) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}

type CreateIndexesIndex struct {
	Key                     bson.D  `bson:"key"`
	Name                    string  `bson:"name"`
	Background              *bool   `bson:"background,omitempty"`
	Unique                  *bool   `bson:"unique,omitempty"`
	PartialFilterExpression bson.D  `bson:"partialFilterExpression,omitempty"`
	Sparse                  *bool   `bson:"sparse,omitempty"`
	ExpireAfterSeconds      *int    `bson:"expireAfterSeconds,omitempty"`
	Hidden                  *bool   `bson:"hidden,omitempty"`
	StorageEngine           bson.D  `bson:"storageEngine,omitempty"`
	Weights                 bson.D  `bson:"weights,omitempty"`
	DefaultLanguage         string  `bson:"default_language,omitempty"`
	LanguageOverride        string  `bson:"language_override,omitempty"`
	TextIndexVersion        int     `bson:"textIndexVersion,omitempty"`
	TwoDSphereIndexVersion  int     `bson:"2dsphereIndexVersion,omitempty"`
	Bits                    int     `bson:"bits,omitempty"`
	Min                     float64 `bson:"min,omitempty"`
	Max                     float64 `bson:"max,omitempty"`
	BucketSize              int     `bson:"bucketSize,omitempty"`
	Collation               bson.D  `bson:"collation,omitempty"`
	WildcardProjection      bson.D  `bson:"wildcardProjection,omitempty"`
}
