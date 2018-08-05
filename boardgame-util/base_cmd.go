package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

//Note: whenever changing these, also change the struct tags in BoardgameUtil.
const cmdBase = "boardgame-util"
const cmdHelp = "help"
const cmdDb = "db"

type BoardgameUtil struct {
	Help Help `command:"help" description:"Prints help for a specific subcommand"`
	Db   Db   `command:"db" alias:"mysql" description:"Configures a mysql database"`
}

func (b *BoardgameUtil) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New("COMMAND is required"))
}
