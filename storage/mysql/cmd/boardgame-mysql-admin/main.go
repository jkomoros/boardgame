/*
 * boardgame-mysql-admin helps create and migrate sql databases for boardgame.
 */
package main

import (
	"database/sql"
	"errors"
	"flag"
	dsnparser "github.com/go-sql-driver/mysql"
	"github.com/jkomoros/boardgame/server/config"
	"github.com/mattes/migrate"
	"github.com/mattes/migrate/database/mysql"
	_ "github.com/mattes/migrate/source/file"
	"log"
	"os"
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
		log.Println("You asked for help!")
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
	_, _ = migrate.NewWithDatabaseInstance(
		"file:///migrations",
		"mysql",
		driver,
	)
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
