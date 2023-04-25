package balancer

import (
	"sync"

	"github.com/Mo-Fatah/mizan/internal/pkg/common"
)

// Weighted Round Robin Balancer
// This is a weighted version of the Round Robin Balancer
// Each server has a weight associated with it, and the load balancer will select the next server based on the weight of each server
// If the weight of server is not specified, it will be set to 1
// TODO (Mo-Fatah): Add support for calculating live connecitons to each server and use that as a weight
type WRR struct {
	servers []*WRRServer
	// The index of the current server
	current uint32
	// Mutex to protect the Servers slice from concurrent writes (when adding new servers with hot reload)
	mu sync.Mutex
	// The current server load counter.
	// When this counter reaches the weight of the current server, the next server will be selected
	currentServerLoadCounter uint32
}

type WRRServer struct {
	server *common.Server
	weight uint32
}

// Next returns the next server to be used based on the weight of each server.
func (wrr *WRR) Next() *common.Server {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	if wrr.currentServerLoadCounter < wrr.servers[wrr.current].weight {
		wrr.currentServerLoadCounter++
		return wrr.servers[wrr.current].server
	}
	wrr.currentServerLoadCounter = 1
	wrr.current = (wrr.current + 1) % uint32(len(wrr.servers))
	return wrr.servers[wrr.current].server
}

func (wrr *WRR) Add(s *common.Server) {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()
	server := &WRRServer{
		server: s,
		weight: s.GetMetaOrDefaultInt("weight", 1),
	}
	wrr.servers = append(wrr.servers, server)
}
