module github.com/wish/mongoproxy

go 1.13

require (
	github.com/HdrHistogram/hdrhistogram-go v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1
	github.com/getsentry/sentry-go v0.9.0
	github.com/jellydator/ttlcache/v2 v2.11.1
	github.com/jessevdk/go-flags v1.4.0
	github.com/json-iterator/go v1.1.11
	github.com/miekg/dns v1.1.41 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/prometheus/client_golang v1.8.0
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/uber/jaeger-client-go v2.27.0+incompatible
	github.com/uber/jaeger-lib v2.4.1+incompatible // indirect
	github.com/wish/discovery v0.0.0-20190510213300-be3745886c68
	go.mongodb.org/mongo-driver v1.5.1
	go.uber.org/automaxprocs v1.3.0
	golang.org/x/net v0.0.0-20210331212208-0fccb6fa2b5c // indirect
	golang.org/x/sync v0.1.0
	golang.org/x/sys v0.0.0-20210426080607-c94f62235c83 // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324
	gopkg.in/fsnotify.v1 v1.4.7
)

replace go.mongodb.org/mongo-driver => github.com/wish/mongo-go-driver v1.5.1-fork
