package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BSONType string

const (
	// BSON types, not all of them supported yet.
	INT             BSONType = "int"
	INT_ARRAY       BSONType = "[]int"
	LONG            BSONType = "long"
	LONG_ARRAY      BSONType = "[]long"
	DOUBLE          BSONType = "double"
	DOUBLE_ARRAY    BSONType = "[]double"
	STRING          BSONType = "string"
	STRING_ARRAY    BSONType = "[]string"
	OBJECT          BSONType = "object"
	OBJECT_ARRAY    BSONType = "[]object"
	BIN_DATA        BSONType = "binData"
	BIN_DATA_ARRAY  BSONType = "[]binData"
	OBJECT_ID       BSONType = "objectID"
	OBJECT_ID_ARRAY BSONType = "[]objectID"
	BOOL            BSONType = "bool"
	BOOL_ARRAY      BSONType = "[]bool"
	DATE            BSONType = "date"
	DATE_ARRAY      BSONType = "[]date"
	NULL            BSONType = "null"
	REGEX           BSONType = "regex"
	DECIMAL128      BSONType = "decimal"

	SKIP_SCHEMA_ANNOTATION = "skipSchema"
)

var OpMap = BuildUpdateOpSet()

type ClusterSchema struct {
	MongosEndpoint       string              `json:"mongosEndpoint"`
	Annotations          map[string]string   `json:"annotations,omitempty"`
	Databases            map[string]Database `json:"dbs"`
	DenyUnknownDatabases bool                `json:"denyUnknownDatabases,omitempty"`
}

func (s *ClusterSchema) UnmarshalJSON(data []byte) error {
	type Alias ClusterSchema
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Walk
	for dbName, db := range s.Databases {
		if strings.ToLower(dbName) != dbName {
			return fmt.Errorf("db names must be lowercase: %s", dbName)
		}
		// Remove dbs which have annotation
		if _, ok := db.Annotations[SKIP_SCHEMA_ANNOTATION]; ok {
			delete(s.Databases, dbName)
			continue
		}

		for collectionName, collection := range db.Collections {
			if strings.ToLower(collectionName) != collectionName {
				return fmt.Errorf("collection names must be lowercase: %s.%s", dbName, collectionName)
			}

			// Remove collections which have annotation
			if _, ok := collection.Annotations[SKIP_SCHEMA_ANNOTATION]; ok {
				delete(db.Collections, collectionName)
				continue
			}

			if err := WalkCollectionFields(collection.Fields, func(fName string, f *CollectionField) error {
				if strings.ToLower(collectionName) != collectionName {
					return fmt.Errorf("field names must be lowercase: %s.%s %s", dbName, collectionName, fName)
				}
				if strings.HasPrefix(string(f.Type), "[]") {
					f.IsArray = true
				}
				// TODO: check against types instead
				if strings.Contains(string(f.Type), ".") && f.remoteCollection == nil {
					ref := strings.Split(string(f.Type), ".")
					// handle remoteCollection array like "[]db.collection"
					ref[0] = strings.TrimPrefix(ref[0], "[]")
					d, ok := s.Databases[ref[0]]
					if !ok {
						return fmt.Errorf("invalid reference %s on %v", ref[0], f)
					}
					c, ok := d.Collections[ref[1]]
					if !ok {
						return fmt.Errorf("invalid reference %s on %v", ref[1], f)
					}
					f.remoteCollection = &c
				}
				return nil
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateInsert will validate the schema of the passed in object.
func (s *ClusterSchema) ValidateInsert(ctx context.Context, database, collection string, obj bson.D) error {
	db, ok := s.Databases[database]
	if !ok {
		if s.DenyUnknownDatabases {
			return fmt.Errorf("unknown DB %v not allowed", database)
		}
		return nil
	}

	return db.ValidateInsert(ctx, collection, obj)
}

// ValidateUpdate will validate the schema of the passed in object.
func (s *ClusterSchema) ValidateUpdate(ctx context.Context, database, collection string, obj bson.D, upsert bool) error {
	db, ok := s.Databases[database]
	if !ok {
		if s.DenyUnknownDatabases {
			return fmt.Errorf("unknown DB %v not allowed", database)
		}
		return nil
	}

	return db.ValidateUpdate(ctx, collection, obj, upsert)
}

type Database struct {
	Annotations            map[string]string     `json:"annotations,omitempty"`
	Collections            map[string]Collection `json:"collections"`
	DenyUnknownCollections bool                  `json:"denyUnknownCollections,omitempty"`
}

// ValidateInsert will validate the schema of the passed in object.
func (d *Database) ValidateInsert(ctx context.Context, collection string, obj bson.D) error {
	c, ok := d.Collections[collection]
	if !ok {
		if d.DenyUnknownCollections {
			return fmt.Errorf("unknown Collection %v not allowed", collection)
		}
		return nil
	}

	return c.ValidateInsert(ctx, obj)
}

// ValidateUpdate will validate the schema of the passed in object.
func (d *Database) ValidateUpdate(ctx context.Context, collection string, obj bson.D, upsert bool) error {
	c, ok := d.Collections[collection]
	if !ok {
		if d.DenyUnknownCollections {
			return fmt.Errorf("unknown Collection %v not allowed", collection)
		}
		return nil
	}

	return c.ValidateUpdate(ctx, obj, upsert)
}

type Collection struct {
	Annotations map[string]string `json:"annotations,omitempty"`
	// All the columns in this table
	Fields map[string]CollectionField `json:"fields"`
	// Whether we should strictly enforce fields or allow others
	DenyUnknownFields bool `json:"denyUnknownFields,omitempty"`
	// Whether we should enforce schema check for this collection
	EnforceSchema bool `json:"enforceSchema,omitempty"`
}

func (c *Collection) GetField(names ...string) *CollectionField {
	var field *CollectionField
	for _, name := range names {
		if field == nil {
			v, ok := c.Fields[name]
			if !ok {
				return nil
			}
			field = &v
		} else {
			var (
				v  CollectionField
				ok bool
			)
			if field.remoteCollection != nil {
				v, ok = field.remoteCollection.Fields[name]
			} else {
				v, ok = field.SubFields[name]
			}
			if !ok {
				logrus.Debugf("can not find field in collection: %s", name)
				return nil
			}
			field = &v
		}
	}
	return field
}

// ValidateInsert will validate the schema of the passed in object.
func (c *Collection) ValidateInsert(ctx context.Context, obj bson.D) error {
	if !c.EnforceSchema {
		return nil
	}
	return Validate(ctx, obj, c.Fields, c.DenyUnknownFields, false)
}

// ValidateUpdate will validate the schema of the passed in object.
func (c *Collection) ValidateUpdate(ctx context.Context, obj bson.D, upsert bool) error {
	/*
		$rename (rename fields -- dot-delimited names)
		$set (set field values -- dot-delimited names)
		$setOnInsert (set  fields on inesert -- dot-delimited names)
		$unset (unset fieds -- dot-delimited names)
	*/
	var (
		setFields    bson.M // Fields with values we have for our update
		insertFields bson.M // Insert fields (if we have them) -- only for upserts
		unsetFields  bson.M // fields being unset
		renameFields bson.M // fields being renamed
	)

	if !c.EnforceSchema {
		return nil
	}
	logrus.Debugf("pending validation object: %s", obj)
	/*
		update command support ,
		* A document that contains update operator expressions,
		* A replacement document with only <field1>: <value1> pairs
		https://www.mongodb.com/docs/v4.2/reference/command/update/#update-statement-documents
	*/
	if !strings.HasPrefix(obj[0].Key, "$") || !SetContain(OpMap, obj[0].Key) {
		m := make(bson.M, len(obj))
		if upsert {
			insertFields = handleObj(obj, m)
			logrus.Debugf("insertFields: %s", insertFields)
		} else {
			setFields = handleObj(obj, m)
			logrus.Debugf("setFields: %s", setFields)
		}
	} else {
		for _, e := range obj {
			// TODO: CHANGE TO DEBUGF
			logrus.Infof("update with operator: %s", e.Key)
			switch e.Key {
			case "$currentDate", "$inc", "$min", "$max", "$mul":
				if setFields == nil {
					setFields = e.Value.(bson.D).Map()
				} else {
					for _, item := range e.Value.(bson.D) {
						setFields[item.Key] = item.Value
					}
				}
			case "$rename":
				renameFields = e.Value.(bson.D).Map()
			case "$set", "$pull", "$pullAll":
				if setFields == nil {
					setFields = Mapify(e.Value.(bson.D))
				} else {
					for _, item := range e.Value.(bson.D) {
						item := processArray(item)
						setFields[item.Key] = item.Value
					}
				}
			case "$addToSet", "$push":
				if setFields == nil {
					setFields = make(bson.M, len(e.Value.(bson.D)))
				}
				setFields = MapifyWithOp(e.Value.(bson.D), setFields)
			case "$setOnInsert":
				insertFields = Mapify(e.Value.(bson.D))
			case "$unset":
				unsetFields = Mapify(e.Value.(bson.D))
			default:
				return fmt.Errorf("cannot recognize key: %s, value: %f, in obj: %s", e.Key, e.Value, e)
			}
		}
	}

	// Verify that unset fields aren't required
	for k := range unsetFields {
		f := c.GetField(strings.Split(k, ".")...)
		if c.DenyUnknownFields && f == nil {
			return fmt.Errorf("cannot unset unknown field: %s", k)
		}
		if f != nil && f.Required {
			return fmt.Errorf("cannot unset required field %s", k)
		}
	}

	// Verify that rename fields aren't required
	// Verify that rename fields types match (before and after)
	for oldK, newKRaw := range renameFields {
		newK, ok := newKRaw.(string)
		if !ok {
			return fmt.Errorf("malformed rename of %s", oldK)
		}
		// Check that the old field exists
		oldF := c.GetField(strings.Split(oldK, ".")...)
		if c.DenyUnknownFields && oldF == nil {
			return fmt.Errorf("cannot rename unknown field: %s", oldK)
		}
		// Check that the new field exists
		newF := c.GetField(strings.Split(newK, ".")...)
		if c.DenyUnknownFields && newF == nil {
			return fmt.Errorf("cannot rename unknown field: %s", newK)
		}
		// Check if the old field is required
		if oldF != nil && oldF.Required {
			return fmt.Errorf("cannot unset required field %s", oldK)
		}

		// Ensure matched types
		if oldF != nil && newF != nil {
			if oldF.Type != newF.Type {
				return fmt.Errorf("cannot rename %s -> %s; mismatched type %s -> %s", oldK, newK, oldF.Type, newF.Type)
			}
		}
	}

	// verify setFields are of the correct type
	for k, v := range setFields {
		f := c.GetField(strings.Split(k, ".")...)
		if c.DenyUnknownFields && f == nil {
			return fmt.Errorf("cannot set unknown field: %s", k)
		}
		//verify that setFields are required before validate
		if f != nil && f.Required && v == nil {
			return fmt.Errorf("cannot set a required field with nil value: %f", v)
		}
		if f != nil && v != nil {
			if err := f.Validate(ctx, v, c.DenyUnknownFields, true); err != nil {
				return err
			}
		}
	}

	// if upsert, check an insert as well
	if upsert {
		doc := make(bson.M, len(setFields)+len(insertFields))
		logrus.Debugf("upsert doc built")
		for k, v := range setFields {
			if err := SetValue(doc, strings.Split(k, "."), v); err != nil {
				return err
			}
		}
		logrus.Debugf("finished setField setValue")

		for k, v := range insertFields {
			if err := SetValue(doc, strings.Split(k, "."), v); err != nil {
				return err
			}
		}
		if err := Validate(ctx, ToBsonD(doc), c.Fields, c.DenyUnknownFields, true); err != nil {
			return err
		}
		logrus.Debugf("finished Validate upsert true")
	}
	return nil
}

type CollectionField struct {
	Name             string            `json:"alias,omitempty"`
	Type             BSONType          `json:"type"`
	remoteCollection *Collection       // Pointer to remote collection (for fields if the type is "foo.bar")
	Annotations      map[string]string `json:"annotations,omitempty"`

	// Various configuration options
	Required bool `json:"required,omitempty"`
	//Default interface{} `json:"default,omitempty"`

	// Field is a array type
	IsArray bool

	// Optional subfields
	SubFields map[string]CollectionField `json:"subfields,omitempty"`
}

// Validate elements type inside an array
func (c *CollectionField) ValidateElement(ctx context.Context, d interface{}, validateType BSONType) error {
	ok := false
	switch validateType {
	case INT:
		switch d.(type) {
		case int, int64, int32:
			ok = true
		}
	case LONG:
		switch d.(type) {
		case int, int64, int32:
			ok = true
		}
	case DOUBLE:
		switch d.(type) {
		case int, int32, int64, float32, float64:
			ok = true
		}
	case STRING:
		switch d.(type) {
		case string:
			ok = true
		}
	case BIN_DATA:
		switch d.(type) {
		case primitive.Binary:
			ok = true
		}
	case OBJECT_ID:
		switch d.(type) {
		case primitive.ObjectID:
			ok = true
		}
	case BOOL:
		switch d.(type) {
		case bool:
			ok = true
		}
	case DATE:
		switch d.(type) {
		case int64, primitive.DateTime:
			ok = true
		}
	}

	if !ok {
		return fmt.Errorf("wrong element data type: %T", d)
	}
	return nil
}

// ValidateInsert will validate the schema of the passed in object.
func (c *CollectionField) Validate(ctx context.Context, v interface{}, denyUnknownFields, isUpdate bool) error {
	validateType := c.Type
	interfaceType := fmt.Sprint(reflect.TypeOf(v))
	if isUpdate { // array update is validating a scalar instead of []
		if !strings.HasPrefix(interfaceType, "[]") && interfaceType != "primitive.A" {
			validateType = BSONType(strings.Trim(string(validateType), "[]"))
		}
	}
	ok := false
	switch validateType {
	case INT:
		switch v.(type) {
		case int, int64, int32:
			ok = true
		}
	case INT_ARRAY:
		switch vTyped := v.(type) {
		case primitive.A:
			ok = true
			for _, d := range vTyped {
				if err := c.ValidateElement(ctx, d, INT); err != nil {
					return fmt.Errorf("%s: []int has non int element: %s", c.Name, err)
				}
			}
		}
	case LONG:
		switch v.(type) {
		case int, int32, int64:
			ok = true
		}
	case LONG_ARRAY:
		switch vTyped := v.(type) {
		case primitive.A:
			ok = true
			for _, d := range vTyped {
				if err := c.ValidateElement(ctx, d, LONG); err != nil {
					return fmt.Errorf("%s: []long has non long element: %s", c.Name, err)
				}
			}
		}
	case DOUBLE:
		switch v.(type) {
		case int, int32, int64, float32, float64:
			ok = true
		}
	case DOUBLE_ARRAY:
		switch vTyped := v.(type) {
		case primitive.A:
			ok = true
			for _, d := range vTyped {
				if err := c.ValidateElement(ctx, d, DOUBLE); err != nil {
					return fmt.Errorf("%s: []double has non double element: %s", c.Name, err)
				}
			}
		}
	case STRING:
		switch v.(type) {
		case string:
			ok = true
		}
	case STRING_ARRAY:
		switch vTyped := v.(type) {
		case primitive.A:
			ok = true
			for _, d := range vTyped {
				if err := c.ValidateElement(ctx, d, STRING); err != nil {
					return fmt.Errorf("%s: []string has non string element: %s", c.Name, err)
				}
			}
		}
	case OBJECT:
		switch vTyped := v.(type) {
		case bson.D:
			ok = c.Type == OBJECT
			if ok {
				// On update ($set) if we have an OBJECT type it is replacing the document, so we validate as insert
				if err := Validate(ctx, vTyped, c.SubFields, denyUnknownFields, isUpdate); err != nil {
					return err
				}
			}
		}

	case OBJECT_ARRAY:
		switch vTyped := v.(type) {
		case primitive.A:
			ok = true
			for _, d := range vTyped {
				dType, ok := d.(bson.D)
				if !ok {
					return fmt.Errorf("#{dType}: is not object")
				}
				if err := Validate(ctx, dType, c.SubFields, denyUnknownFields, isUpdate); err != nil {
					return err
				}
			}
		}
	case BIN_DATA:
		switch v.(type) {
		case primitive.Binary:
			ok = true
		}
	case BIN_DATA_ARRAY:
		switch vTyped := v.(type) {
		case primitive.A:
			ok = true
			for _, d := range vTyped {
				if err := c.ValidateElement(ctx, d, BIN_DATA); err != nil {
					return fmt.Errorf("%s: []binData has non binary element: %s", c.Name, err)
				}
			}
		}
	case OBJECT_ID:
		switch v.(type) {
		case primitive.ObjectID:
			ok = true
		}
	case OBJECT_ID_ARRAY:
		switch vTyped := v.(type) {
		case primitive.A:
			ok = true
			for _, d := range vTyped {
				if err := c.ValidateElement(ctx, d, OBJECT_ID); err != nil {
					return fmt.Errorf("%s: []objectID has non objectID element: %s", c.Name, err)
				}
			}
		}
	case BOOL:
		switch v.(type) {
		case bool:
			ok = true
		}
	case BOOL_ARRAY:
		switch vTyped := v.(type) {
		case primitive.A:
			ok = true
			for _, d := range vTyped {
				if err := c.ValidateElement(ctx, d, BOOL); err != nil {
					return fmt.Errorf("%s: []bool has non bool element: %s", c.Name, err)
				}
			}
		}
	case DATE:
		switch v.(type) {
		case int64, primitive.DateTime:
			ok = true
		}
	case DATE_ARRAY:
		switch vTyped := v.(type) {
		case primitive.A:
			ok = true
			for _, d := range vTyped {
				if err := c.ValidateElement(ctx, d, DATE); err != nil {
					return fmt.Errorf("%s: []date has non datetime element: %s", c.Name, err)
				}
			}
		}
	case NULL: // valid for all types except required fields?
	case REGEX:
		switch v.(type) {
		case primitive.Regex:
			ok = true
		}
	case DECIMAL128:
		switch v.(type) {
		case primitive.Decimal128:
			ok = true
		}
	default:
		if c.remoteCollection != nil {
			switch vTyped := v.(type) {
			case bson.D:
				ok = true
				if c.IsArray && !isUpdate {
					return fmt.Errorf("%s: field expects an array but gets a scalar", c.Name)
				}
				// On update ($set) if we have an OBJECT type it is replacing the document, so we validate as insert
				if err := Validate(ctx, vTyped, c.remoteCollection.Fields, denyUnknownFields, isUpdate); err != nil {
					return err
				}
			case bson.A:
				ok = true
				for _, d := range v.(bson.A) {
					doc, k := d.(bson.D)
					if !k {
						return fmt.Errorf("%s: bson.A element is not bson.D", c.Name)
					}
					if err := Validate(ctx, doc, c.remoteCollection.Fields, denyUnknownFields, isUpdate); err != nil {
						return err
					}
				}
			}
		} else {
			return fmt.Errorf("%s: unknown type %v", c.Name, c.Type)
		}
	}

	if !ok {
		return fmt.Errorf("wrong data type: expecting a %v for field %s, but got %T with value %s", c.Type, c.Name, v, v)
	}
	return nil
}
