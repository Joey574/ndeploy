package app

import (
	"embed"
	"ndeploy/v2/internal/cli"
	"ndeploy/v2/internal/dbu"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/Joey574/sink/v2/pkg/ds/rb"
	"github.com/Joey574/sink/v2/pkg/sink"
)

const dbName = "sqlite.db"

type WindowFactory func(a *App) (fyne.Window, func())

type App struct {
	Fyne      fyne.App
	windows   map[string]fyne.Window
	factories map[string]WindowFactory

	WorkDir string

	Sink       *sink.Sink
	Dbu        *dbu.Dbu
	RingBuffer *rb.RingBuffer
}

func NewApp(args *cli.Args, options ...func(*App)) (*App, error) {
	a := &App{
		Fyne:      app.NewWithID("ndeploy"),
		windows:   make(map[string]fyne.Window),
		factories: make(map[string]WindowFactory),

		RingBuffer: rb.New(16 * 1024 * 1024), // 16MB ring buffer
		Sink: sink.New(
			sink.EnableStdOut(),
			sink.SetLogLevel(sink.TRACE),
			sink.SetFormat(`[\d] [\t] *`),
		),
	}
	a.Sink.PushSinks(a.RingBuffer)

	// setup workdir
	dir, err := a.setupWorkDir(args)
	if err != nil {
		return nil, err
	}
	a.WorkDir = dir

	// init dbu
	a.Dbu = dbu.New(
		filepath.Join(a.WorkDir, dbName),
		dbu.SetSink(a.Sink),
	)
	a.Sink.PushStores(dbu.NewDBStore("test", a.Dbu))

	for _, o := range options {
		o(a)
	}

	return a, nil
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
		panic("no window registered for id: " + id)
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
	if err := a.Dbu.CreateIfNotExists(schema); err != nil {
		return err
	}

	if err := a.Dbu.Connect(); err != nil {
		return err
	}

	a.Fyne.Run()
	return nil
}
