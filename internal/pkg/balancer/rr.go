package balancer

import (
	"sync"
	"sync/atomic"

	"github.com/Mo-Fatah/mizan/internal/pkg/common"
)

// RoundRobin Balancer will select the next server in the list of servers in a round robin fashion
// This is the default balancer used by Mizan
// This is the same as Weighted Round Robin Balancer with all weights set to 1
type RR struct {
	Servers []*common.Server
	// The index of the current server
	current uint32
	// Mutex to protect the Servers slice from concurrent writes (when adding new servers with hot reload)
	mu sync.Mutex
}

func (rr *RR) Next() *common.Server {
	curr := rr.current
	rr.current = atomic.AddUint32(&rr.current, 1) % uint32(len(rr.Servers))
	return rr.Servers[curr]
}

func (rr *RR) Add(s *common.Server) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.Servers = append(rr.Servers, s)
}
