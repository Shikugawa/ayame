package main

import (
	"flag"
	"fmt"
	"github.com/Shikugawa/netb/pkg/config"
	"github.com/Shikugawa/netb/pkg/net"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

var (
	path = flag.String("config", "", "config path")
)

func main() {
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
	fmt.Println(cfg)

	state := net.State{}

	pairs, err := net.InitVethPairs(cfg.Veth)
	if err != nil {
		fmt.Println(err)
		pairs.Cleanup()
		return
	}

	state.VethPairs = pairs

	ns, err := net.InitNamespaces(cfg.Namespace, pairs)
	if err != nil {
		fmt.Println(err)
		pairs.Cleanup()
		ns.Cleanup()
		return
	}

	state.Namespaces = ns

	b, err := yaml.Marshal(state)
	if err != nil {
		fmt.Println(err)
		pairs.Cleanup()
		ns.Cleanup()
		return
	}

	fmt.Println(string(b))

	pairs.Cleanup()
	ns.Cleanup()

	b, err = yaml.Marshal(state)
	if err != nil {
		fmt.Println(err)
		pairs.Cleanup()
		ns.Cleanup()
		return
	}

	fmt.Println(string(b))
	//net.InactivateNS(ans)
	//net.InactivateVeths(aveths)
}
