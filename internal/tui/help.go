package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m *Model) helpUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "?":
			m.view = "table"
		}
	}
	return m, nil
}

func (m *Model) renderHelpView() string {
	s := titleStyle.Render("  Wombat SSH Helper") + "\n\n"
	s += "  " + lipgloss.NewStyle().Bold(true).Render("Keyboard Shortcuts") + "\n\n"

	s += "  " + actionStyle.Render("Global") + "\n"
	s += "    " + actionStyle.Render("[?]") + helpStyle.Render(" Help ") + "\n"
	s += "    " + actionStyle.Render("[s]") + helpStyle.Render(" Settings ") + "\n"
	s += "    " + actionStyle.Render("[⇥]") + helpStyle.Render(" Select tab ") + "\n"
	s += "    " + actionStyle.Render("[q]") + helpStyle.Render(" Quit ") + "\n\n"

	s += "  " + actionStyle.Render("Tunnels") + "\n"
	s += "    " + actionStyle.Render("[↑/↓]") + helpStyle.Render(" Navigate ") + "\n"
	s += "    " + actionStyle.Render("[↵]") + helpStyle.Render(" Edit ") + "\n"
	s += "    " + actionStyle.Render("[Space]") + helpStyle.Render(" Connect / Disconnect ") + "\n"
	s += "    " + actionStyle.Render("[r]") + helpStyle.Render(" Restart ") + "\n"
	s += "    " + actionStyle.Render("[n]") + helpStyle.Render(" Add ") + "\n"
	s += "    " + actionStyle.Render("[⌫]") + helpStyle.Render(" Delete ") + "\n\n"

	s += "  " + actionStyle.Render("SSH Hosts") + "\n"
	s += "    " + actionStyle.Render("[↑/↓]") + helpStyle.Render(" Navigate ") + "\n"
	s += "    " + actionStyle.Render("[↵]") + helpStyle.Render(" Edit ") + "\n"
	s += "    " + actionStyle.Render("[Space]") + helpStyle.Render(" Connect ") + "\n"
	s += "    " + actionStyle.Render("[t]") + helpStyle.Render(" Test connection ") + "\n"
	s += "    " + actionStyle.Render("[n]") + helpStyle.Render(" Add ") + "\n"
	s += "    " + actionStyle.Render("[⌫]") + helpStyle.Render(" Delete ") + "\n\n"

	s += helpStyle.Render("  Press [esc], [?] or [q] to go back")
	return s
}
