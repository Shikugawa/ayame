package network

import (
	"fmt"

	"github.com/Shikugawa/ayame/pkg/config"
	"github.com/google/uuid"
	"go.uber.org/multierr"

	log "github.com/sirupsen/logrus"
)

type Bridge struct {
	Name      string      `json:"name"`
	VethPairs []*VethPair `json:"veth_pairs"`
}

func InitBridge(cfg *config.LinkConfig, dryrun bool) (*Bridge, error) {
	if cfg.LinkMode != config.ModeDirectLink {
		return nil, fmt.Errorf("invalid mode")
	}

	if err := CreateNewBridge(cfg.Name, dryrun); err != nil {
		return nil, err
	}

	return &Bridge{
		Name: cfg.Name,
	}, nil
}

// TODO: consider error handling
func (d *Bridge) Destroy(dryrun bool) error {
	for _, p := range d.VethPairs {
		if err := p.Destroy(dryrun); err != nil {
			log.Warnf(err.Error())
		}
	}

	if err := DeleteBridge(d.Name, dryrun); err != nil {
		return err
	}

	return nil
}

// TODO: consider error handling
func (d *Bridge) CreateLink(target *Namespace, dryrun bool) error {
	val, _ := uuid.NewRandom()
	conf := VethConfig{
		Name: val.String(),
	}

	pair, err := InitVethPair(conf, dryrun)
	if err != nil {
		return err
	}

	if err := target.Attach(&pair.Left, dryrun); err != nil {
		return err
	}

	if err := LinkBridge(d.Name, &pair.Right, dryrun); err != nil {
		return err
	}

	d.VethPairs = append(d.VethPairs, pair)
	return nil
}

func InitBridges(links []*config.LinkConfig, dryrun bool) []*Bridge {
	var brs []*Bridge
	for _, link := range links {
		if link.LinkMode != config.ModeBridge {
			continue
		}

		br, err := InitBridge(link, dryrun)
		if err != nil {
			log.Errorf("failed to init direct link: %s", link.Name)
			continue
		}

		brs = append(brs, br)
	}

	return brs
}

func CleanupBridges(links []*Bridge, dryrun bool) error {
	var allerr error
	for _, link := range links {
		if err := link.Destroy(dryrun); err != nil {
			allerr = multierr.Append(allerr, err)
		}
	}
	return allerr
}
