package sshconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureSetup_createsNewConfig(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	if err := EnsureSetup(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	mainPath := mainConfigPath()
	data, err := os.ReadFile(mainPath)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "Include config.d/wombat") {
		t.Errorf("expected include line, got: %q", content)
	}
	if !strings.Contains(content, includeMarker) {
		t.Errorf("expected marker, got: %q", content)
	}

	// Idempotent second call
	if err := EnsureSetup(); err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	data2, _ := os.ReadFile(mainPath)
	if string(data2) != content {
		t.Error("expected idempotent setup")
	}
}

func TestEnsureSetup_prependsToExistingConfig(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	if err := os.MkdirAll(filepath.Join(tmp, ".ssh"), 0700); err != nil {
		t.Fatal(err)
	}
	existing := "Host existing\n  HostName 1.2.3.4\n"
	if err := os.WriteFile(mainConfigPath(), []byte(existing), 0600); err != nil {
		t.Fatal(err)
	}

	if err := EnsureSetup(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(mainConfigPath())
	content := string(data)
	if !strings.HasPrefix(content, includeMarker) {
		t.Error("expected marker at beginning")
	}
	if !strings.Contains(content, "Host existing") {
		t.Error("expected existing host preserved")
	}
}
