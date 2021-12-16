package insort

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/sirupsen/logrus"

	"github.com/wish/mongoproxy/pkg/mongoerror"
)

func genericLess(a, b interface{}) bool {
	switch valTyped := a.(type) {
	case primitive.ObjectID:
		switch other := b.(type) {
		case primitive.ObjectID:
			return bytes.Compare(valTyped[:], other[:]) < 0
		case nil:
			return false
		default:
			return false
		}
	case int:
		return valTyped < b.(int)
	case int32:
		return valTyped < b.(int32)
	case int64:
		return valTyped < b.(int64)
	case float64:
		return valTyped < b.(float64)
	case string:
		return valTyped < b.(string)
	case primitive.Timestamp:
		if valTyped.T < b.(primitive.Timestamp).T {
			return true
		}

		if valTyped.T == b.(primitive.Timestamp).T && valTyped.I < b.(primitive.Timestamp).I {
			return true
		}
		return false
	// TODO: fix? This is only the case in expressions where in includes a list such as
	// DEBU[2021-01-06T17:06:19-08:00] IN OP_MSG 1 {"Header":{"MessageLength":176,"RequestID":128,"ResponseTo":0,"OpCode":2013},"Flags":0,"Sections":[{"Document":[{"Key":"find","Value":"expr_index_use"},{"Key":"filter","Value":[{"Key":"$expr","Value":[{"Key":"$in","Value":["$x",[1,3]]}]}]},{"Key":"lsid","Value":[{"Key":"id","Value":{"Subtype":4,"Data":"oDfGGQjzRPGjz7gl25dJdg=="}}]},{"Key":"$db","Value":"test"}]}],"Checksum":0}
	case primitive.A:
		return false
	case primitive.Regex:
		var otherVal string
		switch otherValTyped := b.(type) {
		case string:
			otherVal = otherValTyped
		case primitive.Regex:
			otherVal = otherValTyped.Pattern
		default:
			panic("what")
		}
		return strings.Compare(valTyped.Pattern, otherVal) < 0
	case primitive.Decimal128:
		aF, aS := valTyped.GetBytes()
		bF, bS := b.(primitive.Decimal128).GetBytes()
		if aF < bF {
			return true
		} else if bF < aF {
			return false
		} else {
			return aS < bS
		}
	case primitive.DateTime:
		return valTyped < b.(primitive.DateTime)
	case primitive.Symbol:
		return strings.Compare(string(valTyped), string(b.(primitive.Symbol))) < 0
	case primitive.Binary:
		return bytes.Compare(valTyped.Data, b.(primitive.Binary).Data) < 0

	default:
		logrus.Errorf("sorting unknown for %T", valTyped)
		return false
	}
}

func PreprocessFilter(filter bson.D, inLimit int) error {
	for _, element := range filter {
		switch elemValTyped := element.Value.(type) {
		case bson.D:
			if err := sortInQuery(elemValTyped, inLimit); err != nil {
				return err
			}
		case primitive.A:
			if err := preprocessInterfaceFilter(elemValTyped, inLimit); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

type GenericSlice []interface{}

func (p GenericSlice) Len() int { return len(p) }
func (p GenericSlice) Less(i, j int) bool {
	return genericLess(p[i], p[j])
}
func (p GenericSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p GenericSlice) Sort() { sort.Sort(p) }

type BsonDSlice []bson.E

func (p BsonDSlice) Len() int { return len(p) }
func (p BsonDSlice) Less(i, j int) bool {
	return genericLess(p[i], p[j])
}
func (p BsonDSlice) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

func (p BsonDSlice) Sort() { sort.Sort(p) }

func preprocessInterfaceFilter(filters []interface{}, inLimit int) error {
	for _, element := range filters {
		switch elemValTyped := element.(type) {
		case bson.D:
			if err := PreprocessFilter(elemValTyped, inLimit); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

func sortInQuery(d bson.D, inLimit int) error {
	for _, subElement := range d {
		switch subElement.Key {
		case "$in", "$nin":
			switch valTyped := subElement.Value.(type) {
			case bson.D:
				if len(valTyped) <= 0 {
					continue
				}
				switch subElement.Key {
				case "$in":
					if inLimit > 0 && len(valTyped) > inLimit {
						return InLenError{"$in", len(valTyped)}
					}
				case "$nin":
					if inLimit > 0 && len(valTyped) > inLimit {
						return InLenError{"$nin", len(valTyped)}
					}
				}
				BsonDSlice(valTyped).Sort()
			case primitive.A:
				if len(valTyped) <= 0 {
					continue
				}
				switch subElement.Key {
				case "$in":
					if inLimit > 0 && len(valTyped) > inLimit {
						return InLenError{"$in", len(valTyped)}
					}
				case "$nin":
					if inLimit > 0 && len(valTyped) > inLimit {
						return InLenError{"$nin", len(valTyped)}
					}
				}
				GenericSlice(valTyped).Sort()
			default:
				return fmt.Errorf("$in expression must be list")
			}
		}
	}
	return nil
}

type InLenError struct {
	clause string
	count  int
}

func (i InLenError) Error() string {
	return fmt.Sprintf("%s clause longer than limit of %d", i.clause, i.count)
}

// TODO: better interface?
// TODO: better code?
func (i InLenError) BSONError() bson.D {
	return mongoerror.IllegalOperation.ErrMessage(i.Error())
}
