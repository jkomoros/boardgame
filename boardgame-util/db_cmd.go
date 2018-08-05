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

type DbUp struct {
	baseSubCommand
}

func (d *DbUp) Name() string {
	return "up"
}

func (d *DbUp) Description() string {
	return "Apply all migrations forward (run on an existing db to make sure it's up to date"
}

func (d *DbUp) Run(p writ.Path, positonal []string) {
	p.Last().ExitHelp(errors.New("Not yet implemented"))
}

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
	p.Last().ExitHelp(errors.New("Not yet implemented"))
}

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
	p.Last().ExitHelp(errors.New("Not yet implemented"))
}

type DbVersion struct {
	baseSubCommand
}

func (d *DbVersion) Name() string {
	return "version"
}

func (d *DbVersion) Description() string {
	return "Print version of migration and quit"
}

func (d *DbVersion) Run(p writ.Path, positonal []string) {
	p.Last().ExitHelp(errors.New("Not yet implemented"))
}
