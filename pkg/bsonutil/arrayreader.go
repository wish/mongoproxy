package bsonutil

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewArrayReader(a interface{}, disallow bool) *ArrayReader {
	return &ArrayReader{A: a, disallowUnknownFields: disallow}
}

type ArrayReader struct {
	A                     interface{}
	offset                int
	started               bool
	disallowUnknownFields bool
}

func (r *ArrayReader) ReadValue() (bsonrw.ValueReader, error) {
	if !r.started {
		r.started = true
	} else {
		r.offset++
	}

	switch a := r.A.(type) {
	case primitive.A:
		if r.offset >= len(a) {
			return nil, bsonrw.ErrEOA
		}
		return NewValueReader(a[r.offset], r.disallowUnknownFields), nil
	case []primitive.D:
		if r.offset >= len(a) {
			return nil, bsonrw.ErrEOA
		}

		return NewValueReader(a[r.offset], r.disallowUnknownFields), nil
	case []int64:
		if r.offset >= len(a) {
			return nil, bsonrw.ErrEOA
		}

		return NewValueReader(a[r.offset], r.disallowUnknownFields), nil
	}
	return nil, fmt.Errorf("unknown type")
}
