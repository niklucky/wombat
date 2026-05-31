package sshconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niklucky/wombat/internal/models"
)

func TestWriteHosts_format(t *testing.T) {
	tmp := t.TempDir()
	setTestHome(t, tmp)

	if err := os.MkdirAll(filepath.Join(tmp, ".ssh", "config.d"), 0700); err != nil {
		t.Fatal(err)
	}

	hosts := []models.Host{
		{Name: "web", Address: "10.0.0.1", User: "admin", Port: 2222, KeyPath: "~/.ssh/web"},
		{Name: "db", Address: "10.0.0.2", User: "root"},
	}

	if err := WriteHosts(hosts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(wombatConfigPath())
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	content := string(data)

	if !strings.Contains(content, "Host web") {
		t.Error("expected 'Host web' in output")
	}
	if !strings.Contains(content, "HostName 10.0.0.1") {
		t.Error("expected 'HostName 10.0.0.1' in output")
	}
	if !strings.Contains(content, "User admin") {
		t.Error("expected 'User admin' in output")
	}
	if !strings.Contains(content, "Port 2222") {
		t.Error("expected 'Port 2222' in output")
	}
	if !strings.Contains(content, "IdentityFile ~/.ssh/web") {
		t.Error("expected 'IdentityFile ~/.ssh/web' in output")
	}
	if !strings.Contains(content, "Host db") {
		t.Error("expected 'Host db' in output")
	}
	if strings.Contains(content, "Port") && strings.Contains(content, "Host db") {
		// db has default port 0, so Port should NOT appear for it
		lines := strings.Split(content, "\n")
		inDb := false
		for _, line := range lines {
			if strings.HasPrefix(line, "Host db") {
				inDb = true
				continue
			}
			if strings.HasPrefix(line, "Host ") {
				inDb = false
			}
			if inDb && strings.HasPrefix(line, "Port") {
				t.Error("expected no Port line for default-port host db")
			}
		}
	}
}
