package schema

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

func TestSchema(t *testing.T) {
	d := &SchemaPlugin{}

	if err := d.Configure(bson.D{
		{"schemaPath", "example.json"},
	}); err != nil {
		t.Fatal(err)
	}

	p := plugins.BuildPipeline([]plugins.Plugin{d}, func(context.Context, *plugins.Request) (bson.D, error) {
		return bson.D{
			{"ok", 1},
		}, nil
	})

	tests := []struct {
		cmd bson.D
		ok  int
	}{
		///////////////
		// Insert tests
		///////////////
		// do a valid insert
		{
			cmd: bson.D{{"insert", "requirea"}, {"documents", []bson.D{{{"a", "valid"}}}}, {"$db", "testdb"}},
			ok:  1,
		},
		// Do an invalid insert
		{
			cmd: bson.D{{"insert", "requirea"}, {"documents", []bson.D{{{"a", 2}}}}, {"$db", "testdb"}},
			ok:  0,
		},
		{
			cmd: bson.D{{"insert", "requirea"}, {"documents", []bson.D{{}}}, {"$db", "testdb"}},
			ok:  0,
		},
		// do a valid AND invalid insert
		{
			cmd: bson.D{{"insert", "requirea"}, {"documents", []bson.D{{{"a", "valid"}}, {}}}, {"$db", "testdb"}},
			ok:  0,
		},

		//////////////////////
		// FindAndModify Tests
		//////////////////////
		{
			cmd: bson.D{{"findAndModify", "requirea"}, {"update", bson.D{{"$set", bson.D{{"a", "test"}}}}}, {"$db", "testdb"}},
			ok:  1,
		},
		// Do an invalid FindAndModify
		{
			cmd: bson.D{{"findAndModify", "requirea"}, {"update", bson.D{{"$set", bson.D{{"a", 1}}}}}, {"$db", "testdb"}},
			ok:  0,
		},

		///////////////
		// Update Tests
		///////////////
		{
			cmd: bson.D{{"update", "requirea"}, {"updates", []bson.D{{{"u", bson.D{{"$set", bson.D{{"a", "test"}}}}}}}}, {"$db", "testdb"}},
			ok:  1,
		},
		// Do an invalid Update
		{
			cmd: bson.D{{"update", "requirea"}, {"updates", []bson.D{{{"u", bson.D{{"$set", bson.D{{"a", 1}}}}}}}}, {"$db", "testdb"}},
			ok:  0,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			cmd, ok := command.GetCommand(test.cmd[0].Key)
			if !ok {
				t.Fatalf("no such command: '" + test.cmd[0].Key + "'")
			}

			if err := cmd.FromBSOND(test.cmd); err != nil {
				t.Fatal(err)
			}

			r := &plugins.Request{
				CC:          plugins.NewClientConnection(),
				CommandName: test.cmd[0].Key,
				Command:     cmd,
			}
			result, err := p(context.TODO(), r)

			if err != nil {
				t.Fatal(err)
			}
			okRaw, ok := bsonutil.Lookup(result, "ok")
			if !ok {
				t.Fatalf("result missing `ok`: %v", result)
			}
			if !reflect.DeepEqual(okRaw, test.ok) {
				t.Fatalf("Mismatch in ok expected=%v actual=%v", test.ok, okRaw)
			}
		})
	}
}
