//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/niklucky/wombat/internal/core"
)

func execSSH(host core.Host) error {
	port := host.Port
	if port == 0 {
		port = 22
	}

	args := []string{"ssh", "-p", fmt.Sprintf("%d", port)}
	if host.KeyPath != "" {
		args = append(args, "-i", host.KeyPath)
	}
	args = append(args, fmt.Sprintf("%s@%s", host.User, host.Address))

	sshPath, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	return syscall.Exec(sshPath, args, os.Environ())
}
