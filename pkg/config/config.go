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

type NamespaceDeviceConfig struct {
	Name string `yaml:"name"`
	Cidr string `yaml:"cidr"`
}

type NamespaceConfig struct {
	Name     string                  `yaml:"name"`
	Devices  []NamespaceDeviceConfig `yaml:"devices"`
	Commands []string                `yaml:"commands"`
}

type LinkMode string

const (
	ModeDirectLink = "direct_link"
	ModeBridge     = "bridge"
)

type LinkConfig struct {
	LinkMode LinkMode `yaml:"mode"`
	Name     string   `yaml:"name"`
}

type Config struct {
	Links      []*LinkConfig      `yaml:"links"`
	Namespaces []*NamespaceConfig `yaml:"namespaces"`
}

func ParseConfig(bytes []byte) (*Config, error) {
	cfg := Config{}
	if err := yaml.Unmarshal(bytes, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %s", err)
	}

	if err := ValidateLinkConfigs(cfg.Links); err != nil {
		return nil, err
	}
	if err := ValidateNamespace(cfg.Namespaces, cfg.Links); err != nil {
		return nil, err
	}

	return &cfg, nil
}
