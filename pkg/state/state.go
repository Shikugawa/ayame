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

package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/Shikugawa/ayame/pkg/config"
	"github.com/Shikugawa/ayame/pkg/network"
)

type State struct {
	VethPairs  []*network.VethPair  `json:"vethpairs"`
	Namespaces []*network.Namespace `json:"namespaces"`
}

var statePath = os.Getenv("HOME") + "/.ayame"

const stateFileName = "state.json"

func InitAll(cfg *config.Config, verbose bool) (*State, error) {
	pairs, err := network.InitVethPairs(cfg.Veth, verbose)
	if err != nil {
		network.CleanupAllVethPairs(&pairs, verbose)
		return nil, err
	}

	ns, err := network.InitNamespaces(cfg.Namespace, &pairs, verbose)
	if err != nil {
		network.CleanupAllVethPairs(&pairs, verbose)
		network.CleanupAllNamespaces(&ns, verbose)
		return nil, err
	}

	return &State{VethPairs: pairs, Namespaces: ns}, nil
}

func LoadStateFromFile() (*State, error) {
	if _, err := os.Stat(statePath + "/" + stateFileName); os.IsNotExist(err) {
		return nil, fmt.Errorf("no saved state")
	}

	b, err := ioutil.ReadFile(statePath + "/" + stateFileName)
	if err != nil {
		return nil, err
	}

	var state State
	if err := json.Unmarshal(b, &state); err != nil {
		return nil, err
	}

	return &state, nil
}

func (s *State) SaveState() error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		if err := os.MkdirAll(statePath, 0644); err != nil {
			return fmt.Errorf("failed to create %s", statePath)
		}
	}

	if err := ioutil.WriteFile(statePath+"/"+stateFileName, b, 0644); err != nil {
		return err
	}
	return nil
}

func (s *State) DisposeResources(verbose bool) error {
	if err := network.CleanupAllVethPairs(&s.VethPairs, verbose); err != nil {
		return err
	}
	if err := network.CleanupAllNamespaces(&s.Namespaces, verbose); err != nil {
		return err
	}
	if err := os.Remove(statePath + "/" + stateFileName); err != nil {
		return err
	}
	return nil
}

func (s *State) DumpAll() (string, error) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
