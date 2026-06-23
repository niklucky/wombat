package tui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/locales"
)

func (m *Model) initHostForm(isCreate bool, host *core.Host) {
	m.view = "host_form"
	m.formIsCreateHost = isCreate
	m.editingHost = host
	m.formFocus = 0

	inputs := make([]textinput.Model, 7)
	bools := make([]bool, 7)
	placeholders := []string{
		locales.T("forms.placeholders.name"),
		locales.T("forms.placeholders.address"),
		locales.T("forms.placeholders.user"),
		locales.T("forms.placeholders.port"),
		locales.T("forms.placeholders.keyPath"),
		locales.T("forms.placeholders.proxyJump"),
		locales.T("forms.placeholders.saveToSSHHosts"),
	}
	values := []string{"", "", "", "22", "", "", ""}
	bools[6] = true

	if host != nil {
		values[0] = host.Name
		values[1] = host.Address
		values[2] = host.User
		if host.Port != 0 {
			values[3] = fmt.Sprintf("%d", host.Port)
		} else {
			values[3] = "22"
		}
		values[4] = host.KeyPath
		values[5] = host.ProxyJump
	}

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = placeholders[i]
		if i == 6 {
			inputs[i].SetValue(boolToYesNo(bools[i]))
		} else {
			inputs[i].SetValue(values[i])
		}
		inputs[i].CharLimit = 100
		inputs[i].Width = 40
	}
	inputs[0].Focus()

	m.formInputs = inputs
	m.formBools = bools
}

func (m *Model) hostFormUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.returnView == "host_form" {
				m.restoreHostForm()
				m.view = "host_form"
				m.returnView = m.hostFormReturnView
				m.hostFormReturnView = ""
				return m, nil
			}
			if m.returnView == "tunnel_form" {
				m.view = "host_source"
				return m, nil
			}
			m.view = "table"
			return m, nil
		case "ctrl+s":
			name := m.formInputs[0].Value()
			if err := m.saveHostForm(); err == nil {
				if m.returnView == "host_form" {
					m.restoreHostForm()
					m.formInputs[5].SetValue(name)
					m.view = "host_form"
					m.returnView = m.hostFormReturnView
					m.hostFormReturnView = ""
					m.refreshHostRows()
					return m, nil
				}
				if m.returnView == "tunnel_form" {
					name := m.formInputs[0].Value()
					m.restoreTunnelForm()
					m.formInputs[1].SetValue(name)
					m.view = "tunnel_form"
					m.returnView = ""
					m.refreshTables()
					return m, nil
				}
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
		case " ":
			if m.formFocus == 6 {
				m.formBools[6] = !m.formBools[6]
				m.formInputs[6].SetValue(boolToYesNo(m.formBools[6]))
				return m, nil
			}
			m.formInputs[m.formFocus], _ = m.formInputs[m.formFocus].Update(msg)
		case "enter":
			if m.formFocus == 5 {
				m.saveHostFormState()
				m.hostFormReturnView = m.returnView
				m.returnView = "host_form"
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

func (m *Model) saveHostFormState() {
	m.savedHostFormInputs = make([]textinput.Model, len(m.formInputs))
	copy(m.savedHostFormInputs, m.formInputs)
	m.savedHostFormBools = make([]bool, len(m.formBools))
	copy(m.savedHostFormBools, m.formBools)
	m.savedHostFormFocus = m.formFocus
	m.savedHostFormIsCreate = m.formIsCreateHost
	m.savedEditingHost = m.editingHost
}

func (m *Model) restoreHostForm() {
	m.formInputs = m.savedHostFormInputs
	m.formBools = make([]bool, len(m.savedHostFormBools))
	copy(m.formBools, m.savedHostFormBools)
	m.formFocus = m.savedHostFormFocus
	m.formIsCreateHost = m.savedHostFormIsCreate
	m.editingHost = m.savedEditingHost
	m.updateFormFocus()
}

func (m *Model) saveHostForm() error {
	name := m.formInputs[0].Value()
	address := m.formInputs[1].Value()
	user := m.formInputs[2].Value()
	rawPort := m.formInputs[3].Value()
	keyPath := m.formInputs[4].Value()
	proxyJump := m.formInputs[5].Value()
	saveHost := m.formBools[6]

	if name == "" || address == "" || user == "" {
		return locales.Errorf("errors.nameAddressUserRequired")
	}

	var port int
	if rawPort == "" {
		port = 22
	} else {
		p, err := strconv.Atoi(rawPort)
		if err != nil {
			return locales.Errorf("errors.invalidPort", err)
		}
		if p < 1 || p > 65535 {
			return locales.Errorf("errors.portRange")
		}
		port = p
	}

	if !saveHost {
		return nil
	}

	host := core.Host{
		Name:      name,
		Address:   address,
		User:      user,
		Port:      port,
		KeyPath:   keyPath,
		ProxyJump: proxyJump,
	}

	if m.formIsCreateHost {
		m.config.AddHost(host)
	} else {
		if m.editingHost != nil && m.editingHost.Name != name {
			m.config.RemoveHost(m.editingHost.Name)
			m.config.AddHost(host)
		} else {
			if m.editingHost != nil {
				m.config.RemoveHost(m.editingHost.Name)
			}
			m.config.AddHost(host)
		}
	}

	return m.config.Save()
}

func (m *Model) hostFormView() string {
	labels := []string{
		locales.T("forms.labels.name"),
		locales.T("forms.labels.address"),
		locales.T("forms.labels.user"),
		locales.T("forms.labels.port"),
		locales.T("forms.labels.keyPath"),
		locales.T("forms.labels.proxyJump"),
		locales.T("forms.labels.saveToSSHHosts"),
	}

	s := formLabelStyle.Render(locales.T("forms.titles.host")) + "\n\n"
	for i := range m.formInputs {
		cursor := "  "
		if i == m.formFocus {
			cursor = "> "
		}
		s += cursor + formLabelStyle.Render(labels[i]) + "\n"
		s += "  " + m.formInputs[i].View() + "\n\n"
	}
	help := locales.T("forms.help.saveEscCancelTab")
	if m.formFocus == 5 {
		help += "  " + locales.T("forms.help.selectProxyJump")
	}
	if m.formFocus == 6 {
		help += "  " + locales.T("forms.help.toggle")
	}
	s += formHelpStyle.Render(help)
	return s
}
