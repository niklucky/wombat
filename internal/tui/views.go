package tui

import "fmt"

func renderHostList(m Model) string {
	s := "  Wombat SSH Helper\n\n"
	s += "  Hosts\n"
	s += "  -----\n"

	for i, host := range m.config.Hosts {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		s += fmt.Sprintf("  %s %s (%s@%s)\n", cursor, host.Name, host.User, host.Address)
	}

	if len(m.config.Hosts) == 0 {
		s += "  No hosts configured.\n"
	}

	s += "\n  [j/k] navigate  [Enter] connect  [q] quit\n"
	return s
}
