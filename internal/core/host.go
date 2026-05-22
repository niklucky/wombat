package core

import "github.com/niklucky/wombat/internal/models"

// Host is an alias for models.Host.
type Host = models.Host

// AddHost appends a host to the configuration.
func (c *Config) AddHost(h Host) {
	c.Hosts = append(c.Hosts, h)
}

// RemoveHost removes a host by name.
func (c *Config) RemoveHost(name string) {
	filtered := c.Hosts[:0]
	for _, h := range c.Hosts {
		if h.Name != name {
			filtered = append(filtered, h)
		}
	}
	c.Hosts = filtered
}
