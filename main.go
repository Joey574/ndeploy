package main

import (
	"fmt"
	"log"
	"ndeploy/v2/internal/cli"
	"ndeploy/v2/internal/dbu"
	"os"

	"github.com/jessevdk/go-flags"
	_ "modernc.org/sqlite"
)

func main() {
	args := cli.NewArgs()
	_, err := args.Parse()
	if err != nil {
		if flags.WroteHelp(err) {
			os.Exit(0)
		}

		log.Fatalln(err)
	}

	if err = run(args); err != nil {
		log.Fatalln(err)
	}
}

func run(args *cli.Args) error {
	fmt.Println("starting ndeploy...")

	workDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println("workdir:", workDir)

	fmt.Println(dbu.DatabaseExists(workDir))
	return nil
}
