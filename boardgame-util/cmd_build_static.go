package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
)

type BuildStatic struct {
	baseSubCommand
}

func (b *BuildStatic) Run(p writ.Path, positional []string) {

	base := b.Base().(*BoardgameUtil)

	dir := dirPositionalOrDefault(positional, false)

	config := base.GetConfig()

	mode := config.Dev

	staticPath, err := build.Static(dir, mode.GamesList)

	if err != nil {
		errAndQuit("Couldn't create static directory: " + err.Error())
	}

	fmt.Println("Created static dir at " + staticPath)
	fmt.Println("You can remove it with `boardgame-util clean static " + dir + "`")

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
