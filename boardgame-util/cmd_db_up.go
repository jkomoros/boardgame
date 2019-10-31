package main

import (
	"fmt"

	"github.com/bobziuchkovski/writ"
	"github.com/mattes/migrate"
)

type dbUp struct {
	baseSubCommand
}

func (d *dbUp) Name() string {
	return "up"
}

func (d *dbUp) Description() string {
	return "Apply all migrations forward (run on an existing db to make sure it's up to date)"
}

func (d *dbUp) Run(p writ.Path, positonal []string) {

	parent := d.Parent().(*Db)

	m := parent.GetMigrate(false)

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("Already up to date")
		} else {
			d.Base().errAndQuit("Couldn't call up on database: " + err.Error())
		}
	}
}
