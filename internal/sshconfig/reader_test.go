package sshconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadHosts_basic(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	content := "Host web\n  HostName 10.0.0.1\n  User admin\n  Port 2222\n  IdentityFile ~/.ssh/web_key\n"
	if err := os.MkdirAll(filepath.Join(tmp, ".ssh", "config.d"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(wombatConfigPath(), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	hosts, err := ReadHosts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	h := hosts[0]
	if h.Name != "web" {
		t.Errorf("expected name web, got %s", h.Name)
	}
	if h.Address != "10.0.0.1" {
		t.Errorf("expected address 10.0.0.1, got %s", h.Address)
	}
	if h.User != "admin" {
		t.Errorf("expected user admin, got %s", h.User)
	}
	if h.Port != 2222 {
		t.Errorf("expected port 2222, got %d", h.Port)
	}
	if h.KeyPath != "~/.ssh/web_key" {
		t.Errorf("expected key ~/.ssh/web_key, got %s", h.KeyPath)
	}
}

func TestReadHosts_missingFile(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	if err := os.MkdirAll(filepath.Join(tmp, ".ssh", "config.d"), 0700); err != nil {
		t.Fatal(err)
	}

	hosts, err := ReadHosts()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 0 {
		t.Errorf("expected 0 hosts, got %d", len(hosts))
	}
}
