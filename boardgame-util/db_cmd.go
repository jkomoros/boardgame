package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type Db struct {
}

func (d *Db) Run(p writ.Path, positional []string) {

	//TODO this should be implemented literally as a sub-command, not a positional arg.

	if len(positional) != 1 {
		p.Last().ExitHelp(errors.New(cmdDb + " requires one argument SUBCOMMAND"))
	}

}
