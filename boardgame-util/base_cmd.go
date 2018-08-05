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

//selectSubcommandObject takes a subcommand object and a path. It verifes the
//first item is us, then identifies the next object to recurse into based on
//Names of SubcommandObjects.
func selectSubcommandObject(s SubcommandObject, p []string) SubcommandObject {

	if s.Name() != p[0] {
		return nil
	}

	if len(p) < 2 {
		return s
	}

	nextCommand := p[1]

	for _, obj := range s.SubcommandObjects() {
		//We don't need to check alises, because the main library already did
		//the command/object matching
		if nextCommand == obj.Name() {
			return selectSubcommandObject(obj, p[1:])
		}
	}

	return nil
}
