package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/niklucky/wombat/internal/core"
	"github.com/niklucky/wombat/internal/locales"
	"github.com/niklucky/wombat/internal/notify"
	"github.com/niklucky/wombat/internal/tunnelmgr"
)

func (m *Model) initSettingsForm() {
	m.view = "settings_form"
	m.formFocus = 0

	home, _ := core.AppHome()

	inputs := make([]textinput.Model, 3)
	bools := make([]bool, 3)
	placeholders := []string{
		locales.T("forms.placeholders.appHomeFolder"),
		locales.T("forms.placeholders.openTrayOnStart"),
		locales.T("forms.placeholders.showNotifications"),
	}
	bools[1] = m.config.OpenTray
	bools[2] = m.config.ShowNotify

	for i := range inputs {
		inputs[i] = textinput.New()
		inputs[i].Placeholder = placeholders[i]
		if i == 1 || i == 2 {
			inputs[i].SetValue(boolToYesNo(bools[i]))
		} else {
			inputs[i].SetValue(home)
		}
		inputs[i].CharLimit = 100
		inputs[i].Width = 50
	}
	inputs[0].Focus()

	m.formInputs = inputs
	m.formBools = bools
}

func boolToYesNo(v bool) string {
	if v {
		return locales.T("common.yes")
	}
	return locales.T("common.no")
}

func yesNoToBool(s string) bool {
	lower := strings.ToLower(strings.TrimSpace(s))
	return lower == "yes" || lower == "y" || lower == "true" ||
		lower == strings.ToLower(locales.T("common.yes"))
}

func (m *Model) settingsFormUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.view = "table"
			return m, nil
		case "ctrl+s":
			homeChanged, err := m.saveSettingsForm()
			if err != nil {
				return m, nil
			}
			if homeChanged {
				m.RestartRequired = true
				return m, tea.Quit
			}
			m.view = "table"
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
			if m.formFocus == 1 || m.formFocus == 2 {
				m.formBools[m.formFocus] = !m.formBools[m.formFocus]
				m.formInputs[m.formFocus].SetValue(boolToYesNo(m.formBools[m.formFocus]))
				return m, nil
			}
			m.formInputs[m.formFocus], _ = m.formInputs[m.formFocus].Update(msg)
		default:
			m.formInputs[m.formFocus], _ = m.formInputs[m.formFocus].Update(msg)
		}
	}
	return m, nil
}

func (m *Model) saveSettingsForm() (bool, error) {
	newHome := strings.TrimSpace(m.formInputs[0].Value())
	openTray := m.formBools[1]
	showNotify := m.formBools[2]

	if newHome == "" {
		return false, locales.Errorf("errors.appHomeEmpty")
	}

	oldHome, _ := core.AppHome()
	homeChanged := filepath.Clean(oldHome) != filepath.Clean(newHome)

	// If app home changed, stop all tunnels and clean up old location
	if homeChanged {
		// Stop all active tunnels
		for _, t := range m.config.Tunnels {
			if tunnelmgr.IsRunning(t.Name) {
				_ = tunnelmgr.StopDaemon(t.Name)
			}
		}

		// Clean up old PID and log directories
		oldPidDir := filepath.Join(oldHome, "pids")
		oldLogDir := filepath.Join(oldHome, "logs")
		_ = os.RemoveAll(oldPidDir)
		_ = os.RemoveAll(oldLogDir)

		// Try to remove old config.json so we don't leave stale data
		oldConfig := filepath.Join(oldHome, "config.json")
		_ = os.Remove(oldConfig)

		// Update pointer
		if err := core.SetAppHome(newHome); err != nil {
			return false, locales.Errorf("errors.setAppHome", err)
		}

		notify.Notify(locales.T("app.title"), fmt.Sprintf(locales.T("messages.appHomeMoved"), newHome))
	}

	m.config.OpenTray = openTray
	m.config.ShowNotify = showNotify

	return homeChanged, m.config.Save()
}

func (m *Model) settingsFormView() string {
	labels := []string{
		locales.T("forms.labels.appHomeFolder"),
		locales.T("forms.labels.openTrayOnStart"),
		locales.T("forms.labels.showNotifications"),
	}

	s := formLabelStyle.Render(locales.T("forms.titles.settings")) + "\n\n"
	for i := range m.formInputs {
		cursor := "  "
		if i == m.formFocus {
			cursor = "> "
		}
		s += cursor + formLabelStyle.Render(labels[i]) + "\n"
		s += "  " + m.formInputs[i].View() + "\n\n"
	}
	help := locales.T("forms.help.saveEscCancelTab")
	if m.formFocus == 1 || m.formFocus == 2 {
		help += "  " + locales.T("forms.help.toggle")
	}
	s += formHelpStyle.Render(help)
	return s
}
