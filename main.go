package main

import (
	"embed"
	"log"
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/cli"
	"ndeploy/v2/internal/ui/addnode"
	"ndeploy/v2/internal/ui/ids"
	"ndeploy/v2/internal/ui/logviewer"
	"ndeploy/v2/internal/ui/mainwindow"

	"github.com/jessevdk/go-flags"
	_ "modernc.org/sqlite"
)

//go:embed sql/schema
var schema embed.FS

func main() {
	args := cli.NewArgs()
	_, err := args.Parse()
	if err != nil {
		if flags.WroteHelp(err) {
			return
		}

		log.Fatalln(err)
	}

	a, err := app.NewApp(args)
	if err != nil {
		log.Fatalln(err)
	}

	a.Register(ids.MainWindowID, mainwindow.New)
	a.Register(ids.LogViewerID, logviewer.New)
	a.Register(ids.AddNodeID, addnode.New)

	a.OpenOrFocus(ids.MainWindowID)
	if err := a.Run(args, schema); err != nil {
		log.Fatalln(err)
	}
}
