package dedupe

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/wish/mongoproxy/pkg/command"
	"github.com/wish/mongoproxy/pkg/mongoproxy/plugins"
)

func TestDedupeKey(t *testing.T) {
	cmds := []*command.Find{
		{
			Filter: bson.D{{"a", 1}},
		},
		{
			Filter: bson.D{{"a", 2}},
		},
	}

	keys := make(map[string]int)
	for _, cmd := range cmds {
		keys[DedupeKey(cmd)]++
	}

	if len(keys) != len(cmds) {
		fmt.Println(keys)
		t.Fatalf("duplicated key!")
	}
}

func BenchmarkDedupe(b *testing.B) {
	count := 0

	d := &DedupePlugin{}
	d.Configure(nil)

	p := plugins.BuildPipeline([]plugins.Plugin{d}, func(context.Context, *plugins.Request) (bson.D, error) {
		time.Sleep(time.Millisecond) // mimic slow-ish backend
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
					Collection: "foo",
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
		b.Run(strconv.Itoa(i), func(b *testing.B) {
			count = 0

			wg := sync.WaitGroup{}
			startCh := make(chan struct{})
			for x := 0; x < test.inN; x++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-startCh
					for y := 0; y < b.N; y++ {
						p(context.TODO(), &test.r)
					}
				}()
			}

			// start
			close(startCh)
			wg.Wait()
		})
	}
}
