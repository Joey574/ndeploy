package cli

import (
	"github.com/jessevdk/go-flags"
)

type Args struct {
}

func NewArgs() *Args {
	return &Args{}
}

func (a *Args) Parse() ([]string, error) {
	positional, err := flags.Parse(a)
	if err != nil {
		return positional, err
	}

	return positional, nil
}
