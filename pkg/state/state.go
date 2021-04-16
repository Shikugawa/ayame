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

func InitAll(cfg *config.Config) (*State, error) {
	pairs, err := network.InitVethPairs(cfg.Veth)
	if err != nil {
		network.CleanupAllVethPairs(&pairs)
		return nil, err
	}

	ns, err := network.InitNamespaces(cfg.Namespace, &pairs)
	if err != nil {
		network.CleanupAllVethPairs(&pairs)
		network.CleanupAllNamespaces(&ns)
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

func (s *State) DisposeResources() error {
	if err := network.CleanupAllVethPairs(&s.VethPairs); err != nil {
		return err
	}
	if err := network.CleanupAllNamespaces(&s.Namespaces); err != nil {
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
