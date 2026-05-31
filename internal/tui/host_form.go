package tui

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/niklucky/wombat/internal/core"
)

func (m *Model) initHostForm(isCreate bool, host *core.Host) {
	m.view = "host_form"
	m.formIsCreateHost = isCreate
	m.editingHost = host
	m.formFocus = 0

	inputs := make([]textinput.Model, 6)
	placeholders := []string{"Name", "Address", "User", "Port", "Key path", "Save to SSH-hosts"}
	values := []string{"", "", "", "22", "", "yes"}

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
	}

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = placeholders[i]
		inputs[i].SetValue(values[i])
		inputs[i].CharLimit = 100
		inputs[i].Width = 40
	}
	inputs[0].Focus()

	m.formInputs = inputs
}

func (m *Model) hostFormUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.returnView == "tunnel_form" {
				m.view = "host_source"
				return m, nil
			}
			m.view = "table"
			return m, nil
		case "ctrl+s":
			if err := m.saveHostForm(); err == nil {
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
			if m.formFocus == 5 {
				if m.formInputs[5].Value() == "yes" {
					m.formInputs[5].SetValue("no")
				} else {
					m.formInputs[5].SetValue("yes")
				}
				return m, nil
			}
			m.formInputs[m.formFocus], _ = m.formInputs[m.formFocus].Update(msg)
		default:
			m.formInputs[m.formFocus], _ = m.formInputs[m.formFocus].Update(msg)
		}
	}
	return m, nil
}

func (m *Model) saveHostForm() error {
	name := m.formInputs[0].Value()
	address := m.formInputs[1].Value()
	user := m.formInputs[2].Value()
	rawPort := m.formInputs[3].Value()
	keyPath := m.formInputs[4].Value()
	saveHost := yesNoToBool(m.formInputs[5].Value())

	if name == "" || address == "" || user == "" {
		return fmt.Errorf("name, address and user are required")
	}

	var port int
	if rawPort == "" {
		port = 22
	} else {
		p, err := strconv.Atoi(rawPort)
		if err != nil {
			return fmt.Errorf("invalid port: %v", err)
		}
		if p < 1 || p > 65535 {
			return fmt.Errorf("port must be between 1 and 65535")
		}
		port = p
	}

	if !saveHost {
		return nil
	}

	host := core.Host{
		Name:    name,
		Address: address,
		User:    user,
		Port:    port,
		KeyPath: keyPath,
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
	labels := []string{"Name", "Address", "User", "Port", "Key path", "Save to SSH-hosts"}

	s := formLabelStyle.Render("Host") + "\n\n"
	for i := range m.formInputs {
		cursor := "  "
		if i == m.formFocus {
			cursor = "> "
		}
		s += cursor + formLabelStyle.Render(labels[i]) + "\n"
		s += "  " + m.formInputs[i].View() + "\n\n"
	}
	help := "[ctrl+s] save  [esc] cancel  [tab] next field"
	if m.formFocus == 5 {
		help += "  [space] toggle"
	}
	s += formHelpStyle.Render(help)
	return s
}
