//go:build darwin || linux

package main

import (
	"strings"
	"testing"

	"github.com/niklucky/wombat/internal/core"
)

// TestExecSSH_sshNotInPath verifies that execSSH returns a descriptive error
// when the ssh binary cannot be found in PATH.
func TestExecSSH_sshNotInPath(t *testing.T) {
	t.Setenv("PATH", "")

	host := core.Host{
		Name:    "testhost",
		Address: "192.168.1.1",
		User:    "admin",
		Port:    22,
	}

	err := execSSH(host)
	if err == nil {
		t.Fatal("expected an error when ssh is not in PATH, got nil")
	}
	if !strings.Contains(err.Error(), "ssh not found in PATH") {
		t.Errorf("expected error to contain %q, got: %q", "ssh not found in PATH", err.Error())
	}
}

// TestExecSSH_defaultPortUsed verifies that when port is 0 the function still
// proceeds to LookPath (i.e., does not short-circuit before looking for ssh).
// We confirm this by observing a "not found" error rather than any port-related error.
func TestExecSSH_defaultPortUsed(t *testing.T) {
	t.Setenv("PATH", "")

	host := core.Host{
		Name:    "testhost",
		Address: "example.com",
		User:    "user",
		Port:    0, // should default to 22 internally
	}

	err := execSSH(host)
	if err == nil {
		t.Fatal("expected error when ssh is not in PATH")
	}
	// The error must be about ssh lookup, not about an invalid port
	if !strings.Contains(err.Error(), "ssh not found in PATH") {
		t.Errorf("unexpected error: %q", err.Error())
	}
}

// TestExecSSH_withKeyPath verifies that a non-empty KeyPath does not cause a
// different error path before the LookPath call (argument construction is correct).
func TestExecSSH_withKeyPath(t *testing.T) {
	t.Setenv("PATH", "")

	host := core.Host{
		Name:    "myhost",
		Address: "10.0.0.1",
		User:    "deploy",
		Port:    2222,
		KeyPath: "/home/user/.ssh/id_ed25519",
	}

	err := execSSH(host)
	if err == nil {
		t.Fatal("expected error when ssh is not in PATH")
	}
	if !strings.Contains(err.Error(), "ssh not found in PATH") {
		t.Errorf("expected ssh-not-found error, got: %q", err.Error())
	}
}

// TestExecSSH_withoutKeyPath verifies that when KeyPath is empty the function
// still reaches LookPath without panicking or returning an unexpected error.
func TestExecSSH_withoutKeyPath(t *testing.T) {
	t.Setenv("PATH", "")

	host := core.Host{
		Name:    "myhost",
		Address: "10.0.0.2",
		User:    "ops",
		Port:    22,
		KeyPath: "", // no identity file
	}

	err := execSSH(host)
	if err == nil {
		t.Fatal("expected error when ssh is not in PATH")
	}
	if !strings.Contains(err.Error(), "ssh not found in PATH") {
		t.Errorf("expected ssh-not-found error, got: %q", err.Error())
	}
}