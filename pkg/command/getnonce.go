package command

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
)

func init() {
	Register("getnonce", func() Command {
		return &GetNonce{}
	})
}

// the struct for the 'find' command.
type GetNonce struct {
	GetNonce int `bson:"getnonce"`

	Common `bson:",inline"`
}

func (m *GetNonce) FromBSOND(d bson.D) error {
	dec, err := bson.NewDecoder(bsonutil.NewStrictValueReader(d))
	if err != nil {
		return err
	}

	if err := dec.Decode(&m); err != nil {
		return err
	}

	return nil
}
