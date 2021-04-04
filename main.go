package main

import (
	"flag"
	"fmt"
	"github.com/Shikugawa/netb/pkg/config"
	"github.com/Shikugawa/netb/pkg/net"
	"io/ioutil"
)

var (
	path = flag.String("config", "", "config path")
)

func main()  {
	flag.Parse()

	bytes, err := ioutil.ReadFile(*path)
	if err != nil {
		fmt.Println(err)
		return
	}
	cfg, err := config.ParseConfig(bytes)
	if err != nil {
		fmt.Println(err)
		return
	}

	aveths, err := net.ActivateVeth(cfg.Veth)
	if err != nil {
		fmt.Println(err)
		return
	}
	net.InactivateVeths(aveths)

	ans, err := net.ActivateNS(cfg.Namespace)
	if err != nil {
		fmt.Println(err)
		return
	}
	net.InactivateNS(ans)
}
