package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/Mo-Fatah/mizan/internal/mizan"
)

var (
	configFile = flag.String("config-path", "", "Path to config file")
)

func main() {
	flag.Parse()

	// check if the config file exists
	_, err := os.Stat(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	mizan := mizan.NewMizan(*configFile)

	// handle interrupts and gracefully shutdown the server
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-interrupt
		log.Info("Received interrupt signal, shutting down...")
		mizan.ShutDown()
	}()

	mizan.Start()
	wg.Wait()
}
