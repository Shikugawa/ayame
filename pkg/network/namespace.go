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
	"os/exec"
	"regexp"
	"strings"

	"github.com/Shikugawa/ayame/pkg/config"
	log "github.com/sirupsen/logrus"
	"go.uber.org/multierr"
)

type RegisteredDeviceConfig struct {
	config.NamespaceDeviceConfig `json:"device_config"`
	AttachedVeth                 string `json:"attached_veth"`
}

type Namespace struct {
	Name                   string                   `json:"name"`
	RegisteredDeviceConfig []RegisteredDeviceConfig `json:"registered_device_config"`
}

func InitNamespace(config *config.NamespaceConfig, dryrun bool) (*Namespace, error) {
	var configs []RegisteredDeviceConfig
	for _, c := range config.Devices {
		tmp := RegisteredDeviceConfig{
			AttachedVeth: "",
		}
		tmp.NamespaceDeviceConfig = c
		configs = append(configs, tmp)
	}

	ns := &Namespace{
		Name:                   config.Name,
		RegisteredDeviceConfig: configs,
	}

	if err := RunIpNetnsAdd(config.Name, dryrun); err != nil {
		return nil, err
	}

	log.Infof("succeeded to create ns %s\n", config.Name)
	return ns, nil
}

func (n *Namespace) Destroy(dryrun bool) error {
	// namespaces don't exist anymore after host shutted down. Here ignores the closed netns.
	if !CheckIpNetnsExists(n.Name, dryrun) {
		log.Infof("%s doesn't exist\n", n.Name)
		return nil
	}

	if err := RunIpNetnsDelete(n.Name, dryrun); err != nil {
		return err
	}

	log.Infof("succeeded to delete ns %s\n", n.Name)
	return nil
}

func (n *Namespace) Attach(veth *Veth, dryrun bool) error {
	if veth.Attached {
		return fmt.Errorf("device %s is already attached", veth.Name)
	}

	targetCfgIdx := -1
	for idx, config := range n.RegisteredDeviceConfig {
		if !strings.HasPrefix(veth.Name, config.Name) {
			continue
		}

		if len(config.AttachedVeth) != 0 {
			return fmt.Errorf("device %s has been attached to namexpace %s", config.NamespaceDeviceConfig.Name, n.Name)
		}

		targetCfgIdx = idx
		break
	}

	if targetCfgIdx == -1 {
		return fmt.Errorf("proposed device %s can't be attached to %s", veth.Name, n.Name)
	}

	targetCfg := n.RegisteredDeviceConfig[targetCfgIdx]

	_, _, err := net.ParseCIDR(targetCfg.Cidr)
	if err != nil {
		return fmt.Errorf("failed to parse CIDR %s in namespace %s device %s: %s\n",
			targetCfg.Cidr, n.Name, targetCfg.Name, err)
	}

	if err := RunIpLinkSetNamespaces(veth.Name, n.Name, dryrun); err != nil {
		return fmt.Errorf("failed to set device %s in namespace %s: %s", targetCfg.Name, n.Name, err)
	}

	if err := RunAssignCidrToNamespaces(veth.Name, n.Name, targetCfg.Cidr, dryrun); err != nil {
		return fmt.Errorf("failed to assign CIDR %s to ns %s on %s", targetCfg.Cidr, n.Name, veth.Name)
	}

	log.Infof("succeeded to attach CIDR %s to dev %s on ns %s\n", targetCfg.Cidr, veth.Name, n.Name)

	n.RegisteredDeviceConfig[targetCfgIdx].AttachedVeth = veth.Name
	veth.Attached = true
	return nil
}

func (n *Namespace) RunCommands(commands []string, dryrun bool) {
	for _, command := range commands {
		netnsCmd, err := n.buildCommand(command)
		if err != nil {
			log.Warn(err.Error())
			continue
		}
		name := netnsCmd[0]
		rest := netnsCmd[1:]
		cmd := exec.Command(name, rest...)
		log.Infof("execute %s", cmd.String())

		if dryrun {
			continue
		}
		res, err := cmd.Output()
		if err != nil {
			log.Warn(err.Error())
			continue
		}

		log.Infof("\n%s", string(res))
	}
}

func (n *Namespace) buildCommand(command string) ([]string, error) {
	splited := strings.Split(command, " ")
	if len(splited) == 0 {
		return nil, fmt.Errorf("malformed command: %s", command)
	}

	netnsCmd := []string{}
	netnsCmd = append(netnsCmd, "ip")
	netnsCmd = append(netnsCmd, "netns")
	netnsCmd = append(netnsCmd, "exec")
	netnsCmd = append(netnsCmd, n.Name)

	re := regexp.MustCompile(`[0-9a-zA-Z]*`)

	for _, s := range splited {
		newCmd := s
		if res, err := regexp.MatchString(`\$\(.*\)`, newCmd); res && err == nil {
			devName := re.FindString(newCmd)
			for _, dev := range n.RegisteredDeviceConfig {
				if len(dev.AttachedVeth) != 0 && strings.HasPrefix(dev.Name, devName) {
					newCmd = dev.AttachedVeth
				}
			}
		}

		netnsCmd = append(netnsCmd, newCmd)
	}

	return netnsCmd, nil
}

func InitNamespaces(conf []*config.NamespaceConfig, dryrun bool) ([]*Namespace, error) {
	var namespaces []*Namespace

	// Setup namespaces
	for _, c := range conf {
		ns, err := InitNamespace(c, dryrun)
		if err != nil {
			return nil, err
		}

		namespaces = append(namespaces, ns)
	}

	return namespaces, nil
}

func InitNamespacesLinks(namespaces []*Namespace, links map[string]*DirectLink, dryrun bool) error {
	netLinks := make(map[string][]int)

	for i, ns := range namespaces {
		for _, devConf := range ns.RegisteredDeviceConfig {
			if len(devConf.AttachedVeth) != 0 {
				continue
			}

			if _, ok := links[devConf.Name]; !ok {
				continue
			}

			if _, ok := netLinks[devConf.Name]; !ok {
				netLinks[devConf.Name] = []int{}
			}
			netLinks[devConf.Name] = append(netLinks[devConf.Name], i)
		}
	}

	for linkName, idxs := range netLinks {
		if len(idxs) != 2 {
			return fmt.Errorf("%s should have only 2 link in %s\n", linkName, namespaces[idxs[0]].Name)
		}

		targetLink, ok := links[linkName]
		if !ok {
			return fmt.Errorf("can't find device %s in configured links", linkName)
		}

		if err := targetLink.CreateLink(namespaces[idxs[0]], namespaces[idxs[1]], dryrun); err != nil {
			return fmt.Errorf("failed to create links %s: %s", linkName, err.Error())
		}
	}

	return nil
}

func InitNamespacesBridges(namespaces []*Namespace, bridges map[string]*Bridge, dryrun bool) error {
	for _, ns := range namespaces {
		for _, dev := range ns.RegisteredDeviceConfig {
			if len(dev.AttachedVeth) != 0 {
				continue
			}

			targetLink, ok := bridges[dev.Name]
			if !ok {
				continue
			}

			if err := targetLink.CreateLink(ns, dryrun); err != nil {
				return fmt.Errorf("failed to link %s to bridge %s", ns.Name, targetLink.Name)
			}
		}
	}

	return nil
}

func CleanupNamespaces(nss []*Namespace, dryrun bool) error {
	var allerr error
	for _, n := range nss {
		if err := n.Destroy(dryrun); err != nil {
			allerr = multierr.Append(allerr, err)
		}
	}
	return allerr
}
