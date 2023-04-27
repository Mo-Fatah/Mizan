package balancer

import (
	"sync"
	"time"

	"github.com/Mo-Fatah/mizan/internal/pkg/common"
	"github.com/Mo-Fatah/mizan/internal/pkg/health"
)

// Weighted Round Robin Balancer
// This is a weighted version of the Round Robin Balancer
// Each server has a weight associated with it, and the load balancer will select the next server based on the weight of each server
// If the weight of server is not specified, it will be set to 1
// TODO (Mo-Fatah): Add support for calculating live connecitons to each server and use that as a weight
type WRR struct {
	servers []*common.Server
	// Mutex to protect the Servers slice from concurrent writes (when adding new servers with hot reload)
	mu *sync.Mutex
	// The index of the current server
	current uint32
	// The current server load counter.
	// When this counter reaches the weight of the current server, the next server will be selected
	currentServerLoadCounter uint32

	hc *health.HealthChecker
}

func NewWRR() *WRR {
	return &WRR{
		servers: []*common.Server{},
		mu:      &sync.Mutex{},
	}
}

// Next returns the next server to be used based on the weight of each server.
func (wrr *WRR) Next() (*common.Server, error) {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()
	var server *common.Server
	start := time.Now()
	for {
		if wrr.currentServerLoadCounter < wrr.servers[wrr.current].GetWeight() {
			wrr.currentServerLoadCounter++
			server = wrr.servers[wrr.current]
		} else {
			wrr.currentServerLoadCounter = 1
			wrr.current = (wrr.current + 1) % uint32(len(wrr.servers))
			server = wrr.servers[wrr.current]
		}

		if server.IsAlive() {
			break
		}

		if time.Since(start) > timeout {
			return nil, ErrNoAliveServers
		}
	}
	return server, nil
}

func (wrr *WRR) Add(s *common.Server) {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()
	s.SetWeight(s.GetMetaOrDefaultInt("weight", 1))
	wrr.servers = append(wrr.servers, s)
}
