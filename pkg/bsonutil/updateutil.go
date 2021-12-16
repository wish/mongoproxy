package bsonutil

import (
	"go.mongodb.org/mongo-driver/bson"
)

type UpdateFields struct {
	Create []string
	Update []string
	Delete []string
}

// ExpandUpdate expands the update to a set of fields that will be CRUD
func ExpandUpdate(d bson.D, ups *bool) UpdateFields {
	f := UpdateFields{
		Create: make([]string, 0, 10),
		Update: make([]string, 0, 10),
		Delete: make([]string, 0, 10),
	}
	upsert := GetBoolDefault(ups, false)
	for _, e := range d {
		switch e.Key {
		case "$currentDate", "$inc", "$min", "$max", "$mul":
			for _, item := range e.Value.(bson.D) {
				f.Update = append(f.Update, item.Key)
				if upsert {
					f.Create = append(f.Create, item.Key)
				}
			}
		case "$set":
			for _, item := range e.Value.(bson.D) {
				f.Create = append(f.Create, item.Key)
				f.Update = append(f.Update, item.Key)
			}
		case "$setOnInsert":
			for _, item := range e.Value.(bson.D) {
				f.Create = append(f.Create, item.Key)
			}
		case "$unset":
			for _, item := range e.Value.(bson.D) {
				f.Delete = append(f.Delete, item.Key)
			}
		case "$rename":
			for _, item := range e.Value.(bson.D) {
				f.Delete = append(f.Delete, item.Key)
				f.Update = append(f.Update, item.Value.(string))
				f.Create = append(f.Create, item.Value.(string))
			}

		// For all other keys this is just a regular set
		default:
			if upsert {
				f.Create = append(f.Create, e.Key)
			}
			f.Update = append(f.Update, e.Key)
		}
	}

	return f
}
