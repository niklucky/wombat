package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/niklucky/wombat/internal/locales"
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
	s := titleStyle.Render("  "+locales.T("app.title")) + "\n\n"
	s += "  " + lipgloss.NewStyle().Bold(true).Render(locales.T("help.title")) + "\n\n"

	s += "  " + actionStyle.Render(locales.T("help.global")) + "\n"
	s += "    " + actionStyle.Render("[?]") + helpStyle.Render(" "+locales.T("actions.help")+" ") + "\n"
	s += "    " + actionStyle.Render("[s]") + helpStyle.Render(" "+locales.T("actions.settings")+" ") + "\n"
	s += "    " + actionStyle.Render("[⇥]") + helpStyle.Render(" "+locales.T("actions.selectTab")+" ") + "\n"
	s += "    " + actionStyle.Render("[q]") + helpStyle.Render(" "+locales.T("actions.quit")+" ") + "\n\n"

	s += "  " + actionStyle.Render(locales.T("help.tunnels")) + "\n"
	s += "    " + actionStyle.Render("[↑/↓]") + helpStyle.Render(" "+locales.T("keys.navigate")+" ") + "\n"
	s += "    " + actionStyle.Render("[↵]") + helpStyle.Render(" "+locales.T("keys.edit")+" ") + "\n"
	s += "    " + actionStyle.Render("[␣]") + helpStyle.Render(" "+locales.T("keys.connectDisconnect")+" ") + "\n"
	s += "    " + actionStyle.Render("[r]") + helpStyle.Render(" "+locales.T("keys.restart")+" ") + "\n"
	s += "    " + actionStyle.Render("[n]") + helpStyle.Render(" "+locales.T("keys.add")+" ") + "\n"
	s += "    " + actionStyle.Render("[⌫]") + helpStyle.Render(" "+locales.T("keys.delete")+" ") + "\n\n"

	s += "  " + actionStyle.Render(locales.T("help.hosts")) + "\n"
	s += "    " + actionStyle.Render("[↑/↓]") + helpStyle.Render(" "+locales.T("keys.navigate")+" ") + "\n"
	s += "    " + actionStyle.Render("[↵]") + helpStyle.Render(" "+locales.T("keys.edit")+" ") + "\n"
	s += "    " + actionStyle.Render("[␣]") + helpStyle.Render(" "+locales.T("keys.connect")+" ") + "\n"
	s += "    " + actionStyle.Render("[t]") + helpStyle.Render(" "+locales.T("keys.testConnection")+" ") + "\n"
	s += "    " + actionStyle.Render("[n]") + helpStyle.Render(" "+locales.T("keys.add")+" ") + "\n"
	s += "    " + actionStyle.Render("[⌫]") + helpStyle.Render(" "+locales.T("keys.delete")+" ") + "\n\n"

	s += helpStyle.Render("  " + locales.T("help.goBack"))
	return s
}
