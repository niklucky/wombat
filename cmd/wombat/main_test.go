package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// setupTempHome redirects all file-system-dependent calls to a temporary
// directory for the duration of the test. It sets HOME (used by
// sshDir/UserHomeDir) and XDG_CONFIG_HOME (used by UserConfigDir on Linux)
// so that EnsureSetup and core.AppHome() never touch the real user home.
func setupTempHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()

	t.Setenv("HOME", tmp)
	// On Linux, os.UserConfigDir uses $XDG_CONFIG_HOME when set.
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	// On Windows, os.UserConfigDir uses %APPDATA% and os.UserHomeDir uses %USERPROFILE%.
	t.Setenv("APPDATA", filepath.Join(tmp, "AppData", "Roaming"))
	t.Setenv("USERPROFILE", tmp)

	// Pre-create SSH directories that EnsureSetup expects.
	sshDir := filepath.Join(tmp, ".ssh")
	if err := os.MkdirAll(filepath.Join(sshDir, "config.d"), 0700); err != nil {
		t.Fatalf("failed to create temp ssh dirs: %v", err)
	}
	// Minimal ~/.ssh/config already containing the include so EnsureSetup is a no-op.
	sshConfig := filepath.Join(sshDir, "config")
	if err := os.WriteFile(sshConfig, []byte("# Wombat managed hosts\nInclude config.d/wombat\n"), 0600); err != nil {
		t.Fatalf("failed to create temp ssh config: %v", err)
	}

	return tmp
}

// resetFlags resets all flags on cmd and its sub-commands to their default
// values and clears the "Changed" marker. This is necessary because cobra
// reuses the global command tree across Execute() calls within the same test
// binary, so flags set by a previous test would otherwise bleed into later
// tests, defeating required-flag validation.
func resetFlags(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		_ = f.Value.Set(f.DefValue)
	})
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		f.Changed = false
		_ = f.Value.Set(f.DefValue)
	})
	for _, child := range cmd.Commands() {
		resetFlags(child)
	}
}

// runWithSilence executes fn with cobra's error/usage output suppressed on
// rootCmd and restores the original values afterwards. It also resets all
// command flags before fn runs to prevent stale state from previous tests.
func runWithSilence(fn func()) {
	resetFlags(rootCmd)
	rootCmd.SilenceErrors = true
	rootCmd.SilenceUsage = true
	defer func() {
		rootCmd.SilenceErrors = false
		rootCmd.SilenceUsage = false
	}()
	fn()
}

// ---------------------------------------------------------------------------
// version command
// ---------------------------------------------------------------------------

func TestVersionCmd_versionVariableIsDevByDefault(t *testing.T) {
	// The package-level `version` variable defaults to "dev" in source builds.
	if version != "dev" {
		t.Errorf("expected default version %q, got %q", "dev", version)
	}
}

func TestVersionCmd_versionVariableNotEmpty(t *testing.T) {
	if version == "" {
		t.Error("version variable must not be empty")
	}
}

func TestVersionCmd_executesWithoutError(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	runWithSilence(func() {
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("unexpected error running version: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// list command
// ---------------------------------------------------------------------------

func TestListCmd_emptyConfigReturnsNoError(t *testing.T) {
	setupTempHome(t)

	rootCmd.SetArgs([]string{"list"})
	runWithSilence(func() {
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("expected no error for empty host list, got: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// add-host command – required-flag validation
// ---------------------------------------------------------------------------

func TestAddHostCmd_missingAddressFlagReturnsError(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"add-host", "myserver", "--user", "admin"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when --address flag is missing")
	}
}

func TestAddHostCmd_missingUserFlagReturnsError(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"add-host", "myserver", "--address", "10.0.0.1"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when --user flag is missing")
	}
}

func TestAddHostCmd_missingNameArgReturnsError(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"add-host", "--address", "10.0.0.1", "--user", "admin"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when name argument is missing for add-host")
	}
}

// ---------------------------------------------------------------------------
// remove-host argument validation
// ---------------------------------------------------------------------------

func TestRemoveHostCmd_requiresExactlyOneArg(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"remove-host"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when remove-host called with no arguments")
	}
}

// ---------------------------------------------------------------------------
// add-tunnel command – required-flag validation
// ---------------------------------------------------------------------------

func TestAddTunnelCmd_missingHostFlagReturnsError(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"add-tunnel", "mytunnel",
		"--local-port", "8080",
		"--remote-port", "80",
	})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when --host flag is missing for add-tunnel")
	}
}

func TestAddTunnelCmd_missingLocalPortReturnsError(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"add-tunnel", "mytunnel",
		"--host", "myserver",
		"--remote-port", "80",
	})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when --local-port flag is missing for add-tunnel")
	}
}

func TestAddTunnelCmd_missingRemotePortReturnsError(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"add-tunnel", "mytunnel",
		"--host", "myserver",
		"--local-port", "8080",
	})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when --remote-port flag is missing for add-tunnel")
	}
}

func TestAddTunnelCmd_missingNameArgReturnsError(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"add-tunnel",
		"--host", "myserver",
		"--local-port", "8080",
		"--remote-port", "80",
	})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when tunnel name argument is missing")
	}
}

// ---------------------------------------------------------------------------
// remove-tunnel argument validation
// ---------------------------------------------------------------------------

func TestRemoveTunnelCmd_requiresExactlyOneArg(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"remove-tunnel"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when remove-tunnel called with no arguments")
	}
}

// ---------------------------------------------------------------------------
// tunnel-list command
// ---------------------------------------------------------------------------

func TestTunnelListCmd_emptyConfigReturnsNoError(t *testing.T) {
	setupTempHome(t)

	rootCmd.SetArgs([]string{"tunnel-list"})
	runWithSilence(func() {
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("expected no error for empty tunnel list, got: %v", err)
		}
	})
}

// ---------------------------------------------------------------------------
// tunnel-stop command
// ---------------------------------------------------------------------------

func TestTunnelStopCmd_notRunningReturnsError(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"tunnel-stop", "ghost-tunnel"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Fatal("expected error when stopping a tunnel that is not running")
	}
	if !strings.Contains(err.Error(), "not running") {
		t.Errorf("expected error to contain %q, got: %q", "not running", err.Error())
	}
}

func TestTunnelStopCmd_errorMessageIncludesTunnelName(t *testing.T) {
	setupTempHome(t)

	const tunnelName = "my-special-tunnel"
	var err error
	rootCmd.SetArgs([]string{"tunnel-stop", tunnelName})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Fatal("expected error when stopping a non-existent tunnel")
	}
	if !strings.Contains(err.Error(), tunnelName) {
		t.Errorf("expected error message to include %q, got: %q", tunnelName, err.Error())
	}
}

// ---------------------------------------------------------------------------
// tunnel-start command
// ---------------------------------------------------------------------------

func TestTunnelStartCmd_requiresExactlyOneArg(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"tunnel-start"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when tunnel-start called with no arguments")
	}
}

// ---------------------------------------------------------------------------
// tunnel-restart command
// ---------------------------------------------------------------------------

func TestTunnelRestartCmd_requiresExactlyOneArg(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"tunnel-restart"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when tunnel-restart called with no arguments")
	}
}

func TestTunnelRestartCmd_unknownTunnelReturnsError(t *testing.T) {
	if runtime.GOOS == "windows" {
		// Spawning the test binary as a subprocess on Windows causes file-handle
		// locks that interfere with t.TempDir() cleanup. The restart logic is
		// covered by internal/tunnelmgr tests on all platforms.
		t.Skip("skipping on Windows: subprocess file-lock conflict")
	}
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"tunnel-restart", "nonexistent-tunnel"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Fatal("expected error when restarting an unknown tunnel")
	}
	if !strings.Contains(err.Error(), "failed to start") {
		t.Errorf("expected 'failed to start' in error, got: %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// tunnel-daemon command
// ---------------------------------------------------------------------------

func TestTunnelDaemonCmd_unknownTunnelReturnsError(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"tunnel-daemon", "nonexistent-tunnel"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Fatal("expected error for unknown tunnel in tunnel-daemon")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %q", err.Error())
	}
}

// ---------------------------------------------------------------------------
// connect command
// ---------------------------------------------------------------------------

func TestConnectCmd_unknownHostReturnsError(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"connect", "does-not-exist"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Fatal("expected error when connecting to unknown host")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %q", err.Error())
	}
}

func TestConnectCmd_requiresExactlyOneArg(t *testing.T) {
	setupTempHome(t)

	var err error
	rootCmd.SetArgs([]string{"connect"})
	runWithSilence(func() { err = rootCmd.Execute() })

	if err == nil {
		t.Error("expected error when connect called with no arguments")
	}
}

// ---------------------------------------------------------------------------
// tui command
// ---------------------------------------------------------------------------

func TestTuiCmd_isRegistered(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "tui" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'tui' command to be registered on rootCmd")
	}
}

func TestTuiCmd_hasEditTunnelFlag(t *testing.T) {
	f := tuiCmd.Flags().Lookup("edit-tunnel")
	if f == nil {
		t.Fatal("expected --edit-tunnel flag on tui command")
	}
	if f.DefValue != "" {
		t.Errorf("expected default value %q, got %q", "", f.DefValue)
	}
}

// ---------------------------------------------------------------------------
// Port default value in list output (regression)
// ---------------------------------------------------------------------------

// TestListCmd_portDefaultsTo22 verifies that a host with Port==0 is displayed
// with port 22 and the command completes without error.
func TestListCmd_portDefaultsTo22(t *testing.T) {
	tmp := setupTempHome(t)

	// Write a wombat SSH config entry without a Port directive.
	sshConfigD := filepath.Join(tmp, ".ssh", "config.d", "wombat")
	content := "Host myserver\n  HostName 10.0.0.1\n  User admin\n"
	if err := os.WriteFile(sshConfigD, []byte(content), 0600); err != nil {
		t.Fatalf("write wombat ssh config: %v", err)
	}

	// Capture stdout via os.Pipe because listCmd writes via fmt.Printf.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w

	rootCmd.SetArgs([]string{"list"})
	execErr := rootCmd.Execute()

	w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	if execErr != nil {
		t.Fatalf("list command failed: %v", execErr)
	}

	output := buf.String()
	if !strings.Contains(output, "myserver") {
		t.Errorf("expected output to contain host name 'myserver', got: %q", output)
	}
	if !strings.Contains(output, ":22") {
		t.Errorf("expected port 22 in output for host with no explicit port, got: %q", output)
	}
}
