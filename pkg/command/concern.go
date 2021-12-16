package command

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

type WriteConcern struct {
	W        interface{}   `bson:"w"`
	J        bool          `bson:"j"`
	WTimeout time.Duration `bson:"wtimeout"`
}

type ReadConcern struct {
	Level string `bson:"level,omitempty"`
}

type ReadPreference struct {
	Mode         string `bson:"mode"`
	TagSet       bson.A `bson:"tagSet,omitempty"`
	HedgeOptions `bson:"hedgeOptions,omitempty"`
}

type HedgeOptions struct {
	Enabled bool `bson:"enabled"`
}
