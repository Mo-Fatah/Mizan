package main

import (
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"

	"github.com/Mo-Fatah/mizan/internal/mizan"
	"github.com/Mo-Fatah/mizan/internal/pkg/config"
)

var (
	configFile = flag.String("config-path", "", "Path to config file")
)

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

	mizan := mizan.NewMizan(config)

	// handle interrupts and gracefully shutdown the server
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-interrupt
		log.Info("Shutting down...")
		mizan.ShutDown()
	}()

	mizan.Start()
	wg.Wait()
}
