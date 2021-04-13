package net

import (
	"fmt"
	"github.com/Shikugawa/netb/pkg/config"
	"log"
	"net"
	"os/exec"
)

type Veth struct {
	Name     string `yaml:"name""`
	Attached bool   `yaml:"attached"`
}

type AttachedDevice struct {
	Dev  *Veth      `yaml:"device"`
	Cidr *net.IPNet `yaml:cidr`
}

func (v *Veth) Attach(ns *Namespace, cidr *net.IPNet) (*AttachedDevice, error) {
	if v.Attached {
		return nil, fmt.Errorf("%s is already attached", v.Name)
	}

	cmd := exec.Command("ip", "link", "set", v.Name, "netns", ns.Name)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to attach device %s to ns %s", v.Name, ns.Name)
	}

	cmd = exec.Command("ip", "netns", "exec", ns.Name, "ip", "addr", "add", cidr.String(), "dev", v.Name)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to assign CIDR %s to ns %s on %s", cidr.String(), ns.Name, v.Name)
	}

	v.Attached = true
	return &AttachedDevice{Dev: v, Cidr: cidr}, nil
}

type VethPair struct {
	Left   *Veth `yaml:"veth_left"`
	Right  *Veth `yaml:"veth_right"`
	Active bool  `yaml:"is_active"`
}

func CreateVethPair(conf config.Veth) (*VethPair, error) {
	pair := VethPair{
		Left:   &Veth{Name: conf.Left, Attached: false},
		Right:  &Veth{Name: conf.Right, Attached: false},
		Active: false,
	}

	if err := pair.Create(); err != nil {
		return &pair, err
	}

	return &pair, nil
}

func (v *VethPair) Create() error {
	if v.Active {
		return fmt.Errorf("%s@%s is already created", v.Left.Name, v.Right.Name)
	}

	cmd := exec.Command("ip", "link", "add", "name", v.Left.Name, "type", "veth", "peer", v.Right.Name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create veth name %s@%s: %s", v.Left.Name, v.Right.Name, err)
	}

	v.Active = true
	log.Printf("succeeded to create %s@%s", v.Left.Name, v.Right.Name)

	return nil
}

func (v *VethPair) Destroy() error {
	if !v.Active {
		return fmt.Errorf("%s@%s doesn't exist", v.Left.Name, v.Right.Name)
	}

	var cmd *exec.Cmd
	if !v.Left.Attached {
		cmd = exec.Command("ip", "link", "delete", v.Left.Name)
	}

	if cmd == nil && !v.Right.Attached {
		cmd = exec.Command("ip", "link", "delete", v.Right.Name)
	}

	if cmd == nil {
		log.Printf("veth-pair %s@%s is invisible from host")
		return nil
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete device %s@%s: %s", v.Left.Name, v.Right.Name, err)
	}

	v.Active = false
	log.Printf("succeeded to delete %s@%s", v.Left.Name, v.Right.Name)

	return nil
}

type ActiveVethPairs struct {
	VethPairs []*VethPair
}

func InitVethPairs(conf []config.Veth) (*ActiveVethPairs, error) {
	activeVethPairs := ActiveVethPairs{}

	for _, c := range conf {
		vethPair, err := CreateVethPair(c)
		activeVethPairs.VethPairs = append(activeVethPairs.VethPairs, vethPair)

		if err != nil {
			return &activeVethPairs, err
		}
	}

	return &activeVethPairs, nil
}

func (a *ActiveVethPairs) Cleanup() {
	for _, v := range a.VethPairs {
		if err := v.Destroy(); err != nil {
			log.Println(err)
			continue
		}
	}
}
