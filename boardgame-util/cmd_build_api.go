package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/api"
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"strings"
)

type BuildApi struct {
	baseSubCommand

	Storage string

	Prod bool
}

func effectiveStorageType(base *BoardgameUtil, m *config.ConfigMode, storageOverride string) api.StorageType {

	//Use storage type from command line option, then from DefaultStorageType
	//in config, then just fallback on defaultStorageType.
	storageTypeString := storageOverride

	if storageTypeString == "" {
		storageTypeString = m.DefaultStorageType
	}

	//It's OK if storageTypeString is "", that will just mean TypeDefault.
	storage := api.StorageTypeFromString(storageTypeString)

	if storage == api.StorageInvalid {
		base.errAndQuit("Invalid storage type provided (" + storageOverride + "). Must be one of {" + strings.Join(api.ValidStorageTypeStrings(), ",") + "}.")
	}

	return storage
}

func (b *BuildApi) Run(p writ.Path, positional []string) {

	dir := dirPositionalOrDefault(b.Base(), positional, false)

	config := b.Base().GetConfig(false)

	mode := config.Dev
	if b.Prod {
		mode = config.Prod
	}

	storage := effectiveStorageType(b.Base(), config.Dev, b.Storage)

	pkgs, err := mode.AllGamePackages()

	if err != nil {
		b.Base().errAndQuit("Not all game packages were valid: " + err.Error())
	}

	binaryPath, err := api.Build(dir, pkgs, storage)

	if err != nil {
		b.Base().errAndQuit("Couldn't generate binary: " + err.Error())
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

	return b.Name() + ` generates an api server binary based on the config.json in use. It creates the binary in a folder called 'api' within the given DIR.

If DIR is not provided, defaults to "."`
}

func (b *BuildApi) WritOptions() []*writ.Option {
	return []*writ.Option{
		{
			Names:       []string{"storage", "s"},
			Decoder:     writ.NewOptionDecoder(&b.Storage),
			Description: "Which storage subsystem to use. One of {" + strings.Join(api.ValidStorageTypeStrings(), ",") + "}. If not provided, falls back on the DefaultStorageType from config, or as a final fallback just the deafult storage type.",
		},
		{
			Names:       []string{"prod", "p"},
			Description: "If provided, will use prod settings from config.json instead of dev",
			Decoder:     writ.NewFlagDecoder(&b.Prod),
			Flag:        true,
		},
	}
}
