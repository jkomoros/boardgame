package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type Db struct {
	baseSubCommand
}

func (d *Db) Run(p writ.Path, positional []string) {

	//TODO this should be implemented literally as a sub-command, not a positional arg.

	if len(positional) != 1 {
		p.Last().ExitHelp(errors.New(d.Name() + " requires one argument SUBCOMMAND"))
	}

}

func (d *Db) Name() string {
	return "db"
}

func (d *Db) Aliases() []string {
	return []string{
		"mysql",
	}
}

func (d *Db) Description() string {
	return "Configures a mysql database"
}

func (d *Db) Usage() string {
	return "SUBCOMMAND"
}
