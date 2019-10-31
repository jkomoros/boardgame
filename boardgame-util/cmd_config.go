package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bobziuchkovski/writ"
)

//configCmd is named that instead of config to avoid colliding with the config
//package name.
type configCmd struct {
	baseSubCommand

	ConfigNormalize configNormalize
	ConfigSet       configSet
	ConfigAdd       configAdd
	ConfigRemove    configRemove
	ConfigInit      configInit
}

func (c *configCmd) Run(p writ.Path, positional []string) {

	if len(positional) > 0 {
		p.Last().ExitHelp(errors.New(c.Name() + " doesn't take any positional parameters"))
	}

	config := c.Base().GetConfig(false)

	fmt.Println("Path: " + config.Path())
	if secretPath := config.SecretPath(); secretPath != "" {
		fmt.Println("Secret path: " + secretPath)
	} else {
		fmt.Println("NO secret path in use")
	}

	if _, err := config.Dev.AllGamePackages(); err != nil {
		fmt.Println("Not all DEV game packages are valid: " + err.Error())
	}

	if _, err := config.Prod.AllGamePackages(); err != nil {
		fmt.Println("Not all PROD game packages are valid: " + err.Error())
	}

	devBlob, err := json.MarshalIndent(config.Dev, "", "\t")

	if err != nil {
		c.Base().errAndQuit("Couldn't marshal dev: " + err.Error())
	}

	prodBlob, err := json.MarshalIndent(config.Prod, "", "\t")

	if err != nil {
		c.Base().errAndQuit("Couldn't marshal prod: " + err.Error())
	}

	fmt.Println("Derived dev configuration:")
	fmt.Println(string(devBlob))

	fmt.Println("Derived prod configuration: ")
	fmt.Println(string(prodBlob))

}

func (c *configCmd) Name() string {
	return "config"
}

func (c *configCmd) Description() string {
	return "Allows viewing and modifying configuration files"
}

func (c *configCmd) HelpText() string {
	return c.Name() + ` run without arguments prints the derived config in use and the path that is being used. It's a good way to debug config issues.

GENERAL INFORMATION ON CONFIG

configuration is provided for boardgame-util and other libraries within boardgame via a JSON configuration file, typically called "config.json". boardgame libraries search for that file in the current directory, and if they don't find it they walk up the directory hierarchy until they find one. You can also pass an override config parameter to specify a specific file or folder to search in.

It is reasonable to modify the file directly by hand, although the "config" command and its sub-commands can modify the files for you directly to ensure the syntax is correct.

The config file contains information to derive both a Dev and Prod configuration. The file has three sections: base, dev, and prod, each of which is optional.

base is the base values for all items. dev and prod both extend base, overriding any set values. Typically you set values in base and only override them in dev or prod if they differ. For that reason, if you don't pass --dev or --prod options, by default this command will modify the base.

The fields for each section are described in README.md in "boardgame-util/lib/config".

Typically the configuration file is checked into source control. However, some settings (especially the connection strings for databases) might contain sensitive information that should not be checked into version control. For that reason it's also possible to define a "config.SECRET.json" in the same directory as the non-secret config. If your .gitignore file has the pattern "*.SECRET.*" within it, then that file cannot be accidentally committed to version control (of course, it's then your responsibility to distribute it). Any values defined in the secret config override any values defined in the non-secret config. If you want to modify the secret config, pass --secret option. Some fields are assumed to be sensitive, so the tool will ask for explicit confirmation if you try to set them for prod without also passing --secret.

If you use any of the "boardgame-util config set" commands on a configuration file that does not exist, one will be created for you at a default location in the current directory, or in the given directory if you passed a config parameter.`

}

func (c *configCmd) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&c.ConfigSet,
		&c.ConfigAdd,
		&c.ConfigRemove,
		&c.ConfigNormalize,
		&c.ConfigInit,
	}
}
