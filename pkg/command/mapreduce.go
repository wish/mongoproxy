package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("mapReduce", func() Command {
		return &MapReduce{}
	})

	Register("mapreduce", func() Command {
		return &MapReduce{}
	})
}

// MapReduce mongo command
type MapReduce struct {
	// TODO: split types?
	Collection               string        `bson:"mapReduce,omitempty"`
	CollectionLegacy         string        `bson:"mapreduce,omitempty"`
	Map                      interface{}   `bson:"map"`
	Reduce                   interface{}   `bson:"reduce"`
	Out                      interface{}   `bson:"out"`
	Query                    bson.D        `bson:"query,omitempty"`
	Sort                     bson.D        `bson:"sort,omitempty"`
	Limit                    *int64        `bson:"limit,omitempty"`
	Finalize                 interface{}   `bson:"finalize,omitempty"`
	Scope                    bson.D        `bson:"scope,omitempty"`
	JSMode                   *bool         `bson:"jsMode,omitempty"`
	Verbose                  *bool         `bson:"verbose,omitempty"`
	BypassDocumentValidation *bool         `bson:"bypassDocumentValidation,omitempty"`
	Collation                *Collation    `bson:"collation,omitempty"`
	WriteConcern             *WriteConcern `bson:"writeConcern,omitempty"`
	Comment                  string        `bson:"comment,omitempty"`

	Common `bson:",inline"`
}

// GetCollection returns the collection name for this Command
func (m *MapReduce) GetCollection() string { return m.Collection }

// From BSOND loads a command from a bson.D
func (m *MapReduce) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
