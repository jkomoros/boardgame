package main

import (
	"fmt"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/static"
)

type buildStatic struct {
	baseSubCommand

	CopyFiles bool
	Prod      bool
}

func (b *buildStatic) Run(p writ.Path, positional []string) {

	dir := dirPositionalOrDefault(b.Base(), positional, false)

	config := b.Base().GetConfig(false)

	mode := config.Dev

	if b.Prod {
		mode = config.Prod
	}

	pkgs, err := mode.AllGamePackages()

	if err != nil {
		b.Base().errAndQuit("Not all games packages were legal: " + err.Error())
	}

	staticPath, err := static.Build(dir, pkgs, config.Client(b.Prod), b.Prod, b.CopyFiles, mode.OfflineDevMode)

	if err != nil {
		b.Base().errAndQuit("Couldn't create static directory: " + err.Error())
	}

	fmt.Println("Created static dir at " + staticPath)
	fmt.Println("You can remove it with `boardgame-util clean static " + dir + "`")

}

func (b *buildStatic) Name() string {
	return "static"
}

func (b *buildStatic) Description() string {
	return "Generates a folder for all static assets for the games in config"
}

func (b *buildStatic) Usage() string {
	return "DIR"
}

func (b *buildStatic) HelpText() string {

	return b.Name() + ` generates a folder of static server assets based on the config.json in use. It creates the binary in a folder called 'static' within the given DIR.

If DIR is not provided, defaults to "."`
}

func (b *buildStatic) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"prod", "p"},
			Description: "If provided, will created bundled build directory for static resources.",
			Decoder:     writ.NewFlagDecoder(&b.Prod),
			Flag:        true,
		},
		{
			Names:       []string{"copy-files"},
			Description: "If provided, will copy files instead of symlinking them.",
			Decoder:     writ.NewFlagDecoder(&b.CopyFiles),
			Flag:        true,
		},
	}
}
