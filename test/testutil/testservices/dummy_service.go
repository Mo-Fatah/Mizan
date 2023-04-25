package testservices

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

const (
	BASE_PORT = 9090
)

type Service interface {
	Run() error
}

type DummyService struct {
	Ch   chan struct{}
	Port int
}

func (ds *DummyService) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(200)
	resp.Write([]byte(fmt.Sprintf("OK from %d", ds.Port)))
}

func (ds *DummyService) Run() error {
	// either ListenAndServe or Shutdown if received a close signal
	server := http.Server{
		Addr:    fmt.Sprintf(":%d", ds.Port),
		Handler: ds,
	}

	go func() {
		<-ds.Ch
		server.Shutdown(context.TODO())
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

type DummyServiceGen struct {
	replicas int
	ch       chan struct{}
}

func NewDummyServiceGen(replicas int) *DummyServiceGen {
	return &DummyServiceGen{
		replicas: replicas,
		ch:       make(chan struct{}),
	}
}

func (dsg *DummyServiceGen) Stop() {
	close(dsg.ch)
}

func (dsg *DummyServiceGen) Start() {
	for i := 0; i < dsg.replicas; i++ {
		ds := &DummyService{
			Ch:   dsg.ch,
			Port: BASE_PORT + i,
		}
		go ds.Run()
	}
}
