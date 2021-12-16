package mongo

import (
	"reflect"
	"unsafe"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
	"go.mongodb.org/mongo-driver/x/mongo/driver/operation"
	"go.mongodb.org/mongo-driver/x/mongo/driver/topology"
)

func ErrorToDoc(err error) (bson.D, error) {
	switch e := err.(type) {
	case mongo.WriteException:
		writeErrorDocs := make([]bson.D, len(e.WriteErrors))
		for i, wErr := range e.WriteErrors {
			writeErrorDocs[i] = []primitive.E{
				{"code", int(wErr.Code)},
				{"errmsg", wErr.Message},
				{"index", wErr.Index},
			}
		}

		ret := []primitive.E{
			{"n", 0},
			{"writeErrors", writeErrorDocs},
		}

		if e.WriteConcernError != nil {
			subDoc, err := ErrorToDoc(e.WriteConcernError)
			if err != nil {
				return nil, err
			}
			ret = append(ret, bson.E{"writeConcernError", subDoc})
		}

		if len(e.Labels) > 0 {
			ret = append(ret, bson.E{"errorLabels", e.Labels})
		}

		return ret, nil
	case mongo.CommandError:
		ret := []primitive.E{
			{"code", int(e.Code)},
			{"codeName", e.Name},
			{"errmsg", e.Message},
		}

		if len(e.Labels) > 0 {
			ret = append(ret, bson.E{"errorLabels", e.Labels})
		}

		return ret, nil

		// If we got an error doing a write we need to convert to a bson.D of the error
		// This is not a query error (so ok==1) but it is a writeError
	case driver.WriteCommandError:
		writeErrorDocs := make([]bson.D, len(e.WriteErrors))
		for i, wErr := range e.WriteErrors {
			writeErrorDocs[i] = []primitive.E{
				{"code", int(wErr.Code)},
				{"errmsg", wErr.Message},
				{"index", wErr.Index},
			}
		}

		ret := []primitive.E{
			{"n", 0},
			{"writeErrors", writeErrorDocs},
		}

		if e.WriteConcernError != nil {
			subDoc, err := ErrorToDoc(e.WriteConcernError)
			if err != nil {
				return nil, err
			}
			ret = append(ret, bson.E{"writeConcernError", subDoc})
		}

		if len(e.Labels) > 0 {
			ret = append(ret, bson.E{"errorLabels", e.Labels})
		}

		return ret, nil

	case mongo.BulkWriteException:
		writeErrorDocs := make([]bson.D, len(e.WriteErrors))
		for i, wErr := range e.WriteErrors {
			writeErrorDocs[i] = []primitive.E{
				{"code", int(wErr.Code)},
				{"errmsg", wErr.Message},
				{"index", wErr.Index},
			}
		}

		ret := []primitive.E{
			{"ok", 1},
			{"n", 0},
			{"writeErrors", writeErrorDocs},
		}

		if e.WriteConcernError != nil {
			subDoc, err := ErrorToDoc(e.WriteConcernError)
			if err != nil {
				return nil, err
			}
			ret = append(ret, bson.E{"writeConcernError", subDoc})
		}

		if len(e.Labels) > 0 {
			ret = append(ret, bson.E{"errorLabels", e.Labels})
		}

		return ret, nil
	case *mongo.WriteConcernError:
		return bson.D{
			{"code", int(e.Code)},
			{"codeName", e.Name},
			{"errmsg", e.Message},
		}, nil

	case *driver.WriteConcernError:
		return bson.D{
			{"code", int(e.Code)},
			{"codeName", e.Name},
			{"errmsg", e.Message},
		}, nil

	case driver.Error:
		ret := []primitive.E{
			{"code", int(e.Code)},
			{"codeName", e.Name},
			{"errmsg", e.Message},
		}

		if len(e.Labels) > 0 {
			ret = append(ret, bson.E{"errorLabels", e.Labels})
		}

		return ret, nil

	default:
		return nil, err
	}
}

func extractTopology(c *mongo.Client) *topology.Topology {
	e := reflect.ValueOf(c).Elem()
	d := e.FieldByName("deployment")
	d = reflect.NewAt(d.Type(), unsafe.Pointer(d.UnsafeAddr())).Elem() // #nosec G103
	return d.Interface().(*topology.Topology)
}

func extractServer(c *operation.Command) driver.Server {
	e := reflect.ValueOf(c).Elem()
	d := e.FieldByName("srvr")
	d = reflect.NewAt(d.Type(), unsafe.Pointer(d.UnsafeAddr())).Elem() // #nosec G103
	if d.IsNil() {
		return nil
	}
	return d.Interface().(driver.Server)
}
