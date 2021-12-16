package command

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("distinct", func() Command {
		return &Distinct{}
	})
}

// the struct for the 'find' command.
type Distinct struct {
	Collection  string       `bson:"distinct"`
	Collation   *Collation   `bson:"collation,omitempty"`
	Key         string       `bson:"key"`
	MaxTimeMS   *int64       `bson:"maxTimeMS,omitempty"`
	Query       bson.D       `bson:"query"`
	ReadConcern *ReadConcern `bson:"readConcern,omitempty"`

	Common `bson:",inline"`
}

func (m *Distinct) GetCollection() string { return m.Collection }

func (m *Distinct) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		if decErr, ok := err.(*bsoncodec.DecodeError); ok {

			keys := decErr.Keys()
			if len(keys) == 1 {
				switch keys[0] {
				case "query":
					return mongo.CommandError{
						Code:    14,
						Message: fmt.Sprintf("Invalid query: %s", decErr.Unwrap().Error()),
					}
				case "key":
					return mongo.CommandError{
						Code:    14,
						Message: fmt.Sprintf("Invalid Key: %s", decErr.Unwrap().Error()),
					}
				}
			}
		}
		return err
	}

	return nil
}
