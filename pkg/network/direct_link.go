package network

import (
	"fmt"

	"github.com/Shikugawa/ayame/pkg/config"
	log "github.com/sirupsen/logrus"
	"go.uber.org/multierr"
)

type DirectLink struct {
	VethPair `json:"veth_pair"`
	Name     string `json:"name"`
	Busy     bool   `json:"busy"`
}

func InitDirectLink(cfg *config.LinkConfig, dryrun bool) (*DirectLink, error) {
	if cfg.LinkMode != config.ModeDirectLink {
		return nil, fmt.Errorf("invalid mode")
	}

	conf := VethConfig{
		Name: cfg.Name,
	}

	pair, err := InitVethPair(conf, dryrun)
	if err != nil {
		return nil, err
	}

	return &DirectLink{
		VethPair: *pair,
		Name:     cfg.Name,
		Busy:     false,
	}, nil
}

// TODO: consider error handling
func (d *DirectLink) Destroy(dryrun bool) error {
	if !d.Busy {
		return fmt.Errorf("%s is not busy\n", d.Name)
	}

	return d.VethPair.Destroy(dryrun)
}

// TODO: consider error handling
func (d *DirectLink) CreateLink(left *Namespace, right *Namespace, dryrun bool) error {
	if d.Busy {
		return fmt.Errorf("%s has been already busy\n", d.Name)
	}

	if err := (*left).Attach(&d.VethPair.Left, dryrun); err != nil {
		return err
	}

	if err := (*right).Attach(&d.VethPair.Right, dryrun); err != nil {
		// TODO: add error handling if left succeeded but right failed.
		return err
	}

	d.Busy = true
	return nil
}

func InitDirectLinks(links []*config.LinkConfig, dryrun bool) []*DirectLink {
	var dlinks []*DirectLink
	for _, link := range links {
		if link.LinkMode != config.ModeDirectLink {
			continue
		}

		dlink, err := InitDirectLink(link, dryrun)
		if err != nil {
			log.Errorf("failed to init direct link: %s", link.Name)
			continue
		}

		dlinks = append(dlinks, dlink)
	}

	return dlinks
}

func CleanupDirectLinks(links []*DirectLink, dryrun bool) error {
	var allerr error
	for _, link := range links {
		if err := link.Destroy(dryrun); err != nil {
			allerr = multierr.Append(allerr, err)
		}
	}
	return allerr
}
