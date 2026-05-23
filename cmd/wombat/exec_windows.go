//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/niklucky/wombat/internal/core"
)

func execSSH(host core.Host) error {
	port := host.Port
	if port == 0 {
		port = 22
	}

	args := []string{"-p", fmt.Sprintf("%d", port)}
	if host.KeyPath != "" {
		args = append(args, "-i", host.KeyPath)
	}
	args = append(args, fmt.Sprintf("%s@%s", host.User, host.Address))

	cmd := exec.Command("ssh.exe", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	os.Exit(0)
	return nil
}
