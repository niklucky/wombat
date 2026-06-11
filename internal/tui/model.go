package tui

import (
	"fmt"
	"os"
	"os/exec"
	"time"

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
	config    core.Config
	activeTab string
	view      string

	tunnelTable table.Model
	hostTable   table.Model

	formInputs    []textinput.Model
	formFocus     int
	formIsCreate  bool
	editingTunnel *core.Tunnel

	formIsCreateHost bool
	editingHost      *core.Host

	confirmItem     string
	confirmItemType string
	SelectedHost    *core.Host
	RestartRequired bool

	// tunnel host-selection flow state
	returnView         string
	savedFormInputs    []textinput.Model
	savedFormFocus     int
	savedFormIsCreate  bool
	savedEditingTunnel *core.Tunnel
	hostSourceCursor   int
}

// NewModel creates a new TUI model with the given config.
func NewModel(cfg core.Config) Model {
	return NewModelWithEdit(cfg, "")
}

// NewModelWithEdit creates a new TUI model and, if editTunnelName is non-empty
// and the tunnel exists, opens directly into its edit form.
func NewModelWithEdit(cfg core.Config, editTunnelName string) Model {
	m := Model{
		config:    cfg,
		activeTab: "tunnels",
		view:      "table",
	}
	m.initTables()
	if editTunnelName != "" {
		for i, t := range cfg.Tunnels {
			if t.Name == editTunnelName {
				m.tunnelTable.SetCursor(i)
				m.initForm(false, &cfg.Tunnels[i])
				break
			}
		}
	}
	return m
}

// refreshTickMsg triggers a background refresh of tunnel rows.
type refreshTickMsg time.Time

func (m Model) refreshTickCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return refreshTickMsg(t)
	})
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return m.refreshTickCmd()
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case refreshTickMsg:
		m.refreshTunnelRows()
		return m, m.refreshTickCmd()
	}

	switch m.view {
	case "tunnel_form":
		return m.formUpdate(msg)
	case "host_form":
		return m.hostFormUpdate(msg)
	case "host_source":
		return m.hostSourceUpdate(msg)
	case "host_select":
		return m.hostSelectUpdate(msg)
	case "settings_form":
		return m.settingsFormUpdate(msg)
	case "confirm_delete":
		return m.confirmUpdate(msg)
	case "help":
		return m.helpUpdate(msg)
	default:
		return m.tableUpdate(msg)
	}
}

// View implements tea.Model.
func (m Model) View() string {
	switch m.view {
	case "tunnel_form":
		return m.formView()
	case "host_form":
		return m.hostFormView()
	case "host_source":
		return m.hostSourceView()
	case "host_select":
		return m.hostSelectView()
	case "settings_form":
		return m.settingsFormView()
	case "confirm_delete":
		return renderConfirmDelete(m.confirmItemType, m.confirmItem)
	case "help":
		return m.renderHelpView()
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
		case "n":
			if m.activeTab == "tunnels" {
				m.initForm(true, nil)
			} else {
				m.initHostForm(true, nil)
			}
		case "s":
			m.initSettingsForm()
		case "?":
			m.view = "help"
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
						m.initHostForm(false, &h)
					}
				}
			}
		case " ":
			if m.activeTab == "tunnels" {
				if len(m.config.Tunnels) > 0 {
					idx := m.tunnelTable.Cursor()
					if idx < len(m.config.Tunnels) {
						t := m.config.Tunnels[idx]
						if tunnelmgr.IsRunning(t.Name) {
							m.stopSelectedTunnel()
						} else {
							m.startSelectedTunnel()
						}
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
		case "r":
			if m.activeTab == "tunnels" {
				m.restartSelectedTunnel()
			}
		case "backspace":
			if m.activeTab == "tunnels" {
				if len(m.config.Tunnels) > 0 {
					idx := m.tunnelTable.Cursor()
					if idx < len(m.config.Tunnels) {
						m.confirmItem = m.config.Tunnels[idx].Name
						m.confirmItemType = "tunnel"
						m.view = "confirm_delete"
					}
				}
			} else {
				if len(m.config.Hosts) > 0 {
					idx := m.hostTable.Cursor()
					if idx < len(m.config.Hosts) {
						m.confirmItem = m.config.Hosts[idx].Name
						m.confirmItemType = "host"
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
			if m.confirmItemType == "host" {
				m.config.RemoveHost(m.confirmItem)
			} else {
				m.config.RemoveTunnel(m.confirmItem)
			}
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

func (m *Model) restartSelectedTunnel() {
	if len(m.config.Tunnels) == 0 {
		return
	}
	idx := m.tunnelTable.Cursor()
	if idx >= len(m.config.Tunnels) {
		return
	}
	t := m.config.Tunnels[idx]
	start := func(name string) error {
		exe, err := os.Executable()
		if err != nil {
			exe = "wombat"
		}
		cmd := exec.Command(exe, "tunnel-start", name)
		cmd.SysProcAttr = platform.DaemonSysProcAttr()
		return cmd.Run()
	}
	if err := tunnelmgr.RestartTunnel(t.Name, start); err != nil {
		notify.Alert("Wombat", fmt.Sprintf("Failed to restart %s: %v", t.Name, err))
	} else {
		notify.Notify("Wombat", fmt.Sprintf("Tunnel %s restarted", t.Name))
		m.refreshTunnelRows()
	}
}
