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

type Veth struct {
	Name     string `json:"name"`
	Attached bool   `json:"attached"`
}

func (v *Veth) Attach(ns *Namespace, cidr *net.IPNet, verbose bool) (*AttachedDevice, error) {
	if v.Attached {
		return nil, fmt.Errorf("%s is already attached", v.Name)
	}

	if err := RunIpLinkSetNamespaces(v.Name, ns.Name, verbose); err != nil {
		return nil, err
	}

	if err := RunAssignCidrToNamespaces(v.Name, ns.Name, cidr, verbose); err != nil {
		return nil, fmt.Errorf("failed to assign CIDR %s to ns %s on %s", cidr.String(), ns.Name, v.Name)
	}

	v.Attached = true
	return &AttachedDevice{Dev: v, Cidr: cidr.String()}, nil
}

type VethPair struct {
	Left   *Veth `yaml:"veth_left"`
	Right  *Veth `yaml:"veth_right"`
	Active bool  `yaml:"is_active"`
}

func CreateVethPair(conf config.Veth, verbose bool) (*VethPair, error) {
	pair := VethPair{
		Left:   &Veth{Name: conf.Left, Attached: false},
		Right:  &Veth{Name: conf.Right, Attached: false},
		Active: false,
	}

	if err := pair.Create(verbose); err != nil {
		return &pair, err
	}

	return &pair, nil
}

func (v *VethPair) Create(verbose bool) error {
	if v.Active {
		return fmt.Errorf("%s@%s is already created", v.Left.Name, v.Right.Name)
	}

	if err := RunIpLinkCreate(v.Left.Name, v.Right.Name, verbose); err != nil {
		return err
	}

	v.Active = true
	log.Printf("succeeded to create %s@%s", v.Left.Name, v.Right.Name)

	return nil
}

func (v *VethPair) Destroy(verbose bool) error {
	if !v.Active {
		return fmt.Errorf("%s@%s doesn't exist", v.Left.Name, v.Right.Name)
	}

	deleted := false

	if !v.Left.Attached {
		if err := RunIpLinkDelete(v.Left.Name, verbose); err != nil {
			return err
		}

		deleted = true
	}

	if !deleted && !v.Right.Attached {
		if err := RunIpLinkDelete(v.Right.Name, verbose); err != nil {
			return err
		}

		deleted = true
	}

	if !deleted {
		log.Printf("veth-pair %s@%s is invisible from host", v.Left.Name, v.Right.Name)
		return nil
	}

	v.Active = false
	log.Printf("succeeded to delete %s@%s", v.Left.Name, v.Right.Name)

	return nil
}

func InitVethPairs(conf []config.Veth, verbose bool) ([]*VethPair, error) {
	var activeVethPairs []*VethPair

	for _, c := range conf {
		vethPair, err := CreateVethPair(c, verbose)
		activeVethPairs = append(activeVethPairs, vethPair)

		if err != nil {
			return activeVethPairs, err
		}
	}

	return activeVethPairs, nil
}

func CleanupAllVethPairs(vps *[]*VethPair, verbose bool) error {
	var allerr error
	for _, v := range *vps {
		if err := v.Destroy(verbose); err != nil {
			allerr = multierr.Append(allerr, err)
		}
	}
	return allerr
}
