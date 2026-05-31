package core

import (
	"path/filepath"
	"strings"
	"testing"
)

// setTestHome sets the environment variables that os.UserHomeDir and
// os.UserConfigDir consult on both Unix and Windows.
func setTestHome(t *testing.T, tmp string) {
	t.Helper()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("APPDATA", filepath.Join(tmp, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(tmp, "AppData", "Local"))
}

func TestAppHome_defaultsToConfigDir(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

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
	setTestHome(t, tmp)

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
	setTestHome(t, tmp)

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
