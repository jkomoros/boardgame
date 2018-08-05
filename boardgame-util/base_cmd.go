package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type BoardgameUtil struct {
	Help Help
	Db   Db
}

func (b *BoardgameUtil) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New("COMMAND is required"))
}

func (b *BoardgameUtil) Name() string {
	return "boardgame-util"
}

func (b *BoardgameUtil) SubcommandConfig() []*writ.Command {
	//TODO :iterate through this automatically based on SubcommandObjects[]
	return []*writ.Command{
		&writ.Command{
			Name: b.Help.Name(),
			//TODO: pop this out to b.Help.Description().
			Description: b.Help.Description(),
		},
		&writ.Command{
			Name:        b.Db.Name(),
			Aliases:     b.Db.Aliases(),
			Description: b.Db.Description(),
		},
	}
}
