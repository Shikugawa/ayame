package network

// import (
// 	"fmt"
// 	"log"

// 	"github.com/Shikugawa/ayame/pkg/config"
// 	"github.com/google/uuid"
// )

// type Bridge struct {
// 	Name   string  `yaml:"name"`
// 	Links  []*Veth `yaml:"links"`
// 	Active bool    `yaml:"is_active"`
// }

// func (b *Bridge) Create(verbose bool) error {
// 	if b.Active {
// 		return fmt.Errorf("%s is already created", b.Name)
// 	}

// 	if err := CreateNewBridge(b.Name, verbose); err != nil {
// 		return err
// 	}

// 	b.Active = true
// 	log.Printf("succeeded to create %s", b.Name)

// 	return nil
// }

// func (b *Bridge) Destroy(verbose bool) error {
// 	if !b.Active {
// 		return fmt.Errorf("%s doesn't exist", b.Name)
// 	}

// 	// TODO: unlink
// 	if err := DeleteBridge(b.Name, verbose); err != nil {
// 		return err
// 	}

// 	b.Active = false
// 	log.Printf("succeeded to delete %s", b.Name)

// 	return nil
// }

// func (b *Bridge) Link(ns *Namespace, verbose bool) error {
// 	// Create veth pair
// 	uuid, _ := uuid.NewRandom()
// 	conf := config.Link{
// 		Name: fmt.Sprintf("bridge-link-%s", uuid),
// 	}

// 	if verbose {
// 		log.Printf("init link of bridge: %s with config as follows: %s", b.Name, conf)
// 	}

// 	vpair, err := CreateLink(conf, verbose)
// 	if err != nil {
// 		return err
// 	}

// 	if err := LinkBridge(b.Name, vpair.Left, verbose); err != nil {
// 		return err
// 	}

// 	// TODO: implement link to namespace

// 	return nil
// }
