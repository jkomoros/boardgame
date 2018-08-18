package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type Help struct {
	baseSubCommand
}

func (h *Help) Run(p writ.Path, positional []string) {

	if h.base == nil {
		p.Last().ExitHelp(errors.New("BUG: help didn't have reference to base command"))
	}

	if len(positional) == 0 {
		h.base.WritCommand().ExitHelp(nil)
	}

	subCmd := selectSubcommandObject(h.base, append([]string{h.base.Name()}, positional...))

	if subCmd == nil {
		p.Last().ExitHelp(errors.New(positional[0] + " is not a valid subcommand"))
	}

	subCmd.WritCommand().ExitHelp(nil)

}

func (h *Help) Usage() string {
	return "SUBCOMMAND"
}

func (h *Help) Name() string {
	return "help"
}

func (h *Help) Description() string {
	return "Prints help for a specific subcommand"
}