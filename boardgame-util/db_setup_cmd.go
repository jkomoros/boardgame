package main

import (
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
	parent := d.Parent().(*Db)

	m := parent.GetMigrate(true)

	if err := m.Up(); err != nil {
		errAndQuit("Couldn't call up on database: " + err.Error())
	}
}
