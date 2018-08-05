package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type BoardgameUtil struct {
	Help Help `command:"help" description:"Prints help for a specific subcommand"`
}

func (b *BoardgameUtil) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New("COMMAND is required"))
}
