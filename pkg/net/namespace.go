package net

import (
	"fmt"
	"github.com/Shikugawa/netb/pkg/config"
	"log"
	"os/exec"
)

type ActiveNS struct {
	Names []string
}

func ActivateNS(conf []config.Namespace) (*ActiveNS, error) {
	ans := ActiveNS{}

	for _, c := range conf {
		if ans.Contains(c) {
			InactivateNS(&ans)
			return nil, fmt.Errorf("config has duplicated ns name %s", c.Name)
		}
		cmd := exec.Command("ip", "netns", "add", c.Name)
		if err := cmd.Run(); err != nil {
			InactivateNS(&ans)
			return nil, fmt.Errorf("failed to create ns %s", c.Name)
		}
		log.Printf("succeeded to create ns %s", c.Name)
		ans.Names = append(ans.Names, c.Name)
	}

	return &ans, nil
}

func (a *ActiveNS) Contains(conf config.Namespace) bool {
	for _, name := range a.Names {
		if conf.Name == name {
			return true
		}
	}
	return false
}

func InactivateNS(ans *ActiveNS) {
	for _, name := range ans.Names {
		cmd := exec.Command("ip", "netns", "delete", name)
		if err := cmd.Run(); err != nil {
			log.Printf("failed to delete ns %s", name)
			continue
		}
		log.Printf("succeeded to delete ns %s", name)
	}
}
