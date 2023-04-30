package mizan

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	balancer "github.com/Mo-Fatah/mizan/internal/pkg/balancer"
	"github.com/Mo-Fatah/mizan/internal/pkg/common"
	"github.com/Mo-Fatah/mizan/internal/pkg/config"
	"github.com/Mo-Fatah/mizan/internal/pkg/health"
	log "github.com/sirupsen/logrus"
)

type Mizan struct {
	// The configuration loaded from the config file
	// TODO (Mo-Fatah): Should add hot reload for config
	config *config.Config
	// Servers is a map of service matcher to a list of servers/replicas
	serverMap map[string]balancer.Balancer
	// Ports to which Mizan will listen on
	ports []int
	// The shutdown channel
	shutDown chan struct{}
}

func NewMizan(conf *config.Config) *Mizan {
	shutdown := make(chan struct{}, 1)
	serversMap := make(map[string]balancer.Balancer)
	for _, service := range conf.Services {
		servers := make([]*common.Server, 0)
		for _, replica := range service.Replicas {
			server := common.NewServer(replica, service.Name)
			servers = append(servers, server)
		}
		serversMap[service.Matcher] = newBalancer(servers, conf.Strategy)

		serversMap[service.Matcher].SetHealthChecker(health.NewHealthChecker(servers, service.Name))
	}
	// Start health checker
	for _, serviceBalancer := range serversMap {
		go serviceBalancer.HealthChecker().Start()
	}

	return &Mizan{
		config:    conf,
		serverMap: serversMap,
		ports:     conf.Ports,
		shutDown:  shutdown,
	}

}

func newBalancer(servers []*common.Server, strategy string) balancer.Balancer {
	switch strings.ToLower(strategy) {
	case "rr":
		return balancer.NewRR(servers)
	case "wrr":
		return balancer.NewWRR(servers)
	default:
		return balancer.NewRR(servers)
	}
}

func (m *Mizan) Start() {
	wg := sync.WaitGroup{}
	for _, port := range m.ports {
		wg.Add(1)
		go m.startServer(port, &wg)
	}
	wg.Wait()
}

func (m *Mizan) IsReady() bool {
	for _, port := range m.ports {
		_, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			return false
		}
	}
	return true
}

func (m *Mizan) startServer(port int, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Info("Starting http server on port ", port)
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: m,
	}

	go func() {
		// Wait for shutdown signal
		<-m.shutDown
		if err := server.Shutdown(context.TODO()); err != nil {
			log.Error(err)
		}
		log.Info("Shutting down server on port ", port)
		// Send shutdown complete signal
		m.shutDown <- struct{}{}
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Error(err)
	}
}

func (m *Mizan) findService(path string) (balancer.Balancer, error) {
	if _, ok := m.serverMap[path]; !ok {
		return nil, fmt.Errorf("couldn't find path %s", path)
	}
	return m.serverMap[path], nil
}

func (m *Mizan) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	service := r.URL.Path
	log.Infof("Request received from address %s to service %s", r.RemoteAddr, service)
	sl, err := m.findService(service)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err)
		return
	}

	server, err := sl.Next()
	if err != nil {
		// All servers are down
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("All servers are down for service %s", service)
		return
	}

	log.Infof("Proxying request to %s", server.GetUrl().String())
	server.Proxy(w, r)
}

func (m *Mizan) ShutDown() bool {
	// Send shutdown signal to all health checkers
	for _, serviceBalancer := range m.serverMap {
		serviceBalancer.HealthChecker().ShutDown()
	}

	// Send shutdown signal to all servers
	for range m.ports {
		// Send shutdown signal
		m.shutDown <- struct{}{}
		// Wait for shutdown to complete
		<-m.shutDown
	}

	log.Info("All servers are shutdown")
	return true
}
