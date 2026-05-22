package sshconfig

import (
	"os"
	"path/filepath"
)

// sshDir returns the user's ~/.ssh directory.
func sshDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh")
}

// mainConfigPath returns ~/.ssh/config.
func mainConfigPath() string {
	return filepath.Join(sshDir(), "config")
}

// wombatConfigPath returns ~/.ssh/config.d/wombat.
func wombatConfigPath() string {
	return filepath.Join(sshDir(), "config.d", "wombat")
}
