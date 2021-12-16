package insort

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type preprocessorBench func(*testing.B, bson.D)

func BenchmarkPreprocess(b *testing.B) {
	intarr := []int{}
	for i := 0; i < 10000; i++ {
		intarr = append(intarr, (i+9999)%10000)
	}
	filter := map[string]interface{}{"_id": map[string]interface{}{"$in": intarr}}
	bd, _ := bson.Marshal(filter)
	var bsonFilter bson.D
	bson.Unmarshal(bd, &bsonFilter)

	inputs := map[string]bson.D{"int": bsonFilter}

	oidarr := []primitive.ObjectID{}
	for i := 0; i < 10000; i++ {
		oidarr = append(oidarr, primitive.NewObjectID())
	}
	for i := 0; i < 5000; i++ {
		oidarr[i], oidarr[(i+9999)%10000] = oidarr[(i+9999)%10000], oidarr[i]
	}
	filter = map[string]interface{}{"_id": map[string]interface{}{"$in": oidarr}}
	bd, _ = bson.Marshal(filter)
	bson.Unmarshal(bd, &bsonFilter)
	inputs["oid"] = bsonFilter

	filter = map[string]interface{}{}
	bd, _ = bson.Marshal(filter)
	bson.Unmarshal(bd, &bsonFilter)
	inputs["empty"] = bsonFilter

	filter = map[string]interface{}{"$or": []interface{}{map[string]interface{}{"_id": map[string]interface{}{"$in": intarr}}, map[string]interface{}{"_id": map[string]interface{}{"$in": intarr}}}}
	bd, _ = bson.Marshal(filter)
	bson.Unmarshal(bd, &bsonFilter)
	inputs["or"] = bsonFilter

	tests := map[string]preprocessorBench{
		"int":   benchmarkPreprocess,
		"oid":   benchmarkPreprocess,
		"empty": benchmarkPreprocess,
		"or":    benchmarkPreprocess,
	}

	for name, f := range tests {
		b.Run(name, func(b *testing.B) {
			f(b, inputs[name])
		})
	}
}

func benchmarkPreprocess(b *testing.B, bsonFilter bson.D) {
	for n := 0; n < b.N; n++ {
		PreprocessFilter(bsonFilter, -1)
	}
}
