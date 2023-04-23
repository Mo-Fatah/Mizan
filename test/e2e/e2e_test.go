package e2e

import (
	"net/http"
	"strings"
	"testing"

	dummyservice "github.com/Mo-Fatah/Mizan/test/testutil/dummy_service"
	"github.com/stretchr/testify/assert"
)

var (
	replicas = 8
	dsg      *dummyservice.DummyServiceGen
)

// A very simple/sloppy test to check if the proxy is working as expected
func TestE2E(t *testing.T) {
	envSetup(t)
	for i := 0; i < 10; i++ {
		resp, err := http.Get("http://localhost:8080/api/v1")
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, 200)
		body := make([]byte, 100)
		resp.Body.Read(body)
		assert.True(t, strings.Contains(string(body), "All is Good from server"))
	}
	tearDown(t)
}

func envSetup(t *testing.T) {
	//mizanServer := mizan.NewMizan()

	dsg = dummyservice.NewDummyServiceGen(replicas)
	dsg.Start()
}

func tearDown(t *testing.T) {
	dsg.Stop()
}
