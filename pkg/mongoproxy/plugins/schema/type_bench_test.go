package schema

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"
)

var benchErr error

func BenchmarkSchemaInsert(b *testing.B) {
	var schema ClusterSchema

	buf, err := ioutil.ReadFile("example.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(buf, &schema); err != nil {
		panic(err)
	}

	for i, test := range insertTests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				benchErr = schema.ValidateInsert(context.TODO(), test.DB, test.Collection, test.In)
			}
		})
	}
}

func BenchmarkSchemaUpdate(b *testing.B) {
	var schema ClusterSchema

	buf, err := ioutil.ReadFile("example.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(buf, &schema); err != nil {
		panic(err)
	}

	for i, test := range updateTests {
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				benchErr = schema.ValidateUpdate(context.TODO(), test.DB, test.Collection, test.In, test.Upsert)
			}
		})
	}
}
