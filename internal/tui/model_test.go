package tui

import (
	"testing"

	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/models"
)

func TestNewModel_emptyConfigStartsInTableView(t *testing.T) {
	cfg := core.DefaultConfig()
	m := NewModel(cfg)

	if m.view != "table" {
		t.Errorf("expected view %q, got %q", "table", m.view)
	}
	if m.activeTab != "tunnels" {
		t.Errorf("expected activeTab %q, got %q", "tunnels", m.activeTab)
	}
}

func TestNewModelWithEdit_emptyNameStartsInTableView(t *testing.T) {
	cfg := core.Config{
		Tunnels: []models.Tunnel{
			{Name: "web", HostName: "server1", LocalPort: 8080, RemoteHost: "localhost", RemotePort: 80},
		},
	}
	m := NewModelWithEdit(cfg, "")

	if m.view != "table" {
		t.Errorf("expected view %q, got %q", "table", m.view)
	}
}

func TestNewModelWithEdit_matchingTunnelOpensEditForm(t *testing.T) {
	cfg := core.Config{
		Tunnels: []models.Tunnel{
			{Name: "api", HostName: "server1", LocalPort: 3000, RemoteHost: "localhost", RemotePort: 3000},
			{Name: "web", HostName: "server2", LocalPort: 8080, RemoteHost: "localhost", RemotePort: 80},
		},
	}
	m := NewModelWithEdit(cfg, "web")

	if m.view != "tunnel_form" {
		t.Errorf("expected view %q, got %q", "tunnel_form", m.view)
	}
	if m.formIsCreate {
		t.Error("expected formIsCreate to be false when editing")
	}
	if m.editingTunnel == nil {
		t.Fatal("expected editingTunnel to be set")
	}
	if m.editingTunnel.Name != "web" {
		t.Errorf("expected editing tunnel name %q, got %q", "web", m.editingTunnel.Name)
	}
	if m.editingTunnel.HostName != "server2" {
		t.Errorf("expected editing tunnel host %q, got %q", "server2", m.editingTunnel.HostName)
	}
	// Cursor should be positioned on the matching tunnel.
	if m.tunnelTable.Cursor() != 1 {
		t.Errorf("expected cursor at index 1, got %d", m.tunnelTable.Cursor())
	}
}

func TestNewModelWithEdit_unknownTunnelFallsBackToTableView(t *testing.T) {
	cfg := core.Config{
		Tunnels: []models.Tunnel{
			{Name: "api", HostName: "server1", LocalPort: 3000, RemoteHost: "localhost", RemotePort: 3000},
		},
	}
	m := NewModelWithEdit(cfg, "missing")

	if m.view != "table" {
		t.Errorf("expected view %q for unknown tunnel, got %q", "table", m.view)
	}
}
