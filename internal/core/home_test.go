package core

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAppHome_defaultsToConfigDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	home, err := AppHome()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(home, "wombat") {
		t.Errorf("expected path to contain 'wombat', got %q", home)
	}
}

func TestAppHome_readsPointerFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	custom := filepath.Join(tmp, "custom-wombat")
	if err := SetAppHome(custom); err != nil {
		t.Fatalf("set app home failed: %v", err)
	}

	home, err := AppHome()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if home != custom {
		t.Errorf("expected %q, got %q", custom, home)
	}
}

func TestConfigPath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(path, "config.json") {
		t.Errorf("expected path to contain 'config.json', got %q", path)
	}
	if !strings.Contains(path, "wombat") {
		t.Errorf("expected path to contain 'wombat', got %q", path)
	}
}
