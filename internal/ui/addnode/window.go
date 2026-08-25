package addnode

import (
	"context"
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/dbu"
	"ndeploy/v2/internal/sink"
	"ndeploy/v2/internal/ui/ids"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

func New(a *app.App) (fyne.Window, func()) {
	w := a.Fyne.NewWindow(ids.AddNodeID)

	userEntry := widget.NewEntry()
	userEntry.SetPlaceHolder("remote user ex admin")

	hostEntry := widget.NewEntry()
	hostEntry.SetPlaceHolder("remote host ex 192.168.1.10")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "user", Widget: userEntry},
			{Text: "host", Widget: hostEntry},
		},
		OnSubmit: func() {
			ctx := context.Background()
			user := userEntry.Text
			host := hostEntry.Text

			sink.Printf(sink.DEBUG, "add node request, user='%s', host='%s'\n", user, host)
			queries, err := dbu.ConnectTo(a.WorkDir)
			if err != nil {
				sink.Printf(sink.ERROR, "add node error: %v\n", err)
				return
			}

			_, err = queries.CreateNode(
				ctx, db.CreateNodeParams{
					User: user,
					Host: host,
				},
			)

			if err != nil {
				sink.Printf(sink.ERROR, "add node db error: %v\n", err)
				return
			}

			w.Close()
		},
	}

	w.SetContent(form)
	return w, nil
}
