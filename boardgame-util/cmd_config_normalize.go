package main

import (
	"github.com/bobziuchkovski/writ"
)

type configNormalize struct {
	baseSubCommand
}

func (c *configNormalize) Run(p writ.Path, positional []string) {

	config := c.Base().GetConfig(false)

	if _, err := config.Dev.AllGamePackages(); err != nil {
		c.Base().errAndQuit("At least one DEV game package invalid: " + err.Error())
	}

	if _, err := config.Prod.AllGamePackages(); err != nil {
		c.Base().errAndQuit("At least one PROD game package invalid: " + err.Error())
	}

	err := config.Save()

	if err != nil {
		c.Base().errAndQuit("Couldn't save: " + err.Error())
	}

}

func (c *configNormalize) Name() string {
	return "normalize"
}

func (c *configNormalize) Description() string {
	return "Loads and saves config.json so it's in canonical shape"
}

func (c *configNormalize) HelpText() string {
	return c.Name() + ` simply parses and then saves the config.json in use.

This guarantees that it is in canonical structure, so in the future small changes to the config via "boardgame-util config set" should have only a few line diffs.` +

		"\n\n" +

		"See help for the parent command for more information about configuration in general."
}
