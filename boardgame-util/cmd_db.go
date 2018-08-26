package main

import (
	"errors"
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/storage/mysql/connect"
	"github.com/mattes/migrate"
)

type Db struct {
	baseSubCommand
	Up      DbUp
	Down    DbDown
	Setup   DbSetup
	Version DbVersion
	Prod    bool
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

func (d *Db) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"prod", "p"},
			Flag:        true,
			Description: "If true, uses prod settings instead of dev settings",
			Decoder:     writ.NewFlagDecoder(&d.Prod),
		},
	}
}

func (d *Db) SubcommandObjects() []SubcommandObject {

	return []SubcommandObject{
		&d.Up,
		&d.Down,
		&d.Setup,
		&d.Version,
	}

}

func (d *Db) prodConfirm() bool {
	if !d.Prod {
		return true
	}
	fmt.Println("You have selected a destructive action on prod. Are you sure? (y/N)")
	var response string
	fmt.Scanln(&response)
	yesResponses := []string{"Yes", "Y", "yes"}
	for _, responseToTest := range yesResponses {
		if response == responseToTest {
			return true
		}
	}
	return false
}

func (d *Db) GetMigrate(createDb bool) *migrate.Migrate {

	base := d.Base().(*BoardgameUtil)
	config := base.GetConfig(false)

	if !d.prodConfirm() {
		msgAndQuit("Didn't agree to operate on prod")
	}

	mode := config.Dev

	if d.Prod {
		mode = config.Prod
	}

	dsn, ok := mode.Storage["mysql"]

	if !ok {
		errAndQuit("No mysql config provided")
	}

	db, err := connect.Db(dsn, false, createDb)

	if err != nil {
		errAndQuit("Couldn't connect to database: " + err.Error())
	}

	m, err := connect.Migrations(db)

	if err != nil {
		errAndQuit("Couldn't get migrations handle: " + err.Error())
	}

	return m

}
