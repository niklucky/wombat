package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/niklucky/wombat/internal/core"
)

// Model is the root Bubble Tea model.
type Model struct {
	config       core.Config
	cursor       int
	view         string
	SelectedHost *core.Host
}

// NewModel creates a new TUI model with the given config.
func NewModel(cfg core.Config) Model {
	return Model{
		config: cfg,
		cursor: 0,
		view:   "hosts",
	}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.config.Hosts)-1 {
				m.cursor++
			}
		case "enter":
			if m.cursor < len(m.config.Hosts) {
				h := m.config.Hosts[m.cursor]
				m.SelectedHost = &h
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

// View implements tea.Model.
func (m Model) View() string {
	switch m.view {
	case "hosts":
		return renderHostList(m)
	default:
		return "Unknown view\n"
	}
}
