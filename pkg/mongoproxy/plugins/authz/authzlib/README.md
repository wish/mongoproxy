## MongoProxy Authorization

MongoProxy Authorization is a feature in MongoProxy that can limit the ability of a user to CRUD permissions on collections/documents/fields based on the users' business needs.
<br />

### API

See `types.go` for blackbox functions.

### Schema Structure

The general schema design is as follows:

```
schema/
    policies/
        policy_1.json
        ...
        policy_n.json
    roles.json
```
<br />

### Code

Related code can be found in the following files:

* `types.go` (Defines `Authorization`, `AuthorizationMethod` and `AuthorizationQuerier` interfaces)
* `authz.go` (Loads configs for querier in `Authz` struct)
* `querier.go` (Authorization piece + implements `AuthzSchema`)
* `policies.go` (Implements policies to be queried by the querier)
* `roles.go` (Implements roles to be queried by the querier)
* `enforce.go` (Handles enforce > log > authorized > default precedence for helping with authorization piece)
* `utils.go` (Some useful helper functions)
* `authz_test.go` (Unit tests)
