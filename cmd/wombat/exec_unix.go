//go:build darwin || linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/niklucky/wombat/internal/core"
)

// execSSH constructs an ssh command from the provided host and replaces the current process with that ssh invocation.
//
// The function uses host.Port (defaults to 22 when zero), includes host.KeyPath as a `-i` identity flag when non-empty,
// and targets the remote as `user@address`. It resolves the `ssh` binary via PATH and returns an error if `ssh` is not found
// or if the process replacement fails.
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
