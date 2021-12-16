package mongoproxy

import (
	"net"
	"sync/atomic"
	"time"
)

type ConnState int

const (
	StateNew ConnState = iota
	StateActive
	StateIdle
	StateClosed
)

var stateName = map[ConnState]string{
	StateNew:    "new",
	StateActive: "active",
	StateIdle:   "idle",
	StateClosed: "closed",
}

func (c ConnState) String() string {
	return stateName[c]
}

type conn struct {
	p *Proxy
	c net.Conn

	curState struct{ atomic uint64 } // packed (unixtime<<8|uint8(ConnState))
}

func (c *conn) setState(state ConnState) {
	switch state {
	case StateNew:
		c.p.trackConn(c, true)

	case StateClosed:
		c.p.trackConn(c, false)

	}

	if state > 0xff || state < 0 {
		panic("internal error")
	}

	packedState := uint64(time.Now().Unix()<<8) | uint64(state)
	atomic.StoreUint64(&c.curState.atomic, packedState)
}

func (c *conn) getState() (state ConnState, unixSec int64) {
	packedState := atomic.LoadUint64(&c.curState.atomic)
	return ConnState(packedState & 0xff), int64(packedState >> 8)
}
