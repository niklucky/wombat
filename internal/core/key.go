package core

import "github.com/niklucky/wombat/internal/models"

// Key is an alias for models.Key.
type Key = models.Key

// AddKey appends a key to the configuration.
func (c *Config) AddKey(k Key) {
	c.Keys = append(c.Keys, k)
}

// RemoveKey removes a key by path.
func (c *Config) RemoveKey(path string) {
	filtered := c.Keys[:0]
	for _, k := range c.Keys {
		if k.Path != path {
			filtered = append(filtered, k)
		}
	}
	c.Keys = filtered
}
