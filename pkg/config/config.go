// Copyright 2021 Rei Shimizu

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"

	"gopkg.in/yaml.v2"
)

type Veth struct {
	Left  string `yaml:"left"`
	Right string `yaml:"right"`
}

type Device struct {
	Name string `yaml:"name"`
	Cidr string `yaml:"cidr"`
}

type Namespace struct {
	Name   string   `yaml:"name"`
	Device []Device `yaml:"device"`
}

type Config struct {
	Veth      []Veth      `yaml:"veth"`
	Namespace []Namespace `yaml:"namespace"`
}

func ParseConfig(bytes []byte) (*Config, error) {
	cfg := Config{}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %s", err)
	}
	return &cfg, nil
}
