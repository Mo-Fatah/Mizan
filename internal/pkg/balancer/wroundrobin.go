package balancer

import "github.com/Mo-Fatah/mizan/internal/pkg/common"

// Weighted Round Robin Balancer
// This balancer will use the weight of each server to determine the probability of it being selected
type WeightedRoundRobin struct {
}

func (wrb *WeightedRoundRobin) Next() *common.Server {
	return nil
}

func (wrb *WeightedRoundRobin) Add(s *common.Server) {
}
