package app

import (
	"ndeploy/v2/internal/cli"
	"os"

	"github.com/Joey574/sink/v2/pkg/sink"
)

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

	a.Sink.Printf(sink.DEBUG, "work directory: %s\n", workDir)
	return workDir, os.MkdirAll(workDir, 0o750)
}
