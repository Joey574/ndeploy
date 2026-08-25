package addnode

import (
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/sink"
	"ndeploy/v2/internal/ui/ids"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func New(a *app.App) (fyne.Window, func()) {
	w := a.Fyne.NewWindow(ids.AddNodeID)

	userEntry := widget.NewEntry()
	hostEntry := widget.NewEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "user", Widget: userEntry},
			{Text: "host", Widget: hostEntry},
		},
		OnSubmit: func() {
			sink.Printf(sink.DEBUG, "add node request, user='%s', host='%s'\n", userEntry.Text, hostEntry.Text)
		},
	}

	w.SetContent(form)
	return w, nil
}
