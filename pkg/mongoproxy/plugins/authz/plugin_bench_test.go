package authz

import (
	"context"
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

func BenchmarkPlugin(b *testing.B) {
	d := &AuthzPlugin{}

	if err := d.Configure(bson.D{
		{"paths", primitive.A{"authzlib/schema/"}},
	}); err != nil {
		b.Fatal(err)
	}

	p := plugins.BuildPipeline([]plugins.Plugin{d}, func(context.Context, *plugins.Request) (bson.D, error) {
		return bson.D{
			{"ok", 1},
		}, nil
	})

	idents := map[string][]plugins.ClientIdentity{
		"role1": {
			&stubClientIdentity{U: "role1", R: []string{"role1"}},
		},
		"glob1": {
			&stubClientIdentity{U: "glob1", R: []string{"glob1"}},
		},
		"createDB": {
			&stubClientIdentity{U: "createDB", R: []string{"createDB"}},
		},
		"global": {
			&stubClientIdentity{U: "global", R: []string{"global"}},
		},
		"deleteDB": {
			&stubClientIdentity{U: "deleteDB", R: []string{"deleteDB"}},
		},
		"dbCollectionAll": {
			&stubClientIdentity{U: "dbCollectionAll", R: []string{"dbCollectionAll"}},
		},
		"authzRole": {
			&stubClientIdentity{U: "authzRole", R: []string{"authzRole"}},
		},
	}

	tests := []struct {
		cmd  bson.D
		good [][]plugins.ClientIdentity
		bad  [][]plugins.ClientIdentity
	}{
		/////////////
		// aggregate tests
		/////////////
		{
			cmd:  bson.D{{"aggregate", 1}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["dbAll"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"aggregate", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["dbCollectionAll"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"aggregate", "authzcol1"}, {"$db", "authzcolcr"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"aggregate", "authzcol1"}, {"$db", "authzcolcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"aggregate", "authzcol1"}, {"$db", "authzdbcr"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"aggregate", "authzcoll"}, {"$db", "authzdb"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"aggregate", "authzcoll"}, {"$db", "authzcolcru"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// collstats tests
		/////////////
		{
			cmd:  bson.D{{"collStats", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["dbCollectionAll"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"collStats", "authzcol1"}, {"$db", "authzcolcr"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"collStats", "authzcol1"}, {"$db", "authzcolcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// count tests
		/////////////
		{
			cmd:  bson.D{{"count", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["dbCollectionAll"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"count", "authzcol1"}, {"$db", "authzcolcr"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"count", "authzcol1"}, {"$db", "authzcolcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// create tests
		/////////////
		{
			cmd:  bson.D{{"create", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["createDB"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd: bson.D{{"create", "authzcol1"}, {"$db", "authzcolcr"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"create", "authzcol1"}, {"$db", "authzcolcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"create", "authzcol1"}, {"$db", "authzdbcu"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"create", "authzcol1"}, {"$db", "authzdbcr"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// createIndexes tests
		/////////////
		{
			cmd:  bson.D{{"createIndexes", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["createDB"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd: bson.D{{"createIndexes", "authzcol1"}, {"$db", "authzcolcr"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"createIndexes", "authzcol1"}, {"$db", "authzcolcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"createIndexes", "authzcol1"}, {"$db", "authzdbcu"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"createIndexes", "authzcol1"}, {"$db", "authzdbcr"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// currentOp tests
		/////////////
		{
			cmd:  bson.D{{"currentOp", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["global"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"currentOp", "authzcol1"}, {"$db", "authzcolcr"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"currentOp", "authzcol1"}, {"$db", "authzcolcu"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"currentOp", "authzcol1"}, {"$db", "authzdbcu"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"currentOp", "authzcol1"}, {"$db", "authzdbcr"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// delete tests
		/////////////
		{
			cmd:  bson.D{{"delete", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["deleteDB"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd: bson.D{{"delete", "authzcol1"}, {"$db", "authzcolcr"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"delete", "authzcol1"}, {"$db", "authzcolcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"delete", "authzcol1"}, {"$db", "authzdbcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"delete", "authzcol1"}, {"$db", "authzdbcr"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"delete", "authzcol1"}, {"$db", "authzcolcd"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"delete", "authzcol1"}, {"$db", "authzdbcd"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// deleteIndexes tests
		/////////////
		{
			cmd:  bson.D{{"deleteIndexes", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["deleteDB"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd: bson.D{{"deleteIndexes", "authzcol1"}, {"$db", "authzcolcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"deleteIndexes", "authzcol1"}, {"$db", "authzdbcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"deleteIndexes", "authzcol1"}, {"$db", "authzdbcr"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"deleteIndexes", "authzcol1"}, {"$db", "authzcolcd"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"deleteIndexes", "authzcol1"}, {"$db", "authzdbcd"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// distinct tests
		/////////////
		{
			cmd:  bson.D{{"distinct", "coll"}, {"key", "col1"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["glob1"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"distinct", "coll"}, {"key", "field"}, {"$db", "authzdb"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"distinct", "coll"}, {"key", "field"}, {"$db", "authzcolcr"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// dropDatabase tests
		/////////////
		{
			cmd:  bson.D{{"dropDatabase", 1}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["global"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"dropDatabase", 1}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// drop tests
		/////////////
		{
			cmd:  bson.D{{"drop", "col"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["deleteDB"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"drop", "coll"}, {"$db", "authzdbcd"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"drop", "coll"}, {"$db", "authzdb"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// dropIndexes tests
		/////////////
		{
			cmd:  bson.D{{"dropIndexes", "col"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["deleteDB"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"dropIndexes", "coll"}, {"$db", "authzcolcd"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"dropIndexes", "coll"}, {"$db", "authzdb"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// endSessions tests
		/////////////
		{
			cmd:  bson.D{{"endSessions", nil}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["global"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"endSessions", nil}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// explain tests
		/////////////
		{
			cmd:  bson.D{{"explain", bson.D{{"find", "coll"}, {"$db", "db"}}}, {"verbosity", "queryPlanner"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["dbCollectionAll"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd: bson.D{{"explain", bson.D{{"find", "authzcol1"}, {"$db", "authzcolcr"}}}, {"$db", "authzcolcr"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"explain", bson.D{{"find", "coll"}, {"$db", "authzdball"}}}, {"verbosity", "queryPlanner"}, {"$db", "authzdball"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"explain", bson.D{{"find", "authzcol1"}, {"$db", "authzcolcu"}}}, {"$db", "authzcolcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// findAndModify tests
		/////////////
		{
			cmd:  bson.D{{"findAndModify", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["dbCollectionAll"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd: bson.D{{"findAndModify", "coll"}, {"$db", "authzcolcru"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"findAndModify", "coll"}, {"$db", "authzcolcd"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"findAndModify", "coll"}, {"$db", "authzdb"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"findAndModify", "coll"}, {"$db", "authzcru"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// Find tests
		/////////////
		{
			cmd:  bson.D{{"find", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["dbCollectionAll"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		// TODO: Right now this requires `*` access; but in reality if there are
		// only the fields that we have access to this is fine; so once schema integration
		// is done we can handle this
		// User that has access to some fields but not all
		{
			cmd: bson.D{{"find", "coll1"}, {"$db", "db1"}},
			bad: [][]plugins.ClientIdentity{idents["role1"]},
		},
		// Read a field we have permissions to
		{
			cmd:  bson.D{{"find", "coll1"}, {"projection", bson.D{{"field1", 1}, {"field2", 1}, {"field3", 1}}}, {"$db", "db1"}},
			good: [][]plugins.ClientIdentity{nil, idents["role1"]},
		},
		{
			cmd:  bson.D{{"find", "coll1"}, {"projection", bson.D{{"field1", 1}, {"nope", 0}}}, {"$db", "db1"}},
			good: [][]plugins.ClientIdentity{nil, idents["role1"]},
		},
		{
			cmd: bson.D{{"find", "coll1"}, {"projection", bson.D{{"field1", 1}, {"nope", 1}}}, {"$db", "db1"}},
			bad: [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd: bson.D{{"find", "coll"}, {"projection", bson.D{{"field1", 1}}}, {"$db", "authzcolcd"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"find", "coll"}, {"projection", bson.D{{"field1", 1}}}, {"$db", "authzcolcru"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"find", "coll"}, {"projection", bson.D{{"field", 1}}}, {"$db", "authzdb"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// insert tests
		/////////////
		{
			cmd:  bson.D{{"insert", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["dbCollectionAll"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"insert", "coll"}, {"$db", "authzcolcru"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"find", "coll"}, {"$db", "authzdb"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// killAllSessions tests
		/////////////
		{
			cmd:  bson.D{{"killAllSessions", nil}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["global"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"killAllSessions", nil}, {"$db", "authzcolcru"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// killCursors tests
		/////////////
		{
			cmd:  bson.D{{"killCursors", ""}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["global"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"killCursors", ""}, {"$db", "authzcolcru"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// killOp tests
		/////////////
		{
			cmd:  bson.D{{"killOp", 1}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["global"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"killOp", 1}, {"$db", "authzcolcru"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// listCollections tests
		/////////////
		{
			cmd:  bson.D{{"listCollections", 1}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["readDB"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd: bson.D{{"listCollections", 1}, {"$db", "authzcolcru"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"listCollections", 1}, {"$db", "authzdbcd"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd:  bson.D{{"listCollections", 1}, {"$db", "authzdbcr"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// listDatabases tests
		/////////////
		{
			cmd:  bson.D{{"listDatabases", 1}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["global"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"listDatabases", 1}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// listIndexes tests
		/////////////
		{
			cmd:  bson.D{{"listIndexes", "coll"}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["readDB"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"listIndexes", "authzcol1"}, {"$db", "authzcolcr"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},
		{
			cmd: bson.D{{"listIndexes", "authzcol1"}, {"$db", "authzcolcu"}},
			bad: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// serverStatus tests
		/////////////
		{
			cmd:  bson.D{{"serverStatus", 1}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["global"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
		{
			cmd:  bson.D{{"serverStatus", 1}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{idents["authzRole"]},
		},

		/////////////
		// serverStatus tests
		/////////////
		{
			cmd:  bson.D{{"update", "coll"}, {"updates", []bson.D{{{"u", bson.D{{"$set", bson.D{{"a", "test"}}}}}}}}, {"$db", "db"}},
			good: [][]plugins.ClientIdentity{nil, idents["dbCollectionAll"]},
			bad:  [][]plugins.ClientIdentity{idents["role1"]},
		},
	}

	for i, test := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			cmd, ok := command.GetCommand(test.cmd[0].Key)
			if !ok {
				b.Fatalf("no such command: '" + test.cmd[0].Key + "'")
			}

			if err := cmd.FromBSOND(test.cmd); err != nil {
				b.Fatal(err)
			}

			r := &plugins.Request{
				CC:          plugins.NewClientConnection(),
				CommandName: test.cmd[0].Key,
				Command:     cmd,
			}

			for x, g := range test.good {
				r.CC.Identities = g
				b.Run("good:"+strconv.Itoa(x), func(b *testing.B) {
					for xx := 0; xx < b.N; xx++ {
						p(context.TODO(), r)
					}
				})
			}
			for x, bad := range test.bad {
				r.CC.Identities = bad
				b.Run("bad:"+strconv.Itoa(x), func(b *testing.B) {
					for xx := 0; xx < b.N; xx++ {
						p(context.TODO(), r)
					}
				})
			}
		})
	}
}

func BenchmarkPluginDefaultPolicies(b *testing.B) {
	d := &AuthzPlugin{}

	if err := d.Configure(bson.D{
		{"paths", primitive.A{"authzlib/schema/"}},
		{"denyByDefault", true},
		{"denyByDefaultNamespaces", bson.D{
			{"db.deny", true},
			{"db.allow", false},
		}},
	}); err != nil {
		b.Fatal(err)
	}

	p := plugins.BuildPipeline([]plugins.Plugin{d}, func(_ context.Context, r *plugins.Request) (bson.D, error) {
		return bson.D{
			{"ok", 1},
		}, nil
	})

	tests := []struct {
		cmd bson.D
		ok  bool
	}{
		{
			cmd: bson.D{{"find", "unknown"}, {"$db", "db"}},
			ok:  false,
		},
		{
			cmd: bson.D{{"find", "deny"}, {"$db", "db"}},
			ok:  false,
		},
		{
			cmd: bson.D{{"find", "allow"}, {"$db", "db"}},
			ok:  true,
		},
	}

	for i, test := range tests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			cmd, ok := command.GetCommand(test.cmd[0].Key)
			if !ok {
				b.Fatalf("no such command: '" + test.cmd[0].Key + "'")
			}

			if err := cmd.FromBSOND(test.cmd); err != nil {
				b.Fatal(err)
			}

			r := &plugins.Request{
				CC:          plugins.NewClientConnection(),
				CommandName: test.cmd[0].Key,
				Command:     cmd,
			}
			r.CC.Identities = []plugins.ClientIdentity{&stubClientIdentity{U: "unknown", R: []string{"unknown"}}}

			b.ResetTimer()
			for x := 0; x < b.N; x++ {
				p(context.TODO(), r)
			}
		})
	}
}
