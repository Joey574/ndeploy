package app

import (
	"embed"
	"ndeploy/v2/internal/cli"
	"ndeploy/v2/internal/dbu"
	"ndeploy/v2/internal/sink"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
)

type WindowFactory func(a *App) (fyne.Window, func())

type App struct {
	Fyne      fyne.App
	windows   map[string]fyne.Window
	factories map[string]WindowFactory

	WorkDir    string
	RingBuffer *sink.RingBuffer
}

func NewApp() *App {
	return &App{
		Fyne:       app.NewWithID("ndeploy"),
		windows:    make(map[string]fyne.Window),
		factories:  make(map[string]WindowFactory),
		RingBuffer: sink.NewRingBuffer(16 * 1024 * 1024), // 16MB ring buffer
	}
}

func (a *App) Register(id string, factory WindowFactory) {
	a.factories[id] = factory
}

func (a *App) OpenOrFocus(id string) fyne.Window {
	if w, ok := a.windows[id]; ok {
		w.RequestFocus()
		return w
	}

	factory, ok := a.factories[id]
	if !ok {
		panic("no window registered for idd: " + id)
	}

	w, cleanup := factory(a)
	a.windows[id] = w
	w.SetOnClosed(func() {
		delete(a.windows, id)
		if cleanup != nil {
			cleanup()
		}
	})
	w.Show()
	return w
}

func (a *App) Run(args *cli.Args, schema embed.FS) error {
	var err error
	sink.PushSinks(a.RingBuffer)

	a.WorkDir, err = a.setupWorkDir(args)
	if err != nil {
		return err
	}

	if err := dbu.CreateIfNotExists(a.WorkDir, schema); err != nil {
		return err
	}

	a.Fyne.Run()
	return nil
}
