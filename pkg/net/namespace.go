package net

import (
	"fmt"
	"github.com/Shikugawa/netb/pkg/config"
	"log"
	"net"
	"os/exec"
)

type Namespace struct {
	Name    string            `yaml:"name"`
	Devices []*AttachedDevice `yaml:"devices"`
	Active  bool              `yaml:"is_active"`
}

func InitNamespace(conf config.Namespace, pairs *ActiveVethPairs) (*Namespace, error) {
	ns := Namespace{
		Name:   conf.Name,
		Active: false,
	}

	cmd := exec.Command("ip", "netns", "add", ns.Name)
	if err := cmd.Run(); err != nil {
		return &ns, fmt.Errorf("failed to create ns %s", ns.Name)
	}

	for _, dev := range conf.Device {
		var attachedDevice *AttachedDevice

		for _, pair := range pairs.VethPairs {
			if dev.Name == pair.Left.Name && !pair.Left.Attached {
				_, ipNet, err := net.ParseCIDR(dev.Cidr)
				if err != nil {
					return &ns, fmt.Errorf("failed to parse CIDR %s: %s", dev.Cidr, err)
				}

				attachedDevice, err = pair.Left.Attach(&ns, ipNet)
				if err != nil {
					return &ns, err
				}

				log.Printf("succeeded to attach CIDR %s to dev %s on ns %s",
					attachedDevice.Cidr.String(), attachedDevice.Dev.Name, ns.Name)

				break
			}

			if dev.Name == pair.Right.Name && !pair.Right.Attached {
				_, ipNet, err := net.ParseCIDR(dev.Cidr)
				if err != nil {
					return &ns, fmt.Errorf("failed to parse CIDR %s: %s", dev.Cidr, err)
				}

				attachedDevice, err = pair.Right.Attach(&ns, ipNet)
				if err != nil {
					return &ns, err
				}

				log.Printf("succeeded to attach CIDR %s to dev %s on ns %s",
					attachedDevice.Cidr.String(), attachedDevice.Dev.Name, ns.Name)

				break
			}
		}

		if attachedDevice == nil {
			return &ns, fmt.Errorf("device %s not found", dev.Name)
		}

		ns.Devices = append(ns.Devices, attachedDevice)
	}

	log.Printf("succeeded to create ns %s", ns.Name)

	ns.Active = true
	return &ns, nil
}

func (n *Namespace) Destroy() error {
	if !n.Active {
		log.Printf("%s is already inactive", n.Name)
		return nil
	}

	cmd := exec.Command("ip", "netns", "delete", n.Name)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete ns %s", n.Name)
	}

	for _, dev := range n.Devices {
		dev.Dev.Attached = false
	}

	log.Printf("succeeded to delete ns %s", n.Name)
	return nil
}

type ActiveNamespaces struct {
	namespaces []*Namespace `yaml:"namespaces"`
}

func InitNamespaces(conf []config.Namespace, pairs *ActiveVethPairs) (*ActiveNamespaces, error) {
	activeNamespaces := ActiveNamespaces{}

	for _, c := range conf {
		ns, err := InitNamespace(c, pairs)
		activeNamespaces.namespaces = append(activeNamespaces.namespaces, ns)

		if err != nil {
			return &activeNamespaces, err
		}
	}

	return &activeNamespaces, nil
}

func (a *ActiveNamespaces) Cleanup() {
	for _, n := range a.namespaces {
		if err := n.Destroy(); err != nil {
			log.Println(err)
			continue
		}
	}
}
