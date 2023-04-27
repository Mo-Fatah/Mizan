package balancer

import (
	"sync"

	"github.com/Mo-Fatah/mizan/internal/pkg/common"
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
}

func NewRR() *RR {
	return &RR{
		servers: []*common.Server{},
		mu:      &sync.Mutex{},
	}
}

func (rr *RR) Next() *common.Server {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	curr := rr.current
	rr.current = (rr.current + 1) % uint32(len(rr.servers))
	return rr.servers[curr]
}

func (rr *RR) Add(s *common.Server) {
	rr.mu.Lock()
	defer rr.mu.Unlock()
	rr.servers = append(rr.servers, s)
}
