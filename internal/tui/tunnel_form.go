package tui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/locales"
)

func (m *Model) initForm(isCreate bool, tunnel *core.Tunnel) {
	m.view = "tunnel_form"
	m.formIsCreate = isCreate
	m.editingTunnel = tunnel
	m.formFocus = 0

	inputs := make([]textinput.Model, 5)
	placeholders := []string{
		locales.T("forms.placeholders.name"),
		locales.T("forms.placeholders.hostAlias"),
		locales.T("forms.placeholders.localPort"),
		locales.T("forms.placeholders.remoteHost"),
		locales.T("forms.placeholders.remotePort"),
	}
	values := []string{"", "", "", "localhost", ""}

	if tunnel != nil {
		values[0] = tunnel.Name
		values[1] = tunnel.HostName
		values[2] = fmt.Sprintf("%d", tunnel.LocalPort)
		values[3] = tunnel.RemoteHost
		values[4] = fmt.Sprintf("%d", tunnel.RemotePort)
	}

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = placeholders[i]
		inputs[i].SetValue(values[i])
		inputs[i].CharLimit = 50
		inputs[i].Width = 40
	}
	inputs[0].Focus()

	m.formInputs = inputs
}

func (m *Model) formUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.view = "table"
			return m, nil
		case "ctrl+s":
			if err := m.saveForm(); err == nil {
				m.view = "table"
				m.refreshTables()
			}
			return m, nil
		case "tab":
			m.formFocus = (m.formFocus + 1) % len(m.formInputs)
			m.updateFormFocus()
		case "shift+tab":
			m.formFocus--
			if m.formFocus < 0 {
				m.formFocus = len(m.formInputs) - 1
			}
			m.updateFormFocus()
		case "enter":
			if m.formFocus == 1 {
				m.saveTunnelFormState()
				m.returnView = "tunnel_form"
				m.hostSourceCursor = 0
				m.view = "host_source"
				return m, nil
			}
			m.formInputs[m.formFocus], _ = m.formInputs[m.formFocus].Update(msg)
		default:
			m.formInputs[m.formFocus], _ = m.formInputs[m.formFocus].Update(msg)
		}
	}
	return m, nil
}

func (m *Model) updateFormFocus() {
	for i := range m.formInputs {
		if i == m.formFocus {
			m.formInputs[i].Focus()
		} else {
			m.formInputs[i].Blur()
		}
	}
}

func (m *Model) saveTunnelFormState() {
	m.savedFormInputs = make([]textinput.Model, len(m.formInputs))
	copy(m.savedFormInputs, m.formInputs)
	m.savedFormFocus = m.formFocus
	m.savedFormIsCreate = m.formIsCreate
	m.savedEditingTunnel = m.editingTunnel
}

func (m *Model) restoreTunnelForm() {
	m.formInputs = m.savedFormInputs
	m.formFocus = m.savedFormFocus
	m.formIsCreate = m.savedFormIsCreate
	m.editingTunnel = m.savedEditingTunnel
	m.updateFormFocus()
}

func (m *Model) saveForm() error {
	name := m.formInputs[0].Value()
	hostName := m.formInputs[1].Value()
	localPort, _ := strconv.Atoi(m.formInputs[2].Value())
	remoteHost := m.formInputs[3].Value()
	remotePort, _ := strconv.Atoi(m.formInputs[4].Value())

	if name == "" || hostName == "" || localPort == 0 || remotePort == 0 {
		return locales.Errorf("errors.allFieldsRequired")
	}
	if remoteHost == "" {
		remoteHost = "localhost"
	}

	tunnel := core.Tunnel{
		Name:       name,
		HostName:   hostName,
		LocalPort:  localPort,
		RemoteHost: remoteHost,
		RemotePort: remotePort,
	}

	if m.formIsCreate {
		m.config.AddTunnel(tunnel)
	} else {
		// If name changed, remove old and add new
		if m.editingTunnel != nil && m.editingTunnel.Name != name {
			m.config.RemoveTunnel(m.editingTunnel.Name)
			m.config.AddTunnel(tunnel)
		} else {
			// Update in place: remove old, add new
			if m.editingTunnel != nil {
				m.config.RemoveTunnel(m.editingTunnel.Name)
			}
			m.config.AddTunnel(tunnel)
		}
	}

	return m.config.Save()
}

func (m *Model) formView() string {
	labels := []string{
		locales.T("forms.labels.name"),
		locales.T("forms.labels.hostAlias"),
		locales.T("forms.labels.localPort"),
		locales.T("forms.labels.remoteHost"),
		locales.T("forms.labels.remotePort"),
	}

	s := formLabelStyle.Render(locales.T("forms.titles.tunnel")) + "\n\n"
	for i := range m.formInputs {
		cursor := "  "
		if i == m.formFocus {
			cursor = "> "
		}
		s += cursor + formLabelStyle.Render(labels[i]) + "\n"
		s += "  " + m.formInputs[i].View() + "\n\n"
	}
	help := locales.T("forms.help.saveEscCancelTab")
	if m.formFocus == 1 {
		help += "  " + locales.T("forms.help.selectHost")
	}
	s += formHelpStyle.Render(help)
	return s
}
