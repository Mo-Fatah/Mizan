package mizan

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	balancer "github.com/Mo-Fatah/mizan/internal/pkg/balancer"
	"github.com/Mo-Fatah/mizan/internal/pkg/common"
	"github.com/Mo-Fatah/mizan/internal/pkg/config"
	log "github.com/sirupsen/logrus"
)

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
	servers := make(map[string]balancer.Balancer)

	for _, service := range conf.Services {
		for _, replica := range service.Replicas {
			serverUrl, err := url.Parse(replica.Url)
			if err != nil {
				log.Fatal(err)
			}

			metaData := make(map[string]string)
			for k, v := range replica.MetaData {
				metaData[k] = v
			}
			server := common.Server{
				Url:      *serverUrl,
				Proxy:    httputil.NewSingleHostReverseProxy(serverUrl),
				MetaData: metaData,
			}

			if _, ok := servers[service.Matcher]; !ok {
				servers[service.Matcher] = NewBalancer(conf.Strategy, metaData)
			}
			servers[service.Matcher].Add(&server)
		}
	}
	return &Mizan{Config: conf, ServerMap: servers, Ports: conf.Ports, shutDown: shutdown}
}

func NewBalancer(strategy string, metaData map[string]string) balancer.Balancer {
	switch strings.ToLower(strategy) {
	case "roundrobin":
		return &balancer.RoundRobin{}
	case "weightedroundrobin":
		if _, ok := metaData["weight"]; !ok {
			log.Fatal("Weighted Round Robin strategy requires a weight to be specified in the metadata")
		}
		return &balancer.WeightedRoundRobin{}
	default:
		return &balancer.RoundRobin{}
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
	server := sl.Next()

	log.Infof("Proxying request to %s", server.Url.String())
	server.Proxy.ServeHTTP(w, r)
}

func (m *Mizan) ShutDown() {
	close(m.shutDown)
}
