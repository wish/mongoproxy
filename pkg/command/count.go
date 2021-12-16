package command

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("count", func() Command {
		return &Count{}
	})
}

// the struct for the 'update' command.
type Count struct {
	Collection  string       `bson:"count"`
	Query       bson.D       `bson:"query"`
	MaxTimeMS   *int64       `bson:"maxTimeMS,omitempty"`
	Skip        *int64       `bson:"skip,omitempty"`
	Limit       *int64       `bson:"limit,omitempty"`
	Hint        interface{}  `bson:"hint,omitempty"`
	Collation   *Collation   `bson:"collation,omitempty"`
	ReadConcern *ReadConcern `bson:"readConcern,omitempty"`
	Fields      bson.D       `bson:"fields,omitempty"` // Mongo shell sends this for w/e reason, but can't find this in the docs anywhere

	Common `bson:",inline"`
}

func (m *Count) GetCollection() string { return m.Collection }

func (m *Count) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		if decErr, ok := err.(*bsoncodec.DecodeError); ok {

			keys := decErr.Keys()
			if len(keys) == 1 && keys[0] == "count" {
				return mongo.CommandError{
					Code:    73,
					Message: fmt.Sprintf("Invalid Namespace: %s", decErr.Unwrap().Error()),
				}
			}
		}
		return err
	}

	if len(m.Fields) > 0 {
		return mongo.CommandError{
			Code:    9,
			Message: "FailedToParse: \"fields\" should be empty",
		}
	}

	return nil
}
