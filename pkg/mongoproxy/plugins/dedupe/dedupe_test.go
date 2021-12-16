package dedupe

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

func TestDedupe(t *testing.T) {
	tr := true
	count := 0

	d := &DedupePlugin{}
	d.Configure(nil)

	p := plugins.BuildPipeline([]plugins.Plugin{d}, func(context.Context, *plugins.Request) (bson.D, error) {
		time.Sleep(time.Millisecond * 10) // mimic slow-ish backend
		count++
		return bson.D{
			{"a", 1},
		}, nil
	})

	tests := []struct {
		r    plugins.Request
		inN  int
		outN int
	}{
		{
			r: plugins.Request{
				CC:          plugins.NewClientConnection(),
				CommandName: "find",
				Command: &command.Find{
					Collection:  "foo",
					SingleBatch: &tr,
					Common: command.Common{
						ReadPreference: &command.ReadPreference{
							Mode: "secondary",
						},
					},
				},
			},
			inN:  5,
			outN: 1,
		},

		{
			r: plugins.Request{
				CC:          plugins.NewClientConnection(),
				CommandName: "find",
				Command: &command.Find{
					Collection: "foo",
					Common: command.Common{
						ReadPreference: &command.ReadPreference{
							Mode: "secondary",
						},
					},
				},
			},
			inN:  5,
			outN: 5,
		},

		{
			r: plugins.Request{
				CC:          plugins.NewClientConnection(),
				CommandName: "find",
				Command: &command.Find{
					Collection: "foo",
					Common: command.Common{
						ReadPreference: &command.ReadPreference{
							Mode: "primary",
						},
					},
				},
			},
			inN:  5,
			outN: 5,
		},
	}

	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			count = 0

			wg := sync.WaitGroup{}
			startCh := make(chan struct{})
			for x := 0; x < test.inN; x++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-startCh
					p(context.TODO(), &test.r)
				}()
			}

			// start
			close(startCh)
			wg.Wait()

			if count != test.outN {
				t.Fatalf("Unexpected N; expected=%d actual=%d", test.outN, count)
			}
		})
	}
}
