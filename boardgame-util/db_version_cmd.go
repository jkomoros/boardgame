package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type DbVersion struct {
	baseSubCommand
}

func (d *DbVersion) Name() string {
	return "version"
}

func (d *DbVersion) Description() string {
	return "Print what version of the migrations have been applied to db so far and quit"
}

func (d *DbVersion) Run(p writ.Path, positonal []string) {
	p.Last().ExitHelp(errors.New("Not yet implemented"))
}
