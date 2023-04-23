package mizan

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/Mo-Fatah/mizan/internal/pkg/config"
	log "github.com/sirupsen/logrus"
)

type Mizan struct {
	// The configuration loaded from the config file
	// TODO (Mo-Fatah): Should add hot reload for config
	Config *config.Config
	// Servers is a map of service matcher to a list of servers/replicas
	ServerList map[string]*config.ServerList
	Ports      []int
	shutDown   chan bool
}

func NewMizan(conf *config.Config) *Mizan {
	shutdown := make(chan bool)
	servers := make(map[string]*config.ServerList)
	for _, service := range conf.Services {
		for _, replica := range service.Replicas {
			serverUrl, err := url.Parse(replica)
			if err != nil {
				log.Fatal(err)
			}
			server := config.Server{
				Url:   *serverUrl,
				Proxy: httputil.NewSingleHostReverseProxy(serverUrl),
			}

			if _, ok := servers[service.Matcher]; !ok {
				servers[service.Matcher] = &config.ServerList{}
			}
			servers[service.Matcher].Add(&server)
		}
	}
	return &Mizan{Config: conf, ServerList: servers, Ports: conf.Ports, shutDown: shutdown}
}

func (m *Mizan) Start() {
	var wg sync.WaitGroup
	for _, port := range m.Ports {
		go m.startServer(port, &wg)
		wg.Add(1)
	}
	wg.Wait()
}

func (m *Mizan) ShutDown() {
	close(m.shutDown)
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

func (m *Mizan) findService(path string) (*config.ServerList, error) {
	if _, ok := m.ServerList[path]; !ok {
		return nil, fmt.Errorf("couldn't find path %s", path)
	}
	return m.ServerList[path], nil
}

func (m *Mizan) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Infof("Received request from address %s", r.RemoteAddr)
	sl, err := m.findService(r.URL.Path)
	if err != nil {
		log.Error(err)
		return
	}
	server := sl.Next()

	log.Infof("Proxying request to %s", server.Url.String())
	server.Proxy.ServeHTTP(w, r)
}
