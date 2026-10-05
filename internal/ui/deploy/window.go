package deploy

import (
	"context"
	"errors"
	"fmt"
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/db"
	"ndeploy/v2/internal/engine"
	"ndeploy/v2/internal/ui/prompts"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Joey574/sink/v2/pkg/sink"
)

const (
	buildOnTarget     = "Build on target"
	buildOnController = "Build locally"

	maxLines      = 5000
	flushInterval = 100 * time.Millisecond
)

type Deploy struct {
	a    *app.App
	w    fyne.Window
	sink sink.Sink

	ctx    context.Context
	cancel context.CancelFunc

	nodes []db.Node

	target, builder, action *widget.Select
	upgrade, rollback, sudo *widget.Check

	config       *widget.Entry
	browse       *widget.Button
	remember     *widget.Check
	deploy, stop *widget.Button
	status       *widget.Label

	output *widget.TextGrid
	scroll *container.Scroll

	mx    sync.Mutex
	lines []string
	dirty bool
	run   *engine.Run
}

func New(a *app.App) app.Window {
	w := a.Fyne.NewWindow(ids.DeployID)
	w.Resize(fyne.NewSize(960, 640))

	ctx, cancel := context.WithCancel(context.Background())

	d := &Deploy{
		a:      a,
		w:      w,
		ctx:    ctx,
		cancel: cancel,
		sink: sink.New(
			sink.Wrap(a.Sink),
			sink.SetName(ids.DeployID),
		),
	}

	nodes, err := a.Dbu.Queries().ListNodes(ctx)
	if err != nil {
		d.sink.Printf(sink.ERROR, "list nodes: %v\n", err)
	}
	d.nodes = nodes

	labels := make([]string, len(nodes))
	for i, n := range nodes {
		labels[i] = nodeLabel(n)
	}

	d.target = widget.NewSelect(labels, d.onTargetChanged)
	d.target.PlaceHolder = "select a node"

	d.builder = widget.NewSelect(append([]string{buildOnTarget, buildOnController}, labels...), nil)
	d.builder.SetSelected(buildOnController)

	actions := make([]string, 0, len(engine.DeploymentTypes()))
	for _, t := range engine.DeploymentTypes() {
		actions = append(actions, t.String())
	}

	d.action = widget.NewSelect(actions, nil)
	d.action.SetSelected(engine.Switch.String())

	d.config = widget.NewEntry()
	d.config.SetPlaceHolder("configuration.nix for the target")
	d.browse = widget.NewButton("Browse...", d.browseConfig)

	d.upgrade = widget.NewCheck("Upgrade channels", nil)
	d.rollback = widget.NewCheck("Roll back to previous generation", func(rollback bool) {
		if rollback {
			d.config.Disable()
			d.browse.Disable()
		} else {
			d.config.Enable()
			d.browse.Enable()
		}
	})

	d.sudo = widget.NewCheck("Use sudo on target", nil)
	d.sudo.SetChecked(true)

	d.remember = widget.NewCheck("Save config path to node", nil)
	d.remember.SetChecked(true)

	d.status = widget.NewLabel("")
	d.status.Wrapping = fyne.TextWrapWord

	d.deploy = widget.NewButton("Deploy", d.onDeploy)
	d.deploy.Importance = widget.HighImportance

	d.stop = widget.NewButton("Cancel", d.onCancel)
	d.stop.Importance = widget.DangerImportance
	d.stop.Disable()

	d.output = widget.NewTextGrid()
	d.scroll = container.NewScroll(d.output)

	form := widget.NewForm(
		widget.NewFormItem("target", d.target),
		widget.NewFormItem("build on", d.builder),
		widget.NewFormItem("action", d.action),
		widget.NewFormItem("config", container.NewBorder(nil, nil, nil, d.browse, d.config)),
		widget.NewFormItem("options", container.NewVBox(d.sudo, d.upgrade, d.rollback, d.remember)),
	)

	top := container.NewVBox(
		form,
		container.NewBorder(
			nil, nil, nil,
			container.NewHBox(d.stop, d.deploy),
			d.status,
		),
	)

	w.SetContent(container.NewBorder(top, nil, nil, nil, d.scroll))
	if len(nodes) == 0 {
		d.status.SetText("no nodes yet, add one first")
		d.deploy.Disable()
	} else {
		d.status.SetText(fmt.Sprintf("%d node(s) available", len(nodes)))
	}

	go d.flushLoop()
	return d
}

func (d *Deploy) Window() fyne.Window {
	return d.w
}

func (d *Deploy) Close() error {
	d.cancel()

	d.mx.Lock()
	run := d.run
	d.mx.Unlock()

	if run != nil && run.Running() {
		d.sink.Printf(sink.INFO, "deploy window closed while deployment %d is running, it keeps going\n", run.ID)
	}

	return nil
}

func nodeLabel(n db.Node) string {
	return n.User + "@" + n.Host
}

func (d *Deploy) onTargetChanged(string) {
	n, ok := d.selectedNode(d.target)
	if !ok {
		return
	}

	if strings.TrimSpace(d.config.Text) == "" || d.configBelongsToNode() {
		d.config.SetText(n.ConfigFile)
	}

	d.sudo.SetChecked(n.User != "root")
}

func (d *Deploy) configBelongsToNode() bool {
	text := strings.TrimSpace(d.config.Text)
	for _, n := range d.nodes {
		if n.ConfigFile != "" && n.ConfigFile == text {
			return true
		}
	}

	return false
}

func (d *Deploy) selectedNode(sel *widget.Select) (*db.Node, bool) {
	for i := range d.nodes {
		if nodeLabel(d.nodes[i]) == sel.Selected {
			return &d.nodes[i], true
		}
	}

	return nil, false
}

func (d *Deploy) browseConfig() {
	picker := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(err, d.w)
			return
		}

		if file == nil {
			return
		}
		defer file.Close()

		d.config.SetText(file.URI().Path())
	}, d.w)

	picker.SetFilter(storage.NewExtensionFileFilter([]string{".nix"}))

	if current := strings.TrimSpace(d.config.Text); current != "" {
		if dir, err := storage.ListerForURI(storage.NewFileURI(filepath.Dir(current))); err == nil {
			picker.SetLocation(dir)
		}
	}

	picker.Resize(fyne.NewSize(760, 520))
	picker.Show()
}

func (d *Deploy) deployment() (*engine.Deployment, error) {
	target, ok := d.selectedNode(d.target)
	if !ok {
		return nil, errors.New("select a target node")
	}

	t, ok := engine.ParseDeploymentType(d.action.Selected)
	if !ok {
		return nil, fmt.Errorf("unknown action %q", d.action.Selected)
	}

	dep := &engine.Deployment{
		Type:     t,
		Config:   strings.TrimSpace(d.config.Text),
		Upgrade:  d.upgrade.Checked,
		Rollback: d.upgrade.Checked,
		Sudo:     d.sudo.Checked,
		Target:   target,
		Output:   d,
	}

	switch d.builder.Selected {
	case buildOnTarget:
		dep.Builder = target
	case buildOnController:
		dep.Builder = nil
	default:
		builder, ok := d.selectedNode(d.builder)
		if !ok {
			return nil, errors.New("select a build node")
		}
		dep.Builder = builder
	}

	if err := dep.Validate(); err != nil {
		return nil, err
	}

	return dep, nil
}

func (d *Deploy) onDeploy() {
	dep, err := d.deployment()
	if err != nil {
		d.status.SetText(err.Error())
		return
	}

	if dep.Type.Activates() && !dep.Rollback {
		dialog.ShowConfirm(
			"Deploy?",
			fmt.Sprintf("nixos-rebuild %s on %s@%s\n\n%s", dep.Type, dep.Target.User, dep.Target.Host, dep.CommandLine()),
			func(ok bool) {
				if ok {
					d.start(dep)
				}
			},
			d.w,
		)
		return
	}

	d.start(dep)
}

func (d *Deploy) start(dep *engine.Deployment) {
	d.clearOutput()
	d.setBusy(true)
	d.status.SetText("starting " + dep.CommandLine())

	remember := d.remember.Checked && !dep.Rollback && dep.Config != dep.Target.ConfigFile
	target := *dep.Target

	go func() {
		run, err := d.a.Engine.StartDeployment(context.Background(), dep)
		if err != nil {
			d.fail(err)
			fyne.Do(func() { d.setBusy(false) })
			return
		}

		d.mx.Lock()
		d.run = run
		d.mx.Unlock()

		d.setStatus(fmt.Sprintf("deployment %d running: %s", run.ID, run.Command))

		if remember {
			err := d.a.Dbu.Queries().UpdateNodeConfigFile(context.Background(), db.UpdateNodeConfigFileParams{
				ID:         target.ID,
				ConfigFile: dep.Config,
			})
			if err != nil {
				d.sink.Printf(sink.WARN, "saving config path for node %d: %v\n", target.ID, err)
			} else {
				fyne.Do(func() { dep.Target.ConfigFile = dep.Config })
			}
		}

		<-run.Done()
		d.onFinished(run)
	}()
}

func (d *Deploy) onFinished(run *engine.Run) {
	elapsed := time.Since(run.StartedAt).Round(time.Second)

	var text string
	switch {
	case run.Cancelled():
		text = fmt.Sprintf("deployment %d cancelled after %s", run.ID, elapsed)
	case run.Err() != nil:
		text = fmt.Sprintf("deployment %d failed after %s: %v", run.ID, elapsed, run.Err())
	default:
		text = fmt.Sprintf("deployment %d finished in %s", run.ID, elapsed)
	}

	d.mx.Lock()
	if d.run == run {
		d.run = nil
	}
	d.mx.Unlock()

	fyne.Do(func() {
		d.status.SetText(text)
		d.setBusy(false)
	})
}

func (d *Deploy) onCancel() {
	d.mx.Lock()
	run := d.run
	d.mx.Unlock()

	if run == nil {
		return
	}

	d.status.SetText(fmt.Sprintf("cancelling deployment %d...", run.ID))
	d.stop.Disable()
	run.Cancel()
}

func (d *Deploy) Write(p []byte) (int, error) {
	if d.ctx.Err() != nil {
		return len(p), nil
	}

	text := strings.TrimRight(string(p), "\n")

	d.mx.Lock()
	defer d.mx.Unlock()

	d.lines = append(d.lines, strings.Split(text, "\n")...)
	if len(d.lines) > maxLines {
		d.lines = append([]string(nil), d.lines[len(d.lines)-maxLines:]...)
	}
	d.dirty = true
	return len(p), nil
}

func (d *Deploy) clearOutput() {
	d.mx.Lock()
	defer d.mx.Unlock()
	d.lines = nil
	d.dirty = true
}

func (d *Deploy) flushLoop() {
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.mx.Lock()
			if !d.dirty {
				d.mx.Unlock()
				continue
			}

			text := strings.Join(d.lines, "\n")
			d.dirty = false
			d.mx.Unlock()

			fyne.Do(func() {
				d.output.SetText(text)
				d.scroll.ScrollToBottom()
			})
		}
	}
}

func (d *Deploy) fail(err error) {
	if errors.Is(err, context.Canceled) {
		return
	}

	d.sink.Printf(sink.ERROR, "%v\n", err)
	d.setStatus(err.Error())
	prompts.ShowError(d.w, err)
}

func (d *Deploy) setStatus(text string) {
	fyne.Do(func() { d.status.SetText(text) })
}

func (d *Deploy) controls() []fyne.Disableable {
	return []fyne.Disableable{
		d.target, d.builder, d.action, d.config, d.browse,
		d.upgrade, d.rollback, d.sudo, d.remember, d.deploy,
	}
}

func (d *Deploy) setBusy(busy bool) {
	for _, c := range d.controls() {
		if busy {
			c.Disable()
		} else {
			c.Enable()
		}
	}

	if busy {
		d.stop.Enable()
	} else {
		d.stop.Disable()
		if d.rollback.Checked {
			d.config.Disable()
			d.browse.Disable()
		}
	}
}
