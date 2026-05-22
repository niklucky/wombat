//go:build gui

package gui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/niklucky/wombat/internal/core"
)

// Run starts the Fyne GUI application.
func Run(cfg core.Config) {
	a := app.New()
	w := a.NewWindow("Wombat SSH Helper")
	w.Resize(fyne.NewSize(600, 400))

	list := widget.NewList(
		func() int { return len(cfg.Hosts) },
		func() fyne.CanvasObject { return widget.NewLabel("template") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			if i < len(cfg.Hosts) {
				o.(*widget.Label).SetText(cfg.Hosts[i].Name)
			}
		},
	)

	w.SetContent(list)
	w.ShowAndRun()
}
