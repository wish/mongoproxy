package bsonutil

import (
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestArrayReader(t *testing.T) {
	tests := []struct {
		a   interface{}
		n   int
		err bool
	}{
		// nil value; ensure we get an error
		{
			a:   nil,
			err: true,
		},
		// basic 1 value array
		{
			a: primitive.A{int64(1)},
			n: 1,
		},
		// basic 1 value array
		{
			a: []primitive.D{{{"a", int64(1)}}},
			n: 1,
		},
		// basic 2 value array
		{
			a: primitive.A{int64(1), int64(2)},
			n: 2,
		},

		// invalid array; ensure we get an error
		{
			a:   1,
			err: true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r := NewArrayReader(test.a, false)

			var (
				n       = 0
				testErr error
			)

			for {
				_, err := r.ReadValue()
				if err != nil {
					if err != bsonrw.ErrEOA {
						testErr = err
					}
					break
				}
				n++
			}

			if (testErr == nil) == test.err {
				t.Fatalf("mismatch in error err=%v expected=%v", testErr, test.err)
			}

			if n != test.n {
				t.Fatalf("Mismatch in items: expected %d actual %d", test.n, n)
			}

		})
	}
}
