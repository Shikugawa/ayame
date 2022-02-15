package network

import (
	"fmt"

	"github.com/Shikugawa/ayame/pkg/config"
)

type DirectLink struct {
	pair VethPair `json:"veth_pair"`
	busy bool     `json:"busy"`
}

func InitDirectLink(config config.LinkConfig) (*DirectLink, error) {
	conf := VethConfig{
		Name: config.Name,
	}

	pair, err := InitVethPair(conf)
	if err != nil {
		return nil, err
	}

	return &DirectLink{
		pair: *pair,
		busy: false,
	}, nil
}

func (d *DirectLink) Destroy() error {
	if !d.busy {
		return fmt.Errorf("%s is not busy\n", d.pair.Name)
	}

	return d.pair.Destroy()
}

func (d *DirectLink) CreateLink(left Attacheable, right Attacheable) error {
	if d.busy {
		return fmt.Errorf("%s has been already busy\n", d.pair.Name)
	}
	if err := left.Attach(d.pair.Left); err != nil {
		return err
	}

	if err := right.Attach(d.pair.Right); err != nil {
		return err
	}

	d.busy = true
	return nil
}

func (d *DirectLink) Name() string {
	return d.pair.Name
}
