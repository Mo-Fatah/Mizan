package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {

	yamlReader := strings.NewReader(`
services:
  - name: test service
    matcher: "api/v1"
    replicas:
      - localhost:8081
      - localhost:8082
ports:
  - 8080
  - 8081
  - 8082
strategy: RoundRobin
`)
	expected := &Config{
		Services: []Service{
			{
				Name:     "test service",
				Replicas: []string{"localhost:8081", "localhost:8082"},
				Matcher:  "api/v1",
			},
		},
		Strategy: "RoundRobin",
		Ports:    []int{8080, 8081, 8082},
	}
	result, err := LoadConfig(yamlReader)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
