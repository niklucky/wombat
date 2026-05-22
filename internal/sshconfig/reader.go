package sshconfig

import (
	"fmt"
	"os"
	"strconv"

	"github.com/kevinburke/ssh_config"
	"github.com/niklucky/wombat/internal/models"
)

// ReadHosts reads wombat-managed hosts from ~/.ssh/config.d/wombat.
func ReadHosts() ([]models.Host, error) {
	path := wombatConfigPath()
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Host{}, nil
		}
		return nil, fmt.Errorf("open wombat ssh config: %w", err)
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("parse wombat ssh config: %w", err)
	}

	var hosts []models.Host
	for _, h := range cfg.Hosts {
		for _, pat := range h.Patterns {
			name := pat.String()
			if name == "" || name == "*" {
				continue
			}
			host := models.Host{Name: name}

			if val, err := cfg.Get(name, "HostName"); err == nil && val != "" {
				host.Address = val
			}
			if val, err := cfg.Get(name, "User"); err == nil && val != "" {
				host.User = val
			}
			if val, err := cfg.Get(name, "Port"); err == nil && val != "" {
				if p, err := strconv.Atoi(val); err == nil {
					host.Port = p
				}
			}
			if val, err := cfg.Get(name, "IdentityFile"); err == nil && val != "" {
				host.KeyPath = val
			}

			hosts = append(hosts, host)
		}
	}

	return hosts, nil
}
