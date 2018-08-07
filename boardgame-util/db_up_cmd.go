package main

import (
	"github.com/bobziuchkovski/writ"
)

type DbUp struct {
	baseSubCommand
}

func (d *DbUp) Name() string {
	return "up"
}

func (d *DbUp) Description() string {
	return "Apply all migrations forward (run on an existing db to make sure it's up to date)"
}

func (d *DbUp) Run(p writ.Path, positonal []string) {

	parent := d.Parent().(*Db)

	m := parent.GetMigrate(false)

	if err := m.Up(); err != nil {
		errAndQuit("Couldn't call up on database: " + err.Error())
	}
}
