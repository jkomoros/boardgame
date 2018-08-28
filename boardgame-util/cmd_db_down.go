package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/mattes/migrate"
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
	parent := d.Parent().(*Db)

	m := parent.GetMigrate(false)

	if err := m.Down(); err != nil {
		if err == migrate.ErrNoChange {
			fmt.Println("Already at version 0")
		} else {
			d.Base().errAndQuit("Couldn't call down on database: " + err.Error())
		}
	}
}
