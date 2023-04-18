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
    replicas:
      - localhost:8081
      - localhost:8082
strategy: RoundRobin
`)
	expected := &Config{
		Services: []Service{
			{
				Name:     "test service",
				Replicas: []string{"localhost:8081", "localhost:8082"},
			},
		},
		Strategy: "RoundRobin",
	}
	result, err := LoadConfig(yamlReader)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
