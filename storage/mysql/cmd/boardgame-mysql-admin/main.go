/*
 * boardgame-mysql-admin helps create and migrate sql databases for boardgame.
 */
package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	dsnparser "github.com/go-sql-driver/mysql"
	"github.com/jkomoros/boardgame/server/config"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/mysql"
	_ "github.com/mattes/migrate/source/file"
	"log"
	"os"
)

const (
	pathToMigrations = "$GOPATH/src/github.com/jkomoros/boardgame/storage/mysql/migrations/"
)

type appOptions struct {
	Help    bool
	flagSet *flag.FlagSet
}

func defineFlags(options *appOptions) {
	options.flagSet.BoolVar(&options.Help, "help", false, "If true, will print help and exit.")
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

	path := os.ExpandEnv(pathToMigrations)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Println("The migrations path does not appear to exist")
		return
	}

	cfg, err := config.Get()

	if err != nil {
		log.Println("invalid config: " + err.Error())
		return
	}

	configToUse := cfg.Dev

	if configToUse.StorageConfig["mysql"] == "" {
		log.Println("No connection string configured for mysql")
		return
	}

	dsn, err := getDSN(configToUse.StorageConfig["mysql"])

	if err != nil {
		log.Println(err)
		return
	}

	db, _ := sql.Open("mysql", dsn)
	driver, _ := mysql.WithInstance(db, &mysql.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+path,
		"mysql",
		driver,
	)

	if err != nil {
		log.Println("Couldnt' create migration instance: " + err.Error())
		return
	}

	switch options.flagSet.Arg(0) {
	case "up":
		if prodConfirm(true) {
			doUp(m)
		}

	case "down":
		if prodConfirm(true) {
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
	help := `Commands: 'version', 'up', 'down'`
	log.Println(help)
}

func doUp(m *migrate.Migrate) {
	if err := m.Up(); err != nil {
		log.Println("Up failed: " + err.Error())
	}
}

func doDown(m *migrate.Migrate) {
	if err := m.Down(); err != nil {
		log.Println("Down failed: " + err.Error())
	}
}

func doVersion(m *migrate.Migrate) {
	version, _, _ := m.Version()
	log.Println("Version: ", version)
}

func getDSN(config string) (string, error) {

	//Substantially recreated in mysql/main.go

	parsedDSN, err := dsnparser.ParseDSN(config)

	if err != nil {
		return "", errors.New("config provided was not valid DSN: " + err.Error())
	}

	parsedDSN.Collation = "utf8mb4_unicode_ci"
	parsedDSN.MultiStatements = true

	return parsedDSN.FormatDSN(), nil
}
