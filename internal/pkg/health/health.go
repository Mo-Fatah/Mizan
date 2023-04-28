package health

import (
	"net"
	"time"

	"github.com/Mo-Fatah/mizan/internal/pkg/common"
	log "github.com/sirupsen/logrus"
)

var (
	// The period of time after which the health checker will check the health of all replicas of a service
	period = 10 * time.Second
	// The timeout after which the health checker will consider a server unhealthy
	timeout = 5 * time.Second
)

// HealthChecker is a struct that is responsible for checking the health of servers
// It will periodically check the health of servers and update the status of each server
// It is the only entity that can update the status of a server
type HealthChecker struct {
	servers     []*common.Server
	shutdown    chan bool
	period      time.Duration
	serviceName string
}

func NewHealthChecker(servers []*common.Server, serviceName string, ch chan bool) *HealthChecker {
	// Check the health of servers before returning to Initialize the status of servers
	if len(servers) == 0 {
		log.Fatalf("No servers provided for service: %s", serviceName)
	}

	for _, server := range servers {
		checkHealth(server)
	}

	return &HealthChecker{
		servers: servers,
		// shutdown channel is used to signal the health checker to stop checking the health of servers
		shutdown:    ch,
		serviceName: serviceName,
	}
}

func (hc *HealthChecker) SetPeriod(period time.Duration) {
	hc.period = period
}

func (hc *HealthChecker) Start() {
	log.Infof("Starting Health checker for service: %s", hc.serviceName)
	ticker := time.NewTicker(period)
	defer ticker.Stop()
	for {
		select {
		case <-hc.shutdown:
			log.Infof("Shutting down health checker for service: %s ", hc.serviceName)
			return
		case <-ticker.C:
			for _, server := range hc.servers {
				go checkHealth(server)
			}
		}
	}
}

func checkHealth(s *common.Server) {
	_, err := net.DialTimeout("tcp", s.GetUrl().Host, timeout)
	if err != nil {
		log.Errorf("Could not connect to server %s of service %s", s.GetUrl().String(), s.GetServiceName())
		oldState := s.SetLiveness(false)
		if oldState {
			log.Errorf("Transitioned server %s to unhealthy", s.GetUrl().String())
		}
		return
	}

	oldState := s.SetLiveness(true)
	if !oldState {
		log.Infof("Transitioned server %s to healthy", s.GetUrl().String())
	}
}
