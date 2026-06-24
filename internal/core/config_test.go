package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklucky/wombat/internal/models"
	"github.com/niklucky/wombat/internal/sshconfig"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.StorageVersion != CurrentStorageVersion {
		t.Errorf("expected StorageVersion %d, got %d", CurrentStorageVersion, cfg.StorageVersion)
	}
	if !cfg.OpenTray {
		t.Error("expected OpenTray to be true")
	}
	if !cfg.ShowNotify {
		t.Error("expected ShowNotify to be true")
	}
	if cfg.Hosts == nil {
		t.Error("expected Hosts to be non-nil")
	}
	if cfg.Keys == nil {
		t.Error("expected Keys to be non-nil")
	}
	if cfg.Tunnels == nil {
		t.Error("expected Tunnels to be non-nil")
	}
}

func TestFindHost_found(t *testing.T) {
	cfg := Config{Hosts: []models.Host{{Name: "web", Address: "10.0.0.1", User: "root"}}}
	h := cfg.FindHost("web")
	if h == nil {
		t.Fatal("expected to find host")
	}
	if h.Address != "10.0.0.1" {
		t.Errorf("expected address 10.0.0.1, got %s", h.Address)
	}
}

func TestFindHost_missing(t *testing.T) {
	cfg := Config{Hosts: []models.Host{{Name: "web"}}}
	if cfg.FindHost("db") != nil {
		t.Error("expected nil for missing host")
	}
}

func TestFindTunnel_found(t *testing.T) {
	cfg := Config{Tunnels: []models.Tunnel{{Name: "api", HostName: "web", LocalPort: 3000}}}
	tu := cfg.FindTunnel("api")
	if tu == nil {
		t.Fatal("expected to find tunnel")
	}
	if tu.HostName != "web" {
		t.Errorf("expected host name web, got %s", tu.HostName)
	}
}

func TestFindTunnel_missing(t *testing.T) {
	cfg := Config{Tunnels: []models.Tunnel{{Name: "api"}}}
	if cfg.FindTunnel("web") != nil {
		t.Error("expected nil for missing tunnel")
	}
}

func TestAddHost(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddHost(Host{Name: "web", Address: "10.0.0.1", User: "root"})
	if len(cfg.Hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(cfg.Hosts))
	}
	if cfg.Hosts[0].Name != "web" {
		t.Errorf("expected name web, got %s", cfg.Hosts[0].Name)
	}
}

func TestRemoveHost(t *testing.T) {
	cfg := Config{Hosts: []models.Host{{Name: "web"}, {Name: "db"}, {Name: "cache"}}}
	cfg.RemoveHost("db")
	if len(cfg.Hosts) != 2 {
		t.Fatalf("expected 2 hosts, got %d", len(cfg.Hosts))
	}
	for _, h := range cfg.Hosts {
		if h.Name == "db" {
			t.Error("expected db to be removed")
		}
	}
}

func TestRemoveHost_noMatch(t *testing.T) {
	cfg := Config{Hosts: []models.Host{{Name: "web"}}}
	cfg.RemoveHost("missing")
	if len(cfg.Hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(cfg.Hosts))
	}
}

func TestAddTunnel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddTunnel(Tunnel{Name: "api", HostName: "web", LocalPort: 3000, RemotePort: 3000})
	if len(cfg.Tunnels) != 1 {
		t.Fatalf("expected 1 tunnel, got %d", len(cfg.Tunnels))
	}
	if cfg.Tunnels[0].Name != "api" {
		t.Errorf("expected name api, got %s", cfg.Tunnels[0].Name)
	}
}

func TestRemoveTunnel(t *testing.T) {
	cfg := Config{Tunnels: []models.Tunnel{{Name: "a"}, {Name: "b"}, {Name: "c"}}}
	cfg.RemoveTunnel("b")
	if len(cfg.Tunnels) != 2 {
		t.Fatalf("expected 2 tunnels, got %d", len(cfg.Tunnels))
	}
	for _, tu := range cfg.Tunnels {
		if tu.Name == "b" {
			t.Error("expected b to be removed")
		}
	}
}

func TestAddKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AddKey(Key{Path: "/home/user/.ssh/id_rsa"})
	if len(cfg.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(cfg.Keys))
	}
}

func TestRemoveKey(t *testing.T) {
	cfg := Config{Keys: []models.Key{{Path: "/a"}, {Path: "/b"}}}
	cfg.RemoveKey("/a")
	if len(cfg.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(cfg.Keys))
	}
	if cfg.Keys[0].Path != "/b" {
		t.Errorf("expected path /b, got %s", cfg.Keys[0].Path)
	}
}

func TestConfigLoadSave_roundTrip(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	cfg := DefaultConfig()
	cfg.OpenTray = false
	cfg.ShowNotify = false
	cfg.Tunnels = []models.Tunnel{{Name: "web", HostName: "server", LocalPort: 8080, RemoteHost: "localhost", RemotePort: 80}}
	cfg.Keys = []models.Key{{Path: "/key"}}
	cfg.Hosts = []models.Host{{Name: "server", Address: "10.0.0.1", User: "root"}}

	if err := cfg.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded := DefaultConfig()
	if err := loaded.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.StorageVersion != CurrentStorageVersion {
		t.Errorf("expected StorageVersion %d, got %d", CurrentStorageVersion, loaded.StorageVersion)
	}
	if loaded.OpenTray != false {
		t.Errorf("expected OpenTray false, got %v", loaded.OpenTray)
	}
	if loaded.ShowNotify != false {
		t.Errorf("expected ShowNotify false, got %v", loaded.ShowNotify)
	}
	if len(loaded.Tunnels) != 1 || loaded.Tunnels[0].Name != "web" {
		t.Errorf("expected tunnel web, got %+v", loaded.Tunnels)
	}
	if len(loaded.Keys) != 1 || loaded.Keys[0].Path != "/key" {
		t.Errorf("expected key /key, got %+v", loaded.Keys)
	}
	if len(loaded.Hosts) != 1 || loaded.Hosts[0].Name != "server" {
		t.Errorf("expected host server, got %+v", loaded.Hosts)
	}
}

func TestConfigLoad_missingFileStartsEmpty(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	cfg := DefaultConfig()
	if err := cfg.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if cfg.StorageVersion != CurrentStorageVersion {
		t.Errorf("expected StorageVersion %d for fresh config, got %d", CurrentStorageVersion, cfg.StorageVersion)
	}
	if len(cfg.Tunnels) != 0 {
		t.Errorf("expected 0 tunnels, got %d", len(cfg.Tunnels))
	}
}

func TestConfigLoad_backwardsCompatibleWithOldFormats(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	// Pre-localization config: no language field. Also simulate the very first
	// format by omitting open_tray/show_notifications/storage_version to ensure
	// unknown/missing fields do not cause data loss.
	oldConfig := `{
  "keys": [
    {"path": "/home/user/.ssh/id_rsa", "fingerprint": "abc", "is_agent_loaded": true}
  ],
  "tunnels": [
    {"name": "web", "host_name": "server", "local_port": 8080, "remote_host": "localhost", "remote_port": 80, "active": false}
  ]
}`

	home, err := AppHome()
	if err != nil {
		t.Fatalf("app home: %v", err)
	}
	configPath := filepath.Join(home, "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(oldConfig), 0600); err != nil {
		t.Fatalf("write old config: %v", err)
	}

	cfg := DefaultConfig()
	if err := cfg.Load(); err != nil {
		t.Fatalf("load old config: %v", err)
	}

	if cfg.StorageVersion != CurrentStorageVersion {
		t.Errorf("expected StorageVersion upgraded to %d, got %d", CurrentStorageVersion, cfg.StorageVersion)
	}
	if cfg.Language != "" {
		t.Errorf("expected empty language for old config, got %q", cfg.Language)
	}
	if len(cfg.Keys) != 1 || cfg.Keys[0].Path != "/home/user/.ssh/id_rsa" {
		t.Errorf("expected key to load, got %+v", cfg.Keys)
	}
	if len(cfg.Tunnels) != 1 || cfg.Tunnels[0].Name != "web" {
		t.Errorf("expected tunnel to load, got %+v", cfg.Tunnels)
	}

	// Migration should have created a backup of the legacy config.
	if _, err := os.Stat(configPath + ".v0"); err != nil {
		t.Errorf("expected legacy config backup: %v", err)
	}
}

func TestConfigMigration_createsBackupsAndUpgradesV0(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	oldConfig := `{
  "open_tray": false,
  "keys": [
    {"path": "/key", "fingerprint": "abc", "is_agent_loaded": false}
  ]
}`
	oldHosts := `Host server
    HostName 10.0.0.1
    User root
`

	home, err := AppHome()
	if err != nil {
		t.Fatalf("app home: %v", err)
	}
	configPath := filepath.Join(home, "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(oldConfig), 0600); err != nil {
		t.Fatalf("write old config: %v", err)
	}

	hostsPath := sshconfig.WombatConfigPath()
	if err := os.MkdirAll(filepath.Dir(hostsPath), 0755); err != nil {
		t.Fatalf("mkdir ssh dir: %v", err)
	}
	if err := os.WriteFile(hostsPath, []byte(oldHosts), 0600); err != nil {
		t.Fatalf("write old hosts: %v", err)
	}

	cfg := DefaultConfig()
	if err := cfg.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if cfg.StorageVersion != CurrentStorageVersion {
		t.Errorf("expected StorageVersion %d, got %d", CurrentStorageVersion, cfg.StorageVersion)
	}
	if cfg.OpenTray != false {
		t.Errorf("expected OpenTray false, got %v", cfg.OpenTray)
	}
	if len(cfg.Hosts) != 1 || cfg.Hosts[0].Name != "server" {
		t.Errorf("expected host server, got %+v", cfg.Hosts)
	}

	if _, err := os.Stat(configPath + ".v0"); err != nil {
		t.Errorf("expected config backup config.json.v0: %v", err)
	}
	if _, err := os.Stat(hostsPath + ".v0"); err != nil {
		t.Errorf("expected hosts backup wombat.v0: %v", err)
	}

	// The rewritten config should contain the current storage version.
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read rewritten config: %v", err)
	}
	if !strings.Contains(string(data), `"storage_version": 1`) {
		t.Errorf("rewritten config missing current storage version: %s", string(data))
	}
}

func TestConfigMigration_noOpWhenVersionMatches(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	cfg := DefaultConfig()
	cfg.Tunnels = []models.Tunnel{{Name: "web", HostName: "server", LocalPort: 8080, RemotePort: 80}}
	if err := cfg.Save(); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	home, err := AppHome()
	if err != nil {
		t.Fatalf("app home: %v", err)
	}
	configPath := filepath.Join(home, "config.json")
	hostsPath := sshconfig.WombatConfigPath()

	if err := os.Remove(configPath + ".v1"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("cleanup config backup: %v", err)
	}
	if err := os.Remove(hostsPath + ".v1"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("cleanup hosts backup: %v", err)
	}

	loaded := DefaultConfig()
	if err := loaded.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if _, err := os.Stat(configPath + ".v1"); !os.IsNotExist(err) {
		t.Error("expected no config backup when version matches")
	}
	if _, err := os.Stat(hostsPath + ".v1"); !os.IsNotExist(err) {
		t.Error("expected no hosts backup when version matches")
	}
}

func TestBackupFile_avoidsOverwriting(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "config.json")
	if err := os.WriteFile(src, []byte("original"), 0600); err != nil {
		t.Fatalf("write src: %v", err)
	}
	if err := os.WriteFile(src+".v0", []byte("backup1"), 0600); err != nil {
		t.Fatalf("write first backup: %v", err)
	}

	dst, err := backupFile(src, ".v0")
	if err != nil {
		t.Fatalf("backup failed: %v", err)
	}
	want := src + ".v0.1"
	if dst != want {
		t.Errorf("expected backup path %q, got %q", want, dst)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(data) != "original" {
		t.Errorf("expected backup content 'original', got %q", string(data))
	}
}
