package tui

import (
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func TestHostSourceUpdate_cursorBounds(t *testing.T) {
	m := Model{hostSourceCursor: 0}
	m.hostSourceUpdate(tea.KeyMsg{Type: tea.KeyUp})
	if m.hostSourceCursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.hostSourceCursor)
	}

	m.hostSourceUpdate(tea.KeyMsg{Type: tea.KeyDown})
	if m.hostSourceCursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.hostSourceCursor)
	}

	m.hostSourceUpdate(tea.KeyMsg{Type: tea.KeyDown})
	if m.hostSourceCursor != 1 {
		t.Errorf("expected cursor 1, got %d", m.hostSourceCursor)
	}

	m.hostSourceUpdate(tea.KeyMsg{Type: tea.KeyUp})
	if m.hostSourceCursor != 0 {
		t.Errorf("expected cursor 0, got %d", m.hostSourceCursor)
	}
}

func TestHostSourceUpdate_enterSelectsExistingHosts(t *testing.T) {
	m := Model{hostSourceCursor: 0, view: "host_source"}
	m.hostSourceUpdate(tea.KeyMsg{Type: tea.KeyEnter})
	if m.view != "host_select" {
		t.Errorf("expected view host_select, got %s", m.view)
	}
}

func TestHostSourceUpdate_enterSelectsCustomHost(t *testing.T) {
	m := Model{hostSourceCursor: 1, view: "host_source"}
	m.hostSourceUpdate(tea.KeyMsg{Type: tea.KeyEnter})
	if m.view != "host_form" {
		t.Errorf("expected view host_form, got %s", m.view)
	}
}

func TestHostSourceUpdate_escReturnsToTable(t *testing.T) {
	m := Model{view: "host_source", returnView: ""}
	m.hostSourceUpdate(tea.KeyMsg{Type: tea.KeyEsc})
	if m.view != "table" {
		t.Errorf("expected view table, got %s", m.view)
	}
}

func TestHostSourceUpdate_escRestoresTunnelForm(t *testing.T) {
	m := Model{view: "host_source", returnView: "tunnel_form"}
	m.savedFormInputs = make([]textinput.Model, 5)
	for i := range m.savedFormInputs {
		m.savedFormInputs[i] = textinput.New()
	}
	m.savedFormFocus = 1
	m.hostSourceUpdate(tea.KeyMsg{Type: tea.KeyEsc})
	if m.view != "tunnel_form" {
		t.Errorf("expected view tunnel_form, got %s", m.view)
	}
	if m.returnView != "" {
		t.Error("expected returnView cleared")
	}
}
