package authzlib

import (
	"context"
	"fmt"
)

// AuthzSchema implements AuthorizationQuerier. It stores the
// information to be queried from.
type AuthzSchema struct {
	Roles    map[string][]string  // Role name -> list of Policy names
	Policies map[string]*policies // Policy name -> policies object
}

func (q *AuthzSchema) getSchema(path string) error {
	var err error

	var r map[string][]string
	if r, err = getRoles(path); err != nil {
		return err
	}

	var p map[string]*policies
	if p, err = getPolicies(path); err != nil {
		return err
	}

	q.combineRoles(r)
	if err = q.combinePolicies(p); err != nil {
		return err
	}

	return nil
}

// Authorize will authorize the given request based on the URI
// passed in. The URI might be a subset (e.g. DB/Collection) in
// cases where we want to pre-check permissions (e.g. if no
// permissions on anything, just fail to avoid the subsequent
// lookups

//   (policyType, effectType, identity/role, policy)

func (q *AuthzSchema) Authorize(ctx context.Context, identities []string, method AuthorizationMethod, resource Resource) AuthorizeResult {
	resources := expandResource(resource)
	pols := make(map[string]string) // store reverse map of policy -> identity
	for _, identity := range identities {
		if p, ok := q.Roles[identity]; ok {
			for _, item := range p {
				pols[item] = identity
			}
		}
	}
	if len(pols) == 0 {
		return AuthorizeResult{
			AuthorizationMethod: method,
			Resource:            resource,
		}
	}

	// TODO: accumulate all rules across the board for LogOnlyRules
	var (
		allowResult        AuthorizeResult
		denyResult         AuthorizeResult
		resultLogOnlyRules []Rule
	)

	for policy, identity := range pols {
		p, ok := q.Policies[policy]
		if !ok {
			continue
		}
		for _, r := range resources {
			resultLogOnlyRules = append(resultLogOnlyRules, p.getLogOnlyRules(method, r)...)

			// If we have a deny, no need to pull the "real" rules
			if denyResult.Rule != nil {
				continue
			}

			rule := p.getRule(method, r)
			// If we didn't find a rule then we continue to attempt to find another
			if rule == nil {
				continue
			}
			// If the rule is deny, then we return immediately as deny always takes precedence
			if denyResult.Rule == nil && rule.Effect.IsDeny() {
				denyResult = AuthorizeResult{
					IdentityName: identity,
					Rule:         rule,
				}
			}

			// as an allow we just need the first one; we want to move through the rest of the policies to ensure there aren't deny rules
			if allowResult.Rule == nil {
				allowResult = AuthorizeResult{
					IdentityName: identity,
					Rule:         rule,
				}
			}
		}
	}

	if denyResult.Rule != nil {
		denyResult.LogOnlyRules = resultLogOnlyRules
		denyResult.Resource = resource
		denyResult.AuthorizationMethod = method
		return denyResult
	}
	if allowResult.Rule == nil {
		return AuthorizeResult{
			AuthorizationMethod: method,
			Resource:            resource,
			LogOnlyRules:        resultLogOnlyRules,
		}
	}
	allowResult.LogOnlyRules = resultLogOnlyRules
	allowResult.Resource = resource
	allowResult.AuthorizationMethod = method
	return allowResult
}

func (q *AuthzSchema) String() string {
	return fmt.Sprintf("AuthzSchema:\n\tRoles: %+v\n\tPolicies: %+v", q.Roles, q.Policies)
}
