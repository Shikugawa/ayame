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
	log "github.com/sirupsen/logrus"
)

type State struct {
	DirectLinks []*network.DirectLink `json:"direct_links"`
	Bridges     []*network.Bridge     `json:"bridges"`
	Namespaces  []*network.Namespace  `json:"namespaces"`
}

var statePath = os.Getenv("HOME") + "/.ayame"

const stateFileName = "state.json"

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

	log.Info("succeeded to save state")

	return nil
}

func (s *State) DumpAll() (string, error) {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ResourcesSaved() bool {
	if _, err := os.Stat(statePath + "/" + stateFileName); os.IsNotExist(err) {
		return false
	}
	return true
}

func LoadResources() *State {
	if !ResourcesSaved() {
		return nil
	}

	b, err := ioutil.ReadFile(statePath + "/" + stateFileName)
	if err != nil {
		return nil
	}

	return LoadStateFromBytes(b)
}

func LoadStateFromBytes(bytes []byte) *State {
	var state State
	if err := json.Unmarshal(bytes, &state); err != nil {
		return nil
	}

	return &state
}

func DisposeResources() error {
	state := LoadResources()
	if state == nil {
		return fmt.Errorf("resources have already cleared.")
	}

	if err := network.CleanupDirectLinks(state.DirectLinks, false); err != nil {
		return err
	}
	if err := network.CleanupBridges(state.Bridges, false); err != nil {
		return err
	}
	if err := network.CleanupNamespaces(state.Namespaces, false); err != nil {
		return err
	}

	if err := os.Remove(statePath + "/" + stateFileName); err != nil {
		return err
	}
	return nil
}

// TODO: consider error handling
func InitResources(cfg *config.Config, dryrun bool) (*State, error) {
	state := LoadResources()
	if state != nil {
		return nil, fmt.Errorf("resources have already existed.")
	}

	state = &State{Namespaces: nil, DirectLinks: nil, Bridges: nil}

	// Init links
	dlinks, err := network.InitDirectLinks(cfg.Links, dryrun)
	if err != nil {
		return nil, err
	}

	// Init Bridges
	brs, err := network.InitBridges(cfg.Links, dryrun)
	if err != nil {
		network.CleanupDirectLinks(dlinks, dryrun)
		return nil, err
	}

	// Init namespaces
	ns, err := network.InitNamespaces(cfg.Namespaces, dryrun)
	if err != nil {
		network.CleanupDirectLinks(dlinks, dryrun)
		network.CleanupBridges(brs, dryrun)
		return nil, err
	}

	// Link (Direct Links) Namespaces
	if err := network.InitNamespacesLinks(ns, dlinks, dryrun); err != nil {
		network.CleanupDirectLinks(dlinks, dryrun)
		network.CleanupBridges(brs, dryrun)
		return nil, err
	}

	state.DirectLinks = dlinks
	state.Bridges = brs
	state.Namespaces = ns

	return state, nil
}
