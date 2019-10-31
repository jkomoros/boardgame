package main

import (
	"errors"
	"fmt"

	"github.com/bobziuchkovski/writ"
)

type configInit struct {
	baseSubCommand
}

func (c *configInit) Run(p writ.Path, positional []string) {

	if len(positional) > 1 {
		p.Last().ExitHelp(errors.New("Only one positional argument may be provided"))
	}

	typ := ""

	if len(positional) == 1 {
		typ = positional[0]
	}

	cfg, err := c.Base().starterConfigForType(typ)

	if err != nil {
		p.Last().ExitHelp(err)
	}

	if err := cfg.Save(); err != nil {
		c.Base().errAndQuit("Couldn't save config: " + err.Error())
	}

	fmt.Println("Saved " + cfg.Path())

}

func (c *configInit) Name() string {
	return "init"
}

func (c *configInit) Usage() string {
	return "[default|sample|minimal]"
}

func (c *configInit) Description() string {
	return "Creates a starter config object"
}

func (c *configInit) HelpText() string {

	return c.Name() + " creates a starter config object. Uses the location of --config, otherwise defaults to reasonable default in current directory.\n\n" +
		"TYPE may be one of:\n" +
		"* default - A reasonable default starter value. Used if no TYPE is provided\n" +
		"* minimal - A minimal starter config, if you want only the most basic.\n" +
		"* sample - A full-fledged sample, equivalent to `boardgame/config.SAMPLE.json`\n"

}
