package config

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_RoundRobin(t *testing.T) {

	yamlReader := strings.NewReader(`
strategy: "RoundRobin"
ports:
  - 8080
  - 8081
  - 8082
services: 
  - matcher: "/api/v1"
    name: "test service"
    replicas:
      - url: "http://localhost:9090"
      - url: "http://localhost:9091"
      - url: "http://localhost:9092"
`)
	expected := &Config{
		Services: []Service{
			{
				Name: "test service",
				Replicas: []Replica{
					{
						Url: "http://localhost:9090",
					},
					{
						Url: "http://localhost:9091",
					},
					{
						Url: "http://localhost:9092",
					},
				},
				Matcher: "/api/v1",
			},
		},
		Strategy: "RoundRobin",
		Ports:    []int{8080, 8081, 8082},
	}
	result, err := LoadConfig(yamlReader)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestLoadConfig_WeightedRoundRobin(t *testing.T) {

	yamlReader := strings.NewReader(`
strategy: "WeightedRoundRobin"
ports:
  - 8080
  - 8081
  - 8082
services: 
  - matcher: "/api/v1"
    name: "test service"
    replicas:
      - url: "http://localhost:9090"
        metadata:
          weight: 1
      - url: "http://localhost:9091"
        metadata:
          weight: 1
      - url: "http://localhost:9092"
        metadata:
          weight: 1
`)
	expected := &Config{
		Services: []Service{
			{
				Name: "test service",
				Replicas: []Replica{
					{
						Url:      "http://localhost:9090",
						MetaData: map[string]string{"weight": "1"},
					},
					{
						Url:      "http://localhost:9091",
						MetaData: map[string]string{"weight": "1"},
					},
					{
						Url:      "http://localhost:9092",
						MetaData: map[string]string{"weight": "1"},
					},
				},
				Matcher: "/api/v1",
			},
		},
		Strategy: "WeightedRoundRobin",
		Ports:    []int{8080, 8081, 8082},
	}
	result, err := LoadConfig(yamlReader)

	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
