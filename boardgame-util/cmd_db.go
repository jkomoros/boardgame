package main

import (
	"errors"
	"fmt"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/storage/mysql/connect"
	"github.com/mattes/migrate"
)

type db struct {
	baseSubCommand
	Up      dbUp
	Down    dbDown
	Setup   dbSetup
	Version dbVersion
	Prod    bool
}

func (d *db) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New("SUBCOMMAND is required"))
}

func (d *db) Name() string {
	return "db"
}

func (d *db) Aliases() []string {
	return []string{
		"mysql",
	}
}

func (d *db) Description() string {
	return "Configures a mysql database"
}

func (d *db) HelpText() string {
	return d.Name() +

		` helps set up and administer mysql databases for use with boardgame, both
locally and in in prod. 

Reads configuration to connect to the mysql databse from config.json. See
README.md for more about configuring that file.

` + d.Base().Name() +
		` deploy often runs "db up", and "db setup" automatically.`

}

func (d *db) Usage() string {
	return "SUBCOMMAND"
}

func (d *db) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"prod", "p"},
			Flag:        true,
			Description: "If true, uses prod settings instead of dev settings",
			Decoder:     writ.NewFlagDecoder(&d.Prod),
		},
	}
}

func (d *db) SubcommandObjects() []SubcommandObject {

	return []SubcommandObject{
		&d.Up,
		&d.Down,
		&d.Setup,
		&d.Version,
	}

}

func baseConfirm(message string) bool {
	fmt.Println(message + " Are you sure? (y/N)")
	var response string
	fmt.Scanln(&response)
	yesResponses := []string{"Yes", "Y", "yes", "y"}
	for _, responseToTest := range yesResponses {
		if response == responseToTest {
			return true
		}
	}
	return false
}

func (d *db) prodConfirm() bool {
	if !d.Prod {
		return true
	}
	return baseConfirm("You have selected a destructive action on prod.")
}

func (d *db) GetMigrate(createDb bool) *migrate.Migrate {

	config := d.Base().GetConfig(false)

	if !d.prodConfirm() {
		d.Base().msgAndQuit("Didn't agree to operate on prod")
	}

	mode := config.Dev

	if d.Prod {
		mode = config.Prod
	}

	dsn, ok := mode.Storage["mysql"]

	if !ok {
		d.Base().errAndQuit("No mysql config provided")
	}

	db, err := connect.Db(dsn, false, createDb)

	if err != nil {
		d.Base().errAndQuit("Couldn't connect to database: " + err.Error())
	}

	m, err := connect.Migrations(db)

	if err != nil {
		d.Base().errAndQuit("Couldn't get migrations handle: " + err.Error())
	}

	return m

}
