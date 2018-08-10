package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
)

type BuildApi struct {
	baseSubCommand
}

func (b *BuildApi) Run(p writ.Path, positional []string) {

	base := b.Base().(*BoardgameUtil)

	if len(positional) > 1 {
		errAndQuit(b.Name() + " called with more than one positional argument")
	}

	dir := "."

	if len(positional) > 0 {
		dir = positional[0]
	}

	config := base.GetConfig()

	mode := config.Dev

	//TODO: allow switching the type of storage via a command line config.
	binaryPath, err := build.Api(dir, mode.GamesList, build.StorageMysql)

	if err != nil {
		errAndQuit("Couldn't generate binary: " + err.Error())
	}

	fmt.Println("Created api binary at " + binaryPath)
	fmt.Println("You can remove it with `boardgame-util clean api " + dir + "`")

}

func (b *BuildApi) Name() string {
	return "api"
}

func (b *BuildApi) Description() string {
	return "Generates an api server binary for config"
}

func (b *BuildApi) Usage() string {
	return "DIR"
}

func (b *BuildApi) HelpText() string {

	return b.Name() + ` generates an
api server binary based on the config.json in use. It creates the binary in a
folder called 'api' within the given DIR.

If DIR is not provided, defaults to "."

`
}
