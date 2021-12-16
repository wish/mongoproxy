package authzlib

import (
	"context"
	"flag"
	"log"
	"testing"
)

/*
Unit tests for MongoProxy Authorization.

To run tests with printing config:
	go test -run Authz -config true

Without:
	go test -run Authz

To measure benchmark:
	go test -bench Authz
*/

var printer = flag.Bool("config", false, "Print the config")

func loadConfig(t *testing.T) (context.Context, Authz) {
	var a Authz
	ctx := context.Background()
	paths := make([]string, 0)
	paths = append(paths, "schema")
	var q SchemaQuerier

	if err := a.LoadConfig(ctx, paths, &q); err != nil {
		t.Fatal(err)
	}
	return ctx, a
}

type authorizeTestCaseResult struct {
	IdentityName string
	PolicyName   string
	Effect       effectType

	LogOnly bool
}

func authorizeHelper(ctx context.Context, t *testing.T, a *Authz, roles []string, method AuthorizationMethod, uri string, expected authorizeTestCaseResult) {
	resource, err := resourceFromURI(uri)
	if err != nil {
		t.Error(err)
	}

	val := a.GetSchema().Authorize(ctx, roles, method, resource)
	if val.Rule == nil {
		if expected.IdentityName == "" {
			return
		}

		t.Errorf("%s - %v - %s - Expected a result=%v got none", method, roles, uri, expected)
		return
	}

	if expected.IdentityName == "" {
		if val.Rule == nil {
			return
		}

		t.Errorf("%s - %v - %s - Expected no val=%v got none", method, roles, uri, val)
		return
	}

	if val.IdentityName != expected.IdentityName {
		t.Errorf("%s - %v - %s - IdentityName got %v - expected %v", method, roles, uri, val.IdentityName, expected.IdentityName)
	}

	if val.Rule.PolicyName != expected.PolicyName {
		t.Errorf("%s - %v - %s - PolicyName got %v - expected %v", method, roles, uri, val.Rule.PolicyName, expected.PolicyName)
	}

	if val.Rule.Effect != expected.Effect {
		t.Errorf("%s - %v - %s - Rule.Effect got %v - expected %v", method, roles, uri, val.Rule.Effect, expected.Effect)
	}

	if (len(val.LogOnlyRules) > 0) != expected.LogOnly {
		t.Errorf("%s - %v - %s - Rule.Policy got %v - expected %v", method, roles, uri, len(val.LogOnlyRules), expected.LogOnly)
	}
}

func TestAuthzPrint(t *testing.T) {
	if *printer {
		_, a := loadConfig(t)
		log.Println(a.querier)
	}
}

func TestAuthzSimple(t *testing.T) {
	ctx, a := loadConfig(t)

	role1 := []string{"role1"}
	authorizeHelper(ctx, t, &a, role1, Read, "db1/coll2/field2", authorizeTestCaseResult{"role1", "policy2", allowE, false})
	authorizeHelper(ctx, t, &a, role1, Delete, "db1/coll2/field3", authorizeTestCaseResult{"role1", "policy2", denyE, false})
}

func TestAuthzDefault(t *testing.T) {
	ctx, a := loadConfig(t)

	role1 := []string{"role1"}
	authorizeHelper(ctx, t, &a, role1, Read, "db/coll/field", authorizeTestCaseResult{})
	authorizeHelper(ctx, t, &a, role1, Read, "db/coll", authorizeTestCaseResult{})
	authorizeHelper(ctx, t, &a, role1, Read, "db", authorizeTestCaseResult{})
}

func TestAuthzNotRole(t *testing.T) {
	ctx, a := loadConfig(t)

	notRole := []string{"not a role"}
	authorizeHelper(ctx, t, &a, notRole, Read, "db/coll/field", authorizeTestCaseResult{})
}

func TestAuthzGlobFieldBlock(t *testing.T) {
	ctx, a := loadConfig(t)

	glob1 := []string{"glob1"}
	authorizeHelper(ctx, t, &a, glob1, Update, "db/coll/allowed1", authorizeTestCaseResult{"glob1", "policy3", allowE, false})
	authorizeHelper(ctx, t, &a, glob1, Delete, "db/coll/allowed1", authorizeTestCaseResult{"glob1", "policy3", denyE, false})
}

func TestAuthzGlobFieldSearch(t *testing.T) {
	ctx, a := loadConfig(t)

	glob1 := []string{"glob1"}
	authorizeHelper(ctx, t, &a, glob1, Read, "db/coll/*", authorizeTestCaseResult{"glob1", "policy3", allowE, false})
	authorizeHelper(ctx, t, &a, glob1, Delete, "db/coll/*", authorizeTestCaseResult{"glob1", "policy3", denyE, false})
	authorizeHelper(ctx, t, &a, glob1, Update, "db/coll/*", authorizeTestCaseResult{})
}

func TestAuthzGlobFieldAllow(t *testing.T) {
	ctx, a := loadConfig(t)

	glob1 := []string{"glob1"}
	authorizeHelper(ctx, t, &a, glob1, Read, "db/coll/allowed", authorizeTestCaseResult{"glob1", "policy3", allowE, false})
	authorizeHelper(ctx, t, &a, glob1, Read, "db/coll/denied", authorizeTestCaseResult{"glob1", "policy3", denyE, false})
}

func TestAuthzGlobCollBlock(t *testing.T) {
	ctx, a := loadConfig(t)

	glob2 := []string{"glob2"}

	// Glob affects field that is not in config
	authorizeHelper(ctx, t, &a, glob2, Update, "db/random/random", authorizeTestCaseResult{"glob2", "policy4", denyE, false})
	authorizeHelper(ctx, t, &a, glob2, Read, "db/random/random", authorizeTestCaseResult{"glob2", "policy4", allowE, false})

	// Field overridden by glob case
	authorizeHelper(ctx, t, &a, glob2, Update, "db/coll/field", authorizeTestCaseResult{"glob2", "policy4", denyE, false})
	authorizeHelper(ctx, t, &a, glob2, Create, "db/coll/field", authorizeTestCaseResult{"glob2", "policy4", allowE, false})
}

func TestAuthzGlobCollSearch(t *testing.T) {
	ctx, a := loadConfig(t)

	glob2 := []string{"glob2"}
	authorizeHelper(ctx, t, &a, glob2, Update, "db/*", authorizeTestCaseResult{"glob2", "policy4", denyE, false})
	authorizeHelper(ctx, t, &a, glob2, Read, "db/*", authorizeTestCaseResult{"glob2", "policy4", allowE, false})
	authorizeHelper(ctx, t, &a, glob2, Create, "db/*", authorizeTestCaseResult{})
}

func TestAuthzGlobDBBlock(t *testing.T) {
	ctx, a := loadConfig(t)

	glob3 := []string{"glob3"}

	// Glob affects field that is not in config
	authorizeHelper(ctx, t, &a, glob3, Update, "random/random/random", authorizeTestCaseResult{"glob3", "policy5", denyE, false})
	authorizeHelper(ctx, t, &a, glob3, Read, "random/random/random", authorizeTestCaseResult{"glob3", "policy5", allowE, false})
	authorizeHelper(ctx, t, &a, glob3, Update, "random/random", authorizeTestCaseResult{"glob3", "policy5", denyE, false})
	authorizeHelper(ctx, t, &a, glob3, Read, "random/random", authorizeTestCaseResult{"glob3", "policy5", allowE, false})

	// Field overridden by glob case
	authorizeHelper(ctx, t, &a, glob3, Update, "db/coll/field", authorizeTestCaseResult{"glob3", "policy5", denyE, false})
	authorizeHelper(ctx, t, &a, glob3, Create, "db/coll/field", authorizeTestCaseResult{"glob3", "policy5", allowE, false})
}

func TestAuthzGlobDBSearch(t *testing.T) {
	ctx, a := loadConfig(t)

	glob3 := []string{"glob3"}
	authorizeHelper(ctx, t, &a, glob3, Update, "*", authorizeTestCaseResult{"glob3", "policy5", denyE, false})
	authorizeHelper(ctx, t, &a, glob3, Read, "*", authorizeTestCaseResult{"glob3", "policy5", allowE, false})
	authorizeHelper(ctx, t, &a, glob3, Create, "*", authorizeTestCaseResult{})
}

func TestAuthzLogSimple(t *testing.T) {
	ctx, a := loadConfig(t)

	log1 := []string{"log1"}
	authorizeHelper(ctx, t, &a, log1, Create, "db/coll/field", authorizeTestCaseResult{LogOnly: true})
	authorizeHelper(ctx, t, &a, log1, Read, "db/coll/field", authorizeTestCaseResult{LogOnly: true})
	authorizeHelper(ctx, t, &a, log1, Update, "db/coll/field", authorizeTestCaseResult{"log1", "policy6", denyE, true})
	authorizeHelper(ctx, t, &a, log1, Delete, "db/coll/field", authorizeTestCaseResult{"log1", "policy7", denyE, true})
}

func TestAuthzGlobal(t *testing.T) {
	ctx, a := loadConfig(t)

	global := []string{"global"}
	authorizeHelper(ctx, t, &a, global, Create, "-", authorizeTestCaseResult{})
	authorizeHelper(ctx, t, &a, global, Delete, "-", authorizeTestCaseResult{"global", "policy8", allowE, false})
}

func TestAuthzMultipleRoles(t *testing.T) {
	ctx, a := loadConfig(t)

	multiple := []string{"multiple_roles1", "multiple_roles2"}
	authorizeHelper(ctx, t, &a, multiple, Create, "db/coll/field", authorizeTestCaseResult{"multiple_roles1", "policy9", allowE, false})
	authorizeHelper(ctx, t, &a, multiple, Delete, "db/coll/field", authorizeTestCaseResult{"multiple_roles2", "policy10", denyE, false})
}

var authorize Authz
var querier AuthorizationQuerier
var contx context.Context

func BenchmarkAuthzLoadConfig(b *testing.B) {
	contx = context.Background()
	paths := make([]string, 0)
	paths = append(paths, "schema")

	for i := 0; i < b.N; i++ {
		var q SchemaQuerier
		authorize.LoadConfig(contx, paths, &q)
	}

	querier = authorize.Querier()
}

func BenchmarkAuthzSimple(b *testing.B) {
	resource, err := resourceFromURI("db1/coll2/field2")
	if err != nil {
		b.Error(err)
	}
	role1 := []string{"role1"}
	for i := 0; i < b.N; i++ {
		querier.Authorize(contx, role1, Read, resource)
	}
}

func BenchmarkAuthzDefault(b *testing.B) {
	resource, err := resourceFromURI("db/coll/field")
	if err != nil {
		b.Error(err)
	}
	role1 := []string{"role1"}
	for i := 0; i < b.N; i++ {
		querier.Authorize(contx, role1, Read, resource)
	}
}

func BenchmarkAuthzNotRole(b *testing.B) {
	resource, err := resourceFromURI("db/coll/field")
	if err != nil {
		b.Error(err)
	}
	notRole := []string{"not a role"}
	for i := 0; i < b.N; i++ {
		querier.Authorize(contx, notRole, Read, resource)
	}
}

func BenchmarkAuthzGlobFieldBlock(b *testing.B) {
	resource, err := resourceFromURI("db/coll/allowed1")
	if err != nil {
		b.Error(err)
	}
	glob1 := []string{"glob1"}
	for i := 0; i < b.N; i++ {
		querier.Authorize(contx, glob1, Delete, resource)
	}
}

func BenchmarkAuthzGlobFieldSearch(b *testing.B) {
	resource, err := resourceFromURI("db/coll/*")
	if err != nil {
		b.Error(err)
	}
	glob1 := []string{"glob1"}
	for i := 0; i < b.N; i++ {
		querier.Authorize(contx, glob1, Delete, resource)
	}
}

func BenchmarkAuthzGlobFieldAllow(b *testing.B) {
	resource, err := resourceFromURI("db/coll/denied")
	if err != nil {
		b.Error(err)
	}
	glob1 := []string{"glob1"}
	for i := 0; i < b.N; i++ {
		querier.Authorize(contx, glob1, Read, resource)
	}
}

func BenchmarkAuthzGlobCollBlock(b *testing.B) {
	resource, err := resourceFromURI("db/coll/field")
	if err != nil {
		b.Error(err)
	}
	glob2 := []string{"glob2"}
	for i := 0; i < b.N; i++ {
		querier.Authorize(contx, glob2, Update, resource)
	}
}

func BenchmarkAuthzGlobDB(b *testing.B) {
	resource, err := resourceFromURI("*")
	if err != nil {
		b.Error(err)
	}
	glob3 := []string{"glob3"}
	for i := 0; i < b.N; i++ {
		querier.Authorize(contx, glob3, Update, resource)
	}
}
