package authzlib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
)

type Rule struct {
	PolicyName string // Name of the policy this rule came from
	RuleNumber int    // Ordinal (0-N) of rule within the given policy

	Effect    effectType
	Policy    policyType
	Condition map[string]string
	Message   string
}

func (r *Rule) String() string {
	return fmt.Sprintf("Effect: %s, Policy: %s, Condition: %v", r.Effect, r.Policy, r.Condition)
}

// RuleSlice implements Interface for a []Rule, sorting in Effect Order (Deny first)
type RuleSlice []Rule

func (x RuleSlice) Len() int { return len(x) }

// Less reports whether x[i] should be ordered before x[j], as required by the sort Interface.
func (x RuleSlice) Less(i, j int) bool {
	if x[i].Effect == x[j].Effect {
		return false
	}
	if x[i].Effect == denyE {
		return true
	}
	return false
}
func (x RuleSlice) Swap(i, j int) { x[i], x[j] = x[j], x[i] }

type ResourceRules struct {
	// set of Rules for each action
	Create []Rule
	Read   []Rule
	Update []Rule
	Delete []Rule

	LogOnlyCreate []Rule
	LogOnlyRead   []Rule
	LogOnlyUpdate []Rule
	LogOnlyDelete []Rule
}

// SortRules simply sorts rules based on their Effect
func (r *ResourceRules) SortRules() {
	sort.Sort(RuleSlice(r.Create))
	sort.Sort(RuleSlice(r.Read))
	sort.Sort(RuleSlice(r.Update))
	sort.Sort(RuleSlice(r.Delete))

	sort.Sort(RuleSlice(r.LogOnlyCreate))
	sort.Sort(RuleSlice(r.LogOnlyRead))
	sort.Sort(RuleSlice(r.LogOnlyUpdate))
	sort.Sort(RuleSlice(r.LogOnlyDelete))
}

func (r *ResourceRules) combine(other *ResourceRules) error {
	if r == nil {
		return fmt.Errorf("collections cannot be <nil>")
	}
	if r == other || other == nil {
		return nil
	}

	r.Create = append(r.Create, other.Create...)
	r.Read = append(r.Read, other.Read...)
	r.Update = append(r.Update, other.Update...)
	r.Delete = append(r.Delete, other.Delete...)

	r.LogOnlyCreate = append(r.LogOnlyCreate, other.LogOnlyCreate...)
	r.LogOnlyRead = append(r.LogOnlyRead, other.LogOnlyRead...)
	r.LogOnlyUpdate = append(r.LogOnlyUpdate, other.LogOnlyUpdate...)
	r.LogOnlyDelete = append(r.LogOnlyDelete, other.LogOnlyDelete...)

	r.SortRules()
	return nil
}

func (r *ResourceRules) String() string {
	return fmt.Sprintf("[Create: %+v, Read: %+v, Update: %+v, Delete: %+v]", r.Create, r.Read, r.Update, r.Delete)
}

// getRule returns THE matching rule and any log-only rules associated
func (r *ResourceRules) getRule(method AuthorizationMethod) *Rule {
	var rls []Rule
	switch method {
	case Create:
		rls = r.Create
	case Read:
		rls = r.Read
	case Update:
		rls = r.Update
	case Delete:
		rls = r.Delete
	default:
		return nil // TODO: this should be an error; this shouldn't be possible
	}
	if len(rls) > 0 {
		return &rls[0]
	}
	return nil
}

// getRule returns THE matching rule and any log-only rules associated
func (r *ResourceRules) getLogOnlyRules(method AuthorizationMethod) []Rule {
	switch method {
	case Create:
		return r.LogOnlyCreate
	case Read:
		return r.LogOnlyRead
	case Update:
		return r.LogOnlyUpdate
	case Delete:
		return r.LogOnlyDelete
	default:
		return nil
	}
}

type policies struct {
	// maps uri -> *ResourceRules
	Resources map[Resource]*ResourceRules
}

func (p *policies) unpackInterface(policyName string, pols interface{}) error {
	var slicePerms []interface{}
	var okay bool
	if slicePerms, okay = pols.([]interface{}); !okay {
		return fmt.Errorf("could not create []interface{} from interface{}")
	}

	p.Resources = make(map[Resource]*ResourceRules)
	for x, permission := range slicePerms {
		var perm map[string]interface{}
		if perm, okay = permission.(map[string]interface{}); !okay {
			return fmt.Errorf("could not create map[string]interface{} from interface{}")
		}

		var rule Rule
		var str string

		rule.PolicyName = policyName
		rule.RuleNumber = x
		if msg, ok := perm["Message"]; ok {
			str, ok := msg.(string)
			if !ok {
				return fmt.Errorf("message must be a string")
			}
			rule.Message = str
		}

		// Effect
		if str, okay = perm["Effect"].(string); !okay {
			return fmt.Errorf("could not get effect from interface{}")
		}
		rule.Effect = getEffect(str)

		// Policy
		if policyRaw, ok := perm["Policy"]; ok {
			if str, okay = policyRaw.(string); !okay {
				return fmt.Errorf("could not get policy from interface{}")
			}
			rule.Policy = getPolicy(str)
		}

		// Condition
		var cond map[string]interface{}
		rule.Condition = make(map[string]string)
		if cond, okay = perm["Condition"].(map[string]interface{}); !okay {
			return fmt.Errorf("could not get condition from interface{}")
		}
		for k, v := range cond {
			// TODO
			return fmt.Errorf("conditions are not implemented")
			rule.Condition[k] = v.(string)
		}

		// Resource -> Actions
		var rescs []interface{}
		if rescs, okay = perm["Resource"].([]interface{}); !okay {
			return fmt.Errorf("could not create resources slice from interface{}")
		}
		for _, r := range rescs {
			var resc ResourceRules
			resc.Create = make([]Rule, 0)
			resc.Read = make([]Rule, 0)
			resc.Update = make([]Rule, 0)
			resc.Delete = make([]Rule, 0)

			resc.LogOnlyCreate = make([]Rule, 0)
			resc.LogOnlyRead = make([]Rule, 0)
			resc.LogOnlyUpdate = make([]Rule, 0)
			resc.LogOnlyDelete = make([]Rule, 0)

			var a []interface{}
			if a, okay = perm["Action"].([]interface{}); !okay {
				return fmt.Errorf("could not create slice of actions from interface{}")
			}
			for _, action := range a {
				if rule.Policy.IsLogOnly() {
					switch action.(string) {
					case create:
						resc.LogOnlyCreate = append(resc.LogOnlyCreate, rule)
					case read:
						resc.LogOnlyRead = append(resc.LogOnlyRead, rule)
					case update:
						resc.LogOnlyUpdate = append(resc.LogOnlyUpdate, rule)
					case delete:
						resc.LogOnlyDelete = append(resc.LogOnlyDelete, rule)
					default:
						return fmt.Errorf("received a non-CRUD permission")
					}
				} else {
					switch action.(string) {
					case create:
						resc.Create = append(resc.Create, rule)
					case read:
						resc.Read = append(resc.Read, rule)
					case update:
						resc.Update = append(resc.Update, rule)
					case delete:
						resc.Delete = append(resc.Delete, rule)
					default:
						return fmt.Errorf("received a non-CRUD permission")
					}
				}
			}
			rMapInterface := r.(map[string]interface{})
			rMap := make(map[string]string)
			for key, value := range rMapInterface {
				rMap[key] = value.(string)
			}

			reso, err := getResource(rMap)
			if err != nil {
				return err
			}

			resc.SortRules()
			if _, ok := p.Resources[reso]; ok {
				p.Resources[reso].combine(&resc)
			} else {
				p.Resources[reso] = &resc
			}
		}
	}
	return nil
}

func getPolicies(p string) (map[string]*policies, error) {
	file := path.Join(p, "policies.json")
	byteValue, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	mapPolicies := make(map[string]*policies)
	var temp map[string]interface{}

	if err = json.Unmarshal(byteValue, &temp); err != nil {
		return nil, err
	}

	for policyName, policiesInterface := range temp {
		var p policies
		if err := p.unpackInterface(policyName, policiesInterface); err != nil {
			return nil, err
		}
		mapPolicies[policyName] = &p
	}
	return mapPolicies, nil
}

// combinePolicies acts as a way to take multiple policies and turn
// them into one, easily queriable, policy
func (q *AuthzSchema) combinePolicies(p map[string]*policies) error {
	if len(q.Policies) == 0 {
		q.Policies = p
	} else {
		for k, v := range p {
			if _, ok := q.Policies[k]; ok {
				if err := q.Policies[k].combine(v); err != nil {
					return err
				}
			} else {
				q.Policies[k] = v
			}
		}
	}
	return nil
}

func (p *policies) combine(other *policies) error {
	if p == nil {
		return fmt.Errorf("policies cannot be <nil>")
	}
	if p == other || other == nil {
		return nil
	}
	if len(p.Resources) == 0 {
		p.Resources = other.Resources
	} else {
		for k, v := range other.Resources {
			if _, ok := p.Resources[k]; ok {
				if err := p.Resources[k].combine(v); err != nil {
					return err
				}
			} else {
				p.Resources[k] = v
			}
		}
	}
	return nil
}

func (p *policies) String() string {
	return fmt.Sprintf("%+v", p.Resources)
}

// getRule returns THE matching rule and any log-only rules associated
func (p *policies) getRule(method AuthorizationMethod, resource Resource) *Rule {
	r, ok := p.Resources[resource]
	if !ok {
		return nil
	}
	return r.getRule(method)
}

func (p *policies) getLogOnlyRules(method AuthorizationMethod, resource Resource) []Rule {
	r, ok := p.Resources[resource]
	if !ok {
		return nil
	}
	return r.getLogOnlyRules(method)
}
