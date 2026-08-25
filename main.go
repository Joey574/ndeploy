package main

import (
	"embed"
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/cli"
	"ndeploy/v2/internal/sink"
	"ndeploy/v2/internal/ui/ids"
	"ndeploy/v2/internal/ui/logviewer"
	"ndeploy/v2/internal/ui/mainwindow"
	"os"

	"github.com/jessevdk/go-flags"
	_ "modernc.org/sqlite"
)

//go:embed sql/schema
var schema embed.FS

func main() {
	sink.SetFormat(`[\d] [\t] *`)
	sink.SetLogLevel(sink.TRACE) // TODO : let user define this
	sink.PushSinks(os.Stdout)

	args := cli.NewArgs()
	_, err := args.Parse()
	if err != nil {
		if flags.WroteHelp(err) {
			return
		}

		sink.Fatalln(err)
	}

	a := app.NewApp()
	a.Register(ids.MainWindowID, mainwindow.New)
	a.Register(ids.LogViewerID, logviewer.New)

	a.OpenOrFocus(ids.MainWindowID)
	if err := a.Run(args, schema); err != nil {
		sink.Fatalln(err)
	}
}
