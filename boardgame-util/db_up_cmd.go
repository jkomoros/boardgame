package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type DbUp struct {
	baseSubCommand
}

func (d *DbUp) Name() string {
	return "up"
}

func (d *DbUp) Description() string {
	return "Apply all migrations forward (run on an existing db to make sure it's up to date"
}

func (d *DbUp) Run(p writ.Path, positonal []string) {
	p.Last().ExitHelp(errors.New("Not yet implemented"))
}
