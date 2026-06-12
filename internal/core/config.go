package core

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/niklucky/wombat/internal/models"
	"github.com/niklucky/wombat/internal/sshconfig"
)

// Config holds the application configuration.
// Hosts are stored in ~/.ssh/config.d/wombat; tunnels and keys stay in JSON.
type Config struct {
	OpenTray     bool            `json:"open_tray"`
	ShowNotify   bool            `json:"show_notifications"`
	Language     string          `json:"language,omitempty"`
	Hosts        []models.Host   `json:"-"`
	Keys         []models.Key    `json:"keys"`
	Tunnels      []models.Tunnel `json:"tunnels"`
}

// DefaultConfig returns an empty default configuration.
func DefaultConfig() Config {
	return Config{
		OpenTray:   true,
		ShowNotify: true,
		Hosts:      []models.Host{},
		Keys:       []models.Key{},
		Tunnels:    []models.Tunnel{},
	}
}

// ConfigPath returns the path to the wombat JSON config file.
// It lives inside the configured AppHome.
func ConfigPath() (string, error) {
	home, err := AppHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "config.json"), nil
}

// Load reads tunnels/keys from JSON and hosts from SSH config.
func (c *Config) Load() error {
	// Load tunnels and keys from JSON
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		// File doesn't exist — start empty
		*c = DefaultConfig()
	} else {
		if err := json.Unmarshal(data, c); err != nil {
			return err
		}
	}

	// Load hosts from SSH config
	hosts, err := sshconfig.ReadHosts()
	if err != nil {
		return err
	}
	c.Hosts = hosts
	return nil
}

// Save writes tunnels/keys to JSON and hosts to SSH config.
func (c *Config) Save() error {
	// Save tunnels and keys to JSON
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	jsonCfg := struct {
		OpenTray   bool            `json:"open_tray"`
		ShowNotify bool            `json:"show_notifications"`
		Language   string          `json:"language,omitempty"`
		Keys       []models.Key    `json:"keys"`
		Tunnels    []models.Tunnel `json:"tunnels"`
	}{
		OpenTray:   c.OpenTray,
		ShowNotify: c.ShowNotify,
		Language:   c.Language,
		Keys:       c.Keys,
		Tunnels:    c.Tunnels,
	}
	data, err := json.MarshalIndent(jsonCfg, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}

	// Save hosts to SSH config
	return sshconfig.WriteHosts(c.Hosts)
}

// FindHost returns a host by name, or nil if not found.
func (c *Config) FindHost(name string) *models.Host {
	for i := range c.Hosts {
		if c.Hosts[i].Name == name {
			return &c.Hosts[i]
		}
	}
	return nil
}

// FindTunnel returns a tunnel by name, or nil if not found.
func (c *Config) FindTunnel(name string) *models.Tunnel {
	for i := range c.Tunnels {
		if c.Tunnels[i].Name == name {
			return &c.Tunnels[i]
		}
	}
	return nil
}
