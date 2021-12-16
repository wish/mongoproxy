package bsonutil

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const MaxBsonObjectSize = 16777216

func Lookup(in bson.D, keys ...string) (interface{}, bool) {
	if len(keys) < 1 {
		return nil, false
	}
	for i, k := range keys {
		for _, item := range in {
			if item.Key == k {
				// last
				if i == len(keys)-1 {
					return item.Value, true
				} else {
					newIn, ok := item.Value.(bson.D)
					if !ok {
						return nil, false
					}
					in = newIn
					break
				}
			}
		}
	}

	return nil, false
}

func GetBoolDefault(in *bool, d bool) bool {
	if in == nil {
		return d
	}
	return *in
}

func Pop(in bson.D, keys ...string) (bson.D, interface{}, bool) {
	if len(keys) < 1 {
		return in, nil, false
	}

	for i, item := range in {
		if item.Key == keys[0] {
			if len(keys) > 1 {
				newIn, ok := item.Value.(bson.D)
				if !ok {
					return in, nil, false
				}
				result, v, ok := Pop(newIn, keys[1:]...)
				in[i].Value = result
				return in, v, ok
			} else {
				// Remove the element at index i from a.
				copy(in[i:], in[i+1:])        // Shift a[i+1:] left one index.
				in[len(in)-1] = primitive.E{} // Erase last element (write zero value).
				return in[:len(in)-1], item.Value, true
			}
		}
	}

	return in, nil, false
}

// Ok returns the "ok" status of the result. This is required as mongo
// is very inconsistent on the type it uses for "ok"; so this saves all
// of the type switching across the codebase
func Ok(in bson.D) bool {
	if v, ok := Lookup(in, "ok"); ok && v != nil {
		return BoolNumber(v)
	}
	return false
}

// BoolNumber returns the "bool" status of the result (assuming its a number)
func BoolNumber(v interface{}) bool {
	switch vTyped := v.(type) {
	case int:
		return vTyped == 1
	case int32:
		return vTyped == 1
	case int64:
		return vTyped == 1
	case float64:
		return vTyped == 1
	}
	return false
}
