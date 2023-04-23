package config

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Config struct {
	Services []Service `yaml:"services"`
	Strategy string    `yaml:"strategy"`
	// Ports to which Mizan will listen on
	// TODO (Mo-Fatah): Should deal with distributed ports across multiple nodes
	Ports []int `yaml:"ports"`
}

type Service struct {
	Name     string   `yaml:"name"`
	Matcher  string   `yaml:"matcher"`
	Replicas []string `yaml:"replicas"`
}

type Server struct {
	Url   url.URL
	Proxy *httputil.ReverseProxy
}

type ServerList struct {
	Servers []*Server
	current uint32
}

func (sl *ServerList) Next() *Server {
	curr := sl.current
	sl.current = atomic.AddUint32(&sl.current, 1) % uint32(len(sl.Servers))
	return sl.Servers[curr]
}

func (sl *ServerList) Add(s *Server) {
	sl.Servers = append(sl.Servers, s)
}
