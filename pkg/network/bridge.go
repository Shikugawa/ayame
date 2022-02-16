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

func InitBridge(cfg *config.LinkConfig) (*Bridge, error) {
	if cfg.LinkMode != config.ModeDirectLink {
		return nil, fmt.Errorf("invalid mode")
	}

	if err := CreateNewBridge(cfg.Name); err != nil {
		return nil, err
	}

	return &Bridge{
		Name: cfg.Name,
	}, nil
}

// TODO: consider error handling
func (d *Bridge) Destroy() error {
	for _, p := range d.VethPairs {
		if err := p.Destroy(); err != nil {
			log.Warnf(err.Error())
		}
	}

	if err := DeleteBridge(d.Name); err != nil {
		return err
	}

	return nil
}

// TODO: consider error handling
func (d *Bridge) CreateLink(target *Namespace) error {
	val, _ := uuid.NewRandom()
	conf := VethConfig{
		Name: val.String(),
	}

	pair, err := InitVethPair(conf)
	if err != nil {
		return err
	}

	if err := target.Attach(&pair.Left); err != nil {
		return err
	}

	if err := LinkBridge(d.Name, &pair.Right); err != nil {
		return err
	}

	d.VethPairs = append(d.VethPairs, pair)
	return nil
}

func InitBridges(links []*config.LinkConfig) []*Bridge {
	var brs []*Bridge
	for _, link := range links {
		if link.LinkMode != config.ModeBridge {
			continue
		}

		br, err := InitBridge(link)
		if err != nil {
			log.Errorf("failed to init direct link: %s", link.Name)
			continue
		}

		brs = append(brs, br)
	}

	return brs
}

func CleanupBridges(links []*Bridge) error {
	var allerr error
	for _, link := range links {
		if err := link.Destroy(); err != nil {
			allerr = multierr.Append(allerr, err)
		}
	}
	return allerr
}
