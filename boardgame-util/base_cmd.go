package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type BoardgameUtil struct {
	baseSubCommand
	Help Help
	Db   Db
}

func (b *BoardgameUtil) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New("COMMAND is required"))
}

func (b *BoardgameUtil) Name() string {
	return "boardgame-util"
}

func (b *BoardgameUtil) Usage() string {
	return "COMMAND [OPTION]... [ARG]..."
}

func (b *BoardgameUtil) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&b.Help,
		&b.Db,
	}
}
