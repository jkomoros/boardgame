/*
 * boardgame-mysql-admin helps create and migrate sql databases for boardgame.
 */
package main

import (
	"errors"
	"flag"
	"github.com/go-sql-driver/mysql"
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
	log.Println("Hello world!")
}

func getDSN(config string) (string, error) {

	//Substantially recreated in mysql/main.go

	parsedDSN, err := mysql.ParseDSN(config)

	if err != nil {
		return "", errors.New("config provided was not valid DSN: " + err.Error())
	}

	parsedDSN.Collation = "utf8mb4_unicode_ci"
	parsedDSN.MultiStatements = true

	return parsedDSN.FormatDSN(), nil
}
