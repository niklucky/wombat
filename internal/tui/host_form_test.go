package tui

import (
	"path/filepath"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/niklucky/wombat/internal/core"
)

func TestInitHostForm_defaults(t *testing.T) {
	m := Model{}
	m.initHostForm(true, nil)
	if len(m.formInputs) != 6 {
		t.Fatalf("expected 6 inputs, got %d", len(m.formInputs))
	}
	if m.formInputs[5].Value() != "yes" {
		t.Errorf("expected 'Save to SSH-hosts' default yes, got %s", m.formInputs[5].Value())
	}
	if m.formInputs[3].Value() != "22" {
		t.Errorf("expected default port 22, got %s", m.formInputs[3].Value())
	}
}

func TestInitHostForm_prefills(t *testing.T) {
	m := Model{}
	h := &core.Host{Name: "web", Address: "10.0.0.1", User: "admin", Port: 2222, KeyPath: "~/.ssh/web"}
	m.initHostForm(false, h)
	if m.formInputs[0].Value() != "web" {
		t.Errorf("expected name web, got %s", m.formInputs[0].Value())
	}
	if m.formInputs[1].Value() != "10.0.0.1" {
		t.Errorf("expected address 10.0.0.1, got %s", m.formInputs[1].Value())
	}
	if m.formInputs[2].Value() != "admin" {
		t.Errorf("expected user admin, got %s", m.formInputs[2].Value())
	}
	if m.formInputs[3].Value() != "2222" {
		t.Errorf("expected port 2222, got %s", m.formInputs[3].Value())
	}
	if m.formInputs[4].Value() != "~/.ssh/web" {
		t.Errorf("expected key ~/.ssh/web, got %s", m.formInputs[4].Value())
	}
}

func TestSaveHostForm_toggleNoDoesNotSave(t *testing.T) {
	m := Model{config: core.DefaultConfig()}
	m.initHostForm(true, nil)
	m.formInputs[0].SetValue("temp")
	m.formInputs[1].SetValue("10.0.0.1")
	m.formInputs[2].SetValue("root")
	m.formBools[5] = false
	m.formInputs[5].SetValue(boolToYesNo(false))

	if err := m.saveHostForm(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.config.Hosts) != 0 {
		t.Errorf("expected 0 hosts, got %d", len(m.config.Hosts))
	}
}

func TestSaveHostForm_toggleYesSaves(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("USERPROFILE", tmp)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, ".config"))
	t.Setenv("APPDATA", filepath.Join(tmp, "AppData", "Roaming"))
	t.Setenv("LOCALAPPDATA", filepath.Join(tmp, "AppData", "Local"))

	m := Model{config: core.DefaultConfig()}
	m.initHostForm(true, nil)
	m.formInputs[0].SetValue("web")
	m.formInputs[1].SetValue("10.0.0.1")
	m.formInputs[2].SetValue("root")
	m.formBools[5] = true
	m.formInputs[5].SetValue(boolToYesNo(true))

	if err := m.saveHostForm(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.config.Hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(m.config.Hosts))
	}
	if m.config.Hosts[0].Name != "web" {
		t.Errorf("expected name web, got %s", m.config.Hosts[0].Name)
	}
}

func TestSaveHostForm_validation(t *testing.T) {
	m := Model{config: core.DefaultConfig()}
	m.initHostForm(true, nil)
	m.formBools[5] = true
	m.formInputs[5].SetValue(boolToYesNo(true))

	if err := m.saveHostForm(); err == nil {
		t.Error("expected error for empty required fields")
	}
}

func TestSaveHostForm_invalidPort(t *testing.T) {
	m := Model{config: core.DefaultConfig()}
	m.initHostForm(true, nil)
	m.formInputs[0].SetValue("web")
	m.formInputs[1].SetValue("10.0.0.1")
	m.formInputs[2].SetValue("root")
	m.formInputs[3].SetValue("abc")
	m.formBools[5] = true
	m.formInputs[5].SetValue(boolToYesNo(true))

	if err := m.saveHostForm(); err == nil {
		t.Error("expected error for invalid port")
	}
}

func TestSaveHostForm_portOutOfRange(t *testing.T) {
	m := Model{config: core.DefaultConfig()}
	m.initHostForm(true, nil)
	m.formInputs[0].SetValue("web")
	m.formInputs[1].SetValue("10.0.0.1")
	m.formInputs[2].SetValue("root")
	m.formInputs[3].SetValue("99999")
	m.formBools[5] = true
	m.formInputs[5].SetValue(boolToYesNo(true))

	if err := m.saveHostForm(); err == nil {
		t.Error("expected error for out-of-range port")
	}
}

func TestHostFormUpdate_spaceTogglesSaveField(t *testing.T) {
	m := Model{}
	m.initHostForm(true, nil)
	m.formFocus = 5

	m.hostFormUpdate(tea.KeyMsg{Type: tea.KeySpace})
	if m.formInputs[5].Value() != boolToYesNo(false) {
		t.Errorf("expected no after toggle, got %s", m.formInputs[5].Value())
	}
	if m.formBools[5] != false {
		t.Errorf("expected bool false after toggle, got %v", m.formBools[5])
	}

	m.hostFormUpdate(tea.KeyMsg{Type: tea.KeySpace})
	if m.formInputs[5].Value() != boolToYesNo(true) {
		t.Errorf("expected yes after second toggle, got %s", m.formInputs[5].Value())
	}
	if m.formBools[5] != true {
		t.Errorf("expected bool true after second toggle, got %v", m.formBools[5])
	}
}

func TestHostFormUpdate_escFromTunnelFlowReturnsToHostSource(t *testing.T) {
	m := Model{returnView: "tunnel_form", view: "host_form"}
	m.initHostForm(true, nil)
	m.hostFormUpdate(tea.KeyMsg{Type: tea.KeyEsc})
	if m.view != "host_source" {
		t.Errorf("expected view host_source, got %s", m.view)
	}
}

func TestHostFormUpdate_escDirectReturnsToTable(t *testing.T) {
	m := Model{returnView: "", view: "host_form"}
	m.initHostForm(true, nil)
	m.hostFormUpdate(tea.KeyMsg{Type: tea.KeyEsc})
	if m.view != "table" {
		t.Errorf("expected view table, got %s", m.view)
	}
}

func TestHostFormUpdate_ctrlSFromTunnelFlowRestoresForm(t *testing.T) {
	m := Model{returnView: "tunnel_form", view: "host_form"}
	m.initHostForm(true, nil)
	m.formInputs[0].SetValue("web")
	m.formInputs[1].SetValue("10.0.0.1")
	m.formInputs[2].SetValue("root")
	m.formBools[5] = false
	m.formInputs[5].SetValue(boolToYesNo(false))

	// Save tunnel form state so restore works
	m.savedFormInputs = make([]textinput.Model, 5)
	for i := range m.savedFormInputs {
		m.savedFormInputs[i] = textinput.New()
	}
	m.savedFormInputs[1].SetValue("oldhost")
	m.savedFormFocus = 1

	m.hostFormUpdate(tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.view != "tunnel_form" {
		t.Errorf("expected view tunnel_form, got %s", m.view)
	}
	if m.formInputs[1].Value() != "web" {
		t.Errorf("expected host alias web, got %s", m.formInputs[1].Value())
	}
}
