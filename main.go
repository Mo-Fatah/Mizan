package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/Mo-Fatah/Mizan/pkg/config"
)

var (
	port       = flag.String("port", "8080", "Port to listen on")
	configFile = flag.String("config-path", "", "Path to config file")
)

type Mizan struct {
	config  *config.Config
	servers config.ServersList
}

func NewMizan(conf *config.Config) *Mizan {
	servers := make([]*config.Server, 0)
	for _, service := range conf.Services {
		for _, replica := range service.Replicas {
			serverUrl, err := url.Parse(replica)
			if err != nil {
				log.Fatal(err)
			}
			servers = append(servers, &config.Server{
				Url:   *serverUrl,
				Proxy: httputil.NewSingleHostReverseProxy(serverUrl),
			})
		}
	}
	return &Mizan{config: conf, servers: config.ServersList{Servers: servers}}
}

func (m *Mizan) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("Received request from address %s", r.RemoteAddr)
	server := m.servers.Next()
	log.Infof("Proxying request to %s", server.Url.String())
	server.Proxy.ServeHTTP(w, r)
}

func main() {
	flag.Parse()

	file, err := os.Open(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	config, err := config.LoadConfig(file)
	if err != nil {
		log.Fatal(err)
	}

	mizan := NewMizan(config)
	server := http.Server{
		Addr:    fmt.Sprintf(":%s", *port),
		Handler: mizan,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
