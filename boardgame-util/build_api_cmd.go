package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
	"strings"
)

type BuildApi struct {
	baseSubCommand

	Storage string
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

	storage := build.StorageTypeFromString(b.Storage)

	if storage == build.StorageInvalid {
		errAndQuit("Invalid storage type provided (" + b.Storage + "). Must be one of {" + strings.Join(build.ValidStorageTypeStrings(), ",") + "}.")
	}

	binaryPath, err := build.Api(dir, mode.GamesList, storage)

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

If DIR is not provided, defaults to "."`
}

func (b *BuildApi) WritOptions() []*writ.Option {

	defaultStorageType := "bolt"

	return []*writ.Option{
		{
			Names: []string{"storage", "s"},
			Decoder: writ.NewDefaulter(
				writ.NewOptionDecoder(&b.Storage),
				defaultStorageType,
			),
			Description: "Which storage subsystem to use. One of {" + strings.Join(build.ValidStorageTypeStrings(), ",") + "}. Defaults to '" + defaultStorageType + "'.",
		},
	}
}
