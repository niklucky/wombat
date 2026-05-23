//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/niklucky/wombat/internal/core"
)

// execSSH launches the Windows `ssh.exe` client to connect to the provided host and returns any error from running the client.
// It uses host.Port (defaults to 22 when zero), includes host.KeyPath as an `-i` identity option when present, and connects to `host.User@host.Address`.
// Standard input, output and error are wired to the current process. On successful execution the function calls os.Exit(0); otherwise it returns the error from cmd.Run().
func execSSH(host core.Host) error {
	if host.User == "" {
		return fmt.Errorf("missing host user")
	}
	if host.Address == "" {
		return fmt.Errorf("missing host address")
	}

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
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "ssh.exe: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
	return nil
}
