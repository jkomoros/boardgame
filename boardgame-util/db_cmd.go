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

func (d *Db) HelpText() string {
	return d.Name() +

		` helps set up and administer mysql databases for use with boardgame, both
locally and in in prod. 

Reads configuration to connect to the mysql databse from config.json. See
README.md for more about configuring that file.

` + d.Base().Name() +
		` deploy often runs "db up", and "db setup" automatically.`

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
