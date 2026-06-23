package sshconfig

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/niklucky/wombat/internal/models"
)

// WriteHosts writes the given hosts to ~/.ssh/config.d/wombat in SSH config format.
func WriteHosts(hosts []models.Host) error {
	path := wombatConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create config.d: %w", err)
	}

	var buf bytes.Buffer
	for _, h := range hosts {
		fmt.Fprintf(&buf, "Host %s\n", h.Name)
		fmt.Fprintf(&buf, "    HostName %s\n", h.Address)
		fmt.Fprintf(&buf, "    User %s\n", h.User)
		if h.Port != 0 && h.Port != 22 {
			fmt.Fprintf(&buf, "    Port %d\n", h.Port)
		}
		if h.KeyPath != "" {
			fmt.Fprintf(&buf, "    IdentityFile %s\n", h.KeyPath)
		}
		if h.ProxyJump != "" {
			fmt.Fprintf(&buf, "    ProxyJump %s\n", h.ProxyJump)
		}
		fmt.Fprintln(&buf)
	}

	return os.WriteFile(path, buf.Bytes(), 0600)
}
