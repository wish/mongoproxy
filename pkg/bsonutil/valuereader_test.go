package bsonutil

import (
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestDecode(t *testing.T) {

	type testStruct struct {
		String        string                `bson:"string"`
		Document      bson.D                `bson:"document"`
		BoolPtr       *bool                 `bson:"boolPtr"`
		Bool          bool                  `bson:"bool"`
		Int32Ptr      *int32                `bson:"int32Ptr"`
		Int64Ptr      *int64                `bson:"int64Ptr"`
		IntPtr        *int                  `bson:"intPtr"`
		Int32         int32                 `bson:"int32"`
		Int64         int64                 `bson:"int64"`
		Int           int                   `bson:"int"`
		IntArray      []int                 `bson:"intArray"`
		Bytes         []byte                `bson:"bytes"`
		ObjectIDPtr   *primitive.ObjectID   `bson:"objectIDPtr"`
		ObjectID      primitive.ObjectID    `bson:"objectID"`
		Time          time.Time             `bson:"time"`
		Decmial128Ptr *primitive.Decimal128 `bson:"decimal128Ptr"`
		Decmial128    primitive.Decimal128  `bson:"decimal128"`
		DoublePtr     *float64              `bson:"doublePtr"`
		Double        float64               `bson:"double"`
		DArray        []primitive.D         `bson:"darray"`
		Array         primitive.A
		Binary        primitive.Binary
		DateTime      primitive.DateTime
		Timestamp     primitive.Timestamp
		JavaScript    primitive.JavaScript
		Regex         primitive.Regex
		//CodeWithScope primitive.CodeWithScope
		Undefined primitive.Undefined
		DBPointer primitive.DBPointer
		MaxKey    primitive.MaxKey
		MinKey    primitive.MinKey
		Nil       interface{}
	}

	tr := true
	var (
		i32        int32
		i64        int64
		i          int
		f64        float64
		objectID   = primitive.NewObjectID()
		decimal128 = primitive.NewDecimal128(1, 1)
	)
	before := testStruct{
		String:        "string",
		Document:      bson.D{{"a", "b"}},
		BoolPtr:       &tr,
		Bool:          tr,
		Int32Ptr:      &i32,
		Int64Ptr:      &i64,
		IntPtr:        &i,
		Int32:         i32,
		Int64:         i64,
		Int:           i,
		IntArray:      []int{1, 2, 3},
		Bytes:         []byte("foo"),
		ObjectIDPtr:   &objectID,
		ObjectID:      objectID,
		Time:          time.Now().UTC().Truncate(time.Millisecond), // Timestamps are UTC and millisecond
		Decmial128Ptr: &decimal128,
		Decmial128:    decimal128,
		DoublePtr:     &f64,
		Double:        f64,
		DArray:        []primitive.D{{{"a", "b"}}},
		Array:         primitive.A{"a"},
		Binary:        primitive.Binary{Subtype: 4, Data: []byte("foo")},
		DateTime:      primitive.NewDateTimeFromTime(time.Now()),
		Timestamp:     primitive.Timestamp{T: 100},
		JavaScript:    "a",
		Regex:         primitive.Regex{Pattern: ".*"},

		Undefined: primitive.Undefined{},
		DBPointer: primitive.DBPointer{DB: "foo", Pointer: objectID},
		MaxKey:    primitive.MaxKey{},
		MinKey:    primitive.MinKey{},
		Nil:       nil,
	}

	b, err := bson.Marshal(before)
	if err != nil {
		t.Fatal(err)
	}

	var doc bson.D
	if err := bson.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}

	dec, err := bson.NewDecoder(NewValueReader(doc, false))
	if err != nil {
		t.Fatal(err)
	}

	var result testStruct
	if err := dec.Decode(&result); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(before, result) {
		t.Fatalf("Mismatch expected=\n%v \nactual=\n%v", before, result)
	}

}

func TestDecodeExtra(t *testing.T) {
	type AB struct {
		A string `bson:"a"`
		B string `bson:"b"`
	}
	type A struct {
		A string `bson:"a"`
	}

	ab := AB{A: "a", B: "b"}

	b, err := bson.Marshal(ab)
	if err != nil {
		t.Fatal(err)
	}

	var doc bson.D
	if err := bson.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}

	dec, err := bson.NewDecoder(NewValueReader(doc, true))
	if err != nil {
		t.Fatal(err)
	}

	var result A
	err = dec.Decode(&result)
	if err == nil {
		t.Fatalf("missing expected error")
	}
	if err.Error() != "unrecognized field 'b'" {
		t.Fatalf("Incorrect error message")
	}
}
