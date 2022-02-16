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
	"net"
	"strings"

	"github.com/Shikugawa/ayame/pkg/config"
	log "github.com/sirupsen/logrus"
	"go.uber.org/multierr"
)

type Namespace struct {
	Name                  string                         `json:"name"`
	Active                bool                           `json:"is_active"`
	DeviceConfig          []config.NamespaceDeviceConfig `json:"device_config"`
	ConfiguredDeviceNames []string                       `json:"configured_device_names"`
}

func InitNamespace(config *config.NamespaceConfig) (*Namespace, error) {
	ns := &Namespace{
		Name:         config.Name,
		DeviceConfig: config.Devices,
		Active:       false,
	}

	if err := RunIpNetnsAdd(config.Name); err != nil {
		return nil, err
	}

	log.Infof("succeeded to create ns %s\n", config.Name)
	ns.Active = true
	return ns, nil
}

func (n *Namespace) Destroy() error {
	if !n.Active {
		return fmt.Errorf("%s is already inactive\n", n.Name)
	}

	if err := RunIpNetnsDelete(n.Name); err != nil {
		return err
	}

	log.Infof("succeeded to delete ns %s\n", n.Name)
	return nil
}

func (n Namespace) Attach(veth *Veth) error {
	if veth.Attached {
		return fmt.Errorf("device %s is already attached", veth.Name)
	}

	var configuredName string
	for _, config := range n.DeviceConfig {
		if !strings.HasPrefix(veth.Name, config.Name) {
			continue
		}

		_, _, err := net.ParseCIDR(config.Cidr)
		if err != nil {
			log.Warnf("failed to parse CIDR %s in namespace %s device %s: %s\n", config.Cidr, n.Name, config.Name, err)
			continue
		}

		if err := RunIpLinkSetNamespaces(veth.Name, n.Name); err != nil {
			log.Warnf("failed to set device %s in namespace %s: %s", config.Name, n.Name, err)
			continue
		}

		if err := RunAssignCidrToNamespaces(veth.Name, n.Name, config.Cidr); err != nil {
			log.Warnf("failed to assign CIDR %s to ns %s on %s", config.Cidr, n.Name, veth.Name)
			continue
		}

		log.Infof("succeeded to attach CIDR %s to dev %s on ns %s\n", config.Cidr, veth.Name, n.Name)

		veth.Attached = true
		configuredName = veth.Name
		break
	}

	if configuredName == "" {
		return fmt.Errorf("no device configurations find, matches to %s", veth.Name)
	}

	n.ConfiguredDeviceNames = append(n.ConfiguredDeviceNames, configuredName)

	return nil
}

func InitNamespaces(conf []config.NamespaceConfig, links []DirectLink) ([]Namespace, error) {
	var namespaces []Namespace
	netLinks := make(map[string][]int)

	for _, c := range conf {
		ns, err := InitNamespace(&c)
		if err != nil {
			return namespaces, err
		}

		namespaces = append(namespaces, *ns)

		for _, device := range c.Devices {
			if _, ok := netLinks[device.Name]; !ok {
				netLinks[device.Name] = []int{}
			}
			netLinks[device.Name] = append(netLinks[device.Name], len(namespaces)-1)
		}
	}

	// Configure netlinks
	findValidLinkIndex := func(name string) int {
		fmt.Println(name)
		for i, link := range links {
			fmt.Println(link.Name)
			if name == link.Name {
				return i
			}
		}
		return -1
	}

	for linkName, idxs := range netLinks {
		if len(idxs) == 1 {
			log.Warnf("%s have only 1 link in %s\n", linkName, namespaces[idxs[0]].Name)
			continue
		}

		if len(idxs) > 2 {
			log.Warnf("%s has over 3 links despite it is not supported", linkName)
			continue
		}

		linkIdx := findValidLinkIndex(linkName)
		if linkIdx == -1 {
			log.Warnf("can't find device %s in configured links", linkName)
			continue
		}

		targetLink := links[linkIdx]
		if err := targetLink.CreateLink(namespaces[idxs[0]], namespaces[idxs[1]]); err != nil {
			return namespaces, fmt.Errorf("failed to create links %s: %s", linkName, err.Error())
		}
	}

	return namespaces, nil
}

func CleanupNamespaces(nss []Namespace) error {
	var allerr error
	for _, n := range nss {
		if err := n.Destroy(); err != nil {
			allerr = multierr.Append(allerr, err)
		}
	}
	return allerr
}
