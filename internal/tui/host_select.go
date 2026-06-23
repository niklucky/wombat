package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/niklucky/wombat/internal/locales"
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
					switch m.returnView {
					case "host_form":
						m.restoreHostForm()
						m.formInputs[5].SetValue(h.Name)
						m.view = "host_form"
						m.returnView = m.hostFormReturnView
						m.hostFormReturnView = ""
					case "tunnel_form":
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
	s := formLabelStyle.Render(locales.T("forms.titles.selectSSHHost")) + "\n\n"
	if len(m.config.Hosts) == 0 {
		s += "  " + locales.T("messages.noHostsConfigured") + "\n"
	} else {
		s += m.hostTable.View()
	}
	s += "\n" + formHelpStyle.Render("[↑/↓] "+locales.T("keys.navigate")+"  [↵] "+locales.T("keys.selectHost")+"  [esc] "+locales.T("keys.cancel"))
	return s
}
