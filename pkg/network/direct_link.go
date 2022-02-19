package network

import (
	"fmt"

	"github.com/Shikugawa/ayame/pkg/config"
	"go.uber.org/multierr"
)

type DirectLink struct {
	VethPair `json:"veth_pair"`
	Name     string `json:"name"`
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
	}, nil
}

func (d *DirectLink) Destroy(dryrun bool) error {
	return d.VethPair.Destroy(dryrun)
}

// TODO: consider error handling
func (d *DirectLink) CreateLink(left *Namespace, right *Namespace, dryrun bool) error {
	if d.VethPair.Left.Attached && d.VethPair.Right.Attached {
		return fmt.Errorf("%s has been already busy\n", d.Name)
	}

	if err := (*left).Attach(&d.VethPair.Left, dryrun); err != nil {
		return err
	}

	if err := (*right).Attach(&d.VethPair.Right, dryrun); err != nil {
		// TODO: add error handling if left succeeded but right failed.
		return err
	}

	return nil
}

func InitDirectLinks(links []*config.LinkConfig, dryrun bool) (map[string]*DirectLink, error) {
	dlinks := make(map[string]*DirectLink)
	for _, link := range links {
		if link.LinkMode != config.ModeDirectLink {
			continue
		}

		dlink, err := InitDirectLink(link, dryrun)
		if err != nil {
			return nil, fmt.Errorf("failed to init direct link: %s: %s", link.Name, err)
		}

		dlinks[dlink.Name] = dlink
	}

	return dlinks, nil
}

func CleanupDirectLinks(links map[string]*DirectLink, dryrun bool) error {
	var allerr error
	for _, link := range links {
		if err := link.Destroy(dryrun); err != nil {
			allerr = multierr.Append(allerr, err)
		}
	}
	return allerr
}
