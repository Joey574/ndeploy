package identity

import (
	"fmt"
	"ndeploy/v2/internal/ssh"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/Joey574/sink/v2/pkg/sink"
)

const (
	automaticLabel = "Automatic (ssh-agent, then default keys)"
	browseLabel    = "Browse..."
)

type Picker struct {
	Select    *widget.Select
	OnChanged func(path string)

	w     fyne.Window
	sink  sink.Sink
	paths map[string]string
}

func New(w fyne.Window, s sink.Sink) *Picker {
	p := &Picker{
		w:     w,
		sink:  s,
		paths: map[string]string{automaticLabel: ""},
	}

	p.Select = widget.NewSelect(p.options(), p.onSelected)
	p.Select.SetSelected(automaticLabel)

	return p
}

func (p *Picker) Path() string {
	return p.paths[p.Select.Selected]
}

func (p *Picker) SetPath(path string) {
	if path == "" {
		p.Select.SetSelected(automaticLabel)
		return
	}

	for label, known := range p.paths {
		if known == path {
			p.Select.SetSelected(label)
			return
		}
	}

	identity, err := ssh.InspectIdentity(path)
	if err != nil {
		p.sink.Printf(sink.WARN, "stored identity %s: %v\n", path, err)
		identity = ssh.Identity{Path: path}
	}

	p.Select.SetSelected(p.add(identity))
}

func (p *Picker) Enable() {
	p.Select.Enable()
}

func (p *Picker) Disable() {
	p.Select.Disable()
}

func (p *Picker) options() []string {
	options := []string{automaticLabel}

	found, err := ssh.DiscoverIdentities()
	if err != nil {
		p.sink.Printf(sink.WARN, "scanning ~/.ssh failed: %v\n", err)
	}

	for _, identity := range found {
		label := identity.Label()
		p.paths[label] = identity.Path
		options = append(options, label)
	}

	return append(options, browseLabel)
}

func (p *Picker) add(identity ssh.Identity) string {
	label := identity.Label()
	if _, known := p.paths[label]; known {
		return label
	}

	p.paths[label] = identity.Path

	last := len(p.Select.Options) - 1
	options := append([]string{}, p.Select.Options[:last]...)
	p.Select.Options = append(options, label, browseLabel)

	return label
}

func (p *Picker) onSelected(selected string) {
	if selected == browseLabel {
		p.browse()
		return
	}

	if p.OnChanged != nil {
		p.OnChanged(p.paths[selected])
	}
}

func (p *Picker) browse() {
	picker := dialog.NewFileOpen(func(file fyne.URIReadCloser, err error) {
		if err != nil || file == nil {
			if err != nil {
				dialog.ShowError(err, p.w)
			}

			p.Select.SetSelected(automaticLabel)
			return
		}
		defer file.Close()

		identity, err := ssh.InspectIdentity(file.URI().Path())
		if err != nil {
			dialog.ShowError(err, p.w)
			p.Select.SetSelected(automaticLabel)
			return
		}

		p.Select.SetSelected(p.add(identity))
	}, p.w)

	if dir, err := storage.ListerForURI(storage.NewFileURI(ssh.Dir())); err == nil {
		picker.SetLocation(dir)
	}

	picker.Show()
}

func AgentSummary() string {
	keys, err := ssh.AgentIdentities()
	if err != nil {
		return fmt.Sprintf("ssh-agent could not be queries: %v", err)
	}

	if len(keys) == 0 {
		return "ssh-agent holds no keys, automatic will fall back to the default key files."
	}

	return fmt.Sprintf("ssh-agent holds %d key(s) that automatic will try first", len(keys))
}
