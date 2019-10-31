package main

import (
	"fmt"

	"github.com/bobziuchkovski/writ"
	"github.com/mattes/migrate"
)

type dbSetup struct {
	baseSubCommand
}

func (d *dbSetup) Name() string {
	return "setup"
}

func (d *dbSetup) Description() string {
	return "Create the 'boardgame' database and apply all migrations forward"
}

func (d *dbSetup) Run(p writ.Path, positonal []string) {
	parent := d.Parent().(*Db)

	m := parent.GetMigrate(true)

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("Already up to date")
		} else {
			d.Base().errAndQuit("Couldn't call up on database: " + err.Error())
		}
	}
}
