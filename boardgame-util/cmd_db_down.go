package main

import (
	"fmt"

	"github.com/bobziuchkovski/writ"
	"github.com/mattes/migrate"
)

type dbDown struct {
	baseSubCommand
}

func (d *dbDown) Name() string {
	return "down"
}

func (d *dbDown) Description() string {
	return "Apply all migrations downward"
}

func (d *dbDown) Run(p writ.Path, positonal []string) {
	parent := d.Parent().(*db)

	m := parent.GetMigrate(false)

	if err := m.Down(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("Already at version 0")
		} else {
			d.Base().errAndQuit("Couldn't call down on database: " + err.Error())
		}
	}
}
