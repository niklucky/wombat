package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m *Model) hostSourceUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.returnView == "tunnel_form" {
				m.restoreTunnelForm()
				m.view = "tunnel_form"
				m.returnView = ""
				return m, nil
			}
			m.view = "table"
			return m, nil
		case "up", "k":
			if m.hostSourceCursor > 0 {
				m.hostSourceCursor--
			}
		case "down", "j":
			if m.hostSourceCursor < 1 {
				m.hostSourceCursor++
			}
		case "enter":
			if m.hostSourceCursor == 0 {
				m.view = "host_select"
				m.refreshHostRows()
			} else {
				m.initHostForm(true, nil)
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) hostSourceView() string {
	options := []string{"Select from existing SSH hosts", "Enter custom SSH host"}

	s := formLabelStyle.Render("Select host source") + "\n\n"
	for i, opt := range options {
		cursor := "  "
		if i == m.hostSourceCursor {
			cursor = "> "
		}
		s += cursor + opt + "\n"
	}
	s += "\n" + formHelpStyle.Render("[↑/↓] navigate  [↵] select  [esc] cancel")
	return s
}
