package sshconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const includeMarker = "# Wombat managed hosts"

// EnsureSetup makes sure ~/.ssh/config includes our config file.
// It is safe to call multiple times.
func EnsureSetup() error {
	ssh := sshDir()
	if ssh == "" {
		return fmt.Errorf("could not determine SSH directory")
	}

	// Ensure ~/.ssh exists
	if err := os.MkdirAll(ssh, 0700); err != nil {
		return fmt.Errorf("create ~/.ssh: %w", err)
	}

	// Ensure ~/.ssh/config.d exists
	configD := filepath.Join(ssh, "config.d")
	if err := os.MkdirAll(configD, 0700); err != nil {
		return fmt.Errorf("create ~/.ssh/config.d: %w", err)
	}

	mainPath := mainConfigPath()
	includeLine := fmt.Sprintf("Include %s", "config.d/wombat")

	data, err := os.ReadFile(mainPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new ~/.ssh/config with our include
			content := includeMarker + "\n" + includeLine + "\n\n"
			return os.WriteFile(mainPath, []byte(content), 0600)
		}
		return fmt.Errorf("read ~/.ssh/config: %w", err)
	}

	content := string(data)
	if strings.Contains(content, includeLine) {
		return nil // already set up
	}

	// Prepend include at the top
	newContent := includeMarker + "\n" + includeLine + "\n\n" + content
	return os.WriteFile(mainPath, []byte(newContent), 0600)
}
