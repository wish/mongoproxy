package schema

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	insertTests = []struct {
		DB, Collection string
		In             bson.D
		Err            bool
	}{
		// Non-existent
		{DB: "testdb", Collection: "hidden", In: bson.D{}, Err: true},

		// Check empty
		{DB: "testdb", Collection: "nonrequire", In: bson.D{}, Err: false},
		{DB: "testdb", Collection: "testcollection", In: bson.D{}, Err: true},
		// Check string array
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"friends", bson.A{"bob", "alice"}}, {"name", "tom"}}, Err: false},
		// Check string array with wrong type
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"friends", bson.A{"bob", 1}}}, Err: true},
		// Check int array
		{DB: "testdb", Collection: "nonrequire", In: bson.D{{"luckynumbers", bson.A{666, 888}}}, Err: false},
		// Check int array with wrong type
		{DB: "testdb", Collection: "nonrequire", In: bson.D{{"luckynumbers", bson.A{666, "888"}}}, Err: true},

		// Check that required checks empty
		{DB: "testdb", Collection: "requirea", In: bson.D{}, Err: true},
		// Check that required checks wrong type
		{DB: "testdb", Collection: "requirea", In: bson.D{{"a", 1}}, Err: true},
		// Check that required works in expected case
		{DB: "testdb", Collection: "requirea", In: bson.D{{"a", "test"}}, Err: false},
		// Check that required works in expected case with extra field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"a", "test"}, {"b", 1}}, Err: false},

		{DB: "testdb", Collection: "includerequirea", In: bson.D{}, Err: true},
		// Check that required checks wrong type
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"included", 1}}, Err: true},
		// Check that not all required  in expected case
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"included", bson.D{{"a", "test"}}}}, Err: false},
		// Check that required works as expected case with extra field
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"included", bson.D{{"a", "test"}, {"b", 1}}}}, Err: false},

		// Check that missing required works in expected case
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"includedarr", bson.A{bson.D{{"a", "test0"}}, bson.D{{"a", "test1"}}}}}, Err: true},
		//  Check that required works in expected case with array wrong type
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"includedarr", bson.A{bson.D{{"a", "test0"}}, bson.D{{"a", 1}}}}}, Err: true},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"includedarr", bson.D{{"a", "test0"}}}}, Err: true},

		// Check that required checks empty
		{DB: "testdb", Collection: "requireonlya", In: bson.D{}, Err: true},
		// Check that required checks wrong type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"a", 1}}, Err: true},
		// Check that required works in expected case
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"a", "test"}}, Err: false},
		// Check that required works in expected case with extra field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"a", "test"}, {"b", 1}}, Err: true},

		// Check that required checks empty
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{}, Err: true},
		// Check that required checks wrong type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"doc", "test"}}, Err: true},
		// Check that required works in expected case
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"doc", bson.D{{"a", "a"}}}}, Err: false},
		// Check that required works in expected case with extra field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"doc", bson.D{{"a", "a"}, {"b", "b"}}}}, Err: true},
		// Check that non required works in expected case
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"doc", bson.D{{"a", "a"}, {"notrequired", nil}}}}, Err: false},
		// Check that non required works in expected case
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"doc", bson.D{{"a", "a"}, {"notrequired", "1"}}}}, Err: false},
		// Check that non required with invalid type works in expected case
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"doc", bson.D{{"a", "a"}, {"notrequired", 1}}}}, Err: true},
		// Check that required with empty works in expected case
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"doc", bson.D{{"a", ""}, {"notrequired", 1}}}}, Err: true},
		// Check that required with nil works in expected case
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"doc", bson.D{{"a", nil}, {"notrequired", 1}}}}, Err: true},
		// Check that fails if missing subfield
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"doc", bson.D{}}}, Err: true},

		// Check that required checks empty
		{DB: "testdb", Collection: "requireonlysub", In: bson.D{}, Err: true},
		// Check that required checks wrong type
		{DB: "testdb", Collection: "requireonlysub", In: bson.D{{"doc", "test"}}, Err: true},
		// Check that required works in expected case
		{DB: "testdb", Collection: "requireonlysub", In: bson.D{{"doc", bson.D{{"a", "a"}}}}, Err: false},
		// Check that required works in expected case with extra field
		{DB: "testdb", Collection: "requireonlysub", In: bson.D{{"doc", bson.D{{"a", "a"}, {"b", "b"}}}}, Err: true},
		// Check that fails if missing subfield
		{DB: "testdb", Collection: "requireonlysub", In: bson.D{{"doc", bson.D{}}}, Err: true},
	}

	updateTests = []struct {
		DB, Collection string
		In             bson.D
		Upsert         bool
		Err            bool
	}{
		//
		// Set tests
		//
		// set wrong type
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"name", 1}}}}, Err: true},
		// set unknown field
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"unknown", 1}}}}},
		// set correct type
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"name", "name"}}}}},

		// set wrong type
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$set", bson.D{{"a", 1}}}}, Err: true},
		// set unknown field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$set", bson.D{{"unknown", 1}}}}},
		// set correct type
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$set", bson.D{{"a", "name"}}}}},
		// set extra field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$set", bson.D{{"a", "name"}, {"b", 1}}}}},
		// set extra field, don't touch main
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$set", bson.D{{"b", 1}}}}},

		// set wrong type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$set", bson.D{{"a", 1}}}}, Err: true},
		// set unknown field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$set", bson.D{{"unknown", 1}}}}, Err: true},
		// set correct type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$set", bson.D{{"a", "name"}}}}},
		// set extra field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$set", bson.D{{"a", "name"}, {"b", 1}}}}, Err: true},
		// set extra field, don't touch main
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$set", bson.D{{"b", 1}}}}, Err: true},

		// set wrong type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$set", bson.D{{"doc.a", 1}}}}, Err: true},
		// Miss required "a" subfield (since this is an update; it is assumed already set)
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$set", bson.D{{"doc", bson.D{{"notrequired", 1}}}}}}, Err: true},
		// Miss required "a" subfield (since this is an update; it is assumed already set)
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$set", bson.D{{"doc", bson.D{{"a", "test"}}}}}}, Err: false},
		// Miss required "a" subfield (since this is an update; it is assumed already set)
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$set", bson.D{{"doc", bson.D{{"a", 1}}}}}}, Err: true},
		// set unknown field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$set", bson.D{{"doc.unknown", 1}}}}, Err: true},
		// set correct type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$set", bson.D{{"doc.a", "name"}}}}},
		// set correct type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$set", bson.D{{"doc", bson.D{{"a", "name"}}}}}}},
		// set extra field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$set", bson.D{{"doc.a", "name"}, {"doc.b", 1}}}}, Err: true},
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$set", bson.D{{"a", "name"}, {"doc.b", 1}}}}, Err: true},

		//
		// push tests
		//
		// push wrong type
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$push", bson.D{{"name", 1}}}}, Err: true},
		// push unknown field
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$push", bson.D{{"unknown", 1}}}}},
		// push correct type
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$push", bson.D{{"name", "name"}}}}},

		// push wrong type
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$push", bson.D{{"a", 1}}}}, Err: true},
		// push unknown field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$push", bson.D{{"unknown", 1}}}}},
		// push correct type
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$push", bson.D{{"a", "name"}}}}},
		// push extra field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$push", bson.D{{"a", "name"}, {"b", 1}}}}},
		// push extra field, don't touch main
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$push", bson.D{{"b", 1}}}}},

		// push wrong type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$push", bson.D{{"a", 1}}}}, Err: true},
		// push unknown field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$push", bson.D{{"unknown", 1}}}}, Err: true},
		// push correct type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$push", bson.D{{"a", "name"}}}}},
		// push extra field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$push", bson.D{{"a", "name"}, {"b", 1}}}}, Err: true},
		// push extra field, don't touch main
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$push", bson.D{{"b", 1}}}}, Err: true},

		// push wrong type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$push", bson.D{{"doc.a", 1}}}}, Err: true},
		// Miss required "a" subfield (since this is an update; it is assumed already pushed)
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$push", bson.D{{"doc", bson.D{{"notrequired", "name"}}}}}}, Err: false},
		// push unknown field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$push", bson.D{{"doc.unknown", 1}}}}, Err: true},
		// push correct type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$push", bson.D{{"doc.a", "name"}}}}},
		// push correct type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$push", bson.D{{"doc", bson.D{{"a", "name"}}}}}}},
		// push extra field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$push", bson.D{{"doc.a", "name"}, {"doc.b", 1}}}}, Err: true},
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$push", bson.D{{"a", "name"}, {"doc.b", 1}}}}, Err: true},

		//
		// pull tests
		//
		// pull wrong type
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$pull", bson.D{{"name", 1}}}}, Err: true},
		// pull unknown field
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$pull", bson.D{{"unknown", 1}}}}},
		// pull correct type
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$pull", bson.D{{"name", "name"}}}}},

		// pull wrong type
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$pull", bson.D{{"a", 1}}}}, Err: true},
		// pull unknown field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$pull", bson.D{{"unknown", 1}}}}},
		// pull correct type
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$pull", bson.D{{"a", "name"}}}}},
		// pull extra field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$pull", bson.D{{"a", "name"}, {"b", 1}}}}},
		// pull extra field, don't touch main
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$pull", bson.D{{"b", 1}}}}},

		// pull wrong type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$pull", bson.D{{"a", 1}}}}, Err: true},
		// pull unknown field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$pull", bson.D{{"unknown", 1}}}}, Err: true},
		// pull correct type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$pull", bson.D{{"a", "name"}}}}},
		// pull extra field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$pull", bson.D{{"a", "name"}, {"b", 1}}}}, Err: true},
		// pull extra field, don't touch main
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$pull", bson.D{{"b", 1}}}}, Err: true},

		// pull wrong type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$pull", bson.D{{"doc.a", 1}}}}, Err: true},
		// Miss required "a" subfield (since this is an update; it is assumed already pulled)
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$pull", bson.D{{"doc", bson.D{{"notrequired", "name"}}}}}}, Err: false},
		// pull unknown field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$pull", bson.D{{"doc.unknown", 1}}}}, Err: true},
		// pull correct type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$pull", bson.D{{"doc.a", "name"}}}}},
		// pull correct type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$pull", bson.D{{"doc", bson.D{{"a", "name"}}}}}}},
		// pull extra field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$pull", bson.D{{"doc.a", "name"}, {"doc.b", 1}}}}, Err: true},
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$pull", bson.D{{"a", "name"}, {"doc.b", 1}}}}, Err: true},

		//
		// addToSet tests
		//
		// addToSet wrong type
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$addToSet", bson.D{{"name", 1}}}}, Err: true},
		// addToSet unknown field
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$addToSet", bson.D{{"unknown", 1}}}}},
		// addToSet correct type
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$addToSet", bson.D{{"name", "name"}}}}},

		// addToSet wrong type
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$addToSet", bson.D{{"a", 1}}}}, Err: true},
		// addToSet unknown field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$addToSet", bson.D{{"unknown", 1}}}}},
		// addToSet correct type
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$addToSet", bson.D{{"a", "name"}}}}},
		// addToSet extra field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$addToSet", bson.D{{"a", "name"}, {"b", 1}}}}},
		// addToSet extra field, don't touch main
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$addToSet", bson.D{{"b", 1}}}}},

		// addToSet wrong type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$addToSet", bson.D{{"a", 1}}}}, Err: true},
		// addToSet unknown field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$addToSet", bson.D{{"unknown", 1}}}}, Err: true},
		// addToSet correct type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$addToSet", bson.D{{"a", "name"}}}}},
		// addToSet extra field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$addToSet", bson.D{{"a", "name"}, {"b", 1}}}}, Err: true},
		// addToSet extra field, don't touch main
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$addToSet", bson.D{{"b", 1}}}}, Err: true},

		// addToSet wrong type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$addToSet", bson.D{{"doc.a", 1}}}}, Err: true},
		// Miss required "a" subfield
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$addToSet", bson.D{{"doc", bson.D{{"notrequired", "name"}}}}}}, Err: false},
		// addToSet unknown field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$addToSet", bson.D{{"doc.unknown", 1}}}}, Err: true},
		// addToSet correct type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$addToSet", bson.D{{"doc.a", "name"}}}}},
		// addToSet correct type
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$addToSet", bson.D{{"doc", bson.D{{"a", "name"}}}}}}},
		// addToSet extra field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$addToSet", bson.D{{"doc.a", "name"}, {"doc.b", 1}}}}, Err: true},
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$addToSet", bson.D{{"a", "name"}, {"doc.b", 1}}}}, Err: true},

		//
		// rename tests
		//
		// rename; invalid
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$rename", bson.D{{"name", "namenew"}}}}, Err: true},
		// invalid rename (a required)
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$rename", bson.D{{"a", "b"}}}}, Err: true},
		// valid rename: extra fields are allowed
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$rename", bson.D{{"b", "c"}}}}},
		// invalid rename (a required)
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$rename", bson.D{{"a", "b"}}}}, Err: true},
		// invalid rename: extra fields are not allowed
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$rename", bson.D{{"b", "c"}}}}, Err: true},
		// invalid rename (a required)
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$rename", bson.D{{"a", "b"}}}}, Err: true},
		// invalid rename: extra fields are not allowed
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$rename", bson.D{{"doc.a", "c"}}}}, Err: true},

		//
		// unset tests
		//
		// unset required field
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$unset", bson.D{{"name", 1}}}}, Err: true},
		// unset known field
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$unset", bson.D{{"unknown", 1}}}}},
		// unset unknown field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$unset", bson.D{{"a", 1}}}}, Err: true},
		// unset known field
		{DB: "testdb", Collection: "requirea", In: bson.D{{"$unset", bson.D{{"unknown", 1}}}}},
		// unset unknown field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$unset", bson.D{{"a", 1}}}}, Err: true},
		// unset known field
		{DB: "testdb", Collection: "requireonlya", In: bson.D{{"$unset", bson.D{{"unknown", 1}}}}, Err: true},
		// unset unknown field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$unset", bson.D{{"a", 1}}}}, Err: true},
		// unset known field
		{DB: "testdb", Collection: "requireonlysuba", In: bson.D{{"$unset", bson.D{{"doc.a", 1}}}}, Err: true},

		//
		// SetOnInsert tests
		//
		// set wrong type
		{DB: "testdb", Collection: "testcollection", In: bson.D{
			{"$set", bson.D{{"age", 1}, {"friends.0", "linda"}}},
			{"$setOnInsert", bson.D{{"name", 1}, {"friends.0", "alice"}}},
		}, Upsert: true, Err: true},
		// set correct type
		{DB: "testdb", Collection: "testcollection", In: bson.D{
			{"$set", bson.D{{"age", 1}, {"friends.0", "linda"}}},
			{"$setOnInsert", bson.D{{"name", "a"}, {"friends.0", "linda"}}},
		}, Upsert: true},

		// miss required field
		{DB: "testdb", Collection: "requirea", In: bson.D{
			{"$set", bson.D{{"b", 1}}},
			{"$setOnInsert", bson.D{{"c", 1}}},
		}, Upsert: true, Err: false},
		// set correct type
		{DB: "testdb", Collection: "requirea", In: bson.D{
			{"$set", bson.D{{"b", 1}}},
			{"$setOnInsert", bson.D{{"a", "a"}}},
		}, Upsert: true},

		// incorrect type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{
			{"$set", bson.D{{"a", 1}}},
			{"$setOnInsert", bson.D{{"a", 1}}},
		}, Upsert: true, Err: true},
		// set incorrect type
		{DB: "testdb", Collection: "requireonlya", In: bson.D{
			{"$set", bson.D{{"a", 1}}},
			{"$setOnInsert", bson.D{{"a", "a"}}},
		}, Upsert: true, Err: true},
		// set correct type
		{DB: "testdb", Collection: "nonrequire", In: bson.D{
			{"$set", bson.D{{"a", "a"}}},
			{"$setOnInsert", bson.D{{"a", nil}}},
		}, Upsert: true},
		// update without update operator with invalid value
		{DB: "testdb", Collection: "nonrequire", In: bson.D{
			{"age", "9"},
		}, Upsert: true, Err: true},
		// update without update operator with valid value
		{DB: "testdb", Collection: "nonrequire", In: bson.D{
			{"age", 1},
		}, Upsert: true},
		// update without "$" with invalid value
		{DB: "testdb", Collection: "requireadollarsign", In: bson.D{
			{"$id", 1},
		}, Upsert: true, Err: true},
		// update without "$" with valid value
		{DB: "testdb", Collection: "requireadollarsign", In: bson.D{
			{"$id", "a"},
		}, Upsert: true},

		// included schema
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$set", bson.D{{"included.a", "name"}}}}},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$set", bson.D{{"included.a", 1}}}}, Err: true}, // This one
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$rename", bson.D{{"included.a", "namenew"}}}}, Err: true},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$rename", bson.D{{"included.b", "namenew"}}}}},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$unset", bson.D{{"included.a", 1}}}}, Err: true},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$unset", bson.D{{"included.b", 1}}}}},

		//// set wrong type
		//{DB: "testdb", Collection: "includerequirea", In: bson.D{
		//	{"$set", bson.D{{"included.a", 1}}},
		//	{"$setOnInsert", bson.D{{"included.a", 1}}},
		//}, Upsert: true, Err: true},
		//// set correct type
		//{DB: "testdb", Collection: "includerequirea", In: bson.D{
		//	{"$set", bson.D{{"included.a", "1"}}},
		//	{"$setOnInsert", bson.D{{"included.a", "a"}}},
		//}, Upsert: true},

		//
		// array tests
		//
		// https://docs.mongodb.com/manual/reference/operator/update-array/
		// https://docs.mongodb.com/manual/reference/operator/update/positional/
		// https://docs.mongodb.com/manual/reference/operator/update/set/
		// support cases:
		// 1. $
		//        a. { $set: { "grades.$" : 6 } }
		//        b. { $set: { "grades.$.std" : 6 } }
		//        c. { $set: { "grades.$[]" : -2 } }
		//        d. { $set: { "grades.$[].std" : -2 } },
		//        e. { $set: { "myArray.$[element]": 2 } },
		//        f. { $set: { "grades.$[elem].mean" : 100 } },
		// 2. \d
		//        a.{$set: {"tags.1": "rain gear"}},
		//        b.{$set: {"ratings.0.rating": 2}},

		//  set correct type for array
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"friends.$", "linda"}}}}},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"friends.0", "linda"}}}}},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"friends", bson.A{"linda", "test"}}}}}},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"friends.$[]", "linda"}}}}},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"friends.$[foo]", "linda"}}}}},

		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$set", bson.D{{"includedarr.$.a", "linda"}}}}},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$set", bson.D{{"includedarr.0.a", "linda"}}}}},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$set", bson.D{{"includedarr.$[].a", "linda"}}}}},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$set", bson.D{{"includedarr.$[foo].a", "linda"}}}}},

		// set wrong type for array
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"friends.$", 1}}}}, Err: true},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"friends.0", 2}}}}, Err: true},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"friends.$[]", 3}}}}, Err: true},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$set", bson.D{{"friends.$[foo]", 4}}}}, Err: true},

		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$set", bson.D{{"includedarr.$.a", 5}}}}, Err: true},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$set", bson.D{{"includedarr.0.a", 6}}}}, Err: true},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$set", bson.D{{"includedarr.$[].a", 7}}}}, Err: true},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$set", bson.D{{"includedarr.$[foo].a", 8}}}}, Err: true},

		// rename within array is not supported in mongodb

		// unset is follow dot-notion and $ projection as well just like set
		// https://docs.mongodb.com/manual/reference/operator/update/unset/
		//  unset required fields
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$unset", bson.D{{"friends.$", "linda"}}}}, Err: true},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$unset", bson.D{{"friends.0", "linda"}}}}, Err: true},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$unset", bson.D{{"friends.$[]", "linda"}}}}, Err: true},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$unset", bson.D{{"friends.$[foo]", "linda"}}}}, Err: true},

		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$unset", bson.D{{"includedarr.$.a", "linda"}}}}, Err: true},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$unset", bson.D{{"includedarr.0.a", "linda"}}}}, Err: true},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$unset", bson.D{{"includedarr.$[].a", "linda"}}}}, Err: true},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$unset", bson.D{{"includedarr.$[foo].a", "linda"}}}}, Err: true},

		// unset non require fields
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$unset", bson.D{{"luckynumbers.$", 1}}}}},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$unset", bson.D{{"luckynumbers.0", 2}}}}},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$unset", bson.D{{"luckynumbers.$[]", 3}}}}},
		{DB: "testdb", Collection: "testcollection", In: bson.D{{"$unset", bson.D{{"luckynumbers.$[foo]", 4}}}}},

		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$unset", bson.D{{"includedarr.$.b", 5}}}}},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$unset", bson.D{{"includedarr.0.b", 6}}}}},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$unset", bson.D{{"includedarr.$[].b", 7}}}}},
		{DB: "testdb", Collection: "includerequirea", In: bson.D{{"$unset", bson.D{{"includedarr.$[foo].b", 8}}}}},

		// setOninsert supports dot-notion, but no $ projection
		// https://docs.mongodb.com/manual/reference/operator/update/setOnInsert/
		//{DB: "testdb", Collection: "testcollection", In: bson.D{
		//	{"$set", bson.D{{"friends.0", "linda"}}},
		//	{"$setOnInsert", bson.D{{"friends.0", 1}}},
		//}, Upsert: true, Err: true},
		//{DB: "testdb", Collection: "testcollection", In: bson.D{
		//	{"$set", bson.D{{"friends.0", "linda"}}},
		//	{"$setOnInsert", bson.D{{"friends.0", "alice"}}},
		//}, Upsert: true, Err: false},
		//
		//{DB: "testdb", Collection: "includerequirea", In: bson.D{
		//	{"$set", bson.D{{"includedarr.0.a", "linda"}}},
		//	{"$setOnInsert", bson.D{{"includedarr.0.a", 1}}},
		//}, Upsert: true, Err: true},
		//{DB: "testdb", Collection: "includerequirea", In: bson.D{
		//	{"$set", bson.D{{"includedarr.0.a", "linda"}}},
		//	{"$setOnInsert", bson.D{{"includedarr.0.a", "alice"}}},
		//}, Upsert: true, Err: false},
	}
)

type TypeTest struct {
	fieldType string
	valid     []interface{}
	invalid   []interface{}
}

func Test_SchemaInsert(t *testing.T) {
	var schema ClusterSchema

	b, err := ioutil.ReadFile("example.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, &schema); err != nil {
		panic(err)
	}

	for i, test := range insertTests {
		b, _ := json.Marshal(test)
		t.Run(strconv.Itoa(i)+"_"+string(b), func(t *testing.T) {
			err := schema.ValidateInsert(context.TODO(), test.DB, test.Collection, test.In)
			if (err != nil) != test.Err {
				if err == nil {
					t.Errorf("Missing expected err")
				} else {
					t.Errorf("Unexpected Err: %v", err)
				}
			}
		})
	}
}

func Test_SchemaUpdate(t *testing.T) {
	var schema ClusterSchema

	b, err := ioutil.ReadFile("example.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, &schema); err != nil {
		panic(err)
	}

	for i, test := range updateTests {
		b, _ := json.Marshal(test)
		t.Run(strconv.Itoa(i)+"_"+string(b), func(t *testing.T) {
			err := schema.ValidateUpdate(context.TODO(), test.DB, test.Collection, test.In, test.Upsert)
			if (err != nil) != test.Err {
				if err == nil {
					t.Errorf("Missing expected err")
				} else {
					t.Errorf("Unexpected Err: %v", err)
				}
			}
		})
	}
}

func Test_SchemaTypes(t *testing.T) {
	var schema ClusterSchema

	b, err := ioutil.ReadFile("example.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(b, &schema); err != nil {
		panic(err)
	}

	intTypeTests := TypeTest{
		fieldType: "int",
		valid:     []interface{}{1, int32(1), int64(1)},
		invalid:   []interface{}{"1", nil},
	}

	longTypeTests := TypeTest{
		fieldType: "long",
		valid:     []interface{}{1, int32(1), int64(1)},
		invalid:   []interface{}{"1", nil},
	}

	doubleTypeTests := TypeTest{
		fieldType: "double",
		valid:     []interface{}{1, int32(1), int64(1), float32(1.1), float64(1.1)},
		invalid:   []interface{}{"1", nil},
	}

	stringTypeTests := TypeTest{
		fieldType: "string",
		valid:     []interface{}{"1"},
		invalid:   []interface{}{1, nil},
	}

	objectTypeTests := TypeTest{
		fieldType: "object",
		valid:     []interface{}{bson.D{{"string", "b"}, {"int", 1}}},
		invalid:   []interface{}{bson.D{{"string", 1}}, bson.D{{"int", "1"}}},
	}

	bindataTypeTests := TypeTest{
		fieldType: "bindata",
		valid:     []interface{}{primitive.Binary{Subtype: 'a', Data: []byte("foo")}},
		invalid:   []interface{}{1, nil},
	}

	objectidTypeTests := TypeTest{
		fieldType: "objectid",
		valid:     []interface{}{primitive.ObjectID{}},
		invalid:   []interface{}{1, nil, "1"},
	}

	boolTypeTests := TypeTest{
		fieldType: "bool",
		valid:     []interface{}{false},
		invalid:   []interface{}{1, nil, "1"},
	}

	dateTypeTests := TypeTest{
		fieldType: "date",
		valid:     []interface{}{time.Now().Unix()},
		invalid:   []interface{}{nil, "1"},
	}

	regexTypeTests := TypeTest{
		fieldType: "regex",
		valid:     []interface{}{primitive.Regex{Pattern: ".*"}},
		invalid:   []interface{}{nil, "1", 1},
	}

	decimalTypeTests := TypeTest{
		fieldType: "decimal",
		valid:     []interface{}{primitive.NewDecimal128(1, 2)},
		invalid:   []interface{}{nil, "1", 1},
	}

	stringArrayTypeTests := TypeTest{
		fieldType: "[]string",
		valid:     []interface{}{bson.A{"1", "2"}},
		invalid:   []interface{}{bson.A{"1", 2}},
	}

	intArrayTypeTests := TypeTest{
		fieldType: "[]int",
		valid:     []interface{}{bson.A{1, 2}},
		invalid:   []interface{}{bson.A{"1", 2}},
	}

	longArrayTypeTests := TypeTest{
		fieldType: "[]long",
		valid:     []interface{}{bson.A{int32(1), int64(101)}},
		invalid:   []interface{}{bson.A{float32(1.01), nil}},
	}

	doubleArrayTypeTests := TypeTest{
		fieldType: "[]double",
		valid:     []interface{}{bson.A{float32(3.1415), 2}},
		invalid:   []interface{}{bson.A{"1", nil}},
	}

	boolArrayTypeTests := TypeTest{
		fieldType: "[]bool",
		valid:     []interface{}{bson.A{true, false}},
		invalid:   []interface{}{bson.A{1, nil}},
	}

	objectIdArrayTypeTests := TypeTest{
		fieldType: "[]objectID",
		valid:     []interface{}{bson.A{primitive.ObjectID{}, primitive.ObjectID{}}},
		invalid:   []interface{}{bson.A{"1", nil}},
	}

	objectArrayTypeTests := TypeTest{
		fieldType: "[]object",
		valid:     []interface{}{bson.A{bson.D{{"string", "b"}}}},
		invalid:   []interface{}{bson.A{"bar", nil, 1}},
	}

	SchemaTypesTestHelper(t, intTypeTests, "bsonint", schema)
	SchemaTypesTestHelper(t, longTypeTests, "bsonlong", schema)
	SchemaTypesTestHelper(t, doubleTypeTests, "bsondouble", schema)
	SchemaTypesTestHelper(t, stringTypeTests, "bsonstring", schema)
	SchemaTypesTestHelper(t, objectTypeTests, "bsonobject", schema)
	SchemaTypesTestHelper(t, bindataTypeTests, "bsonbindata", schema)
	SchemaTypesTestHelper(t, objectidTypeTests, "bsonobjectid", schema)
	SchemaTypesTestHelper(t, boolTypeTests, "bsonbool", schema)
	SchemaTypesTestHelper(t, dateTypeTests, "bsondate", schema)
	SchemaTypesTestHelper(t, regexTypeTests, "bsonregex", schema)
	SchemaTypesTestHelper(t, decimalTypeTests, "bsondecimal", schema)
	SchemaTypesTestHelper(t, stringArrayTypeTests, "bsonstringarr", schema)
	SchemaTypesTestHelper(t, intArrayTypeTests, "bsonintarr", schema)
	SchemaTypesTestHelper(t, longArrayTypeTests, "bsonlongarr", schema)
	SchemaTypesTestHelper(t, doubleArrayTypeTests, "bsondoublearr", schema)
	SchemaTypesTestHelper(t, boolArrayTypeTests, "bsonboolarr", schema)
	SchemaTypesTestHelper(t, objectIdArrayTypeTests, "bsonobjectidarr", schema)
	SchemaTypesTestHelper(t, objectArrayTypeTests, "bsonobjectarr", schema)

}

func SchemaTypesTestHelper(t *testing.T, typeTest TypeTest, collection string, schema ClusterSchema) {
	t.Run(strconv.Itoa(0)+"_"+typeTest.fieldType, func(t *testing.T) {
		for i, e := range typeTest.valid {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				err := schema.ValidateInsert(context.TODO(), "testdb", collection, bson.D{{typeTest.fieldType, e}})
				if err != nil {
					t.Fatalf("Unexpected err: %v", err)
				}
			})
		}
		for i, e := range typeTest.invalid {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				err := schema.ValidateInsert(context.TODO(), "testdb", collection, bson.D{{typeTest.fieldType, e}})
				if err == nil {
					t.Fatalf("Missing expected error")
				}
			})
		}
	})
}
