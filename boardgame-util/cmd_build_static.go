package main

import (
	"errors"
	"github.com/bobziuchkovski/writ"
)

type BuildStatic struct {
	baseSubCommand
}

func (b *BuildStatic) Run(p writ.Path, positional []string) {

	//Positional dir should use the same thing we use in cmd_codegen. (And
	//buildApi should too)

	p.Last().ExitHelp(errors.New("Not yet implemented"))

}

func (b *BuildStatic) Name() string {
	return "static"
}

func (b *BuildStatic) Description() string {
	return "Generates a folder for all static assets for the games in config"
}

func (b *BuildStatic) Usage() string {
	return "DIR"
}

func (b *BuildStatic) HelpText() string {

	return b.Name() + ` generates a folder of static server assets based on the config.json in use. It creates the binary in a folder called 'static' within the given DIR.

If DIR is not provided, defaults to "."`
}
