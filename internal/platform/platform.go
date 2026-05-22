package platform

import (
	"os"
	"path/filepath"
)

// AgentSocketPath returns the SSH agent socket path, or empty string if not set.
func AgentSocketPath() string {
	if p := os.Getenv("SSH_AUTH_SOCK"); p != "" {
		return p
	}
	return ""
}

// ConfigDir returns the OS-specific directory for Wombat configuration.
func ConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "wombat"), nil
}
