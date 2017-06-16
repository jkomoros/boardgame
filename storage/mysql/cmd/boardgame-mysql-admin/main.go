/*
 * boardgame-mysql-admin helps create and migrate sql databases for boardgame.
 */
package main

import (
	"flag"
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
