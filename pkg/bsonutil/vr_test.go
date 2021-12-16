package bsonutil

import (
	"errors"
	"fmt"
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
)

type SafeBool bool

func (t *SafeBool) UnmarshalBSONValue(typ bsontype.Type, b []byte) error {
	r := bsonrw.NewBSONValueReader(typ, b)
	switch typ {
	case bsontype.String:
		s, err := r.ReadString()
		if err != nil {
			return err
		}
		v, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		*t = SafeBool(v)
	case bsontype.Boolean:
		v, err := r.ReadBoolean()
		if err != nil {
			return err
		}
		*t = SafeBool(v)
	case bsontype.Int32:
		v, err := r.ReadInt32()
		if err != nil {
			return err
		}
		*t = SafeBool(v > 0)
	case bsontype.Int64:
		v, err := r.ReadInt64()
		if err != nil {
			return err
		}
		*t = SafeBool(v > 0)
	case bsontype.Decimal128:
		v, err := r.ReadDecimal128()
		if err != nil {
			return err
		}
		if !v.IsZero() {
			*t = SafeBool(true)
		}

	case bsontype.Double:
		v, err := r.ReadDouble()
		if err != nil {
			return err
		}
		*t = SafeBool(v > 0)
	default:
		return errors.New("Unknown type") // TODO: better error
	}

	return nil
}
func (t SafeBool) String() string {
	if bool(t) {
		return "true"
	}
	return "false"
}

func TestDebug(t *testing.T) {
	doc := bson.D{{"b", "true"}}

	type B struct {
		B *SafeBool `bson:"b"`
	}

	dec, err := bson.NewDecoder(NewValueReader(doc, false))
	if err != nil {
		t.Fatal(err)
	}

	var result B
	if err := dec.Decode(&result); err != nil {
		t.Fatal(err)
	}

	fmt.Println("result", result.B)
}

func TestDebug2(t *testing.T) {
	doc := bson.D{{"b", true}}

	type B struct {
		B *SafeBool `bson:"b"`
	}

	dec, err := bson.NewDecoder(NewValueReader(doc, false))
	if err != nil {
		t.Fatal(err)
	}

	var result B
	if err := dec.Decode(&result); err != nil {
		t.Fatal(err)
	}

	fmt.Println("result", result.B)
}
