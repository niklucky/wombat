package sshconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportFromMainConfig_skipsWildcardsAndIncludes(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := os.MkdirAll(filepath.Join(tmp, ".ssh"), 0700); err != nil {
		t.Fatal(err)
	}
	content := `
Host *
  ForwardAgent yes

Host web
  HostName 10.0.0.1
  User admin

Host config.d*
  HostName ignored

Host db?!?
  HostName 10.0.0.2
`
	if err := os.WriteFile(mainConfigPath(), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	hosts, err := ImportFromMainConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].Name != "web" {
		t.Errorf("expected name web, got %s", hosts[0].Name)
	}
}

func TestImportFromMainConfig_missingFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	hosts, err := ImportFromMainConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 0 {
		t.Errorf("expected 0 hosts, got %d", len(hosts))
	}
}

func TestImportFromMainConfig_fallbackAddress(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	if err := os.MkdirAll(filepath.Join(tmp, ".ssh"), 0700); err != nil {
		t.Fatal(err)
	}
	content := "Host myserver\n  User root\n"
	if err := os.WriteFile(mainConfigPath(), []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	hosts, err := ImportFromMainConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}
	if hosts[0].Address != "myserver" {
		t.Errorf("expected fallback address myserver, got %s", hosts[0].Address)
	}
}
