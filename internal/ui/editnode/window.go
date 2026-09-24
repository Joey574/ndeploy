package editnode

import (
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/ui/ids"

	"fyne.io/fyne/v2"
)

type EditNode struct {
	w fyne.Window
}

func New(a *app.App) app.Window {
	w := a.Fyne.NewWindow(ids.EditNodeID)
	w.Resize(fyne.NewSize(680, 460))

	return &EditNode{
		w: w,
	}
}

func (e *EditNode) Window() fyne.Window {
	return e.w
}

func (e *EditNode) Close() error {
	return nil
}
