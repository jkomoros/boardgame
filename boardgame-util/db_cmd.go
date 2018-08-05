package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type Db struct {
	baseSubCommand
	Up      DbUp
	Down    DbDown
	Setup   DbSetup
	Version DbVersion
}

func (d *Db) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New("SUBCOMMAND is required"))
}

func (d *Db) Name() string {
	return "db"
}

func (d *Db) Aliases() []string {
	return []string{
		"mysql",
	}
}

func (d *Db) Description() string {
	return "Configures a mysql database"
}

func (d *Db) Usage() string {
	return "SUBCOMMAND"
}

func (d *Db) SubcommandObjects() []SubcommandObject {

	return []SubcommandObject{
		&d.Up,
		&d.Down,
		&d.Setup,
		&d.Version,
	}

}
