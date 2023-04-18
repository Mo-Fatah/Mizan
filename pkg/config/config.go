package config

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Config struct {
	Services []Service `yaml:"services"`
	Strategy string    `yaml:"strategy"`
}

type Service struct {
	Name     string   `yaml:"name"`
	Replicas []string `yaml:"replicas"`
}

type Server struct {
	Url   url.URL
	Proxy *httputil.ReverseProxy
}

type ServersList struct {
	Servers []*Server
	current uint32
}

func (sl *ServersList) Next() *Server {
	curr := sl.current
	sl.current = atomic.AddUint32(&sl.current, 1) % uint32(len(sl.Servers))
	return sl.Servers[curr]
}
