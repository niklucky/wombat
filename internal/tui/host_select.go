package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) hostSelectUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.view = "host_source"
			return m, nil
		case "up", "k":
			m.hostTable.MoveUp(1)
		case "down", "j":
			m.hostTable.MoveDown(1)
		case "enter":
			if len(m.config.Hosts) > 0 {
				idx := m.hostTable.Cursor()
				if idx < len(m.config.Hosts) {
					h := m.config.Hosts[idx]
					if m.returnView == "tunnel_form" {
						m.restoreTunnelForm()
						m.formInputs[1].SetValue(h.Name)
						m.view = "tunnel_form"
						m.returnView = ""
					}
				}
			}
		}
	}
	return m, nil
}

func (m *Model) hostSelectView() string {
	s := formLabelStyle.Render("Select SSH host") + "\n\n"
	if len(m.config.Hosts) == 0 {
		s += "  No hosts configured.\n"
	} else {
		s += m.hostTable.View()
	}
	s += "\n" + formHelpStyle.Render("[↑/↓] navigate  [↵] select  [esc] back")
	return s
}
