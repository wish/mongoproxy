package command

import "go.mongodb.org/mongo-driver/bson"

type Command interface {
	GetSession() *Session
	FromBSOND(d bson.D) error
}

type CommandReadPreference interface {
	Command
	GetReadPreference() *ReadPreference
}

type CommandDatabase interface {
	Command
	GetDatabase() string
}

type CommandCollection interface {
	Command
	GetCollection() string
}
