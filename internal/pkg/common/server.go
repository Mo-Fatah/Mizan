package common

import (
	"net/http/httputil"
	"net/url"
	"strconv"
)

type Server struct {
	Url   url.URL
	Proxy *httputil.ReverseProxy
	// MetaData is a map of key-value pairs that can be used by the balancer
	MetaData map[string]string
	// Weight is used by the Weighted Round Robin balancer
	Weight uint32
}

func (s *Server) GetMetaOrDefault(key string, defaultValue string) string {
	if value, ok := s.MetaData[key]; ok {
		return value
	}
	return defaultValue
}

func (s *Server) GetMetaOrDefaultInt(key string, defaultValue int) uint32 {
	if value, ok := s.MetaData[key]; ok {
		v, err := strconv.Atoi(value)
		if err == nil {
			return uint32(v)
		}
	}
	return uint32(defaultValue)
}
