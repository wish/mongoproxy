package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("getMore", func() Command {
		return &GetMore{}
	})
}

// the struct for the 'update' command.
type GetMore struct {
	CursorID   int64  `bson:"getMore"`
	Collection string `bson:"collection"`
	BatchSize  *int32 `bson:"batchSize,omitempty"`

	Common `bson:",inline"`
}

func (m *GetMore) GetCollection() string { return m.Collection }

func (m *GetMore) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
