package net

import (
	"fmt"
	"github.com/Shikugawa/netb/pkg/config"
	"log"
	"os/exec"
)

type VethPair struct {
	Host string
	Target string
}

type ActiveVeths struct {
	VethPairs []VethPair
}

func ActivateVeth(conf []config.Veth) (*ActiveVeths, error) {
	aveths := ActiveVeths{}

	for _, c := range conf {
		if aveths.Contains(c) {
			InactivateVeths(&aveths)
			return nil, fmt.Errorf("config has duplicated veth name %s", c.Name)
		}
		pair := VethPair{}
		pair.Host = c.Name
		pair.Target = c.Name + "-target"
		cmd := exec.Command("ip", "link", "add", "name", pair.Host, "type", "veth", "peer", pair.Target)
		if err := cmd.Run(); err != nil {
			InactivateVeths(&aveths)
			return nil, fmt.Errorf("failed to create veth name %s: %s", pair.Host, err)
		}
		log.Printf("succeeded to create veth name %s", pair.Host)
		aveths.VethPairs = append(aveths.VethPairs, pair)
	}

	return &aveths, nil
}

func (a *ActiveVeths) Contains(conf config.Veth) bool {
	for _, vethPair := range a.VethPairs {
		if conf.Name == vethPair.Host {
			return true
		}
	}
	return false
}

func InactivateVeths(veths *ActiveVeths) {
	for _, vethPair := range veths.VethPairs {
		cmd := exec.Command("ip", "link", "delete", vethPair.Host)
		if err := cmd.Run(); err != nil {
			log.Printf("failed to delete ns %s", vethPair.Host)
			continue
		}
		log.Printf("succeeded to delete %s", vethPair.Host)
	}
}
