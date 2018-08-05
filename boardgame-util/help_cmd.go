package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type Help struct {
}

func (h *Help) Run(p writ.Path, positional []string) {
	if len(positional) != 1 {
		p.Last().ExitHelp(errors.New(cmdHelp + " requires one argument SUBCOMMAND"))
	}

	p.Last().ExitHelp(errors.New("No subcommands yet"))

}
