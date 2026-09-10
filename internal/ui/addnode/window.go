package addnode

import (
	"context"
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/ssh"
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

			_, err := queries.CreateNode(
				ctx, db.CreateNodeParams{
					User: user,
					Host: host,
				},
			)

			if err != nil {
				a.Sink.Printf(sink.ERROR, "add node db error: %v\n", err)
				return
			}

			// at this point any error is considered
			// non-"fatal" and we should close cleanly
			defer w.Close()

			// run an initial test connection with the node
			client, err := ssh.NewClient(user, host, ssh.SetSink(a.Sink))
			if err != nil {
				a.Sink.Printf(sink.ERROR, "new ssh client: %v\n", err)
				return
			}
			defer client.Close()

			c, err := client.Dial()
			if err != nil {
				a.Sink.Printf(sink.ERROR, "ssh dial: %v\n", err)
				return
			}

			c.Close()
			a.Sink.Println(sink.DEBUG, "connection with node succesful")
		},
	}

	w.SetContent(form)
	return w, nil
}
