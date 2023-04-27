package common

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
)

type Server struct {
	url *url.URL
	// Proxy is the reverse proxy that will be used to proxy requests to this server
	proxy *httputil.ReverseProxy
	// MetaData is a map of key-value pairs that can be used by the balancer
	metaData map[string]string
	// weight is used by the Weighted Round Robin balancer, defaults to 1 if not specified
	weight uint32
	// alive is used by the Balancer's Health Checker
	alive bool
}

func NewServer(url url.URL, proxy *httputil.ReverseProxy, metaData map[string]string) *Server {
	return &Server{
		url:      &url,
		proxy:    proxy,
		metaData: metaData,
		weight:   1,
		alive:    false,
	}
}

func (s *Server) Proxy(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func (s *Server) IsAlive() bool {
	return s.alive
}

func (s *Server) SetLiveness(alive bool) bool {
	old := s.alive
	s.alive = alive
	return old
}

func (s *Server) GetUrl() *url.URL {
	return s.url
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
