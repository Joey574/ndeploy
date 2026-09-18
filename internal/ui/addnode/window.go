package addnode

import (
	"context"
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/ui/ids"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/Joey574/sink/v2/pkg/sink"
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

			a.Sink.Printf(sink.DEBUG, "add node request, user='%s', host='%s'\n", user, host)
			queries := a.Dbu.Queries()

			n, err := queries.CreateNode(
				ctx, db.CreateNodeParams{
					User: user,
					Host: host,
				},
			)

			if err != nil {
				a.Sink.Printf(sink.ERROR, "create node failed: %v\n", err)
				return
			}
			defer w.Close()

			if err := a.Engine.TestConnection(&n); err != nil {
				a.Sink.Printf(sink.WARN, "test connection failed: %v\n", err)
			}

			a.Sink.Println(sink.DEBUG, "connection with node succesful")
		},
	}

	w.SetContent(form)
	return w, nil
}
