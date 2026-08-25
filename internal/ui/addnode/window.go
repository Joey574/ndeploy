package addnode

import (
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/ui/ids"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func New(a *app.App) (fyne.Window, func()) {
	w := a.Fyne.NewWindow(ids.AddNodeID)

	entry := widget.NewEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "user", Widget: entry},
			{Text: "host", Widget: entry},
		},
	}

	w.SetContent(form)
	return w, nil
}
