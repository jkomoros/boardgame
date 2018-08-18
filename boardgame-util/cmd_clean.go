package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type Clean struct {
	baseSubCommand

	CleanApi    CleanApi
	CleanStatic CleanStatic
}

func (c *Clean) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New(c.Name() + " cannot be run by itself"))
}

func (c *Clean) Name() string {
	return "clean"
}

func (c *Clean) Description() string {
	return "Cleans up files created by the build command"
}

func (c *Clean) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&c.CleanApi,
		&c.CleanStatic,
	}
}