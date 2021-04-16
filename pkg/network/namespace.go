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

package network

import (
	"fmt"
	"log"
	"net"
	"os/exec"

	"github.com/Shikugawa/ayame/pkg/config"
	"go.uber.org/multierr"
)

type AttachedDevice struct {
	Dev  *Veth  `json:"device"`
	Cidr string `json:"cidr"`
}

type Namespace struct {
	Name    string            `json:"name"`
	Devices []*AttachedDevice `json:"devices"`
	Active  bool              `json:"is_active"`
}

func InitNamespace(conf config.Namespace, pairs *[]*VethPair) (*Namespace, error) {
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

		for _, pair := range *pairs {
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
					attachedDevice.Cidr, attachedDevice.Dev.Name, ns.Name)

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
					attachedDevice.Cidr, attachedDevice.Dev.Name, ns.Name)

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

func InitNamespaces(conf []config.Namespace, pairs *[]*VethPair) ([]*Namespace, error) {
	var activeNamespaces []*Namespace

	for _, c := range conf {
		ns, err := InitNamespace(c, pairs)
		activeNamespaces = append(activeNamespaces, ns)

		if err != nil {
			return activeNamespaces, err
		}
	}

	return activeNamespaces, nil
}

func CleanupAllNamespaces(nss *[]*Namespace) error {
	var allerr error
	for _, n := range *nss {
		if err := n.Destroy(); err != nil {
			allerr = multierr.Append(allerr, err)
		}
	}
	return allerr
}
