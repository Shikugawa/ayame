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
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func RunIpLinkCreate(left string, right string) error {
	cmd := exec.Command("ip", "link", "add", "name", left, "type", "veth", "peer", right)
	log.Infoln("execute ", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create veth name %s@%s: %s", left, right, err)
	}

	return nil
}

func RunIpLinkDelete(name string) error {
	cmd := exec.Command("ip", "link", "delete", name)
	log.Infoln("execute ", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete device %s: %s", name, err)
	}

	return nil
}

func RunIpLinkSetNamespaces(ifname string, nsname string) error {
	cmd := exec.Command("ip", "link", "set", ifname, "netns", nsname)
	log.Infoln("execute ", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach device %s to ns %s", ifname, nsname)
	}

	return nil
}

func RunAssignCidrToNamespaces(ifname string, nsname string, cidr string) error {
	cmd := exec.Command("ip", "netns", "exec", nsname, "ip", "addr", "add", cidr, "dev", ifname)
	log.Infoln("execute ", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to assign CIDR %s to ns %s on %s", cidr, nsname, ifname)
	}

	return nil
}

func RunIpNetnsAdd(nsname string) error {
	cmd := exec.Command("ip", "netns", "add", nsname)
	log.Infoln("execute ", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create ns %s", nsname)
	}

	return nil
}

func RunIpNetnsDelete(nsname string) error {
	cmd := exec.Command("ip", "netns", "delete", nsname)
	log.Infoln("execute ", cmd.String())

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete ns %s", nsname)
	}

	return nil
}
