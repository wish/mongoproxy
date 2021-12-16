package authzlib

import (
	"encoding/json"
	"io/ioutil"
	"path"
)

func getRoles(p string) (map[string][]string, error) {
	file := path.Join(p, "roles.json")
	byteValue, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	roleToPolicies := make(map[string][]string)
	if err = json.Unmarshal(byteValue, &roleToPolicies); err != nil {
		return nil, err
	}

	return roleToPolicies, nil
}

func (q *AuthzSchema) combineRoles(r map[string][]string) {
	if len(q.Roles) == 0 {
		q.Roles = r
	} else {
		for k, v := range r {
			if _, ok := q.Roles[k]; ok {
				q.Roles[k] = appendArrayIfMissing(q.Roles[k], v)
			} else {
				q.Roles[k] = v
			}
		}
	}
}
