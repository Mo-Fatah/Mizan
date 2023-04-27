package mizan

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	balancer "github.com/Mo-Fatah/mizan/internal/pkg/balancer"
	"github.com/Mo-Fatah/mizan/internal/pkg/common"
	"github.com/Mo-Fatah/mizan/internal/pkg/config"
	"github.com/Mo-Fatah/mizan/internal/pkg/health"
	log "github.com/sirupsen/logrus"
)

// TODO (Mo-Fatah): Add support for TLS
// TODO (Mo-Fatah): Add support for HTTP/2
// TODO (Mo-Fatah): Add support for gRPC
// TODO (Mo-Fatah): Add shutdown channel to gracefully shutdown the server
type Mizan struct {
	// The configuration loaded from the config file
	// TODO (Mo-Fatah): Should add hot reload for config
	Config *config.Config
	// Servers is a map of service matcher to a list of servers/replicas
	ServerMap map[string]balancer.Balancer
	// Ports to which Mizan will listen on
	Ports    []int
	shutDown chan bool
}

func NewMizan(conf *config.Config) *Mizan {
	shutdown := make(chan bool)
	serversMap := make(map[string]balancer.Balancer)
	for _, service := range conf.Services {
		servers := make([]*common.Server, 0)
		for _, replica := range service.Replicas {
			server := common.NewServer(replica, service.Name)
			servers = append(servers, server)
		}
		serversMap[service.Matcher] = newBalancer(servers, conf.Strategy)

		serversMap[service.Matcher].SetHealthChecker(health.NewHealthChecker(servers, shutdown))
	}
	// Start health checker
	for _, serviceBalancer := range serversMap {
		go serviceBalancer.HealthChecker().Start()
	}

	return &Mizan{
		Config:    conf,
		ServerMap: serversMap,
		Ports:     conf.Ports,
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
	var wg sync.WaitGroup
	for _, port := range m.Ports {
		go m.startServer(port, &wg)
		wg.Add(1)
	}
	wg.Wait()
}

func (m *Mizan) startServer(port int, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Info("Starting Listening on port ", port)
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: m,
	}

	go func() {
		<-m.shutDown
		log.Info("Shutting down server")
		if err := server.Close(); err != nil {
			log.Error(err)
		}
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Error(err)
	}
}

func (m *Mizan) findService(path string) (balancer.Balancer, error) {
	if _, ok := m.ServerMap[path]; !ok {
		return nil, fmt.Errorf("couldn't find path %s", path)
	}
	return m.ServerMap[path], nil
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

func (m *Mizan) ShutDown() {
	close(m.shutDown)
}
