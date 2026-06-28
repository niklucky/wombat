package tui

import "github.com/charmbracelet/lipgloss"

// const brown = "#B54600"
const primary = "#219ed9"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(primary))

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A0A0A0"))

	tabStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(primary)).
			Bold(true).
			Underline(true).
			Padding(0, 2)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color(primary)).
			Padding(0, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	actionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(primary)).
			Bold(true)

	formLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Bold(true)

	formHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	dialogBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(primary)).
			Padding(1, 2)

	confirmTextStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA"))

	confirmKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(primary)).
			Bold(true)

	versionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))
)
