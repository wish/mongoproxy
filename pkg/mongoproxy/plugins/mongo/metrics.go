package mongo

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.mongodb.org/mongo-driver/event"
)

var (
	connectionsCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_mongo_connection_created_total",
		Help: "The duration of mongo commands",
	}, []string{"address"})

	connectionsClosedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongoproxy_plugins_mongo_connection_closed_total",
		Help: "The duration of mongo commands",
	}, []string{"address"})

	connectionPoolSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mongoproxy_plugins_mongo_connection_pool_available_count",
		Help: "The current number of available connections in the pool",
	}, []string{"address"})

	inUseConnections = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "mongoproxy_plugins_mongo_connection_pool_inuse_count",
		Help: "The current number of in-use connections in the pool",
	}, []string{"address"})
)

// Pool Metrics
var PoolMonitor = event.PoolMonitor{
	Event: func(e *event.PoolEvent) {
		switch e.Type {
		case "ConnectionCreated":
			connectionsCreatedTotal.WithLabelValues(e.Address).Inc()
			connectionPoolSize.WithLabelValues(e.Address).Inc()
		case "ConnectionClosedEvent":
			connectionsClosedTotal.WithLabelValues(e.Address).Inc()
			connectionPoolSize.WithLabelValues(e.Address).Dec()
		case "ConnectionCheckedOut":
			connectionPoolSize.WithLabelValues(e.Address).Dec()
			inUseConnections.WithLabelValues(e.Address).Inc()
		case "ConnectionCheckedIn":
			connectionPoolSize.WithLabelValues(e.Address).Inc()
			inUseConnections.WithLabelValues(e.Address).Dec()
		}
	},
}

// Command metrics
var CommandMonitor = event.CommandMonitor{
	Started: func(ctx context.Context, e *event.CommandStartedEvent) {
	},
	Succeeded: func(ctx context.Context, e *event.CommandSucceededEvent) {
	},
	Failed: func(ctx context.Context, e *event.CommandFailedEvent) {
	},
}

/*
var ServerMonitor = event.ServerMonitor{
	ServerDescriptionChanged: func(e *event.ServerDescriptionChangedEvent) {
		fmt.Println("HandleServerDescriptionChanged", e)
	},
	ServerOpening: func(e *event.ServerOpeningEvent) {
		fmt.Println("HandleServerOpening", e)
	},
	ServerClosed: func(e *event.ServerClosedEvent) {
		fmt.Println("HandleServerClosed", e)
	},
	TopologyDescriptionChanged: func(e *event.TopologyDescriptionChangedEvent) {
		fmt.Println("HandleTopologyDescriptionChanged", e)
	},
	TopologyOpening: func(e *event.TopologyOpeningEvent) {
		fmt.Println("HandleTopologyOpening", e)
	},
	TopologyClosed: func(e *event.TopologyClosedEvent) {
		fmt.Println("HandleTopologyClosed", e)
	},
	ServerHeartbeatStarted: func(e *event.ServerHeartbeatStartedEvent) {
		fmt.Println("HandleServerHeartbeatStarted", e)
	},
	ServerHeartbeatSucceeded: func(e *event.ServerHeartbeatSucceededEvent) {
		fmt.Println("HandleServerHeartbeatSucceeded", e)
	},
	ServerHeartbeatFailed: func(e *event.ServerHeartbeatFailedEvent) {
		fmt.Println("HandleServerHeartbeatFailed", e)
	},
}
*/
