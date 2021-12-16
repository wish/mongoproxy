package gotests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/wish/mongoproxy/integrationtest"
)

var (
	ctx = context.Background()
)

type Trainer struct {
	Name string
	Age  int
	City string
}

// TODO: better test suite (maybe re-use client library tests?)
// tests copied from https://github.com/coinbase/mongobetween/blob/b7344eaf5bbd31ae83b83925282d4d7d7190d80e/proxy/proxy_test.go#L132
func TestProxy(t *testing.T) {
	// TODO: support retry writes?
	client := integrationtest.SetupClient(t, options.Client().SetRetryWrites(false))
	collection := client.Database("test").Collection("trainers")
	_, err := collection.DeleteMany(ctx, bson.D{{}})
	assert.Nil(t, err)

	ash := Trainer{"Ash", 10, "Pallet Town"}
	misty := Trainer{"Misty", 10, "Cerulean City"}
	brock := Trainer{"Brock", 15, "Pewter City"}

	_, err = collection.InsertOne(ctx, ash)
	assert.Nil(t, err)

	_, err = collection.InsertMany(ctx, []interface{}{misty, brock})
	assert.Nil(t, err)

	filter := bson.D{{Key: "name", Value: "Ash"}}
	update := bson.D{
		{Key: "$inc", Value: bson.D{
			{Key: "age", Value: 1},
		}},
	}
	updateResult, err := collection.UpdateOne(ctx, filter, update)
	assert.Nil(t, err)
	assert.Equal(t, int64(1), updateResult.MatchedCount)
	assert.Equal(t, int64(1), updateResult.ModifiedCount)

	var result Trainer
	err = collection.FindOne(ctx, filter).Decode(&result)
	assert.Nil(t, err)
	assert.Equal(t, "Pallet Town", result.City)

	var results []Trainer
	cur, err := collection.Find(ctx, bson.D{}, options.Find().SetLimit(2).SetBatchSize(1))
	assert.Nil(t, err)
	err = cur.All(ctx, &results)
	assert.Nil(t, err)
	assert.Equal(t, "Pallet Town", results[0].City)
	assert.Equal(t, "Cerulean City", results[1].City)

	deleteResult, err := collection.DeleteMany(ctx, bson.D{{}})
	assert.Nil(t, err)
	assert.Equal(t, int64(3), deleteResult.DeletedCount)

	err = client.Disconnect(ctx)
	assert.Nil(t, err)

	// TODO
	// proxy.Shutdown()
}
