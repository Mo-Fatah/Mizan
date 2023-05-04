package e2e

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Mo-Fatah/mizan/internal/mizan"
	testservice "github.com/Mo-Fatah/mizan/test/testutil/testservices"
	"github.com/stretchr/testify/assert"
)

// TODO (Mo-Fatah): Add tests for the following:
// - Load balancing strategies
// - Hot reload
// - Health checking
// - Load balancing strategies with health checking

// TODO (Mo-Fatah): There are some refactoring needed here

var (
	defaultReplicas   = 3
	mizanServer       *mizan.Mizan
	dsg               *testservice.DummyServiceGen
	yamlPathRR        = "./testConfigs/rr.yml"
	yamlPathWRR       = "./testConfigs/wrr.yml"
	yamlPathHotReload = "./testConfigs/hotreload.yml"
)

// Round Robin should rotate on the servers in the order they are defined in the config
// So if we have 3 servers, the first request should go to the first server, the second to the second server and so on
func TestE2E_BasicRoundRobin(t *testing.T) {
	envSetup(defaultReplicas, yamlPathRR)
	defer tearDown()

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
	envSetup(defaultReplicas, yamlPathWRR)
	defer tearDown()

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

func TestE2E_WhenAServiceisDownRR(t *testing.T) {
	// The yaml file has 3 replicas, but we will start only 2 to simulate a down service
	envSetup(2, yamlPathRR)
	defer tearDown()

	ports := []int{9090, 9091}
	portsIndex := 0

	for i := 0; i < 5; i++ {
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

func TestE2E_HotReload(t *testing.T) {
	// start with Round Robin config
	copyFile(yamlPathRR, yamlPathHotReload)

	envSetup(defaultReplicas, yamlPathHotReload)
	defer tearDown()

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
	time.Sleep(4 * time.Second)
	// change the config to Weighted Round Robin
	copyFile(yamlPathWRR, yamlPathHotReload)

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

func envSetup(replicas int, yamlPath string) {
	dsg = testservice.NewDummyServiceGen(replicas)
	dsg.Start()
	for !dsg.IsReady() {
		continue
	}

	mizanServer = mizan.NewMizan(yamlPath)
	go mizanServer.Start()
	for !mizanServer.IsReady() {
		continue
	}
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Create the destination file for writing
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy the contents of the source file to the destination file
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// Flush any buffered data to the destination file
	err = dstFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func tearDown() {
	mizanServer.ShutDown()
	dsg.Stop()
}
