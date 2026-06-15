package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/locales"
)

func TestMain(m *testing.M) {
	if err := locales.SetLanguage("en"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestBoolToYesNo(t *testing.T) {
	if boolToYesNo(true) != "yes" {
		t.Error("expected 'yes' for true")
	}
	if boolToYesNo(false) != "no" {
		t.Error("expected 'no' for false")
	}
}

func TestYesNoToBool(t *testing.T) {
	cases := []struct {
		input string
		want  bool
	}{
		{"yes", true},
		{"YES", true},
		{"y", true},
		{"true", true},
		{"no", false},
		{"NO", false},
		{"n", false},
		{"false", false},
		{"", false},
		{"maybe", false},
	}
	for _, c := range cases {
		got := yesNoToBool(c.input)
		if got != c.want {
			t.Errorf("yesNoToBool(%q) = %v, want %v", c.input, got, c.want)
		}
	}
}

func setTestHome(t *testing.T, tmp string) {
	t.Helper()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("APPDATA", filepath.Join(tmp, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(tmp, "AppData", "Local"))
}

func TestSaveSettingsForm_migratesConfigSafely(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	oldHome := filepath.Join(tmp, ".config", "wombat")
	newHome := filepath.Join(tmp, "new-wombat")

	cfg := core.DefaultConfig()
	cfg.Tunnels = []core.Tunnel{
		{Name: "prod", HostName: "server", LocalPort: 8080, RemoteHost: "localhost", RemotePort: 80},
	}
	cfg.OpenTray = false
	cfg.ShowNotify = true

	if err := core.SetAppHome(oldHome); err != nil {
		t.Fatalf("set old app home: %v", err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("save initial config: %v", err)
	}

	m := NewModel(cfg)
	m.initSettingsForm()
	m.formInputs[0].SetValue(newHome)
	m.formBools[1] = true
	m.formBools[2] = true

	changed, err := m.saveSettingsForm()
	if err != nil {
		t.Fatalf("save settings form: %v", err)
	}
	if !changed {
		t.Fatal("expected home to change")
	}

	// New location should have the tunnel.
	newConfigPath := filepath.Join(newHome, "config.json")
	data, err := os.ReadFile(newConfigPath)
	if err != nil {
		t.Fatalf("read new config: %v", err)
	}
	if !strings.Contains(string(data), `"prod"`) {
		t.Errorf("new config missing tunnel: %s", data)
	}

	// Old location should still have config as backup.
	oldConfigPath := filepath.Join(oldHome, "config.json")
	if _, err := os.Stat(oldConfigPath); err != nil {
		t.Errorf("old config should be preserved as backup: %v", err)
	}

	// Pointer should point to new home.
	currentHome, err := core.AppHome()
	if err != nil {
		t.Fatalf("get app home: %v", err)
	}
	if currentHome != newHome {
		t.Errorf("app home = %q, want %q", currentHome, newHome)
	}
}

func TestSaveSettingsForm_restoresPointerOnSaveFailure(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	oldHome := filepath.Join(tmp, ".config", "wombat")
	// New home is a file, so creating a directory with the same name will fail.
	newHome := filepath.Join(tmp, "not-a-dir")

	cfg := core.DefaultConfig()
	cfg.Tunnels = []core.Tunnel{
		{Name: "prod", HostName: "server", LocalPort: 8080, RemoteHost: "localhost", RemotePort: 80},
	}

	if err := core.SetAppHome(oldHome); err != nil {
		t.Fatalf("set old app home: %v", err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("save initial config: %v", err)
	}

	m := NewModel(cfg)
	m.initSettingsForm()
	m.formInputs[0].SetValue(newHome)

	// Create a file at newHome so MkdirAll for config.json fails.
	if err := os.WriteFile(newHome, []byte("block"), 0600); err != nil {
		t.Fatalf("create blocking file: %v", err)
	}

	_, err := m.saveSettingsForm()
	if err == nil {
		t.Fatal("expected save to fail")
	}

	// Pointer should be restored to old home.
	currentHome, err := core.AppHome()
	if err != nil {
		t.Fatalf("get app home: %v", err)
	}
	if currentHome != oldHome {
		t.Errorf("app home = %q, want %q", currentHome, oldHome)
	}

	// Old config should still exist.
	oldConfigPath := filepath.Join(oldHome, "config.json")
	if _, err := os.Stat(oldConfigPath); err != nil {
		t.Errorf("old config missing: %v", err)
	}

	// New config should not exist.
	newConfigPath := filepath.Join(newHome, "config.json")
	if _, err := os.Stat(newConfigPath); err == nil {
		t.Errorf("new config should not exist")
	}
}
