package balancer

import (
	"sync/atomic"

	"github.com/Mo-Fatah/mizan/internal/pkg/common"
)

// RoundRobin Balancer will select the next server in the list of servers in a round robin fashion
type RoundRobin struct {
	Servers []*common.Server
	current uint32
}

func (rb *RoundRobin) Next() *common.Server {
	curr := rb.current
	rb.current = atomic.AddUint32(&rb.current, 1) % uint32(len(rb.Servers))
	return rb.Servers[curr]
}

func (rb *RoundRobin) Add(s *common.Server) {
	rb.Servers = append(rb.Servers, s)
}
