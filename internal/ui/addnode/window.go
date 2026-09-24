package addnode

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
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Joey574/sink/v2/pkg/sink"
)

const (
	automaticLabel = "Automatic (ssh-agent, then default keys)"
	browseLabel    = "Browse..."
)

type AddNode struct {
	a    *app.App
	w    fyne.Window
	sink sink.Sink

	ctx    context.Context
	cancel context.CancelFunc

	user, host *widget.Entry
	identity   *widget.Select
	status     *widget.Label
	submit     *widget.Button

	identities map[string]string
}

func New(a *app.App) app.Window {
	w := a.Fyne.NewWindow(ids.AddNodeID)
	w.Resize(fyne.NewSize(680, 360))

	ctx, cancel := context.WithCancel(context.Background())

	f := &AddNode{
		a:      a,
		w:      w,
		ctx:    ctx,
		cancel: cancel,
		sink: sink.New(
			sink.Wrap(a.Sink),
			sink.SetName("addnode"),
		),
		identities: map[string]string{automaticLabel: ""},
	}

	f.user = widget.NewEntry()
	f.user.SetPlaceHolder("remote user ex admin")

	f.host = widget.NewEntry()
	f.host.SetPlaceHolder("remote host ex 192.168.1.10 or 192.168.1.10:2222")
	f.host.OnSubmitted = func(s string) {
		f.onSubmit()
	}

	f.identity = widget.NewSelect(f.identityOptions(), f.onIdentityChanged)
	f.identity.SetSelected(automaticLabel)

	f.status = widget.NewLabel(f.agentSummary())
	f.status.Wrapping = fyne.TextWrapWord

	f.submit = widget.NewButton("Add node", f.onSubmit)
	f.submit.Importance = widget.HighImportance

	w.SetContent(container.NewBorder(
		nil,
		container.NewVBox(f.status, f.submit),
		nil, nil,
		widget.NewForm(
			widget.NewFormItem("user", f.user),
			widget.NewFormItem("host", f.host),
			widget.NewFormItem("ssh key", f.identity),
		),
	))

	return f
}

func (f *AddNode) Window() fyne.Window {
	return f.w
}

func (f *AddNode) Close() error {
	f.cancel()
	return nil
}

func (f *AddNode) identityOptions() []string {
	options := []string{automaticLabel}

	found, err := ssh.DiscoverIdentities()
	if err != nil {
		f.sink.Printf(sink.WARN, "scanning ~/.ssh failed: %v\n", err)
	}

	for _, identity := range found {
		label := identity.Label()
		f.identities[label] = identity.Path
		options = append(options, label)
	}

	return append(options, browseLabel)
}

func (f *AddNode) agentSummary() string {
	keys, err := ssh.AgentIdentities()
	if err != nil {
		return fmt.Sprintf("ssh-agent could not be queried: %v", err)
	}

	if len(keys) == 0 {
		return "ssh-agent holds no keys, automatic will fall back to the default key files."
	}

	return fmt.Sprintf("ssh-agent holds %d key(s) that automatic will try first", len(keys))
}

func (f *AddNode) onIdentityChanged(selected string) {
	if selected == browseLabel {
		f.browseIdentity()
	}
}

func (f *AddNode) browseIdentity() {
	picker := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil || file == nil {
			if err != nil {
				dialog.ShowError(err, f.w)
			}

			f.identity.SetSelected(automaticLabel)
			return
		}
		defer file.Close()

		identity, err := ssh.InspectIdentity(file.URI().Path())
		if err != nil {
			dialog.ShowError(err, f.w)
			f.identity.SetSelected(automaticLabel)
			return
		}

		label := identity.Label()
		if _, known := f.identities[label]; !known {
			f.identities[label] = identity.Path

			last := len(f.identity.Options) - 1
			options := append([]string{}, f.identity.Options[:last]...)
			f.identity.Options = append(options, label, browseLabel)
		}

		f.identity.SetSelected(label)
	}, f.w)

	if dir, err := storage.ListerForURI(storage.NewFileURI(ssh.Dir())); err == nil {
		picker.SetLocation(dir)
	}

	picker.Show()
}

func (f *AddNode) onSubmit() {
	user := strings.TrimSpace(f.user.Text)
	host := strings.TrimSpace(f.host.Text)
	identityFile := f.identities[f.identity.Selected]

	if user == "" || host == "" {
		f.status.SetText("user and host are both required")
		return
	}

	f.setBusy(true)
	go f.addNode(user, host, identityFile)
}

func (f *AddNode) addNode(user, host, identityFile string) {
	defer fyne.Do(func() { f.setBusy(false) })

	f.sink.Printf(sink.DEBUG, "add node request, user='%s', host='%s', identity='%s'\n", user, host, identityFile)
	queries := f.a.Dbu.Queries()

	if _, err := queries.GetNodeByHost(f.ctx, host); err == nil {
		f.fail(fmt.Errorf("a node with host %s already exists", host))
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		f.fail(fmt.Errorf("looking up %s failed: %w", host, err))
		return
	}

	f.setStatus("fetching host key of " + host + "...")
	info, err := f.a.Engine.ProbeHost(f.ctx, host)
	if err != nil {
		f.fail(fmt.Errorf("could not reach %s: %w", host, err))
		return
	}

	if info.KnownHosts == ssh.KnownHostsMatch {
		f.sink.Printf(sink.INFO, "host key of %s matches known_hosts, trusting it\n", host)
	} else if !prompts.ConfirmHostKey(f.w, host, info) {
		f.sink.Printf(sink.INFO, "host key of %s was not trusted, node not added\n", host)
		f.setStatus("host key not trusted, node not added")
		return
	}

	node := db.Node{
		User:         user,
		Host:         host,
		HostKey:      ssh.MarshalHostKey(info.Key),
		IdentityFile: identityFile,
	}

	f.setStatus("testing connection to " + host + "...")
	if err := f.a.Engine.TestConnection(f.ctx, &node, prompts.Passphrase(f.w)); err != nil {
		if errors.Is(err, ssh.ErrPromptCanceled) {
			f.setStatus("passphrase prompt cancelled, not not added")
			return
		}

		f.fail(fmt.Errorf("test connection failed: %w", err))
		return
	}

	created, err := queries.CreateNode(f.ctx, db.CreateNodeParams{
		User:         node.User,
		Host:         node.Host,
		HostKey:      node.HostKey,
		IdentityFile: node.IdentityFile,
	})
	if err != nil {
		f.fail(fmt.Errorf("create node failed: %w", err))
		return
	}

	f.sink.Printf(sink.INFO, "added node %d (%s@%s), host key %s\n",
		created.ID, created.User, created.Host, ssh.Fingerprint(info.Key))

	fyne.Do(f.w.Close)
}

func (f *AddNode) fail(err error) {
	if errors.Is(err, context.Canceled) {
		return
	}

	f.sink.Printf(sink.ERROR, "%v\n", err)
	f.setStatus(err.Error())
	prompts.ShowError(f.w, err)
}

func (f *AddNode) setStatus(text string) {
	fyne.Do(func() { f.status.SetText(text) })
}

func (f *AddNode) setBusy(busy bool) {
	for _, control := range []fyne.Disableable{f.user, f.host, f.identity, f.submit} {
		if busy {
			control.Disable()
		} else {
			control.Enable()
		}
	}
}
