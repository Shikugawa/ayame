package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

type Veth struct {
	Name string `yaml:name`
}

type Device struct {
	Attach string `yaml:attach`
	Cidr   string `yaml:cidr`
}

type Namespace struct {
	Name   string   `yaml:name`
	Device []Device `yaml:device`
}

type Config struct {
	Veth []Veth           `yaml:veth`
	Namespace []Namespace `yaml:namespace`
}

func ParseConfig(bytes []byte) (*Config, error) {
	cfg := Config{}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %s", err)
	}
	return &cfg, nil
}
