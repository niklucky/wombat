package sshconfig

import (
	"path/filepath"
	"testing"
)

// setTestHome sets the environment variables that os.UserHomeDir consults on
// Unix (HOME) and Windows (USERPROFILE) so tests use a temporary directory.
func setTestHome(t *testing.T, tmp string) {
	t.Helper()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
}

func TestPaths_underTempHome(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	if got := sshDir(); got != filepath.Join(tmp, ".ssh") {
		t.Errorf("sshDir() = %q, want %q", got, filepath.Join(tmp, ".ssh"))
	}
	if got := mainConfigPath(); got != filepath.Join(tmp, ".ssh", "config") {
		t.Errorf("mainConfigPath() = %q, want %q", got, filepath.Join(tmp, ".ssh", "config"))
	}
	if got := wombatConfigPath(); got != filepath.Join(tmp, ".ssh", "config.d", "wombat") {
		t.Errorf("wombatConfigPath() = %q, want %q", got, filepath.Join(tmp, ".ssh", "config.d", "wombat"))
	}
}

func TestPaths_emptyHome(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("USERPROFILE", "")
	if got := sshDir(); got != "" {
		t.Errorf("sshDir() = %q, want empty string", got)
	}
}
