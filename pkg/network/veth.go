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

	log "github.com/sirupsen/logrus"
)

type VethConfig struct {
	Name string `yaml:"name"`
}

type Veth struct {
	Name     string `json:"name"`
	Attached bool   `json:"attached"`
}

type VethPair struct {
	Left   Veth `json:"veth_left"`
	Right  Veth `json:"veth_right"`
	Active bool `json:"is_active"`
}

func InitVethPair(config VethConfig, dryrun bool) (*VethPair, error) {
	pair := &VethPair{
		Left:   Veth{Name: config.Name + "-left", Attached: false},
		Right:  Veth{Name: config.Name + "-right", Attached: false},
		Active: false,
	}

	if err := pair.Create(dryrun); err != nil {
		return nil, err
	}

	return pair, nil
}

func (v *VethPair) Create(dryrun bool) error {
	if v.Active {
		return fmt.Errorf("%s@%s is already created", v.Left.Name, v.Right.Name)
	}

	if err := RunIpLinkCreate(v.Left.Name, v.Right.Name, dryrun); err != nil {
		return err
	}

	v.Active = true
	log.Infof("succeeded to create %s@%s", v.Left.Name, v.Right.Name)

	return nil
}

func (v *VethPair) Destroy(dryrun bool) error {
	if !v.Active {
		return fmt.Errorf("%s@%s doesn't exist", v.Left.Name, v.Right.Name)
	}

	deleted := false

	if !v.Left.Attached {
		if err := RunIpLinkDelete(v.Left.Name, dryrun); err != nil {
			return err
		}

		deleted = true
	}

	if !deleted && !v.Right.Attached {
		if err := RunIpLinkDelete(v.Right.Name, dryrun); err != nil {
			return err
		}

		deleted = true
	}

	if !deleted {
		log.Infof("veth-pair %s@%s is invisible from host", v.Left.Name, v.Right.Name)
		return nil
	}

	v.Active = false
	log.Infof("succeeded to delete %s@%s", v.Left.Name, v.Right.Name)

	return nil
}
