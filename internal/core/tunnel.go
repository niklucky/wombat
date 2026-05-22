package core

import "github.com/niklucky/wombat/internal/models"

// Tunnel is an alias for models.Tunnel.
type Tunnel = models.Tunnel

// AddTunnel appends a tunnel to the configuration.
func (c *Config) AddTunnel(t Tunnel) {
	c.Tunnels = append(c.Tunnels, t)
}

// RemoveTunnel removes a tunnel by name.
func (c *Config) RemoveTunnel(name string) {
	filtered := c.Tunnels[:0]
	for _, t := range c.Tunnels {
		if t.Name != name {
			filtered = append(filtered, t)
		}
	}
	c.Tunnels = filtered
}
