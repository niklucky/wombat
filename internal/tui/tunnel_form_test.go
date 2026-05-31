package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/niklucky/wombat/internal/core"
)

func TestInitForm_defaults(t *testing.T) {
	m := Model{}
	m.initForm(true, nil)
	if len(m.formInputs) != 5 {
		t.Fatalf("expected 5 inputs, got %d", len(m.formInputs))
	}
	if m.formInputs[3].Value() != "localhost" {
		t.Errorf("expected default remote host localhost, got %s", m.formInputs[3].Value())
	}
}

func TestInitForm_prefills(t *testing.T) {
	m := Model{}
	tu := &core.Tunnel{Name: "api", HostName: "web", LocalPort: 3000, RemoteHost: "localhost", RemotePort: 3000}
	m.initForm(false, tu)
	if m.formInputs[0].Value() != "api" {
		t.Errorf("expected name api, got %s", m.formInputs[0].Value())
	}
	if m.formInputs[1].Value() != "web" {
		t.Errorf("expected host web, got %s", m.formInputs[1].Value())
	}
	if m.formInputs[2].Value() != "3000" {
		t.Errorf("expected local port 3000, got %s", m.formInputs[2].Value())
	}
	if m.formInputs[4].Value() != "3000" {
		t.Errorf("expected remote port 3000, got %s", m.formInputs[4].Value())
	}
}

func TestFormUpdate_enterOnHostAliasOpensSource(t *testing.T) {
	m := Model{view: "tunnel_form"}
	m.initForm(true, nil)
	m.formFocus = 1
	m.formInputs[0].SetValue("mytunnel")
	m.formInputs[2].SetValue("8080")

	m.formUpdate(tea.KeyMsg{Type: tea.KeyEnter})
	if m.view != "host_source" {
		t.Errorf("expected view host_source, got %s", m.view)
	}
	if m.returnView != "tunnel_form" {
		t.Errorf("expected returnView tunnel_form, got %s", m.returnView)
	}
	if len(m.savedFormInputs) != 5 {
		t.Fatalf("expected saved inputs, got %d", len(m.savedFormInputs))
	}
	if m.savedFormInputs[0].Value() != "mytunnel" {
		t.Errorf("expected saved name mytunnel, got %s", m.savedFormInputs[0].Value())
	}
}

func TestSaveTunnelFormState_andRestore(t *testing.T) {
	m := Model{view: "tunnel_form", formIsCreate: true}
	m.initForm(true, nil)
	m.formFocus = 2
	m.formInputs[0].SetValue("t1")
	m.formInputs[1].SetValue("h1")

	m.saveTunnelFormState()
	m.initHostForm(true, nil) // overwrite formInputs
	m.restoreTunnelForm()

	if m.formFocus != 2 {
		t.Errorf("expected restored focus 2, got %d", m.formFocus)
	}
	if m.formIsCreate != true {
		t.Error("expected restored formIsCreate true")
	}
	if m.formInputs[0].Value() != "t1" {
		t.Errorf("expected restored name t1, got %s", m.formInputs[0].Value())
	}
	if m.formInputs[1].Value() != "h1" {
		t.Errorf("expected restored host h1, got %s", m.formInputs[1].Value())
	}
}

func TestSaveForm_validation(t *testing.T) {
	m := Model{config: core.DefaultConfig()}
	m.initForm(true, nil)
	if err := m.saveForm(); err == nil {
		t.Error("expected error for empty required fields")
	}
}

func TestSaveForm_defaultsRemoteHost(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp+"/.config")

	m := Model{config: core.DefaultConfig()}
	m.initForm(true, nil)
	m.formInputs[0].SetValue("t1")
	m.formInputs[1].SetValue("h1")
	m.formInputs[2].SetValue("8080")
	m.formInputs[4].SetValue("80")

	if err := m.saveForm(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.config.Tunnels) != 1 {
		t.Fatalf("expected 1 tunnel, got %d", len(m.config.Tunnels))
	}
	if m.config.Tunnels[0].RemoteHost != "localhost" {
		t.Errorf("expected remote host localhost, got %s", m.config.Tunnels[0].RemoteHost)
	}
}

func TestSaveForm_create(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp+"/.config")

	m := Model{config: core.DefaultConfig()}
	m.initForm(true, nil)
	m.formInputs[0].SetValue("web")
	m.formInputs[1].SetValue("server")
	m.formInputs[2].SetValue("8080")
	m.formInputs[3].SetValue("localhost")
	m.formInputs[4].SetValue("80")

	if err := m.saveForm(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.config.Tunnels) != 1 || m.config.Tunnels[0].Name != "web" {
		t.Errorf("expected tunnel web, got %+v", m.config.Tunnels)
	}
}

func TestSaveForm_editInPlace(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp+"/.config")

	cfg := core.DefaultConfig()
	cfg.AddTunnel(core.Tunnel{Name: "web", HostName: "old", LocalPort: 80, RemotePort: 80})
	m := Model{config: cfg}
	m.initForm(false, &cfg.Tunnels[0])
	m.formInputs[1].SetValue("newhost")

	if err := m.saveForm(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.config.Tunnels) != 1 {
		t.Fatalf("expected 1 tunnel, got %d", len(m.config.Tunnels))
	}
	if m.config.Tunnels[0].HostName != "newhost" {
		t.Errorf("expected host name newhost, got %s", m.config.Tunnels[0].HostName)
	}
}

func TestSaveForm_rename(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", tmp+"/.config")

	cfg := core.DefaultConfig()
	cfg.AddTunnel(core.Tunnel{Name: "old", HostName: "server", LocalPort: 80, RemotePort: 80})
	m := Model{config: cfg}
	m.initForm(false, &cfg.Tunnels[0])
	m.formInputs[0].SetValue("new")

	if err := m.saveForm(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.config.Tunnels) != 1 {
		t.Fatalf("expected 1 tunnel, got %d", len(m.config.Tunnels))
	}
	if m.config.Tunnels[0].Name != "new" {
		t.Errorf("expected name new, got %s", m.config.Tunnels[0].Name)
	}
}
