package main

import (
	"errors"

	"github.com/bobziuchkovski/writ"
)

type build struct {
	baseSubCommand

	BuildAPI    buildAPI
	BuildStatic buildStatic
}

func (b *build) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New(b.Name() + " cannot be run by itself"))
}

func (b *build) Name() string {
	return "build"
}

func (b *build) Description() string {
	return "Builds servers, server static folders, or golden folders"
}

func (b *build) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&b.BuildAPI,
		&b.BuildStatic,
	}
}
