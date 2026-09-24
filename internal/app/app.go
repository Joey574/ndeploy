package app

import (
	"embed"
	"ndeploy/v2/internal/cli"
	"ndeploy/v2/internal/dbu"
	"ndeploy/v2/internal/engine"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/Joey574/sink/v2/pkg/ds/rb"
	"github.com/Joey574/sink/v2/pkg/sink"
)

const dbName = "sqlite.db"

type WindowFactory func(*App) Window

type Window interface {
	Window() fyne.Window
	Close() error
}

type App struct {
	Fyne      fyne.App
	windows   map[string]Window
	factories map[string]WindowFactory

	WorkDir string

	Engine     *engine.Engine
	Sink       sink.Sink
	Dbu        *dbu.Dbu
	RingBuffer *rb.RingBuffer
}

func NewApp(args *cli.Args, options ...func(*App)) (*App, error) {
	rb := rb.New(16 * 1024 * 1024) // 16MB ring buffer
	s := sink.New(
		sink.PushSinks(rb, os.Stdout),
		sink.SetLogLevel(sink.TRACE),
		sink.SetFormat(`[\d] [\t] \s: *`),
		sink.SetName("main"),
		sink.ThreadSafe(),
	)

	a := &App{
		Fyne:      app.NewWithID("ndeploy"),
		windows:   make(map[string]Window),
		factories: make(map[string]WindowFactory),

		Sink:       s,
		RingBuffer: rb,
		Engine: engine.NewEngine(
			engine.SetSink(sink.New(
				sink.Wrap(s),
				sink.SetName("engine"),
			)),
		),
	}

	// setup workdir
	dir, err := a.setupWorkDir(args)
	if err != nil {
		return nil, err
	}
	a.WorkDir = dir

	// init dbu
	a.Dbu = dbu.New(
		filepath.Join(a.WorkDir, dbName),
		dbu.SetSink(sink.New(
			sink.Wrap(a.Sink),
			sink.SetName("dbu"),
		)),
	)

	for _, o := range options {
		o(a)
	}

	return a, nil
}

func (a *App) Register(id string, factory WindowFactory) {
	a.factories[id] = factory
}

func (a *App) OpenOrFocus(id string) Window {
	if w, ok := a.windows[id]; ok {
		w.Window().RequestFocus()
		return w
	}

	factory, ok := a.factories[id]
	if !ok {
		panic("no window registered for id: " + id)
	}

	w := factory(a)
	a.windows[id] = w
	w.Window().SetOnClosed(func() {
		delete(a.windows, id)
		if err := w.Close(); err != nil {
			a.Sink.Printf(sink.ERROR, "on close: %v\n", err)
		}
	})

	w.Window().Show()
	return w
}

func (a *App) Run(args *cli.Args, schema embed.FS) error {
	if err := a.Dbu.CreateIfNotExists(schema); err != nil {
		return err
	}

	if err := a.Dbu.Connect(); err != nil {
		return err
	}
	a.Sink.PushStores(dbu.NewDBStore("dbstore", a.Dbu))

	a.Fyne.Run()
	return nil
}
