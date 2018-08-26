package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"io/ioutil"
	"os"
)

type BoardgameUtil struct {
	baseSubCommand
	Help    Help
	Db      Db
	Codegen Codegen
	Build   Build
	Clean   Clean
	Serve   Serve
	Config  Config

	ConfigPath string

	config *config.Config

	//Dirs to delete on exit
	tempDirs []string
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
the README for more on configuring that configuration file, or run "boardgame-
util help config" to learn more.

See the individual sub-commands for more on what each one does.`

}

func (b *BoardgameUtil) Usage() string {
	return "COMMAND [OPTION]... [ARG]..."
}

func (b *BoardgameUtil) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"config", "c"},
			Decoder:     writ.NewOptionDecoder(&b.ConfigPath),
			Description: "The path to the config file or dir to use. If not provided, searches within current directory for files that could be a config, and then walks upwards until it finds one.",
		},
	}
}

func (b *BoardgameUtil) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&b.Help,
		&b.Db,
		&b.Codegen,
		&b.Build,
		&b.Clean,
		&b.Serve,
		&b.Config,
	}
}

//Do any cleanup tasks as program exits.
func (b *BoardgameUtil) Cleanup() {

	for _, dir := range b.tempDirs {
		os.RemoveAll(dir)
	}

}

//NewTempDir will vend a new temporary dir that will be remove when program exits.
func (b *BoardgameUtil) NewTempDir(prefix string) string {
	dir, err := ioutil.TempDir(".", prefix)

	if err != nil {
		errAndQuit("Couldn't create temporary directory: " + err.Error())
	}

	b.tempDirs = append(b.tempDirs, dir)

	return dir
}

//GetConfig fetches the config, finding it from disk if it hasn't yet. If
//finding the config errors for any reason, program will quit. That is, when
//you call this method we assume that it's required for operation of that
//command.
func (b *BoardgameUtil) GetConfig(createIfNotExist bool) *config.Config {
	if b.config != nil {
		return b.config
	}

	c, err := config.Get(b.ConfigPath, createIfNotExist)

	if err != nil {
		errAndQuit("config is required for this command, but it couldn't be loaded. See README.md for more about structuring config.json.\nError: " + err.Error())
	}

	b.config = c

	return c
}
