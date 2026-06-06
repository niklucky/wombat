package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/niklucky/wombat/internal/tunnelmgr"
)

func (m *Model) initTables() {
	// Tunnel table
	tunnelCols := []table.Column{
		{Title: "#", Width: 4},
		{Title: "", Width: 3},
		{Title: "Name", Width: 22},
		{Title: "Host", Width: 22},
		{Title: "Forward", Width: 22},
	}
	m.tunnelTable = table.New(
		table.WithColumns(tunnelCols),
		table.WithFocused(true),
		table.WithHeight(20),
	)
	m.tunnelTable.SetStyles(table.Styles{
		Header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Padding(0, 1).Border(lipgloss.NormalBorder()).BorderBottom(true).BorderTop(false).BorderLeft(false).BorderRight(false).BorderForeground(lipgloss.Color(primary)),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color(primary)),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	})
	m.refreshTunnelRows()

	// Host table
	hostCols := []table.Column{
		{Title: "#", Width: 4},
		{Title: "Name", Width: 24},
		{Title: "Address", Width: 20},
		{Title: "User", Width: 14},
		{Title: "Port", Width: 8},
	}
	m.hostTable = table.New(
		table.WithColumns(hostCols),
		table.WithFocused(true),
		table.WithHeight(20),
	)
	m.hostTable.SetStyles(table.Styles{
		Header:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Padding(0, 1).Border(lipgloss.NormalBorder()).BorderBottom(true).BorderTop(false).BorderLeft(false).BorderRight(false).BorderForeground(lipgloss.Color(primary)),
		Selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color(primary)),
		Cell:     lipgloss.NewStyle().Padding(0, 1),
	})
	m.refreshHostRows()
}

func (m *Model) refreshTunnelRows() {
	var rows []table.Row
	for i, t := range m.config.Tunnels {
		status := "⚪"
		if tunnelmgr.IsRunning(t.Name) {
			status = "🟢"
		}
		fwd := fmt.Sprintf("%d → %s:%d", t.LocalPort, t.RemoteHost, t.RemotePort)
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			status,
			t.Name,
			t.HostName,
			fwd,
		})
	}
	m.tunnelTable.SetRows(rows)
}

func (m *Model) refreshHostRows() {
	var rows []table.Row
	for i, h := range m.config.Hosts {
		port := h.Port
		if port == 0 {
			port = 22
		}
		rows = append(rows, table.Row{
			fmt.Sprintf("%d", i+1),
			h.Name,
			h.Address,
			h.User,
			fmt.Sprintf("%d", port),
		})
	}
	m.hostTable.SetRows(rows)
}

func (m *Model) refreshTables() {
	m.refreshTunnelRows()
	m.refreshHostRows()
}

func (m *Model) activeTable() *table.Model {
	if m.activeTab == "hosts" {
		return &m.hostTable
	}
	return &m.tunnelTable
}

func (m *Model) renderTableView() string {
	// Title
	s := titleStyle.Render("  Wombat SSH Helper") + "\n"
	s += subtitleStyle.Render("  Cross-platform SSH helper with TUI and system tray") + "\n\n"

	// Actions bar
	s += "  Actions: "
	s += actionStyle.Render("[?]") + helpStyle.Render(" Help ")
	s += actionStyle.Render("[s]") + helpStyle.Render(" Settings ")
	s += actionStyle.Render("[⇥]") + helpStyle.Render(" Select tab ")
	s += actionStyle.Render("[q]") + helpStyle.Render(" Quit ")
	s += "\n\n"

	// Tabs
	tabStr := "  "
	if m.activeTab == "tunnels" {
		tabStr += activeTabStyle.Render("Tunnels")
	} else {
		tabStr += tabStyle.Render("Tunnels")
	}
	tabStr += "  "
	if m.activeTab == "hosts" {
		tabStr += activeTabStyle.Render("SSH Hosts")
	} else {
		tabStr += tabStyle.Render("SSH Hosts")
	}
	s += tabStr + "\n\n"

	// Table content
	if m.activeTab == "tunnels" {
		s += m.tunnelTable.View()
		s += "\n  " + actionStyle.Render("[↵]") + helpStyle.Render(" Edit  ") + actionStyle.Render("[Space]") + helpStyle.Render(" Connect/Disconnect  ") + actionStyle.Render("[r]") + helpStyle.Render(" Restart  ") + actionStyle.Render("[n]") + helpStyle.Render(" Add  ") + actionStyle.Render("[⌫]") + helpStyle.Render(" Delete")
	} else {
		s += m.hostTable.View()
		s += "\n  " + actionStyle.Render("[↵]") + helpStyle.Render(" Edit  ") + actionStyle.Render("[t]") + helpStyle.Render(" Test  ") + actionStyle.Render("[n]") + helpStyle.Render(" Add  ") + actionStyle.Render("[⌫]") + helpStyle.Render(" Delete")
	}

	return s
}
