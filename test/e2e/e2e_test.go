package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/Mo-Fatah/mizan/internal/mizan"
	"github.com/Mo-Fatah/mizan/internal/pkg/config"
	testservice "github.com/Mo-Fatah/mizan/test/testutil/testservices"
	"github.com/stretchr/testify/assert"
)

var (
	replicas    = 8
	mizanServer *mizan.Mizan
	dsg         *testservice.DummyServiceGen
	configYaml  = `
strategy: "RoundRobin"
ports:
  - 8080
  - 8081
  - 8082
services: 
  - matcher: "/api/v1"
    name: "test service"
    replicas:
      - "http://localhost:9090"
      - "http://localhost:9091"
      - "http://localhost:9092"
`
)

// A very simple/sloppy test to check if the proxy is working as expected
func TestE2E(t *testing.T) {
	envSetup(t)
	for i := 0; i < 10; i++ {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", (i%3)+8080, "/api/v1"))
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)
		body := make([]byte, 100)
		resp.Body.Read(body)
		assert.True(t, strings.Contains(string(body), "All is Good from dummy service on port"))
	}
	tearDown(t)
}

func envSetup(t *testing.T) {
	config, err := config.LoadConfig(strings.NewReader(configYaml))
	assert.NoError(t, err)

	mizanServer = mizan.NewMizan(config)
	go mizanServer.Start()

	dsg = testservice.NewDummyServiceGen(replicas)
	dsg.Start()
}

func tearDown(t *testing.T) {
	mizanServer.ShutDown()
	dsg.Stop()
}
