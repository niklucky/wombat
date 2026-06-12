package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"
	"github.com/niklucky/wombat/internal/locales"
	"github.com/niklucky/wombat/internal/tunnelmgr"
)

func (m *Model) initTables() {
	// Tunnel table
	tunnelCols := []table.Column{
		{Title: locales.T("table.columns.number"), Width: 4},
		{Title: "", Width: 3},
		{Title: locales.T("table.columns.name"), Width: 22},
		{Title: locales.T("table.columns.host"), Width: 22},
		{Title: locales.T("table.columns.forward"), Width: 22},
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
		{Title: locales.T("table.columns.number"), Width: 4},
		{Title: locales.T("table.columns.name"), Width: 24},
		{Title: locales.T("table.columns.address"), Width: 20},
		{Title: locales.T("table.columns.user"), Width: 14},
		{Title: locales.T("table.columns.port"), Width: 8},
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
	s := titleStyle.Render("  "+locales.T("app.title")) + "\n"
	s += subtitleStyle.Render("  "+locales.T("app.subtitle")) + "\n\n"

	// Actions bar
	s += "  " + locales.T("actions.label") + " "
	s += actionStyle.Render("[?]") + helpStyle.Render(" " + locales.T("actions.help") + " ")
	s += actionStyle.Render("[s]") + helpStyle.Render(" " + locales.T("actions.settings") + " ")
	s += actionStyle.Render("[⇥]") + helpStyle.Render(" " + locales.T("actions.selectTab") + " ")
	s += actionStyle.Render("[q]") + helpStyle.Render(" " + locales.T("actions.quit") + " ")
	s += "\n\n"

	// Tabs
	tabStr := "  "
	if m.activeTab == "tunnels" {
		tabStr += activeTabStyle.Render(locales.T("tabs.tunnels"))
	} else {
		tabStr += tabStyle.Render(locales.T("tabs.tunnels"))
	}
	tabStr += "  "
	if m.activeTab == "hosts" {
		tabStr += activeTabStyle.Render(locales.T("tabs.hosts"))
	} else {
		tabStr += tabStyle.Render(locales.T("tabs.hosts"))
	}
	s += tabStr + "\n\n"

	// Table content
	if m.activeTab == "tunnels" {
		s += m.tunnelTable.View()
		s += "\n  " + actionStyle.Render("[↵]") + helpStyle.Render(" "+locales.T("keys.edit")+"  ") + actionStyle.Render("[Space]") + helpStyle.Render(" "+locales.T("keys.connectDisconnect")+"  ") + actionStyle.Render("[r]") + helpStyle.Render(" "+locales.T("keys.restart")+"  ") + actionStyle.Render("[n]") + helpStyle.Render(" "+locales.T("keys.add")+"  ") + actionStyle.Render("[⌫]") + helpStyle.Render(" "+locales.T("keys.delete"))
	} else {
		s += m.hostTable.View()
		s += "\n  " + actionStyle.Render("[↵]") + helpStyle.Render(" "+locales.T("keys.edit")+"  ") + actionStyle.Render("[t]") + helpStyle.Render(" "+locales.T("keys.test")+"  ") + actionStyle.Render("[n]") + helpStyle.Render(" "+locales.T("keys.add")+"  ") + actionStyle.Render("[⌫]") + helpStyle.Render(" "+locales.T("keys.delete"))
	}

	return s
}
