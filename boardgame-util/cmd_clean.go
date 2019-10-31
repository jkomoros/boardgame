package main

import (
	"errors"

	"github.com/bobziuchkovski/writ"
)

type clean struct {
	baseSubCommand

	CleanAPI    cleanAPI
	CleanStatic cleanStatic
	CleanCache  cleanCache
}

func (c *clean) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New(c.Name() + " cannot be run by itself"))
}

func (c *clean) Name() string {
	return "clean"
}

func (c *clean) Description() string {
	return "Cleans up files created by the build command"
}

func (c *clean) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&c.CleanAPI,
		&c.CleanStatic,
		&c.CleanCache,
	}
}
