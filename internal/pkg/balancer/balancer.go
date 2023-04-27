package balancer

import (
	"github.com/Mo-Fatah/mizan/internal/pkg/common"
	"github.com/Mo-Fatah/mizan/internal/pkg/health"
)

// Balancer is an interface that defines the behavior of a load balancer
type Balancer interface {
	// Next returns the next server to be used
	Next() (*common.Server, error)
	// Add adds a new server to the balancer
	Add(*common.Server)

	HealthChecker() *health.HealthChecker

	SetHealthChecker(*health.HealthChecker)
}
