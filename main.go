package main

import (
	"fmt"
	"log"
	"nixos-deploy/v2/internal/cli"
	"nixos-deploy/v2/internal/identify"
	"nixos-deploy/v2/internal/remote"
	"os"

	"github.com/jessevdk/go-flags"
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

	remotes := make([]*remote.Remote, len(args.Remote))
	for i := range args.Remote {
		remotes[i] = remote.NewRemote(args.Remote[i])
	}

	for _, r := range remotes {
		deps, err := identify.FindDependencies(r.ConfigPath)
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Println(deps)
	}
}
