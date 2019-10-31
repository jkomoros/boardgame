package main

import (
	"errors"

	"github.com/bobziuchkovski/writ"
)

type help struct {
	baseSubCommand
}

func (h *help) Run(p writ.Path, positional []string) {

	if h.Base() == nil {
		p.Last().ExitHelp(errors.New("BUG: help didn't have reference to base command"))
	}

	if len(positional) == 0 {
		h.Base().WritCommand().ExitHelp(nil)
	}

	subCmd := selectSubcommandObject(h.Base(), append([]string{h.base.Name()}, positional...))

	if subCmd == nil {
		p.Last().ExitHelp(errors.New(positional[0] + " is not a valid subcommand"))
	}

	subCmd.WritCommand().ExitHelp(nil)

}

func (h *help) Usage() string {
	return "SUBCOMMAND"
}

func (h *help) Name() string {
	return "help"
}

func (h *help) Description() string {
	return "Prints help for a specific subcommand"
}
