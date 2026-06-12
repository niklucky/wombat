package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/locales"
	"github.com/niklucky/wombat/internal/models"
)

func TestHostSelectUpdate_escReturnsToHostSource(t *testing.T) {
	m := Model{view: "host_select"}
	m.hostSelectUpdate(tea.KeyMsg{Type: tea.KeyEsc})
	if m.view != "host_source" {
		t.Errorf("expected view host_source, got %s", m.view)
	}
}

func TestHostSelectUpdate_enterSelectsHost(t *testing.T) {
	m := Model{
		view:       "host_select",
		returnView: "tunnel_form",
		config: core.Config{
			Hosts: []models.Host{
				{Name: "web", Address: "10.0.0.1", User: "root"},
				{Name: "db", Address: "10.0.0.2", User: "root"},
			},
		},
	}
	m.initTables()
	m.hostTable.SetCursor(1)

	m.savedFormInputs = make([]textinput.Model, 5)
	for i := range m.savedFormInputs {
		m.savedFormInputs[i] = textinput.New()
	}
	m.savedFormFocus = 1
	m.savedFormIsCreate = true

	m.hostSelectUpdate(tea.KeyMsg{Type: tea.KeyEnter})
	if m.view != "tunnel_form" {
		t.Errorf("expected view tunnel_form, got %s", m.view)
	}
	if m.formInputs[1].Value() != "db" {
		t.Errorf("expected host alias db, got %s", m.formInputs[1].Value())
	}
	if m.returnView != "" {
		t.Error("expected returnView cleared")
	}
}

func TestHostSelectUpdate_enterNoHosts(t *testing.T) {
	m := Model{
		view:       "host_select",
		returnView: "tunnel_form",
		config:     core.Config{Hosts: []models.Host{}},
	}
	m.initTables()
	m.hostSelectUpdate(tea.KeyMsg{Type: tea.KeyEnter})
	if m.view != "host_select" {
		t.Errorf("expected view to remain host_select, got %s", m.view)
	}
}

func TestHostSelectView_emptyHosts(t *testing.T) {
	_ = locales.SetLanguage("en")
	m := Model{config: core.Config{Hosts: []models.Host{}}}
	m.initTables()
	s := m.hostSelectView()
	if !strings.Contains(s, locales.T("messages.noHostsConfigured")) {
		t.Error("expected 'No hosts configured' message")
	}
}

