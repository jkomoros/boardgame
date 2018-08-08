package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
)

type BoardgameUtil struct {
	baseSubCommand
	Help    Help
	Db      Db
	Codegen Codegen
	config  *config.Config
}

func (b *BoardgameUtil) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New("COMMAND is required"))
}

func (b *BoardgameUtil) Name() string {
	return "boardgame-util"
}

func (b *BoardgameUtil) HelpText() string {

	return b.Name() +
		` is a comprehensive CLI tool to make working with
the boardgame framework easy. It has a number of subcommands to help do
everything from generate PropReader interfaces, to building and running a
server.

All of the substantive functionality provided by this utility is also
available as individual utility libraries to use directly if for some reason
this tool doesn't do exactly what you need.

A number of the commands expect some values to be provided in config.json. See
the README for more on configuring that configuration file.

See the individual sub-commands for more on what each one does.`

}

func (b *BoardgameUtil) Usage() string {
	return "COMMAND [OPTION]... [ARG]..."
}

func (b *BoardgameUtil) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&b.Help,
		&b.Db,
		&b.Codegen,
	}
}

//GetConfig fetches the config, finding it from disk if it hasn't yet. If
//finding the config errors for any reason, program will quit. That is, when
//you call this method we assume that it's required for operation of that
//command.
func (b *BoardgameUtil) GetConfig() *config.Config {
	if b.config != nil {
		return b.config
	}

	c, err := config.Get()

	if err != nil {
		errAndQuit("config is required for this command, but it couldn't be loaded. See README.md for more about structuring config.json.\nError: " + err.Error())
	}

	b.config = c

	return c
}
