package schema

import (
	"context"
	"fmt"
	"regexp"

	"go.mongodb.org/mongo-driver/bson"
)

var (
	re = regexp.MustCompile(`\.(\$\[*\w*\]*|\d+)`)
)

// Validate checks datatype aganst CollectionField
// It validates either Insert or Update
// TODO: remove the .Map()?
func Validate(ctx context.Context, obj bson.D, fields map[string]CollectionField, denyUnknownFields, isUpdate bool) error {
	objMap := obj.Map()

	// Check for required fields
	for k, f := range fields {
		objV, ok := objMap[k]
		// If this is an update all `Required` fields should already be set (upper layers
		// ensure that they aren't un-set). Otherwise we'd require all required fields in
		// the update doc which is impractical
		if !isUpdate && f.Required && !ok {
			return fmt.Errorf("missing required field: %s", k)
		}

		if ok {
			if err := f.Validate(ctx, objV, denyUnknownFields, isUpdate); err != nil {
				return err
			}
		}
	}

	if denyUnknownFields {
		for k := range objMap {
			if _, ok := fields[k]; !ok {
				return fmt.Errorf("unknown fields are not allowed")
			}
		}
	}
	return nil
}

func SetValue(m bson.M, key []string, v interface{}) error {
	// Fast path
	if len(key) == 1 {
		m[key[0]] = v
		return nil
	}

	for _, k := range key[:len(key)-1] {
		v, ok := m[k]
		if !ok {
			v = bson.M{}
			m[k] = v
		}
		vM, ok := v.(bson.M)
		if !ok {
			return fmt.Errorf("cannot set %s -> %s as intermediate is already set: %T", key, k, v)
		}
		m = vM
	}

	m[key[len(key)-1]] = v
	return nil
}

func ToBsonD(m bson.M) bson.D {
	r := make(bson.D, 0, len(m))

	for k, v := range m {
		switch vTyped := v.(type) {
		case bson.M:
			r = append(r, bson.E{k, ToBsonD(vTyped)})
		default:
			r = append(r, bson.E{k, v})
		}
	}

	return r
}

// Map creates a map from the elements of the D.
// It makes additional process for arrays
func Mapify(d bson.D) bson.M {
	m := make(bson.M, len(d))
	for _, e := range d {
		e := processArray(e)
		m[e.Key] = e.Value
	}
	return m
}

func processArray(e bson.E) bson.E {
	if match := re.Find([]byte(e.Key)); match != nil { // detect array
		// remove positional info, so tree traverse can proceed
		e.Key = re.ReplaceAllString(e.Key, "")
	}
	return e
}

type WalkCollectionFieldsFunc func(string, *CollectionField) error

func WalkCollectionFields(fields map[string]CollectionField, fn WalkCollectionFieldsFunc) error {
	for k, v := range fields {
		if err := fn(k, &v); err != nil {
			return err
		}
		fields[k] = v
		if v.SubFields != nil {
			if err := WalkCollectionFields(v.SubFields, fn); err != nil {
				return err
			}
			fields[k] = v
		}
	}

	return nil
}