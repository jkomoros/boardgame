package main

import (
	"errors"

	"github.com/bobziuchkovski/writ"
)

type Build struct {
	baseSubCommand

	BuildAPI    buildAPI
	BuildStatic BuildStatic
}

func (b *Build) Run(p writ.Path, positional []string) {
	p.Last().ExitHelp(errors.New(b.Name() + " cannot be run by itself"))
}

func (b *Build) Name() string {
	return "build"
}

func (b *Build) Description() string {
	return "Builds servers, server static folders, or golden folders"
}

func (b *Build) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&b.BuildAPI,
		&b.BuildStatic,
	}
}
