package authzlib

import (
	"fmt"
)

type Resource struct {
	Global     bool
	DB         string
	Collection string
	Field      string
}

// getResource receives a map[string]string and returns the global, dbs, collections and fields
//	it is attempting to access.
func getResource(resource map[string]string) (r Resource, err error) {
	if resource["Global"] == "*" {
		r.Global = true
		return r, nil
	}
	if !r.Global && resource["Database"] == "" {
		err = fmt.Errorf("must specify a db or all dbs ('*') for Database if Global is not '*'")
	} else {
		r.Field = resource["Field"]
		r.Collection = resource["Collection"]
		r.DB = resource["Database"]

		if r.Field != "" { // Field level resource
			if r.DB == "" {
				r.DB = "*"
			}
			if r.Collection == "" {
				r.Collection = "*"
			}
		} else if r.Collection != "" { // Collection level resource
			if r.DB == "" {
				r.DB = "*"
			}
		}
		// else DB level resource (do nothing)
	}
	return r, err
}

func resourceFromURI(uri string) (Resource, error) {
	var r Resource

	db, coll, fld, err := splitURI(uri)
	if err != nil {
		return r, err
	}

	if len(db) > 0 && db[0] == "-" {
		r.Global = true
	} else if len(db) > 0 {
		r.DB = db[0]
	} else {
		r.DB = "*"
	}

	if len(coll) > 0 {
		r.Collection = coll[0]
	} else {
		r.Collection = "*"
	}

	if len(fld) > 0 {
		r.Field = fld[0]
	} else {
		r.Field = "*"
	}

	return r, nil
}

func (r *Resource) String() string {
	var output string
	if r.Global {
		return "-"
	}

	if r.DB != "" {
		output += r.DB
	} else {
		output += "*"
	}

	output += "/"
	if r.Collection != "" {
		output += r.Collection
	} else {
		output += "*"
	}

	output += "/"
	if r.Field != "" {
		output += r.Field
	} else {
		output += "*"
	}

	return output
}
