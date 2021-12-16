package insort

import (
	"reflect"
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestPreprocessInt(t *testing.T) {
	subExpr := map[string]interface{}{"$in": []int32{5, 4, 3, 2, 1}}
	filter := map[string]interface{}{"_id": subExpr}
	b, _ := bson.Marshal(filter)
	var bsonFilter bson.D
	bson.Unmarshal(b, &bsonFilter)
	PreprocessFilter(bsonFilter, -1)
	m := bsonFilter.Map()
	expected := []int32{1, 2, 3, 4, 5}
	actual := m["_id"].(bson.D).Map()["$in"].(primitive.A)
	if len(expected) != len(actual) {
		t.Fatalf("preprocessed $in query length not the same before and after")
	}
	for i := range expected {
		if expected[i] != actual[i].(int32) {
			t.Fatalf("$in query input not sorted: %v", actual)
		}
	}
}

func TestPreprocessObjectId(t *testing.T) {
	input := []primitive.ObjectID{primitive.NewObjectID(), primitive.NewObjectID()}
	reversed := make([]primitive.ObjectID, 2)
	for i, val := range input {
		reversed[(i+1)%2] = val
	}
	subExpr := map[string]interface{}{"$in": reversed}
	filter := map[string]interface{}{"_id": subExpr}
	b, _ := bson.Marshal(filter)
	var bsonFilter bson.D
	bson.Unmarshal(b, &bsonFilter)
	PreprocessFilter(bsonFilter, -1)
	m := bsonFilter.Map()
	expected := input
	actual := m["_id"].(bson.D).Map()["$in"].(primitive.A)
	if len(expected) != len(actual) {
		t.Fatalf("preprocessed $in query length not the same before and after")
	}
	for i := range expected {
		if expected[i] != actual[i].(primitive.ObjectID) {
			t.Fatalf("$in query input not sorted: %v", actual)
		}
	}
}

func TestPreprocessOr(t *testing.T) {
	subExpr := map[string]interface{}{"$in": []int32{5, 4, 3, 2, 1}}
	subExpr2 := map[string]interface{}{"$in": []int32{10, 9, 8, 7, 6}}
	orExpr := map[string]interface{}{"_id": subExpr}
	orExpr2 := map[string]interface{}{"_id": subExpr2}
	filter := map[string]interface{}{"$or": []interface{}{orExpr, orExpr2}}
	b, _ := bson.Marshal(filter)
	var bsonFilter bson.D
	bson.Unmarshal(b, &bsonFilter)
	PreprocessFilter(bsonFilter, -1)
	m := bsonFilter.Map()
	expected := []primitive.A{{int32(1), int32(2), int32(3), int32(4), int32(5)}, {int32(6), int32(7), int32(8), int32(9), int32(10)}}
	actual := m["$or"].(primitive.A)
	for i, actualVal := range actual {
		if len(expected[i]) != len(actualVal.(bson.D)[0].Value.(bson.D)[0].Value.(primitive.A)) {
			t.Fatalf("preprocessed $in query length not the same before and after")
		}
		if !reflect.DeepEqual(expected[i], actualVal.(bson.D)[0].Value.(bson.D)[0].Value.(primitive.A)) {
			t.Fatalf("$in query input not sorted, expected=%T actual=%T", expected[i][0], actualVal.(bson.D)[0].Value.(bson.D)[0].Value.(primitive.A)[0])
		}
	}
}

func TestPreprocessInLimitInt(t *testing.T) {
	subExpr := map[string]interface{}{"$in": []int{5, 4, 3, 2, 1}}
	filter := map[string]interface{}{"_id": subExpr}
	b, _ := bson.Marshal(filter)
	var bsonFilter bson.D
	bson.Unmarshal(b, &bsonFilter)
	if err := PreprocessFilter(bsonFilter, 2); err == nil {
		t.Fatalf("Missing expected error!")
	}
}

func TestGenericSliceSort(t *testing.T) {
	tests := []GenericSlice{
		{"a", "b", "c"},
		{1, 2, 3, 4, 5},
		{int32(1), int32(2), int32(3), int32(4), int32(5)},
		{int64(1), int64(2), int64(3), int64(4), int64(5)},
		{float64(1), float64(2), float64(3), float64(4), float64(5)},
		{nil, primitive.ObjectID{}},
		{primitive.Timestamp{T: 1}, primitive.Timestamp{T: 2}},
		{primitive.Regex{Pattern: "a"}, primitive.Regex{Pattern: "b"}},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			orig := test
			f := func(p GenericSlice) {
				if len(orig) != len(p) {
					t.Fatalf("Mismatch in len: %v %v", len(orig), len(p))
				}

				for i, x := range orig {
					if p[i] != x {
						t.Fatalf("Mismatch in %d: %v %v", i, p[i], x)
					}
				}
			}
			Perm(orig, f)
		})
	}
}

// Perm calls f with each permutation of a.
func Perm(a GenericSlice, f func(GenericSlice)) {
	perm(a, f, 0)
}

// Permute the values at index i to len(a)-1.
func perm(a GenericSlice, f func(GenericSlice), i int) {
	if i > len(a) {
		f(a)
		return
	}
	perm(a, f, i+1)
	for j := i + 1; j < len(a); j++ {
		a[i], a[j] = a[j], a[i]
		perm(a, f, i+1)
		a[i], a[j] = a[j], a[i]
	}
}
