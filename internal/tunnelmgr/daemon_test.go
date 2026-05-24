package tunnelmgr

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

func setupTempHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	return tmp
}

func TestRestartTunnel_nonRunningCallsStart(t *testing.T) {
	setupTempHome(t)

	called := false
	start := func(name string) error {
		called = true
		if name != "test-tunnel" {
			t.Errorf("expected name %q, got %q", "test-tunnel", name)
		}
		return nil
	}

	err := RestartTunnel("test-tunnel", start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected start to be called")
	}
}

func TestRestartTunnel_runningStopsThenStarts(t *testing.T) {
	setupTempHome(t)

	// Spawn a child process that we can safely stop.
	if runtime.GOOS == "windows" {
		t.Skip("skipping: sleep binary is not portable on Windows")
	}
	cmd := exec.Command("sleep", "30")
	if err := cmd.Start(); err != nil {
		t.Skip("cannot spawn test process:", err)
	}
	defer func() {
		if err := cmd.Process.Kill(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			t.Logf("failed to kill test process: %v", err)
		}
		_ = cmd.Wait()
	}()

	// Write its PID to the tunnel PID file.
	pidDir, err := PidDir()
	if err != nil {
		t.Fatalf("failed to get pid dir: %v", err)
	}
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		t.Fatalf("failed to create pid dir: %v", err)
	}
	pidFile := filepath.Join(pidDir, "test-tunnel.pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0600); err != nil {
		t.Fatalf("failed to write pid file: %v", err)
	}

	called := false
	start := func(name string) error {
		called = true
		return nil
	}

	err = RestartTunnel("test-tunnel", start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Error("expected start to be called")
	}

	// Verify PID file was removed.
	if _, err := os.Stat(pidFile); !os.IsNotExist(err) {
		t.Error("expected PID file to be removed after stop")
	}
}

func TestRestartTunnel_startFailureReturnsError(t *testing.T) {
	setupTempHome(t)

	start := func(name string) error {
		return os.ErrNotExist
	}

	err := RestartTunnel("test-tunnel", start)
	if err == nil {
		t.Fatal("expected error when start fails")
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected error to wrap os.ErrNotExist, got: %v", err)
	}
}
