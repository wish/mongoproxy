package mongowire

import (
	"io"

	"github.com/sirupsen/logrus"
)

type Request struct {
	hdr MessageHeader
	crc Crc32c
	r   io.Reader
}

func NewRequestWithHeader(h MessageHeader, c io.Reader) *Request {
	return &Request{
		hdr: h,
		r:   c,
	}
}

func NewRequest(c io.Reader) (*Request, error) {
	req := &Request{}
	req.crc.Init()
	tr := io.TeeReader(c, &req.crc)
	h, err := ReadHeader(tr)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("Header=%s\n", h)
	req.hdr = *h

	req.r = io.LimitReader(tr, int64(h.MessageLength-HeaderLen))
	return req, nil
}

func (req *Request) GetHeader() *MessageHeader {
	return &req.hdr
}

func (req *Request) GetOpQuery() *OP_QUERY {
	q := &OP_QUERY{
		Header: req.hdr,
	}
	q.FromWire(req.r)
	return q
}

func (req *Request) GetOpKillCursors() *OP_KILL_CURSORS {
	q := &OP_KILL_CURSORS{
		Header: req.hdr,
	}
	q.FromWire(req.r)
	return q
}

func (req *Request) GetOpMore() *OP_GETMORE {
	gm := &OP_GETMORE{
		Header: req.hdr,
	}
	gm.FromWire(req.r)
	return gm
}

func (req *Request) GetOpMsg() *OP_MSG {
	o := &OP_MSG{
		Header: req.hdr,
	}
	o.FromWire(req.r, &req.crc, int(req.hdr.MessageLength-HeaderLen))
	return o
}

func (req *Request) GetOpCompressed() *OP_COMPRESSED {
	o := &OP_COMPRESSED{
		Header: req.hdr,
	}
	o.FromWire(req.r)
	return o
}
