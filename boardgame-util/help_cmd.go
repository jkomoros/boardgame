package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type Help struct {
	base *writ.Command
}

func (h *Help) Run(p writ.Path, positional []string) {
	if len(positional) != 1 {
		p.Last().ExitHelp(errors.New(h.Name() + " requires one argument SUBCOMMAND"))
	}

	if h.base == nil {
		p.Last().ExitHelp(errors.New("BUG: help didn't have reference to base command"))
	}

	subCmd := h.base.Subcommand(positional[0])

	if subCmd == nil {
		p.Last().ExitHelp(errors.New(positional[0] + " is not a valid subcommand"))
	}

	subCmd.ExitHelp(nil)

}

func (h *Help) Name() string {
	return "help"
}
