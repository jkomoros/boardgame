package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

//SubcommandObject is a literal struct that implements a subcommand
type SubcommandObject interface {
	//The name the command is registered as
	Name() string
	//The description of the command to register
	Description() string
	//The aliases to register for
	Aliases() []string
	//The rest of the usage string, which will be appened to "NAME "
	Usage() string
	//SubcommandObjects should return the list of sub comamnds, or nil if a
	//terminal command.
	SubcommandObjects() []SubcommandObject
	//The command to actually run
	Run(p writ.Path, positional []string)
}

type baseSubCommand struct{}

func (b *baseSubCommand) Aliases() []string {
	return nil
}

func (b *baseSubCommand) Description() string {
	return ""
}

func (b *baseSubCommand) Usage() string {
	return ""
}

func (b *baseSubCommand) SubcommandObjects() []SubcommandObject {
	return nil
}

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
func selectSubcommandObject(s SubcommandObject, p writ.Path) SubcommandObject {

	if s.Name() != p.First().String() {
		return nil
	}

	if len(p) < 2 {
		return s
	}

	nextCommand := p[1]

	for _, obj := range s.SubcommandObjects() {
		//We don't need to check alises, because the main library already did
		//the command/object matching
		if nextCommand.Name == obj.Name() {
			return selectSubcommandObject(obj, p[1:])
		}
	}

	return nil
}
