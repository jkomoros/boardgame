package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
)

type boardgameUtil struct {
	baseSubCommand
	Help    Help
	Db      db
	Codegen codegen
	Build   build
	Clean   clean
	Serve   Serve
	Config  configCmd
	Stub    Stub
	Golden  goldenCmd

	ConfigPath            string
	OverrideStarterConfig string

	config *config.Config

	//Dirs to delete on exit
	tempDirs []string
}

func (b *boardgameUtil) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New("COMMAND is required"))
}

func (b *boardgameUtil) Name() string {
	return "boardgame-util"
}

func (b *boardgameUtil) HelpText() string {

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

func (b *boardgameUtil) Usage() string {
	return "COMMAND [OPTION]... [ARG]..."
}

func (b *boardgameUtil) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"config", "c"},
			Decoder:     writ.NewOptionDecoder(&b.ConfigPath),
			Description: "The path to the config file or dir to use. If not provided, searches within current directory for files that could be a config, and then walks upwards until it finds one.",
		},
		{
			Names:       []string{"override-starter-config"},
			Decoder:     writ.NewOptionDecoder(&b.OverrideStarterConfig),
			Description: "If provided, the normal config will be ignored and a starter config will be used instead. Useful for running in contexts where you don't have a config.json set up yet. Valid values are the same as for `config init`",
			Placeholder: "TYPE",
		},
	}
}

func (b *boardgameUtil) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&b.Help,
		&b.Serve,
		&b.Config,
		&b.Codegen,
		&b.Stub,
		&b.Db,
		&b.Build,
		&b.Clean,
		&b.Golden,
	}
}

//Do any cleanup tasks as program exits.
func (b *boardgameUtil) Cleanup() {

	for _, dir := range b.tempDirs {
		os.RemoveAll(dir)
	}

}

func (b *boardgameUtil) errAndQuit(message string) {
	fmt.Println(message)
	b.Cleanup()
	os.Exit(1)
}

func (b *boardgameUtil) msgAndQuit(message string) {
	fmt.Println(message)
	b.Cleanup()
	os.Exit(0)
}

//NewTempDir will vend a new temporary dir that will be remove when program exits.
func (b *boardgameUtil) NewTempDir(prefix string) string {
	dir, err := ioutil.TempDir(".", prefix)

	if err != nil {
		b.errAndQuit("Couldn't create temporary directory: " + err.Error())
	}

	b.tempDirs = append(b.tempDirs, dir)

	return dir
}

func (b *boardgameUtil) starterConfigForType(typ string) (*config.Config, error) {

	if typ == "" {
		typ = "default"
	}

	typ = strings.ToLower(typ)

	configPath := b.ConfigPath

	switch typ {
	case "default":
		return config.DefaultStarterConfig(configPath), nil
	case "sample":
		return config.SampleStarterConfig(configPath), nil
	case "minimal":
		return config.MinimalStarterConfig(configPath), nil
	default:
		return nil, errors.New(typ + " is not a legal type")
	}
}

//GetConfig fetches the config, finding it from disk if it hasn't yet. If
//finding the config errors for any reason, program will quit. That is, when
//you call this method we assume that it's required for operation of that
//command.
func (b *boardgameUtil) GetConfig(createIfNotExist bool) *config.Config {
	if b.config != nil {
		return b.config
	}

	var c *config.Config
	var err error

	if b.OverrideStarterConfig != "" {
		fmt.Println("Ignoring normal config, using starter config of type: " + b.OverrideStarterConfig)
		c, err = b.starterConfigForType(b.OverrideStarterConfig)
		if err != nil {
			b.errAndQuit(err.Error())
			return nil
		}
	} else {
		c, err = config.Get(b.ConfigPath, createIfNotExist)
	}

	if err != nil {
		b.errAndQuit("config is required for this command, but it couldn't be loaded. You can create one with `boardgame-util config init`.\nError: " + err.Error())
	}

	b.config = c

	return c
}
