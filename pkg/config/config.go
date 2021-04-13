package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
)

//namespace:
//- name: ns1
//  device:
//	- attach: veth1
//	  cidr: 192.168.100.10/24
//- name: ns2
//  device:
//	- attach: veth1-target
//    cidr: 192.168.100.20/24

type Veth struct {
	Left  string `yaml:left`
	Right string `yaml:right`
}

type Device struct {
	Name string `yaml:name`
	Cidr string `yaml:cidr`
}

type Namespace struct {
	Name   string   `yaml:name`
	Device []Device `yaml:device`
}

type Config struct {
	Veth      []Veth      `yaml:veth`
	Namespace []Namespace `yaml:namespace`
}

func ParseConfig(bytes []byte) (*Config, error) {
	cfg := Config{}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %s", err)
	}
	return &cfg, nil
}
