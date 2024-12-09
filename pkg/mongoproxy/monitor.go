package mongoproxy

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/wish/mongoproxy/pkg/mongoproxy/config"
	"net"
	"sort"
	"sync"
	"time"
)

type ReqMonitor struct {
	LoggingConfig *config.LoggingConfig
	requestCounts sync.Map
}

func NewReqMonitor(cfg *config.Config) *ReqMonitor {
	loggingConfig := cfg.Logging
	monitor := &ReqMonitor{
		requestCounts: sync.Map{},
		LoggingConfig: &loggingConfig,
	}
	if monitor.IsEnabled() {
		go monitor.logTopClients(time.Duration(loggingConfig.LogIntervalSeconds)*time.Second, loggingConfig.TopN)
	}

	return monitor
}

func (rm *ReqMonitor) sortClientsByRequests(clientCounts map[string]int64) []clientInfo {
	var clients []clientInfo
	for addr, count := range clientCounts {
		clients = append(clients, clientInfo{addr: addr, count: count})
	}

	sort.Slice(clients, func(i, j int) bool {
		return clients[i].count > clients[j].count
	})
	return clients
}

func (rm *ReqMonitor) logTopClients(interval time.Duration, topN int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		clientCounts := make(map[string]int64)
		rm.requestCounts.Range(func(key, value interface{}) bool {
			clientCounts[key.(string)] = value.(int64)
			return true
		})
		//reset requests
		rm.requestCounts = sync.Map{}

		sortedClients := rm.sortClientsByRequests(clientCounts)
		var topClients []map[string]interface{}
		for i, client := range sortedClients {
			if i >= topN {
				break
			}
			topClients = append(topClients, map[string]interface{}{
				"address":       client.addr,
				"request_count": client.count,
			})
		}
		topClientsJSON, _ := json.Marshal(topClients)
		logrus.Debugf("Top clients by request frequency: %s", topClientsJSON)
	}
}

type clientInfo struct {
	addr  string
	count int64
}

func (rm *ReqMonitor) IsEnabled() bool {
	return rm.LoggingConfig.Enable
}

func (rm *ReqMonitor) UpdateRequestCount(clientAddr string) {
	if rm.IsEnabled() {
		host := clientAddr
		if parsedHost, _, err := net.SplitHostPort(clientAddr); err == nil {
			host = parsedHost
		}
		count, _ := rm.requestCounts.LoadOrStore(host, int64(0))
		rm.requestCounts.Store(host, count.(int64)+1)
	}
}
