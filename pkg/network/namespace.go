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

	"github.com/Shikugawa/ayame/pkg/config"
	"go.uber.org/multierr"
)

type Namespace struct {
	Name                  string                         `json:"name"`
	Active                bool                           `json:"is_active"`
	deviceConfig          []config.NamespaceDeviceConfig `json:"device_config"`
	configuredDeviceNames []string                       `json:"configured_device_names"`
}

func InitNamespace(config *config.NamespaceConfig, verbose bool) (*Namespace, error) {
	ns := &Namespace{
		Name:         config.Name,
		deviceConfig: config.Devices,
		Active:       false,
	}

	if err := RunIpNetnsAdd(config.Name, verbose); err != nil {
		return nil, err
	}

	log.Printf("succeeded to create ns %s", config.Name)
	ns.Active = true
	return ns, nil
}

func (n *Namespace) Destroy(verbose bool) error {
	if !n.Active {
		return fmt.Errorf("%s is already inactive\n", n.Name)
	}

	if err := RunIpNetnsDelete(n.Name, verbose); err != nil {
		return err
	}

	log.Printf("succeeded to delete ns %s", n.Name)
	return nil
}

func (n Namespace) Attach(veth *Veth, verbose bool) error {
	if veth.Attached {
		return fmt.Errorf("device %s is already attached", veth.Name)
	}

	var configuredName string
	for _, config := range n.deviceConfig {
		if veth.Name != config.Name {
			continue
		}

		_, _, err := net.ParseCIDR(config.Cidr)
		if err != nil {
			return fmt.Errorf("failed to parse CIDR %s: %s", config.Cidr, err)
		}
		if err := RunIpLinkSetNamespaces(veth.Name, n.Name, verbose); err != nil {
			return err
		}

		if err := RunAssignCidrToNamespaces(veth.Name, n.Name, config.Cidr, verbose); err != nil {
			return fmt.Errorf("failed to assign CIDR %s to ns %s on %s", config.Cidr, n.Name, veth.Name)
		}

		if verbose {
			log.Printf("succeeded to attach CIDR %s to dev %s on ns %s",
				config.Cidr, veth.Name, n.Name)
		}

		veth.Attached = true
		configuredName = veth.Name
		break
	}

	if configuredName == "" {
		return fmt.Errorf("no device configurations find, matches to %s", veth.Name)
	}

	n.configuredDeviceNames = append(n.configuredDeviceNames, configuredName)

	return nil
}

func InitNamespaces(conf []config.NamespaceConfig, links []Link, verbose bool) ([]Namespace, error) {
	var namespaces []Namespace
	var netLinks map[string][]int

	for _, c := range conf {
		ns, err := InitNamespace(&c, verbose)
		if err != nil {
			return namespaces, err
		}

		namespaces = append(namespaces, *ns)

		for _, device := range c.Devices {
			if val, ok := netLinks[device.Name]; ok {
				val = append(val, len(namespaces)-1)
			}
		}
	}

	// Configure netlinks
	for k, idxs := range netLinks {
		for _, link := range links {
			if link.Name() == k {
				if len(idxs) == 1 {
					return namespaces, fmt.Errorf("failed to link namespaces; %s only have 1 link\n", link.Name())
				}

				if len(idxs) > 2 {
					return namespaces, fmt.Errorf("> 3 links are not supported")
				}

				if err := link.CreateLink(namespaces[idxs[0]], namespaces[idxs[1]], verbose); err != nil {
					return namespaces, fmt.Errorf("failed to create links %s", link.Name())
				}
			}
		}
	}

	return namespaces, nil
}

func CleanupNamespaces(nss []Namespace, verbose bool) error {
	var allerr error
	for _, n := range nss {
		if err := n.Destroy(verbose); err != nil {
			allerr = multierr.Append(allerr, err)
		}
	}
	return allerr
}
