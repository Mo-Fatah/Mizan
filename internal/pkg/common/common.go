package common

import (
	"net/http/httputil"
	"net/url"
)

type Server struct {
	Url      url.URL
	Proxy    *httputil.ReverseProxy
	MetaData map[string]string
}
