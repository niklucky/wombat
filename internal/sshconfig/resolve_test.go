package sshconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolve_found(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	if err := os.MkdirAll(filepath.Join(tmp, ".ssh"), 0700); err != nil {
		t.Fatal(err)
	}
	content := "Host web\n  HostName 10.0.0.1\n  User admin\n  Port 2222\n  IdentityFile ~/.ssh/web\n"
	if err := os.WriteFile(mainConfigPath(), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	host, err := Resolve("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if host.Name != "web" {
		t.Errorf("expected name web, got %s", host.Name)
	}
	if host.Address != "10.0.0.1" {
		t.Errorf("expected address 10.0.0.1, got %s", host.Address)
	}
	if host.User != "admin" {
		t.Errorf("expected user admin, got %s", host.User)
	}
	if host.Port != 2222 {
		t.Errorf("expected port 2222, got %d", host.Port)
	}
	if host.KeyPath != "~/.ssh/web" {
		t.Errorf("expected key ~/.ssh/web, got %s", host.KeyPath)
	}
}

func TestResolve_missing(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	if err := os.MkdirAll(filepath.Join(tmp, ".ssh"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(mainConfigPath(), []byte("Host other\n  HostName 1.1.1.1\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := Resolve("missing")
	if err == nil {
		t.Error("expected error for missing host")
	}
}

func TestResolve_missingConfig(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	_, err := Resolve("anything")
	if err == nil {
		t.Error("expected error when config missing")
	}
}
