package ioutil

import (
	"strconv"
	"testing"
)

func TestLimitedWriter(t *testing.T) {
	tests := []struct {
		l   int
		in  []string
		out string
	}{
		// Example where everything lines up
		{
			l:   6,
			in:  []string{"foobar"},
			out: "foobar",
		},
		// truncate with multiple writes
		{
			l:   1,
			in:  []string{"f", "o", "o", "b", "a", "r"},
			out: "f",
		},
		// Truncate with single write
		{
			l:   5,
			in:  []string{"foobar"},
			out: "fooba",
		},
		// truncate with multiple writes
		{
			l:   5,
			in:  []string{"f", "o", "o", "b", "a", "r"},
			out: "fooba",
		},
		// too large of buffer
		{
			l:   7,
			in:  []string{"foobar"},
			out: "foobar",
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			w := NewLimitedWriter(make([]byte, test.l))
			for _, in := range test.in {
				w.Write([]byte(in)) // TODO: check return
			}

			if test.out != string(w.Get()) {
				t.Fatalf("Mismatch in expected output expected=%s actual=%s", test.out, w.Get())
			}
		})
	}
}
