package editnode

import (
	"context"
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/ui/ids"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/Joey574/sink/v2/pkg/sink"
)

type EditNode struct {
	w fyne.Window
}

func New(a *app.App) app.Window {
	w := a.Fyne.NewWindow(ids.EditNodeID)
	w.Resize(fyne.NewSize(680, 460))

	s := sink.New(
		sink.Wrap(a.Sink),
		sink.SetName(ids.EditNodeID),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := a.Dbu.Queries()
	nodes, err := q.ListNodes(ctx)
	if err != nil {
		s.Printf(sink.ERROR, "list nodes: %v\n", err)
	}

	data := createTable(nodes)

	list := widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i.Row][i.Col])
		})

	w.SetContent(list)
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

func createTable(nodes []db.Node) [][]string {
	data := make([][]string, len(nodes)+1)

	data[0] = []string{
		"User",
		"Host",
		"SSH Key",
		"Config File",
	}

	for i, n := range nodes {
		data[i+1] = []string{
			n.User,
			n.Host,
			n.IdentityFile,
			n.ConfigFile,
		}
	}

	return data
}
