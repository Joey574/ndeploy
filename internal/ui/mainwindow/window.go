package mainwindow

import (
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/ui/ids"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type MainWindow struct {
	w fyne.Window
}

func New(a *app.App) app.Window {
	w := a.Fyne.NewWindow(ids.MainWindowID)

	w.Resize(fyne.NewSize(600, 400))

	list := container.NewVBox(
		widget.NewButton("Logs", func() {
			a.OpenOrFocus(ids.LogViewerID)
		}),
		widget.NewButton("Add Node", func() {
			a.OpenOrFocus(ids.AddNodeID)
		}),
		widget.NewButton("Edit Node", func() {
			a.OpenOrFocus(ids.EditNodeID)
		}),
	)

	w.SetContent(list)
	return &MainWindow{
		w: w,
	}
}

func (m *MainWindow) Window() fyne.Window {
	return m.w
}

func (m *MainWindow) Close() error {
	return nil
}
