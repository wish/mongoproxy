package authzlib

import (
	"fmt"
	"strings"
)

// appendArrayIfMissing combines two arrays without duplicates
func appendArrayIfMissing(slice []string, other []string) []string {
	newSlice := make([]string, len(slice))
	copy(newSlice, slice)

	for _, i := range other {
		inSlice := false
		for _, ele := range slice {
			if ele == i {
				inSlice = true
				break
			}
		}
		if !inSlice {
			newSlice = append(newSlice, i)
		}
	}
	return newSlice
}

// splitURI receives a uri and returns the dbs, collections and fields
//	it is attempting to access.
//
// uri examples:
//	db/coll/fld -> [db] [coll] [fld]
//	db/coll/* -> [db] [coll] [*]
//	db/coll/fld1,fld2 -> [db] [coll] [fld1 fld2]
//	db/*/fld1,fld2 -> [db] [*] [fld1 fld2]
//
func splitURI(uri string) ([]string, []string, []string, error) {
	parts := strings.Split(uri, "/")
	if len(parts) > 3 {
		return nil, nil, nil, fmt.Errorf("too many parts in uri: %s", uri)
	} else if len(parts) == 3 {
		dbs := strings.Split(parts[0], ",")
		colls := strings.Split(parts[1], ",")
		flds := strings.Split(parts[2], ",")
		return dbs, colls, flds, nil
	} else if len(parts) == 2 {
		dbs := strings.Split(parts[0], ",")
		colls := strings.Split(parts[1], ",")
		flds := []string{"*"}
		return dbs, colls, flds, nil
	} else if len(parts) == 1 {
		dbs := strings.Split(parts[0], ",")
		colls := []string{"*"}
		flds := []string{"*"}
		return dbs, colls, flds, nil
	} else {
		dbs := []string{"*"}
		colls := []string{"*"}
		flds := []string{"*"}
		return dbs, colls, flds, nil
	}
}

// TODO: longer-term we shouldn't be expanding but rather using some datastructure
// (e.g. some tree) which can evaluate the permissions without having to expand to
// all possible options
// expandResource returns a slice of potential resources that can appear
//	in config given a single resource.
//
// Example:
//	db/coll/fld -> [db/coll/fld db/coll/* db/*/fld db/*/*
//					*/coll/fld */*/fld */coll/* */*/*]
func expandResource(r Resource) []Resource {
	if r.Global {
		return []Resource{
			{
				Global: r.Global,
			},
		}
	}
	// Field level permission
	if r.Field != "" {
		return []Resource{
			// Field level perms
			r,
			{
				Global:     r.Global,
				DB:         r.DB,
				Collection: r.Collection,
				Field:      "*",
			},
			{
				Global:     r.Global,
				DB:         r.DB,
				Collection: "*",
				Field:      r.Field,
			},
			{
				Global:     r.Global,
				DB:         r.DB,
				Collection: "*",
				Field:      "*",
			},
			{
				Global:     r.Global,
				DB:         "*",
				Collection: r.Collection,
				Field:      r.Field,
			},
			{
				Global:     r.Global,
				DB:         "*",
				Collection: r.Collection,
				Field:      "*",
			},
			{
				Global:     r.Global,
				DB:         "*",
				Collection: "*",
				Field:      r.Field,
			},
			{
				Global:     r.Global,
				DB:         "*",
				Collection: "*",
				Field:      "*",
			},
		}
	}
	// Collection level permission
	if r.Collection != "" {
		return []Resource{
			r,
			{
				Global:     r.Global,
				DB:         r.DB,
				Collection: "*",
				Field:      r.Field,
			},
			{
				Global:     r.Global,
				DB:         "*",
				Collection: "*",
				Field:      r.Field,
			},
		}
	}
	// DB level permission
	if r.DB != "" && r.DB != "*" {
		return []Resource{
			r,
			{
				Global: r.Global,
				DB:     "*",
			},
		}
	}

	// At this point its db:"*" and everything else is blank
	return []Resource{
		r,
	}
}
