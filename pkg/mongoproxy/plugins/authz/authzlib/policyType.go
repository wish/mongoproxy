package authzlib

import "strings"

type policyType int8

const (
	notSetP policyType = iota
	logP
)

func (p policyType) IsSet() bool {
	return p != notSetP
}

func (p policyType) IsLogOnly() bool {
	return p == logP
}

func getPolicy(str string) (p policyType) {
	switch strings.ToLower(str) {
	case "logonly":
		p = logP
	default:
		p = notSetP
	}
	return p
}

func (p policyType) String() string {
	switch p {
	case logP:
		return "LogOnly"
	}
	return "NotSet"
}
