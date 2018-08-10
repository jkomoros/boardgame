package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/mattes/migrate"
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
		if err == migrate.ErrNoChange {
			fmt.Println("Already up to date")
		} else {
			errAndQuit("Couldn't call up on database: " + err.Error())
		}
	}
}
