package tui

import "fmt"

func renderConfirmDelete(itemType, item string) string {
	s := dialogBoxStyle.Render(
		confirmTextStyle.Render(fmt.Sprintf("Delete %s %q?", itemType, item)) + "\n\n" +
			confirmKeyStyle.Render("y") + confirmTextStyle.Render(" confirm  ") +
			confirmKeyStyle.Render("n") + confirmTextStyle.Render(" cancel"),
	)
	return s
}
