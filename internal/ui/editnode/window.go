package editnode

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/ssh"
	"ndeploy/v2/internal/ui/ids"
	"ndeploy/v2/internal/ui/prompts"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Joey574/sink/v2/pkg/sink"
)

type EditNode struct {
	a    *app.App
	w    fyne.Window
	sink sink.Sink

	ctx    context.Context
	cancel context.CancelFunc

	nodes    []db.Node
	selected int

	list *widget.List

	user, host, config *widget.Entry
	key                *identity.Picker
	browse             *widget.Button
	test, save, remove *widget.Button
	status             *widget.Label
}

func New(a *app.App) app.Window {
	w := a.Fyne.NewWindow(ids.EditNodeID)
	w.Resize(fyne.NewSize(860, 480))

	ctx, cancel := context.WithCancel(context.Background())

	e := &EditNode{
		a:      a,
		w:      w,
		ctx:    ctx,
		cancel: cancel,
		sink: sink.New(
			sink.Wrap(a.Sink),
			sink.SetName(ids.EditNodeID),
		),
		selected: -1,
	}

	e.list = widget.NewList(
		func() int { return len(e.nodes) },
		func() fyne.CanvasObject { return widget.NewLabel("user@host.example") },
		func(lii widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(nodeLabel(e.nodes[i]))
		},
	)

	e.list.OnSelected = e.onSelected
	e.list.OnUnselected = func(widget.ListItemID) { e.onSelected(-1) }

	e.user = widget.NewEntry()
	e.user.SetPlaceHolder("remote user ex admin")

	e.host = widget.NewEntry()
	e.host.SetPlaceHolder("remote host ex 192.169.1.10 or 192.168.1.10:2222")

	e.key = identity.NewPicket(w, e.sink)

	e.config = widget.NewEntry()
	e.config.SetPlaceHolder("configuration.nix for this node, on this machine")
	e.browse = widget.NewButton("Browse...", e.browseConfig)

	e.status = widget.NewLabel("select a node to edit")
	e.status.Wrapping = fyne.TextWrapWord

	e.test = widget.NewButton("Test connection", e.onTest)
	e.save = widget.NewButton("Save", e.onSave)
	e.save.Importance = widget.HighImportance
	e.remove = widget.NewButton("Delete", e.onDelete)
	e.remove.Importance = widget.DangerImportance

	form := widget.NewForm(
		widget.NewFormItem("user", e.user),
		widget.NewFormItem("host", e.host),
		widget.NewFormItem("ssh key", e.key.Select),
		widget.NewFormItem("config", container.NewBorder(nil, nil, nil, e.browse, e.config)),
	)

	buttons := container.NewHBox(e.remove, e.test, e.save)

	editor := container.NewBorder(
		nil,
		container.NewVBox(e.status, buttons),
		nil, nil,
		form,
	)

	split := container.NewHSplit(e.list, editor)
	split.SetOffset(0.3)
	w.SetContent(split)
	e.setEnabled(false)
	e.reload(-1)
	return e
}

func (e *EditNode) Window() fyne.Window {
	return e.w
}

func (e *EditNode) Close() error {
	e.cancel()
	return nil
}

func nodeLabel(n db.Node) string {
	return n.User + "@" + n.Host
}

func (e *EditNode) reload(keep int64) {
	nodes, err := e.a.Dbu.Queries().ListNodes(e.ctx)
	if err != nil {
		e.failNow(fmt.Errorf("list nodes: %w", err))
		return
	}

	e.sink.Printf(sink.DEBUG, "fetched %d nodes\n", len(nodes))
	e.nodes = nodes
	e.list.UnselectAll()
	e.list.Refresh()

	for i, n := range nodes {
		if n.ID == keep {
			e.list.Select(i)
			return
		}
	}

	e.onSelected(-1)
}

func (e *EditNode) onSelected(i int) {
	e.selected = i

	if i == -1 || i >= len(e.nodes) {
		e.selected = -1
		e.user.SetText("")
		e.host.SetText("")
		e.config.SetText("")
		e.key.SetPath("")
		e.setEnabled(false)
		e.status.SetText("select a node to edit")
		return
	}

	n := e.nodes[i]
	e.user.SetText(n.User)
	e.host.SetText(n.Host)
	e.config.SetText(n.ConfigFile)
	e.key.SetPath(n.IdentityFile)
	e.setEnabled(true)

	key, err := ssh.ParseHostKey(n.HostKey)
	if err != nil {
		e.status.SetText("stored host key is invalid, re-test the connection to fix it")
		return
	}

	e.status.SetText(fmt.Sprintf("host key %s %s", key.Type(), ssh.Fingerprint(key)))
}

func (e *EditNode) current() (*db.Node, bool) {
	if e.selected == -1 || e.selected >= len(e.nodes) {
		return nil, false
	}

	n := e.nodes[e.selected]
	return &n, true
}

func (e *EditNode) draft() (*db.Node, error) {
	stored, ok := e.current()
	if !ok {
		return nil, errors.New("no node selected")
	}

	n := *stored
	n.User = strings.TrimSpace(e.user.Text)
	n.Host = strings.TrimSpace(e.host.Text)
	n.IdentityFile = e.key.Path()
	n.ConfigFile = strings.TrimSpace(e.config.Text)

	if n.User == "" || n.Host == "" {
		return nil, errors.New("user and host are both required")
	}

	return &n, nil
}

func (e *EditNode) browseConfig() {
	picker := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, e.w)
			return
		}

		if file == nil {
			return
		}
		defer file.Close()

		e.config.SetText(file.URI().Path())
	}, e.w)

	picker.SetFilter(storage.NewExtensionFileFilter([]string{".nix"}))

	if current := strings.TrimSpace(e.config.Text); current != "" {
		if dir, err := storage.ListerForURI(storage.NewFileURI(filepath.Dir(current))); err == nil {
			picker.SetLocation(dir)
		}
	}

	picker.Resize(fyne.NewSize(760, 520))
	picker.Show()
}

func (e *EditNode) onTest() {
	n, err := e.draft()
	if err != nil {
		e.status.SetText(err.Error())
		return
	}

	stored, _ := e.current()
	e.setBusy(true)
	go func() {
		defer fyne.Do(func() { e.setBusy(false) })

		if err := e.verify(n, stored); err != nil {
			e.fail(err)
			return
		}

		e.setStatus(fmt.Sprintf("connection to %s@%s ok", n.User, n.Host))
	}()
}

func (e *EditNode) onSave() {
	n, err := e.draft()
	if err != nil {
		e.status.SetText(err.Error())
		return
	}

	stored, _ := e.current()

	e.setBusy(true)
	go e.saveNode(n, stored)
}

func (e *EditNode) saveNode(n, stored *db.Node) {
	defer fyne.Do(func() { e.setBusy(false) })

	e.sink.Printf(sink.DEBUG, "save node %d request, user='%s', host='%s', identity='%s', config='%s'\n",
		n.ID, n.User, n.Host, n.IdentityFile, n.ConfigFile)

	if n.User != stored.User || n.Host != stored.Host || n.IdentityFile != stored.IdentityFile {
		if err := e.verify(n, stored); err != nil {
			e.fail(err)
			return
		}
	}

	err := e.a.Dbu.Queries().UpdateNode(e.ctx, db.UpdateNodeParams{
		ID:           n.ID,
		User:         n.User,
		Host:         n.Host,
		HostKey:      n.HostKey,
		IdentityFile: n.IdentityFile,
		ConfigFile:   n.ConfigFile,
	})
	if err != nil {
		e.fail(fmt.Errorf("update node failed: %w", err))
		return
	}

	e.sink.Printf(sink.INFO, "updated node %d (%s@%s)\n", n.ID, n.User, n.Host)

	fyne.Do(func() {
		e.reload(n.ID)
		e.status.SetText(fmt.Sprintf("saved %s@%s", n.User, n.Host))
	})
}

func (e *EditNode) verify(n, stored *db.Node) error {
	q := e.a.Dbu.Queries()

	if n.Host != stored.Host {
		other, err := q.GetNodeByHost(e.ctx, n.Host)
		if err == nil && other.ID != n.ID {
			return fmt.Errorf("a node with host %s already exists", n.Host)
		} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("looking up %s failed: %w", n.Host, err)
		}

		if err := e.probe(n); err != nil {
			return err
		}
	}

	e.setStatus(fmt.Sprintf("testing connection to %s@%s...", n.User, n.Host))
	err := e.a.Engine.TestConnection(e.ctx, n, prompts.Passphrase(e.w))

	var mismatch *ssh.HostKeyMismatchError
	if errors.As(err, &mismatch) {
		e.sink.Printf(sink.WARN, "%v\n", err)
		if err := e.probe(n); err != nil {
			return err
		}

		e.setStatus(fmt.Sprintf("testing conection to %s@%s with new host key...", n.User, n.Host))
		err = e.a.Engine.TestConnection(e.ctx, n, prompts.Passphrase(e.w))
	}

	if errors.Is(err, ssh.ErrPromptCanceled) {
		return errors.New("passphrase prompt cancelled")
	}

	if err != nil {
		return fmt.Errorf("test connection failed: %w", err)
	}

	return nil
}

func (e *EditNode) probe(n *db.Node) error {
	e.setStatus("fetching host key of " + n.Host + "...")

	info, err := e.a.Engine.ProbeHost(e.ctx, n.Host)
	if err != nil {
		return fmt.Errorf("could not reach %s: %w", n.Host, err)
	}

	if info.KnownHosts == ssh.KnownHostsMatch {
		e.sink.Printf(sink.INFO, "host key of %s matches known_hosts, trusting it\n", n.Host)
	} else if !prompts.ConfirmHostKey(e.w, n.Host, info) {
		return errors.New("host key not trusted")
	}

	n.HostKey = ssh.MarshalHostKey(info.Key)
	return nil
}

func (e *EditNode) onDelete() {
	n, ok := e.current()
	if !ok {
		return
	}

	dialog.ShowConfirm(
		"Delete node?",
		fmt.Sprintf("Remove %s@%s and its deployment history?", n.User, n.Host),
		func(ok bool) {
			if !ok {
				return
			}

			if err := e.a.Dbu.Queries().DeleteNode(e.ctx, n.ID); err != nil {
				e.failNow(fmt.Errorf("delete node failed: %w", err))
				return
			}

			e.sink.Printf(sink.INFO, "deleted node %d (%s@%s)\n", n.ID, n.User, n.Host)
			e.reload(-1)
			e.status.SetText(fmt.Sprintf("deleted %s@%s", n.User, n.Host))
		},
		e.w,
	)
}

func (e *EditNode) fail(err error) {
	if errors.Is(err, context.Canceled) {
		return
	}

	e.sink.Printf(sink.ERROR, "%v\n", err)
	e.setStatus(err.Error())
	prompts.ShowError(e.w, err)
}

func (e *EditNode) failNow(err error) {
	e.sink.Printf(sink.ERROR, "%v\n", err)
	e.status.SetText(err.Error())
	dialog.ShowError(err, e.w)
}

func (e *EditNode) setStatus(text string) {
	fyne.Do(func() { e.status.SetText(text) })
}

func (e *EditNode) controls() []fyne.Disableable {
	return []fyne.Disableable{e.user, e.host, e.key.Select, e.config, e.browse, e.test, e.save, e.remove}
}

func (e *EditNode) setEnabled(enabled bool) {
	for _, control := range e.controls() {
		if enabled {
			control.Enable()
		} else {
			control.Disable()
		}
	}
}

func (e *EditNode) setBusy(busy bool) {
	e.setEnabled((!busy && e.selected == -1))
}
