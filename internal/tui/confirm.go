package tui

import "fmt"

func renderConfirmDelete(item string) string {
	s := dialogBoxStyle.Render(
		confirmTextStyle.Render(fmt.Sprintf("Delete tunnel %q?", item)) + "\n\n" +
			confirmKeyStyle.Render("y") + confirmTextStyle.Render(" confirm  ") +
			confirmKeyStyle.Render("n") + confirmTextStyle.Render(" cancel"),
	)
	return s
}
