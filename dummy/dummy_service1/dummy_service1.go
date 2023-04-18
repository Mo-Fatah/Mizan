package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var (
	port = flag.Int("port", 8081, "Port to listen on")
)

type DummyServer struct{}

func (ds *DummyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("All is Good from server %d\n", *port)))
}

func main() {
	flag.Parse()
	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), &DummyServer{}); err != nil {
		log.Fatal(err)
	}
}
