package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type DbDown struct {
	baseSubCommand
}

func (d *DbDown) Name() string {
	return "down"
}

func (d *DbDown) Description() string {
	return "Apply all migrations downward"
}

func (d *DbDown) Run(p writ.Path, positonal []string) {
	p.Last().ExitHelp(errors.New("Not yet implemented"))
}
