package network

import (
	"fmt"

	"github.com/Shikugawa/ayame/pkg/config"
)

type DirectLink struct {
	pair VethPair `json:"veth_pair"`
	busy bool     `json:"busy"`
}

func InitDirectLink(config config.LinkConfig, verbose bool) (*DirectLink, error) {
	conf := VethConfig{
		Name: config.Name,
	}

	pair, err := InitVethPair(conf, verbose)
	if err != nil {
		return nil, err
	}

	return &DirectLink{
		pair: *pair,
		busy: false,
	}, nil
}

func (d *DirectLink) Destroy(verbose bool) error {
	if !d.busy {
		return fmt.Errorf("%s is not busy\n", d.pair.Name)
	}

	return d.pair.Destroy(verbose)
}

func (d *DirectLink) CreateLink(left Attacheable, right Attacheable, verbose bool) error {
	if d.busy {
		return fmt.Errorf("%s has been already busy\n", d.pair.Name)
	}
	if err := left.Attach(d.pair.Left, verbose); err != nil {
		return err
	}

	if err := right.Attach(d.pair.Right, verbose); err != nil {
		return err
	}

	d.busy = true
	return nil
}

func (d *DirectLink) Name() string {
	return d.pair.Name
}
