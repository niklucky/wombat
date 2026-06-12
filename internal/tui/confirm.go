package tui

import "github.com/niklucky/wombat/internal/locales"

func renderConfirmDelete(itemType, item string) string {
	s := dialogBoxStyle.Render(
		confirmTextStyle.Render(locales.T("dialog.deleteTitle", itemType, item)) + "\n\n" +
			confirmKeyStyle.Render("y") + confirmTextStyle.Render(locales.T("dialog.confirm")) +
			confirmKeyStyle.Render("n") + confirmTextStyle.Render(locales.T("dialog.cancel")),
	)
	return s
}
