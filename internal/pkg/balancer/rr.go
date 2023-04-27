package balancer

import (
	"sync"
	"time"

	"github.com/Mo-Fatah/mizan/internal/pkg/common"
	"github.com/Mo-Fatah/mizan/internal/pkg/health"
)

// Round Robin Balancer will select the next server in the list of servers in a round robin fashion.
// This is the default balancer used by Mizan.
// Equivalent to Weighted Round Robin with all weights set to 1.
type RR struct {
	servers []*common.Server
	// Mutex to protect the Servers slice from concurrent writes (when adding new servers with hot reload)
	mu *sync.Mutex
	// The index of the current server
	current uint32

	Hc *health.HealthChecker
}

func NewRR(servers []*common.Server) *RR {
	return &RR{
		servers: servers,
		mu:      &sync.Mutex{},
	}
}

func (rr *RR) Next() (*common.Server, error) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	start := time.Now()
	for {
		curr := rr.current
		rr.current = (rr.current + 1) % uint32(len(rr.servers))
		if rr.servers[curr].IsAlive() {
			return rr.servers[curr], nil
		}
		if time.Since(start) > timeout {
			return nil, ErrNoAliveServers
		}
	}
}

func (rr *RR) Add(s *common.Server) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.servers = append(rr.servers, s)
}

func (rr *RR) HealthChecker() *health.HealthChecker {
	return rr.Hc
}

func (rr *RR) SetHealthChecker(hc *health.HealthChecker) {
	rr.Hc = hc
}
