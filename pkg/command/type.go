package command

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TODO: add IsZero to bson.D ?
// OptionalDoc is a document that is optional but allowed to be empty (by default an empty bson.D is omitempty-able)
type OptionalDoc struct {
	bson.D
}

func (t *OptionalDoc) IsZero() bool {
	return t == nil
}

type Session struct {
	LSID        bson.D       `bson:"lsid,omitempty"`
	TxnNumber   *int64       `bson:"txnNumber,omitempty"`
	StmtIDs     []int32      `bson:"stmtIds,omitempty"`
	ClusterTime *ClusterTime `bson:"$clusterTime,omitempty"`
}

func (s *Session) GetSession() *Session {
	return s
}

type ClusterTime struct {
	ClusterTime primitive.Timestamp `bson:"clusterTime,omitempty"`
	Signature   bson.Raw            `bson:"signature,omitempty"`
}

type Common struct {
	ReadPreference *ReadPreference `bson:"$readPreference,omitempty"`
	Database       string          `bson:"$db,omitempty"`
	Session        `bson:",inline"`
}

func (c *Common) GetDatabase() string                { return c.Database }
func (c *Common) GetReadPreference() *ReadPreference { return c.ReadPreference }

// Cursor encapsulates the separate "cursor" doc found on some commands
type Cursor struct {
	BatchSize *int32 `bson:"batchSize,omitempty"`
}
