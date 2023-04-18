package config

import (
	"io"

	"gopkg.in/yaml.v3"
)

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
