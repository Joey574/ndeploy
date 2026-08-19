package main

import (
	"ndeploy/v2/internal/app"
	"ndeploy/v2/internal/cli"
	"ndeploy/v2/internal/sink"
	"os"

	"github.com/jessevdk/go-flags"
	_ "modernc.org/sqlite"
)

func main() {
	sink.SetFormat(`[\d] [\t] *`)
	sink.PushSinks(os.Stdout)
	sink.SetLogLevel(sink.TRACE) // TODO : let user define this

	args := cli.NewArgs()
	_, err := args.Parse()
	if err != nil {
		if flags.WroteHelp(err) {
			return
		}

		sink.Fatalln(err)
	}

	a := app.NewApp()
	if err := a.Run(args); err != nil {
		sink.Fatalln(err)
	}
}
