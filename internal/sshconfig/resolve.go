package sshconfig

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kevinburke/ssh_config"
	"github.com/niklucky/wombat/internal/models"
)

// Resolve looks up effective SSH settings for an alias across the full SSH config
// (user's main config + wombat include).
func Resolve(alias string) (models.Host, error) {
	mainPath := mainConfigPath()
	f, err := os.Open(mainPath)
	if err != nil {
		if os.IsNotExist(err) {
			return models.Host{}, fmt.Errorf("host %q not found in SSH config", alias)
		}
		return models.Host{}, fmt.Errorf("open ~/.ssh/config: %w", err)
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return models.Host{}, fmt.Errorf("parse ~/.ssh/config: %w", err)
	}

	host := models.Host{Name: alias}

	if val, err := cfg.Get(alias, "HostName"); err == nil && val != "" {
		host.Address = val
	} else {
		return models.Host{}, fmt.Errorf("host %q not found in SSH config", alias)
	}

	if val, err := cfg.Get(alias, "User"); err == nil && val != "" {
		host.User = val
	}
	if val, err := cfg.Get(alias, "Port"); err == nil && val != "" {
		if p, err := strconv.Atoi(val); err == nil {
			host.Port = p
		}
	}
	if val, err := cfg.Get(alias, "IdentityFile"); err == nil && val != "" {
		host.KeyPath = val
	}

	return host, nil
}
