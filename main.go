package main

import (
	"ndeploy/v2/internal/cli"
	"ndeploy/v2/internal/dbu"
	"ndeploy/v2/internal/sink"
	"os"

	"github.com/jessevdk/go-flags"
	_ "modernc.org/sqlite"
)

func main() {
	sink.SetFormat("[\\d] [\\t] *")
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

	if err = run(args); err != nil {
		sink.Fatalln(err)
	}
}

func run(args *cli.Args) error {
	sink.Println(sink.INFO, "starting ndeploy")

	var err error
	var workDir string

	if args.WorkDir != "" {
		workDir = args.WorkDir
	} else {
		workDir, err = os.UserConfigDir()
		if err != nil {
			return err
		}
		workDir += "/ndeploy"
	}

	if err := os.MkdirAll(workDir, 0o750); err != nil {
		return err
	}

	sink.Printf(sink.DEBUG, "work directory: %s\n", workDir)

	if !dbu.DatabaseExists(workDir) {
		sink.Println(sink.TRACE, "creating database")
		if err := dbu.InitDb(workDir); err != nil {
			sink.Fatalln(err)
		}
	}

	return nil
}
