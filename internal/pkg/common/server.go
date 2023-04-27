package common

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"

	"github.com/Mo-Fatah/mizan/internal/pkg/config"
)

type Server struct {
	serviceName string
	url         *url.URL
	// Proxy is the reverse proxy that will be used to proxy requests to this server
	proxy *httputil.ReverseProxy
	// MetaData is a map of key-value pairs that can be used by the balancer
	metaData map[string]string
	// weight is used by the Weighted Round Robin balancer, defaults to 1 if not specified
	weight uint32
	// alive is used by the Balancer's Health Checker
	alive bool

	mu *sync.Mutex
}

// TODO (Mo-Fatah): Refactor this to use the Service struct as a parameter
func NewServer(replica *config.Replica, serviceName string) *Server {
	serverUrl, err := url.Parse(replica.Url)
	if err != nil {
		log.Fatal(err)
	}

	metaData := make(map[string]string)
	for k, v := range replica.MetaData {
		metaData[k] = v
	}

	server := &Server{
		url:         serverUrl,
		proxy:       httputil.NewSingleHostReverseProxy(serverUrl),
		metaData:    metaData,
		alive:       false, /* TODO (Mo-Fatah): Should be false by default */
		serviceName: serviceName,
		mu:          &sync.Mutex{},
	}
	server.weight = server.GetMetaOrDefaultInt("weight", 1)
	return server
}

func (s *Server) Proxy(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func (s *Server) IsAlive() bool {
	return s.alive
}

func (s *Server) SetLiveness(alive bool) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	old := s.alive
	s.alive = alive
	return old
}

func (s *Server) GetUrl() *url.URL {
	return s.url
}

// TODO (Mo-Fatah): Implement this
func (s *Server) GetServiceName() string {
	return s.serviceName
}

func (s *Server) GetMetaOrDefault(key string, defaultValue string) string {
	if value, ok := s.metaData[key]; ok {
		return value
	}
	return defaultValue
}

func (s *Server) GetMetaOrDefaultInt(key string, defaultValue int) uint32 {
	if value, ok := s.metaData[key]; ok {
		v, err := strconv.Atoi(value)
		if err == nil {
			return uint32(v)
		}
	}
	return uint32(defaultValue)
}

func (s *Server) GetWeight() uint32 {
	return s.weight
}

func (s *Server) SetWeight(weight uint32) {
	s.weight = weight
}
