package tui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/notify"
	"github.com/niklucky/wombat/internal/platform"
	"github.com/niklucky/wombat/internal/sshutil"
	"github.com/niklucky/wombat/internal/tunnelmgr"
)

// Model is the root Bubble Tea model.
type Model struct {
	config          core.Config
	activeTab       string
	view            string

	tunnelTable     table.Model
	hostTable       table.Model

	formInputs      []textinput.Model
	formFocus       int
	formIsCreate    bool
	editingTunnel   *core.Tunnel

	confirmItem     string
	SelectedHost    *core.Host
	RestartRequired bool
}

// NewModel creates a new TUI model with the given config.
func NewModel(cfg core.Config) Model {
	m := Model{
		config:    cfg,
		activeTab: "tunnels",
		view:      "table",
	}
	m.initTables()
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m.view {
	case "tunnel_form":
		return m.formUpdate(msg)
	case "settings_form":
		return m.settingsFormUpdate(msg)
	case "confirm_delete":
		return m.confirmUpdate(msg)
	default:
		return m.tableUpdate(msg)
	}
}

// View implements tea.Model.
func (m Model) View() string {
	switch m.view {
	case "tunnel_form":
		return m.formView()
	case "settings_form":
		return m.settingsFormView()
	case "confirm_delete":
		return renderConfirmDelete(m.confirmItem)
	default:
		return m.renderTableView()
	}
}

func (m *Model) tableUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			if m.activeTab == "tunnels" {
				m.activeTab = "hosts"
			} else {
				m.activeTab = "tunnels"
			}
		case "shift+tab":
			if m.activeTab == "hosts" {
				m.activeTab = "tunnels"
			} else {
				m.activeTab = "hosts"
			}
		case "i":
			// Open tray in background
			exe, err := os.Executable()
			if err != nil {
				exe = "wombat"
			}
			cmd := exec.Command(exe, "tray-daemon")
			cmd.SysProcAttr = platform.DaemonSysProcAttr()
			_ = cmd.Start()
		case "a":
			if m.activeTab == "tunnels" {
				m.initForm(true, nil)
			}
		case "n":
			if m.activeTab == "hosts" {
				// TODO: add host form
			}
		case "s":
			m.initSettingsForm()
		case "t":
			if m.activeTab == "hosts" {
				m.testSelectedHost()
			}
		case "enter":
			if m.activeTab == "tunnels" {
				if len(m.config.Tunnels) > 0 {
					idx := m.tunnelTable.Cursor()
					if idx < len(m.config.Tunnels) {
						t := m.config.Tunnels[idx]
						m.initForm(false, &t)
					}
				}
			} else {
				if len(m.config.Hosts) > 0 {
					idx := m.hostTable.Cursor()
					if idx < len(m.config.Hosts) {
						h := m.config.Hosts[idx]
						m.SelectedHost = &h
						return m, tea.Quit
					}
				}
			}
		case "c":
			if m.activeTab == "tunnels" {
				m.startSelectedTunnel()
			}
		case "d":
			if m.activeTab == "tunnels" {
				m.stopSelectedTunnel()
			}
		case "backspace":
			if m.activeTab == "tunnels" {
				if len(m.config.Tunnels) > 0 {
					idx := m.tunnelTable.Cursor()
					if idx < len(m.config.Tunnels) {
						m.confirmItem = m.config.Tunnels[idx].Name
						m.view = "confirm_delete"
					}
				}
			}
		case "up", "k":
			m.activeTable().MoveUp(1)
		case "down", "j":
			m.activeTable().MoveDown(1)
		}
	}
	return m, nil
}

func (m *Model) confirmUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y":
			m.config.RemoveTunnel(m.confirmItem)
			_ = m.config.Save()
			m.refreshTables()
			m.view = "table"
		case "n", "esc":
			m.view = "table"
		}
	}
	return m, nil
}

func (m *Model) testSelectedHost() {
	if len(m.config.Hosts) == 0 {
		return
	}
	idx := m.hostTable.Cursor()
	if idx >= len(m.config.Hosts) {
		return
	}
	h := m.config.Hosts[idx]
	if err := sshutil.TestConnection(h); err != nil {
		notify.Alert("Wombat", fmt.Sprintf("%s: connection failed", h.Name))
	} else {
		notify.Notify("Wombat", fmt.Sprintf("%s: connection OK", h.Name))
	}
}

func (m *Model) startSelectedTunnel() {
	if len(m.config.Tunnels) == 0 {
		return
	}
	idx := m.tunnelTable.Cursor()
	if idx >= len(m.config.Tunnels) {
		return
	}
	t := m.config.Tunnels[idx]
	if tunnelmgr.IsRunning(t.Name) {
		return
	}
	exe, err := os.Executable()
	if err != nil {
		exe = "wombat"
	}
	cmd := exec.Command(exe, "tunnel-start", t.Name)
	cmd.SysProcAttr = platform.DaemonSysProcAttr()
	if err := cmd.Start(); err != nil {
		notify.Alert("Wombat", fmt.Sprintf("Failed to start %s: %v", t.Name, err))
	} else {
		notify.Notify("Wombat", fmt.Sprintf("Tunnel %s started", t.Name))
		m.refreshTunnelRows()
	}
}

func (m *Model) stopSelectedTunnel() {
	if len(m.config.Tunnels) == 0 {
		return
	}
	idx := m.tunnelTable.Cursor()
	if idx >= len(m.config.Tunnels) {
		return
	}
	t := m.config.Tunnels[idx]
	if !tunnelmgr.IsRunning(t.Name) {
		return
	}
	if err := tunnelmgr.StopDaemon(t.Name); err != nil {
		notify.Alert("Wombat", fmt.Sprintf("Failed to stop %s: %v", t.Name, err))
	} else {
		notify.Notify("Wombat", fmt.Sprintf("Tunnel %s stopped", t.Name))
		m.refreshTunnelRows()
	}
}
