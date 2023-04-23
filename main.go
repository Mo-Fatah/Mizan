package main

import (
	"flag"
	"os"

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
	mizan.Start()
}
