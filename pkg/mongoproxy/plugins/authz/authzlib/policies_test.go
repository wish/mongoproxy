package authzlib

import (
	"reflect"
	"sort"
	"strconv"
	"testing"
)

func TestRuleSlice(t *testing.T) {
	tests := []struct {
		in  []Rule
		out []Rule
	}{
		{
			in: []Rule{
				{Effect: allowE},
				{Effect: denyE},
			},
			out: []Rule{
				{Effect: denyE},
				{Effect: allowE},
			},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			sort.Sort(RuleSlice(test.in))
			if !reflect.DeepEqual(test.in, test.out) {
				t.Fatalf("RuleSlice order issues \nexpected=%+v \nactual=%+v", test.out, test.in)
			}
		})
	}
}
