package plugins

import (
	"context"
	"net"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/command"
)

// PipelineFunc is the function type for the built pipeline, and is called
// to begin the pipeline.
type PipelineFunc func(context.Context, *Request) (bson.D, error)

type Plugin interface {
	// Name returns the name for the given plugin
	Name() string

	// Configure configures this plugin with the given configuration object. Returns
	// an error if the configuration is invalid for the plugin.
	Configure(bson.D) error

	// Process is the function executed when a message is called in the pipeline.
	Process(context.Context, *Request, PipelineFunc) (bson.D, error)
}

func NewCursorCacheEntry(id int64) *CursorCacheEntry {
	return &CursorCacheEntry{
		ID:  id,
		Map: map[interface{}]interface{}{},
	}
}

type CursorCache interface {
	CreateCursor(cursorID int64) *CursorCacheEntry
	GetCursor(cursorID int64) *CursorCacheEntry
	CloseCursor(cursorID int64)
}

type CursorCacheEntry struct {
	ID             int64
	CursorConsumed int

	// Map is storage that resets on cursor change
	Map map[interface{}]interface{}
}

// Request encapsulates a mongo request
type Request struct {
	CC *ClientConnection
	CursorCache

	// TODO: add reference to cursor here (we can maintain cursor mapping in core)

	CommandName string
	Command     command.Command

	// Map of arbitrary data for plugins to store stuff in
	Map map[string]interface{}
}

func (r *Request) Close() {}

func NewClientConnection() *ClientConnection {
	return &ClientConnection{
		Map: map[interface{}]interface{}{},
	}
}

type ClientConnection struct {
	// Address of client connection
	Addr net.Addr
	// According to the docs (https://docs.mongodb.com/manual/core/authentication/#authentication-methods) multiple logins should
	// have the credentials for all until a logout happens; for now we aren't doing that.
	Identities []ClientIdentity

	// Map is storage that resets on cursor change
	Map map[interface{}]interface{}
}

func (c *ClientConnection) GetAddr() string {
	if c.Addr == nil {
		return ""
	}
	return c.Addr.String()
}

func (c *ClientConnection) Close() {}

type ClientIdentity interface {
	Type() string // Where the identity came from
	User() string
	Roles() []string
}

func NewStaticIdentity(t, u string, rs ...string) *StaticIdentity {
	return &StaticIdentity{
		T:  t,
		U:  u,
		RS: rs,
	}
}

type StaticIdentity struct {
	T  string   `bson"type"`
	U  string   `bson:"user"`
	RS []string `bson:"roles"`
}

func (i *StaticIdentity) Type() string    { return i.T }
func (i *StaticIdentity) User() string    { return i.U }
func (i *StaticIdentity) Roles() []string { return i.RS }
