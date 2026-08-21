package app

import (
	"embed"
	"ndeploy/v2/internal/cli"
	"ndeploy/v2/internal/dbu"
	"ndeploy/v2/internal/sink"
	"os"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type App struct {
	workDir string
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run(args *cli.Args, schema embed.FS) error {
	var err error

	a.workDir, err = a.setupWorkDir(args)
	if err != nil {
		return err
	}

	if err := dbu.CreateIfNotExists(a.workDir, schema); err != nil {
		return err
	}

	app := app.NewWithID("ndeploy")
	w := app.NewWindow("hello")

	message := widget.NewLabel("welcome")
	button := widget.NewButton("Update", func() {
		formatted := time.Now().Format("Time: 03:04:05")
		message.SetText(formatted)
	})

	w.SetContent(container.NewVBox(message, button))
	w.ShowAndRun()

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
