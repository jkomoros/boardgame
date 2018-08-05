/*

boardgame-mysql-admin helps create and migrate sql databases for boardgame.

*/
package main

import (
	"flag"
	"fmt"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/storage/mysql/connect"
	"github.com/mattes/migrate"
	"log"
	"os"
	"strings"
)

type appOptions struct {
	Help    bool
	Prod    bool
	flagSet *flag.FlagSet
}

func defineFlags(options *appOptions) {
	options.flagSet.BoolVar(&options.Help, "help", false, "If true, will print help and exit.")
	options.flagSet.BoolVar(&options.Prod, "prod", false, "If true will operate on prod. If omitted will default to dev")
}

func getOptions(flagSet *flag.FlagSet, flagArguments []string) *appOptions {
	options := &appOptions{flagSet: flagSet}
	defineFlags(options)
	flagSet.Parse(flagArguments)
	return options
}

func main() {
	flagSet := flag.CommandLine
	process(getOptions(flagSet, os.Args[1:]))
}

func process(options *appOptions) {

	if options.Help {
		doHelp()
		return
	}

	cfg, err := config.Get()

	if err != nil {
		log.Println("invalid config: " + err.Error())
		return
	}

	configToUse := cfg.Dev

	if options.Prod {
		configToUse = cfg.Prod
	}

	if configToUse.StorageConfig["mysql"] == "" {
		log.Println("No connection string configured for mysql")
		return
	}

	createDb := false
	if strings.ToLower(options.flagSet.Arg(0)) == "setup" {
		createDb = true
	}

	db, err := connect.Db(configToUse.StorageConfig["mysql"], false, createDb)

	if err != nil {
		log.Println(err)
		return
	}

	m, err := connect.Migrations(db)

	if err != nil {
		log.Println("Couldnt' create migration instance: " + err.Error())
		return
	}

	switch strings.ToLower(options.flagSet.Arg(0)) {
	case "up", "setup":
		if prodConfirm(options.Prod) {
			doUp(m)
		}

	case "down":
		if prodConfirm(options.Prod) {
			doDown(m)
		}
	default:
		doVersion(m)
	}

}

func prodConfirm(isProd bool) bool {
	if !isProd {
		return true
	}
	log.Println("You have selected a destructive action on prod. Are you sure? (y/N)")
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

func doHelp() {
	help := `Commands: 
* 'version' = Print version of migration and quit.
* 'up' = Apply all migrations forward (run on an existing db to make sure it's up to date)
* 'down' = Apply all migrations downward
* 'setup' = Create the 'boardgame' database and apply all migrations forward`
	log.Println(help)
}

func doUp(m *migrate.Migrate) {
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("Already up to date")
		} else {
			log.Println("Up failed: " + err.Error())
		}
	}
}

func doDown(m *migrate.Migrate) {
	if err := m.Down(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("Already at version 0")
		} else {
			log.Println("Down failed: " + err.Error())
		}
	}
}

func doVersion(m *migrate.Migrate) {
	version, _, _ := m.Version()
	log.Println("Version: ", version)
}
