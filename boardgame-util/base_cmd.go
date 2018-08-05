package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

//TODO: don't use struct comments to do the top-level, so we can have the
//constant names once.

type BoardgameUtil struct {
	Help Help `command:"help" description:"Prints help for a specific subcommand"`
	Db   Db   `command:"db" alias:"mysql" description:"Configures a mysql database"`
	//When chaning these values, also change the associated command's Name() method.
}

func (b *BoardgameUtil) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New("COMMAND is required"))
}

func (b *BoardgameUtil) Name() string {
	return "boardgame-util"
}
