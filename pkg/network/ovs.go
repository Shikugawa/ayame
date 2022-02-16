package network

import (
	"fmt"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

func CreateNewBridge(name string, dryrun bool) error {
	cmd := exec.Command("ovs-vsctl", "add-br", name)

	log.Infof("execute %s", cmd.String())

	if dryrun {
		return nil
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bridge %s", name)
	}

	return nil
}

func DeleteBridge(name string, dryrun bool) error {
	cmd := exec.Command("ovs-vsctl", "del-br", name)

	log.Infof("execute %s", cmd.String())

	if dryrun {
		return nil
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete bridge %s", name)
	}

	return nil
}

func LinkBridge(name string, veth *Veth, dryrun bool) error {
	cmd := exec.Command("ovs-vsctl", "add-port", name, veth.Name)

	log.Infof("execute %s", cmd.String())

	if dryrun {
		return nil
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed link %s to %s", veth.Name, name)
	}

	return nil
}
