package bsonutil

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func NewDocumentReader(v interface{}, disallow bool) *DocumentReader {
	return &DocumentReader{V: v, disallowUnknownFields: disallow}
}

type DocumentReader struct {
	V                     interface{}
	offset                int
	started               bool
	disallowUnknownFields bool
}

func (r *DocumentReader) ReadElement() (string, bsonrw.ValueReader, error) {
	if !r.started {
		r.started = true
	} else {
		r.offset++
	}

	switch v := r.V.(type) {
	case primitive.D:
		if r.offset >= len(v) {
			return "", nil, bsonrw.ErrEOD
		}

		vr := NewValueReader(v[r.offset].Value, r.disallowUnknownFields)
		vr.SetKey(v[r.offset].Key)

		return v[r.offset].Key, vr, nil
	}

	return "", nil, fmt.Errorf("unknown type")

}
