package writeconcernoverride

import (
	"context"
	"reflect"
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/bsonutil"
	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

func TestWritePref(t *testing.T) {
	d := &WriteconcernOverridePlugin{}
	d.Configure(bson.D{
		bson.E{"updateOverride", bson.D{
			{"majority", int64(1)},
			{"2", int64(1)},
		}},
	})

	p := plugins.BuildPipeline([]plugins.Plugin{d}, func(ctx context.Context, request *plugins.Request) (bson.D, error) {
		switch cmd := request.Command.(type) {
		case *command.Update:
			return bson.D{
				bson.E{"writepref", cmd.WriteConcern.W},
			}, nil
		}
		return bson.D{
			{"a", 1},
		}, nil
	})

	tests := []struct {
		r            plugins.Request
		outWritePref interface{}
	}{
		{
			r: plugins.Request{
				CC:          plugins.NewClientConnection(),
				CommandName: "update",
				Command: &command.Update{
					Collection: "foo",
					WriteConcern: &command.WriteConcern{
						W: "majority",
					},
				},
			},
			outWritePref: int64(1),
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			result, err := p(context.TODO(), &test.r)
			if err != nil {
				t.Fatal(err)
			}

			v, ok := bsonutil.Lookup(result, "writepref")
			if !ok {
				t.Fatalf("No writepref found in response!")
			}
			if !reflect.DeepEqual(v, test.outWritePref) {
				t.Fatalf("writepref not overwritten as expected; expected=%v:%T actual=%v:%T", test.outWritePref, test.outWritePref, v, v)
			}
		})
	}
}
