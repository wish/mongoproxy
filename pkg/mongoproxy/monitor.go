package mongoproxy

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"github.com/wish/mongoproxy/pkg/mongoproxy/config"
	"net"
	"sort"
	"strings"
	"sync"
	"time"
)

type ReqMonitor struct {
	LoggingConfig   *config.LoggingConfig
	operationCounts sync.Map
}

func NewReqMonitor(cfg *config.Config) *ReqMonitor {
	loggingConfig := cfg.Logging
	monitor := &ReqMonitor{
		operationCounts: sync.Map{},
		LoggingConfig:   &loggingConfig,
	}
	if monitor.IsEnabled() {
		go monitor.logTopClientOps(time.Duration(loggingConfig.LogIntervalSeconds)*time.Second, loggingConfig.TopN)
	}
	return monitor
}

func (rm *ReqMonitor) UpdateRequestCount(clientAddr, opCode string) {
	if rm.IsEnabled() {
		host := clientAddr
		if parsedHost, _, err := net.SplitHostPort(clientAddr); err == nil {
			host = parsedHost
		}

		clientOpKey := host + "|" + opCode
		opCount, _ := rm.operationCounts.LoadOrStore(clientOpKey, int64(0))
		rm.operationCounts.Store(clientOpKey, opCount.(int64)+1)
	}
}

func (rm *ReqMonitor) logTopClientOps(interval time.Duration, topN int) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		clientOpCounts := make(map[string]int64)
		rm.operationCounts.Range(func(key, value interface{}) bool {
			clientOpCounts[key.(string)] = value.(int64)
			return true
		})
		rm.operationCounts = sync.Map{}

		var clientOps []clientInfo
		for key, count := range clientOpCounts {
			clientOps = append(clientOps, clientInfo{addr: key, count: count})
		}
		sort.Slice(clientOps, func(i, j int) bool {
			return clientOps[i].count > clientOps[j].count
		})

		var topClientOps []map[string]interface{}
		for i, clientOp := range clientOps {
			if i >= topN {
				break
			}
			parts := strings.SplitN(clientOp.addr, "|", 2)
			client := parts[0]
			op := ""
			if len(parts) > 1 {
				op = parts[1]
			}
			topClientOps = append(topClientOps, map[string]interface{}{
				"client":       client,
				"operation":    op,
				"request_count": clientOp.count,
			})
		}

		topClientOpsJSON, _ := json.Marshal(topClientOps)
		logrus.Debugf("Top client+op by request frequency: %s", topClientOpsJSON)
	}
}

type clientInfo struct {
	addr  string
	count int64
}

func (rm *ReqMonitor) IsEnabled() bool {
	return rm.LoggingConfig.Enable
}
