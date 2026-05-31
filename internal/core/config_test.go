package core

import (
	"path/filepath"
	"testing"

	"github.com/niklucky/wombat/internal/models"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
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
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))

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
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))

	cfg := DefaultConfig()
	if err := cfg.Load(); err != nil {
		t.Fatalf("load failed: %v", err)
	}
	if len(cfg.Tunnels) != 0 {
		t.Errorf("expected 0 tunnels, got %d", len(cfg.Tunnels))
	}
}
