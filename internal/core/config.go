package core

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/niklucky/wombat/internal/models"
	"github.com/niklucky/wombat/internal/sshconfig"
)

// CurrentStorageVersion is the independent storage schema version.
// It changes only when the persisted config format changes, not on every app release.
const CurrentStorageVersion = 1

// Config holds the application configuration.
// Hosts are stored in ~/.ssh/config.d/wombat; tunnels and keys stay in JSON.
type Config struct {
	StorageVersion int             `json:"storage_version"`
	OpenTray       bool            `json:"open_tray"`
	ShowNotify     bool            `json:"show_notifications"`
	Language       string          `json:"language,omitempty"`
	Hosts          []models.Host   `json:"-"`
	Keys           []models.Key    `json:"keys"`
	Tunnels        []models.Tunnel `json:"tunnels"`
}

// DefaultConfig returns an empty default configuration.
func DefaultConfig() Config {
	return Config{
		StorageVersion: CurrentStorageVersion,
		OpenTray:       true,
		ShowNotify:     true,
		Hosts:          []models.Host{},
		Keys:           []models.Key{},
		Tunnels:        []models.Tunnel{},
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
// If the on-disk storage version differs from CurrentStorageVersion, it backs up
// both config.json and the wombat-managed SSH host file, then re-saves everything
// in the current format.
func (c *Config) Load() error {
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
		hosts, err := sshconfig.ReadHosts()
		if err != nil {
			return err
		}
		c.Hosts = hosts
		return nil
	}

	// Peek the stored version first. A missing storage_version field means
	// legacy v0; we cannot rely on json.Unmarshal zeroing the field because
	// it preserves existing struct values for absent keys.
	var raw struct {
		StorageVersion *int `json:"storage_version"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	storedVersion := 0
	if raw.StorageVersion != nil {
		storedVersion = *raw.StorageVersion
	}

	if err := json.Unmarshal(data, c); err != nil {
		return err
	}

	// Ensure slices are non-nil for a consistent in-memory state.
	if c.Hosts == nil {
		c.Hosts = []models.Host{}
	}
	if c.Keys == nil {
		c.Keys = []models.Key{}
	}
	if c.Tunnels == nil {
		c.Tunnels = []models.Tunnel{}
	}

	// Load hosts from SSH config (re-import).
	hosts, err := sshconfig.ReadHosts()
	if err != nil {
		return err
	}
	c.Hosts = hosts

	// Migrate if the stored version differs.
	if storedVersion != CurrentStorageVersion {
		if err := c.migrate(path, storedVersion); err != nil {
			return fmt.Errorf("migrate config from v%d: %w", storedVersion, err)
		}
	}

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
		StorageVersion int             `json:"storage_version"`
		OpenTray       bool            `json:"open_tray"`
		ShowNotify     bool            `json:"show_notifications"`
		Language       string          `json:"language,omitempty"`
		Keys           []models.Key    `json:"keys"`
		Tunnels        []models.Tunnel `json:"tunnels"`
	}{
		StorageVersion: CurrentStorageVersion,
		OpenTray:       c.OpenTray,
		ShowNotify:     c.ShowNotify,
		Language:       c.Language,
		Keys:           c.Keys,
		Tunnels:        c.Tunnels,
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

// migrate backs up the existing config and host files, applies any necessary
// format transformations, and re-saves everything with the current version.
func (c *Config) migrate(configPath string, oldVersion int) error {
	suffix := fmt.Sprintf(".v%d", oldVersion)

	if _, err := backupFile(configPath, suffix); err != nil {
		return fmt.Errorf("backup config: %w", err)
	}

	hostsPath := sshconfig.WombatConfigPath()
	if _, err := backupFile(hostsPath, suffix); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("backup hosts: %w", err)
	}

	if err := transformConfig(c, oldVersion); err != nil {
		return fmt.Errorf("transform config: %w", err)
	}

	c.StorageVersion = CurrentStorageVersion
	return c.Save()
}

// backupFile copies src to src{suffix}. If that destination already exists,
// it appends a numeric counter (e.g. .v0.1, .v0.2) until a free name is found.
// The backup file is created with the same permissions as the source file.
// The destination name is reserved atomically with O_EXCL to avoid races.
func backupFile(src, suffix string) (string, error) {
	info, err := os.Stat(src)
	if err != nil {
		if os.IsNotExist(err) {
			return "", err
		}
		return "", err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return "", err
	}

	dst := src + suffix
	for i := 1; ; i++ {
		dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_EXCL, info.Mode().Perm())
		if err == nil {
			_, copyErr := io.Copy(dstFile, srcFile)

			// Explicitly close and surface errors. A failed close on the
			// destination file can indicate an incomplete write, so that error
			// must be checked.
			srcCloseErr := srcFile.Close()
			dstCloseErr := dstFile.Close()

			if copyErr != nil {
				return "", copyErr
			}
			if dstCloseErr != nil {
				return "", dstCloseErr
			}
			if srcCloseErr != nil {
				return "", srcCloseErr
			}

			return dst, nil
		}

		if !os.IsExist(err) {
			srcFile.Close()
			return "", err
		}

		// dst already exists; try the next counter suffix.
		dst = fmt.Sprintf("%s%s.%d", src, suffix, i)
	}
}

// transformConfig applies version-specific transformations when migrating from
// an older storage version to the current one. It is the hook for any future
// schema changes that cannot be handled by JSON unmarshalling alone.
func transformConfig(c *Config, fromVersion int) error {
	switch fromVersion {
	case 0:
		// v0 -> v1: no structural change other than adding storage_version.
		return nil
	default:
		return fmt.Errorf("unknown source storage version %d", fromVersion)
	}
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
