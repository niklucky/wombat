package sshconfig

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kevinburke/ssh_config"
	"github.com/niklucky/wombat/internal/models"
)

// ImportFromMainConfig reads non-wildcard hosts from ~/.ssh/config and returns them.
func ImportFromMainConfig() ([]models.Host, error) {
	mainPath := mainConfigPath()
	f, err := os.Open(mainPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []models.Host{}, nil
		}
		return nil, fmt.Errorf("open ~/.ssh/config: %w", err)
	}
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("parse ~/.ssh/config: %w", err)
	}

	var hosts []models.Host
	for _, h := range cfg.Hosts {
		for _, pat := range h.Patterns {
			name := pat.String()
			if name == "" {
				continue
			}
			// Skip wildcards and patterns
			if strings.ContainsAny(name, "*?!") {
				continue
			}
			// Skip our own include marker sections
			if strings.HasPrefix(name, "config.d") {
				continue
			}

			host := models.Host{Name: name}

			if val, err := cfg.Get(name, "HostName"); err == nil && val != "" {
				host.Address = val
			} else {
				// If no HostName is set, the alias itself is the address
				host.Address = name
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
