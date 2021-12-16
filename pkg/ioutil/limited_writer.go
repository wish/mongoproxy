package ioutil

import (
	"io"
)

func NewLimitedWriter(b []byte) *LimitedWriter {
	return &LimitedWriter{b: b}
}

// LimitedWriter simply is a writer that will return errors once its buffer is filled
// This is esecially useful when you want to limit the output of a given writer
type LimitedWriter struct {
	b []byte
	o int
}

func (l *LimitedWriter) Write(p []byte) (n int, err error) {
	if l.o >= len(l.b) {
		return 0, io.ErrUnexpectedEOF
	}

	n = len(l.b) - l.o
	if len(p) < n {
		n = len(p)
	}

	n = copy(l.b[l.o:], p[:n])
	l.o += n
	return
}

func (l *LimitedWriter) Get() []byte {
	return l.b[:l.o]
}
