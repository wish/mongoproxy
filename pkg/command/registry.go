package command

import (
	"fmt"
)

type CommandFunc func() Command

var Registry = make(map[string]CommandFunc)

func Register(n string, f CommandFunc) {
	if _, ok := Registry[n]; ok {
		msg := fmt.Sprintf("Command named %s already registered", n)
		panic(msg)
	}
	Registry[n] = f
}

func GetCommand(d string) (Command, bool) {
	cmdF, ok := Registry[d]
	if !ok {
		return nil, false
	}

	return cmdF(), true
}
