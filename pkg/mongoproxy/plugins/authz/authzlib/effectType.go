package authzlib

import "strings"

type effectType int8

const (
	notSetE effectType = iota
	denyE
	allowE
)

func (e effectType) IsSet() bool {
	return e != notSetE
}

func (e effectType) IsDeny() bool {
	return e == denyE
}

func (e effectType) IsAllow() bool {
	return e == allowE
}

func getEffect(str string) (e effectType) {
	switch strings.ToLower(str) {
	case "deny":
		e = denyE
	case "allow":
		e = allowE
	default:
		e = notSetE
	}
	return e
}

func (e effectType) String() string {
	switch e {
	case denyE:
		return "Deny"
	case allowE:
		return "Allow"
	}
	return "NotSet"
}
