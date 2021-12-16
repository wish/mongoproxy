package bsonutil

import (
	"reflect"
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		in   bson.D
		keys []string
		out  interface{}
		ok   bool
	}{
		{
			in:   bson.D{{"a", 1}},
			keys: []string{"a"},
			out:  1,
			ok:   true,
		},
		{
			in:   bson.D{{"a", 1}},
			keys: []string{"b"},
			out:  nil,
			ok:   false,
		},
		{
			in:   bson.D{{"a", bson.D{{"b", 1}}}},
			keys: []string{"a", "b"},
			out:  1,
			ok:   true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v, ok := Lookup(test.in, test.keys...)

			if ok != test.ok {
				t.Fatalf("Mismatch in ok: expected=%v actual=%v", test.ok, ok)
			}

			if v != test.out {
				t.Fatalf("Mismatch in value: expected=%v actual=%v", test.out, v)
			}
		})
	}
}

func TestPop(t *testing.T) {
	tests := []struct {
		in      bson.D
		inAfter bson.D
		keys    []string
		out     interface{}
		ok      bool
	}{
		{
			in:      bson.D{{"a", 1}},
			inAfter: bson.D{},
			keys:    []string{"a"},
			out:     1,
			ok:      true,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			inAfter, v, ok := Pop(test.in, test.keys...)

			if ok != test.ok {
				t.Fatalf("Mismatch in ok: expected=%v actual=%v", test.ok, ok)
			}

			if v != test.out {
				t.Fatalf("Mismatch in value: expected=%v actual=%v", test.out, v)
			}

			if !reflect.DeepEqual(inAfter, test.inAfter) {
				t.Fatalf("in not changed expected=%v actual=%v", test.inAfter, inAfter)
			}
		})
	}
}
