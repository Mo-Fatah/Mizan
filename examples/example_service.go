package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	port = flag.Int("port", 9090, "Port to listen on")
)

type ExampleService struct{}

func (es *ExampleService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("All is Good from server %d\n", *port)))
}

func main() {
	flag.Parse()
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), &ExampleService{}); err != nil {
		log.Fatal(err)
	}
}
