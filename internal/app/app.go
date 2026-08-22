package app

import (
	"embed"
	"ndeploy/v2/internal/cli"
	"ndeploy/v2/internal/dbu"
	"ndeploy/v2/internal/sink"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	workDir string
	rb      *sink.RingBuffer
}

func NewApp() *App {
	return &App{
		rb: sink.NewRingBuffer(16 * 1024 * 1024), // 16MB ring buffer
	}
}

func (a *App) Run(args *cli.Args, schema embed.FS) error {
	var err error
	sink.PushSinks(a.rb)

	a.workDir, err = a.setupWorkDir(args)
	if err != nil {
		return err
	}

	if err := dbu.CreateIfNotExists(a.workDir, schema); err != nil {
		return err
	}

	lv := NewLogViewer(a.rb)
	go lv.Run(100 * time.Millisecond)

	fApp := app.NewWithID("ndeploy")
	w1 := fApp.NewWindow("hello")
	w2 := fApp.NewWindow("logs")

	w1.Resize(fyne.NewSize(300, 200))
	w2.Resize(fyne.NewSize(600, 400))

	w2.SetContent(lv.CanvasObject())
	w2.SetOnClosed(func() {
		lv.Stop()
	})

	message := widget.NewLabel("welcome")
	button := widget.NewButton("Update", func() {
		formatted := time.Now().Format("Time: 03:04:05")
		message.SetText(formatted)

		sink.Println(sink.TRACE, "button pressed :)")
	})

	w1.SetContent(container.NewVBox(message, button))
	w1.Show()
	w2.Show()
	fApp.Run()

	return nil
}

func (a *App) setupWorkDir(args *cli.Args) (string, error) {
	var err error
	var workDir string

	if args.WorkDir == "" {
		workDir, err = os.UserConfigDir()
		if err != nil {
			return workDir, err
		}

		workDir += "/ndeploy"
	} else {
		workDir = args.WorkDir
	}

	sink.Printf(sink.DEBUG, "work directory: %s\n", workDir)
	return workDir, os.MkdirAll(workDir, 0o750)
}
