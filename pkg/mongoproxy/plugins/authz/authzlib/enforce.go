package authzlib

type EnforceMethod int8

const (
	_ = iota // ignore first value

	// DefaultCase is when the EnforceMethod is not set
	DefaultCase EnforceMethod = iota

	// EnforceCase is when the effect is to deny and we
	// would like to enforce the outcome.
	EnforceCase

	// LogCase is when the effect is to deny and we
	// would like to log the outcome without enforcing
	LogCase

	// AuthorizedCase is when the effect is to allow
	AuthorizedCase
)

func (m EnforceMethod) String() string {
	switch m {
	case DefaultCase:
		return "default"
	case EnforceCase:
		return "enforce"
	case LogCase:
		return "log"
	case AuthorizedCase:
		return "authorized"
	default:
		return "default"
	}
}
