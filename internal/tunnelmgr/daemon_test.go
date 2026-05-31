package tunnelmgr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTempAppHome(t *testing.T) {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("APPDATA", filepath.Join(tmp, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(tmp, "AppData", "Local"))
}

func TestPidDir(t *testing.T) {
	setupTempAppHome(t)
	dir, err := PidDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(dir, "pids") {
		t.Errorf("expected path to contain 'pids', got %q", dir)
	}
	if !strings.Contains(dir, "wombat") {
		t.Errorf("expected path to contain 'wombat', got %q", dir)
	}
}

func TestLogDir(t *testing.T) {
	setupTempAppHome(t)
	dir, err := LogDir()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(dir, "logs") {
		t.Errorf("expected path to contain 'logs', got %q", dir)
	}
	if !strings.Contains(dir, "wombat") {
		t.Errorf("expected path to contain 'wombat', got %q", dir)
	}
}

func TestPidFilePath(t *testing.T) {
	setupTempAppHome(t)
	path, err := pidFilePath("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !filepath.IsAbs(path) {
		t.Error("expected absolute path")
	}
	if filepath.Base(path) != "web.pid" {
		t.Errorf("expected web.pid, got %s", filepath.Base(path))
	}
}

func TestWritePidFile_and_ReadPid(t *testing.T) {
	setupTempAppHome(t)
	if err := WritePidFile("web"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pid, err := ReadPid("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("expected pid %d, got %d", os.Getpid(), pid)
	}
}

func TestReadPid_missing(t *testing.T) {
	setupTempAppHome(t)
	_, err := ReadPid("ghost")
	if err == nil {
		t.Error("expected error for missing pid file")
	}
}

func TestRemovePidFile(t *testing.T) {
	setupTempAppHome(t)
	if err := WritePidFile("web"); err != nil {
		t.Fatal(err)
	}
	if err := RemovePidFile("web"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, err := ReadPid("web")
	if err == nil {
		t.Error("expected error after removal")
	}
}

func TestOpenLogFile(t *testing.T) {
	setupTempAppHome(t)
	f, err := OpenLogFile("web")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	f.Close()

	path, _ := LogFilePath("web")
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected log file to exist: %v", err)
	}
}

func TestRemoveLogFile(t *testing.T) {
	setupTempAppHome(t)
	f, _ := OpenLogFile("web")
	f.Close()

	if err := RemoveLogFile("web"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	path, _ := LogFilePath("web")
	if _, err := os.Stat(path); err == nil {
		t.Error("expected log file to be removed")
	}
}

func TestIsRunning_notRunning(t *testing.T) {
	setupTempAppHome(t)
	if IsRunning("ghost") {
		t.Error("expected false for non-running tunnel")
	}
}

func TestIsRunning_stalePidCleanup(t *testing.T) {
	setupTempAppHome(t)
	// Write a PID for a process that definitely doesn't exist (PID 99999)
	path, _ := pidFilePath("stale")
	os.MkdirAll(filepath.Dir(path), 0755)
	os.WriteFile(path, []byte("99999"), 0600)

	if IsRunning("stale") {
		t.Error("expected false for stale pid")
	}
	if _, err := os.Stat(path); err == nil {
		t.Error("expected stale pid file to be cleaned up")
	}
}

func TestStopDaemon_notRunning(t *testing.T) {
	setupTempAppHome(t)
	err := StopDaemon("ghost")
	if err == nil {
		t.Error("expected error for non-running tunnel")
	}
}

func TestRestartTunnel(t *testing.T) {
	setupTempAppHome(t)
	calls := 0
	start := func(string) error {
		calls++
		return nil
	}
	if err := RestartTunnel("any", start); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 start call, got %d", calls)
	}
}

func TestRestartTunnel_startFails(t *testing.T) {
	setupTempAppHome(t)
	start := func(string) error { return os.ErrNotExist }
	if err := RestartTunnel("any", start); err == nil {
		t.Error("expected error when start fails")
	}
}

func TestTrayPidFilePath(t *testing.T) {
	setupTempAppHome(t)
	path, err := TrayPidFilePath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filepath.Base(path) != "tray-daemon.pid" {
		t.Errorf("expected tray-daemon.pid, got %s", filepath.Base(path))
	}
}

func TestWriteTrayPidFile_and_RemoveTrayPidFile(t *testing.T) {
	setupTempAppHome(t)
	if err := WriteTrayPidFile(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := RemoveTrayPidFile(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	path, _ := TrayPidFilePath()
	if _, err := os.Stat(path); err == nil {
		t.Error("expected tray pid file to be removed")
	}
}

func TestIsTrayRunning_notRunning(t *testing.T) {
	setupTempAppHome(t)
	if IsTrayRunning() {
		t.Error("expected false for non-running tray")
	}
}
