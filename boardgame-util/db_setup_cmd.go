package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type DbSetup struct {
	baseSubCommand
}

func (d *DbSetup) Name() string {
	return "setup"
}

func (d *DbSetup) Description() string {
	return "Create the 'boardgame' database and apply all migrations forward"
}

func (d *DbSetup) Run(p writ.Path, positonal []string) {
	p.Last().ExitHelp(errors.New("Not yet implemented"))
}
