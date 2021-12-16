package bsonutil

import (
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func newValueReaderItem(v interface{}) *valueReaderItem {
	switch vTyped := v.(type) {
	case primitive.D:
		return &valueReaderItem{v: vTyped}
	case string:
		return &valueReaderItem{
			t: bsontype.String,
			v: vTyped,
		}
	case bool:
		return &valueReaderItem{
			t: bsontype.Boolean,
			v: vTyped,
		}
	case int32:
		return &valueReaderItem{
			t: bsontype.Int32,
			v: vTyped,
		}
	case *int32:
		return &valueReaderItem{
			t: bsontype.Int32,
			v: vTyped,
		}
	case int64:
		return &valueReaderItem{
			t: bsontype.Int64,
			v: vTyped,
		}
	case int:
		return &valueReaderItem{
			t: bsontype.Int64,
			v: vTyped,
		}
	case *int64:
		return &valueReaderItem{
			t: bsontype.Int64,
			v: vTyped,
		}
	case float64:
		return &valueReaderItem{
			t: bsontype.Double,
			v: vTyped,
		}
	case *float64:
		return &valueReaderItem{
			t: bsontype.Double,
			v: vTyped,
		}
	case []primitive.D:
		return &valueReaderItem{
			t: bsontype.Array,
			v: vTyped,
		}
	case primitive.A:
		return &valueReaderItem{
			t: bsontype.Array,
			v: vTyped,
		}
	case []int64:
		return &valueReaderItem{
			t: bsontype.Array,
			v: vTyped,
		}
	case primitive.Binary:
		return &valueReaderItem{
			t: bsontype.Binary,
			v: vTyped,
		}
	case primitive.ObjectID:
		return &valueReaderItem{
			t: bsontype.ObjectID,
			v: vTyped,
		}
	case primitive.DateTime:
		return &valueReaderItem{
			t: bsontype.DateTime,
			v: vTyped,
		}
	case primitive.Timestamp:
		return &valueReaderItem{
			t: bsontype.Timestamp,
			v: vTyped,
		}
	case primitive.Decimal128:
		return &valueReaderItem{
			t: bsontype.Decimal128,
			v: vTyped,
		}
	case *primitive.Decimal128:
		return &valueReaderItem{
			t: bsontype.Decimal128,
			v: vTyped,
		}
	case primitive.JavaScript:
		return &valueReaderItem{
			t: bsontype.JavaScript,
			v: vTyped,
		}
	case primitive.Regex:
		return &valueReaderItem{
			t: bsontype.Regex,
			v: vTyped,
		}
	case primitive.CodeWithScope:
		return &valueReaderItem{
			t: bsontype.CodeWithScope,
			v: vTyped,
		}
	case primitive.Undefined:
		return &valueReaderItem{
			t: bsontype.Undefined,
			v: vTyped,
		}
	case primitive.DBPointer:
		return &valueReaderItem{
			t: bsontype.DBPointer,
			v: vTyped,
		}
	case primitive.MaxKey:
		return &valueReaderItem{
			t: bsontype.MaxKey,
			v: vTyped,
		}
	case primitive.MinKey:
		return &valueReaderItem{
			t: bsontype.MinKey,
			v: vTyped,
		}
	case nil:
		return &valueReaderItem{
			t: bsontype.Null,
			v: vTyped,
		}
	default:
		fmt.Println(v)
		fmt.Println("UNKNOWNrecurse", reflect.TypeOf(v))
		panic("UNKNOWN TYPE")
	}

	return nil
}

type valueReaderItem struct {
	t bsontype.Type

	v interface{} // TODO: other types
}

func NewValueReader(in interface{}, disallow bool) *BSONValueReader {
	readerItem := newValueReaderItem(in)
	if readerItem == nil {
		return nil
	}

	r := &BSONValueReader{
		current:               readerItem,
		disallowUnknownFields: disallow,
	}
	return r
}

func NewStrictValueReader(in interface{}) *BSONValueReader {
	return NewValueReader(in, true)
}

type BSONValueReader struct {
	current               *valueReaderItem
	disallowUnknownFields bool
	key                   string
}

func (r *BSONValueReader) SetKey(k string) {
	r.key = k
}

func (r *BSONValueReader) Type() bsontype.Type {
	return r.current.t
}

// Skip is called when a value is going to be skipped; so this allows us to be "pedantic"
// or return an error if there is an unknown field
func (r *BSONValueReader) Skip() error {
	if r.disallowUnknownFields {
		return fmt.Errorf("unrecognized field '%s'", r.key)
	}
	return nil
}

func (r *BSONValueReader) ReadArray() (bsonrw.ArrayReader, error) {
	switch v := r.current.v.(type) {
	case primitive.A:
		return NewArrayReader(v, r.disallowUnknownFields), nil
	case []primitive.D:
		return NewArrayReader(v, r.disallowUnknownFields), nil
	case []int64:
		return NewArrayReader(v, r.disallowUnknownFields), nil
	}
	return nil, fmt.Errorf("unknown type (ReadArray)")
}

func (r *BSONValueReader) ReadBinary() (b []byte, btype byte, err error) {
	bin := r.current.v.(primitive.Binary)
	return bin.Data, bin.Subtype, nil
}

func (r *BSONValueReader) ReadBoolean() (bool, error) {
	return r.current.v.(bool), nil
}

func (r *BSONValueReader) ReadDocument() (bsonrw.DocumentReader, error) {
	return NewDocumentReader(r.current.v.(primitive.D), r.disallowUnknownFields), nil
}

func (r *BSONValueReader) ReadCodeWithScope() (code string, dr bsonrw.DocumentReader, err error) {
	v := r.current.v.(primitive.CodeWithScope)
	return string(v.Code), NewDocumentReader(v.Scope, r.disallowUnknownFields), nil
}

func (r *BSONValueReader) ReadDBPointer() (ns string, oid primitive.ObjectID, err error) {
	switch v := r.current.v.(type) {
	case primitive.DBPointer:
		return v.DB, v.Pointer, nil
	}

	return "", primitive.ObjectID{}, fmt.Errorf("unknown type (ReadDBPointer)")
}

func (r *BSONValueReader) ReadDateTime() (int64, error) {
	return int64(r.current.v.(primitive.DateTime)), nil
}

func (r *BSONValueReader) ReadDecimal128() (primitive.Decimal128, error) {
	switch v := r.current.v.(type) {
	case primitive.Decimal128:
		return v, nil
	case *primitive.Decimal128:
		return *v, nil
	}
	return primitive.Decimal128{}, fmt.Errorf("unknown type (ReadDecimal128)")
}

func (r *BSONValueReader) ReadDouble() (float64, error) {
	switch v := r.current.v.(type) {
	case float64:
		return v, nil
	case *float64:
		return *v, nil
	}
	return 0, fmt.Errorf("unknown type (ReadDouble)")
}

func (r *BSONValueReader) ReadInt32() (int32, error) {
	switch v := r.current.v.(type) {
	case int32:
		return v, nil
	case *int32:
		return *v, nil
	}
	return 0, fmt.Errorf("unknown type (ReadInt32)")
}

func (r *BSONValueReader) ReadInt64() (int64, error) {
	switch v := r.current.v.(type) {
	case int64:
		return v, nil
	case *int64:
		return *v, nil
	case int:
		return int64(v), nil
	}
	return 0, fmt.Errorf("unknown type (ReadInt64)")
}

func (r *BSONValueReader) ReadJavascript() (code string, err error) {
	return string(r.current.v.(primitive.JavaScript)), nil
}

func (r *BSONValueReader) ReadMaxKey() error {
	switch r.current.v.(type) {
	case primitive.MaxKey:
		return nil
	}

	return fmt.Errorf("unknown type (ReadMaxKey)")
}

func (r *BSONValueReader) ReadMinKey() error {
	switch r.current.v.(type) {
	case primitive.MinKey:
		return nil
	}

	return fmt.Errorf("unknown type (ReadMinKey)")
}

func (r *BSONValueReader) ReadNull() error {
	return nil
}

func (r *BSONValueReader) ReadObjectID() (primitive.ObjectID, error) {
	switch v := r.current.v.(type) {
	case primitive.ObjectID:
		return v, nil
	case *primitive.ObjectID:
		return *v, nil
	}

	return primitive.ObjectID{}, fmt.Errorf("unknown type (ReadObjectID)")
}

func (r *BSONValueReader) ReadRegex() (pattern, options string, err error) {
	v := r.current.v.(primitive.Regex)
	return v.Pattern, v.Options, nil
}

func (r *BSONValueReader) ReadString() (string, error) {
	return r.current.v.(string), nil
}

func (r *BSONValueReader) ReadSymbol() (symbol string, err error) {
	return "", fmt.Errorf("ReadSymbolnot implemented")
}

func (r *BSONValueReader) ReadTimestamp() (t, i uint32, err error) {
	switch v := r.current.v.(type) {
	case primitive.Timestamp:
		return v.T, v.I, nil
	}

	return 0, 0, fmt.Errorf("unknown type (ReadTimestamp)")
}

func (r *BSONValueReader) ReadUndefined() error {
	switch r.current.v.(type) {
	case primitive.Undefined:
		return nil
	}

	return fmt.Errorf("unknown type (ReadUndefined)")
}
