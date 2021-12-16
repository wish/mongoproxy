package authzlib

import (
	"reflect"
	"strconv"
	"testing"
)

func TestExpandResource(t *testing.T) {
	tests := []struct {
		in  Resource
		out []Resource
	}{
		{in: Resource{Global: true}, out: []Resource{
			{Global: true},
		}},
		{in: Resource{false, "db", "col", "field"}, out: []Resource{
			// Field level
			{false, "db", "col", "field"},
			{false, "db", "col", "*"},
			{false, "db", "*", "field"},
			{false, "db", "*", "*"},
			{false, "*", "col", "field"},
			{false, "*", "col", "*"},
			{false, "*", "*", "field"},
			{false, "*", "*", "*"},
		}},
		{in: Resource{false, "db", "col", ""}, out: []Resource{
			{false, "db", "col", ""},
			{false, "db", "*", ""},
			{false, "*", "*", ""},
		}},
		{in: Resource{false, "db", "", ""}, out: []Resource{
			{false, "db", "", ""},
			{false, "*", "", ""},
		}},
		{in: Resource{false, "*", "", ""}, out: []Resource{
			{false, "*", "", ""},
		}},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out := expandResource(test.in)
			if !reflect.DeepEqual(out, test.out) {
				t.Fatalf("Mismatch in out \n\texpected=%v \n\tactual=%v", test.out, out)
			}
		})
	}
}
