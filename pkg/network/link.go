package network

import (
	"fmt"

	"github.com/Shikugawa/ayame/pkg/config"
	"go.uber.org/multierr"
)

type Attacheable interface {
	Attach(*Veth, bool) error
}

type Link interface {
	Destroy(verbose bool) error
	CreateLink(left Attacheable, right Attacheable, verbose bool) error
	Name() string
}

func InitLinks(confs []config.LinkConfig, verbose bool) ([]Link, error) {
	var links []Link

	for _, conf := range confs {
		if conf.LinkMode == config.DirectLink {
			link, err := InitDirectLink(conf, verbose)
			if err != nil {
				return nil, err
			}
			links = append(links, link)
		} else if conf.LinkMode == config.Bridge {
			return nil, fmt.Errorf("not implemented\n")
		}
	}

	return links, nil
}

func CleanupLinks(links []Link, verbose bool) error {
	var allerr error
	for _, link := range links {
		if err := link.Destroy(verbose); err != nil {
			allerr = multierr.Append(allerr, err)
		}
	}
	return allerr
}
