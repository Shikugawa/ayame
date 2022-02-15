package network

import (
	"fmt"
	"os/exec"
)

func CreateNewBridge(name string, verbose bool) error {
	cmd := exec.Command("ovs-vsctl", "add-br", name)

	if verbose {
		fmt.Println(cmd.String())
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bridge %s", name)
	}

	return nil
}

func DeleteBridge(name string, verbose bool) error {
	cmd := exec.Command("ovs-vsctl", "del-br", name)

	if verbose {
		fmt.Println(cmd.String())
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete bridge %s", name)
	}

	return nil
}

func LinkBridge(name string, veth *Veth, verbose bool) error {
	cmd := exec.Command("ovs-vsctl", "add-port", name, veth.Name)

	if verbose {
		fmt.Println(cmd.String())
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed link %s to %s", veth.Name, name)
	}

	return nil
}
