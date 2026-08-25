package mainwindow

import (
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/ui/ids"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func New(a *app.App) (fyne.Window, func()) {
	w := a.Fyne.NewWindow(ids.MainWindowID)

	w.Resize(fyne.NewSize(600, 400))

	list := container.NewVBox(
		widget.NewButton("Logs", func() {
			a.OpenOrFocus(ids.LogViewerID)
		}),
		widget.NewButton("Add Node", func() {
			a.OpenOrFocus(ids.AddNodeID)
		}),
	)

	w.SetContent(list)
	return w, nil
}
