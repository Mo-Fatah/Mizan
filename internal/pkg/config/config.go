package config

import (
	"io"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Services []Service `yaml:"services"`
	Strategy string    `yaml:"strategy"`
	// Ports to which Mizan will listen on
	// TODO (Mo-Fatah): Should deal with distributed ports across multiple nodes
	Ports []int `yaml:"ports"`
}

type Service struct {
	Name     string    `yaml:"name"`
	Matcher  string    `yaml:"matcher"`
	Replicas []Replica `yaml:"replicas"`
}

type Replica struct {
	Url      string            `yaml:"url"`
	MetaData map[string]string `yaml:"metadata"`
}

func LoadConfig(reader io.Reader) (*Config, error) {
	buf, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	config := Config{}

	if err = yaml.Unmarshal(buf, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
