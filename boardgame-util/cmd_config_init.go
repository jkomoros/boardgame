package main

import (
	"errors"
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"strings"
)

type ConfigInit struct {
	baseSubCommand
}

func (c *ConfigInit) Run(p writ.Path, positional []string) {

	typ := "default"

	if len(positional) > 1 {
		p.Last().ExitHelp(errors.New("Only one positional argument may be provided"))
	}

	if len(positional) == 1 {
		typ = strings.ToLower(positional[0])
	}

	var cfg *config.Config

	configPath := c.Base().ConfigPath

	switch typ {
	case "default":
		cfg = config.DefaultStarterConfig(configPath)
	case "sample":
		cfg = config.SampleStarterConfig(configPath)
	case "minimal":
		cfg = config.MinimalStarterConfig(configPath)
	default:
		p.Last().ExitHelp(errors.New(typ + " is not a legal type"))
	}

	if err := cfg.Save(); err != nil {
		c.Base().errAndQuit("Couldn't save config: " + err.Error())
	}

	fmt.Println("Saved " + cfg.Path())

}

func (c *ConfigInit) Name() string {
	return "init"
}

func (c *ConfigInit) Usage() string {
	return "[default|sample|minimal]"
}

func (c *ConfigInit) Description() string {
	return "Creates a starter config object"
}

func (c *ConfigInit) HelpText() string {

	return c.Name() + " creates a starter config object. Uses the location of --config, otherwise defaults to reasonable default in current directory.\n\n" +
		"TYPE may be one of:\n" +
		"* default - A reasonable default starter value. Used if no TYPE is provided\n" +
		"* minimal - A minimal starter config, if you want only the most basic.\n" +
		"* sample - A full-fledged sample, equivalent to `boardgame/config.SAMPLE.json`\n"

}
