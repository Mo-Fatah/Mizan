package e2e

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/Mo-Fatah/mizan/internal/mizan"
	testservice "github.com/Mo-Fatah/mizan/test/testutil/testservices"
	"github.com/stretchr/testify/assert"
)

// TODO (Mo-Fatah): Add tests for the following:
// - Load balancing strategies
// - Hot reload
// - Health checking
// - Load balancing strategies with health checking

var (
	replicas    = 3
	mizanServer *mizan.Mizan
	dsg         *testservice.DummyServiceGen
	yamlPathRR  = "./testConfigs/rr.yml"
	yamlPathWRR = "./testConfigs/wrr.yml"
)

// Round Robin should rotate on the servers in the order they are defined in the config
// So if we have 3 servers, the first request should go to the first server, the second to the second server and so on
func TestE2E_BasicRoundRobin(t *testing.T) {
	envSetup(t, yamlPathRR)
	defer tearDown(t)

	ports := []int{9090, 9091, 9092}
	portsIndex := 0
	for i := 0; i < 10; i++ {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", (i%3)+8080, "/api/v1"))
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)

		body := make([]byte, 100)
		resp.Body.Read(body)
		// Remove null bytes from the body
		bodyStr := strings.Trim(string(body), "\x00")
		assert.True(t, strings.Contains(bodyStr, "OK"))

		servicePort, err := strconv.Atoi(strings.Split(bodyStr, " ")[2])
		assert.NoError(t, err)
		assert.Equal(t, servicePort, ports[portsIndex])
		portsIndex = (portsIndex + 1) % len(ports)
	}
}

// Weighted Round Robin should rotate on the servers considering their weights
// So if we have 2 servers with weights 2 and 1, the first 2 requests should go to the first server and the third to the second server
func TestE2E_BasicWeightedRoundRobin(t *testing.T) {
	envSetup(t, yamlPathWRR)
	defer tearDown(t)

	portsFreq := map[int]int{
		9090: 0,
		9091: 0,
		9092: 0,
	}

	for i := 0; i < 10; i++ {
		resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", (i%3)+8080, "/api/v1"))
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)

		body := make([]byte, 100)
		resp.Body.Read(body)
		// Remove null bytes from the body
		bodyStr := strings.Trim(string(body), "\x00")
		assert.True(t, strings.Contains(bodyStr, "OK"))

		servicePort, err := strconv.Atoi(strings.Split(bodyStr, " ")[2])
		assert.NoError(t, err)
		portsFreq[servicePort]++
	}
	assert.Equal(t, portsFreq[9090], 6)
	assert.Equal(t, portsFreq[9091], 3)
	assert.Equal(t, portsFreq[9092], 1)
}

func envSetup(t *testing.T, yamlPath string) {
	dsg = testservice.NewDummyServiceGen(replicas)
	dsg.Start()

	mizanServer = mizan.NewMizan(yamlPath)
	go mizanServer.Start()
	for !mizanServer.IsReady() {
		continue
	}
}

func tearDown(t *testing.T) {
	mizanServer.ShutDown()
	dsg.Stop()
}
