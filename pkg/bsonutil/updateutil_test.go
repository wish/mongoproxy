package bsonutil

import (
	"reflect"
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestExpandUpdate(t *testing.T) {
	tests := []struct {
		in      bson.D
		upsert  bool
		c, u, d []string
	}{
		// Basic update (no operators)
		{
			in: bson.D{
				{"a", 1},
			},
			c: []string{},
			u: []string{"a"},
			d: []string{},
		},
		// Basic upsert (no operators)
		{
			in: bson.D{
				{"a", 1},
			},
			upsert: true,
			c:      []string{"a"},
			u:      []string{"a"},
			d:      []string{},
		},

		// $ operators
		{
			in: bson.D{
				{"$currentDate", bson.D{
					{"currentDate", 1},
				}},
				{"$inc", bson.D{
					{"inc", 1},
				}},
				{"$min", bson.D{
					{"min", 1},
				}},
				{"$max", bson.D{
					{"max", 1},
				}},
				{"$mul", bson.D{
					{"mul", 1},
				}},
				{"$set", bson.D{
					{"set", 1},
				}},
				{"$setOnInsert", bson.D{
					{"setOnInsert", 1},
				}},
				{"$unset", bson.D{
					{"unset", 1},
				}},
				{"$rename", bson.D{
					{"rename", "renameto"},
				}},
			},
			c: []string{"set", "setOnInsert", "renameto"},
			u: []string{"currentDate", "inc", "min", "max", "mul", "set", "renameto"},
			d: []string{"unset", "rename"},
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			out := ExpandUpdate(test.in, &test.upsert)

			if !reflect.DeepEqual(out.Create, test.c) {
				t.Errorf("Create Mismatch: %v != %v", out.Create, test.c)
			}

			if !reflect.DeepEqual(out.Update, test.u) {
				t.Errorf("Update Mismatch: %v != %v", out.Update, test.u)
			}

			if !reflect.DeepEqual(out.Delete, test.d) {
				t.Errorf("Delete Mismatch: %v != %v", out.Delete, test.d)
			}
		})
	}
}
