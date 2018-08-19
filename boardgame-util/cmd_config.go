package main

import (
	"encoding/json"
	"fmt"
	"github.com/bobziuchkovski/writ"
)

type Config struct {
	baseSubCommand

	ConfigNormalize ConfigNormalize
	ConfigSet       ConfigSet
	ConfigAdd       ConfigAdd
}

func (c *Config) Run(p writ.Path, positional []string) {

	base := c.Base().(*BoardgameUtil)

	config := base.GetConfig()

	fmt.Println("Path: " + config.Path())
	if secretPath := config.SecretPath(); secretPath != "" {
		fmt.Println("Secret path: " + secretPath)
	} else {
		fmt.Println("NO secret path in use")
	}

	devBlob, err := json.MarshalIndent(config.Dev, "", "\t")

	if err != nil {
		errAndQuit("Couldn't marshal dev: " + err.Error())
	}

	prodBlob, err := json.MarshalIndent(config.Prod, "", "\t")

	if err != nil {
		errAndQuit("Couldn't marshal prod: " + err.Error())
	}

	fmt.Println("Derived dev configuration:")
	fmt.Println(string(devBlob))

	fmt.Println("Derived prod configuration: ")
	fmt.Println(string(prodBlob))

}

func (c *Config) Name() string {
	return "config"
}

func (c *Config) Description() string {
	return "Allows viewing and modifying configuration files"
}

func (c *Config) HelpText() string {
	return c.Name() + ` run without arguments prints the derived config in use and the path that is being used.

It's a good way to debug config issues.`
}

func (c *Config) SubcommandObjects() []SubcommandObject {
	return []SubcommandObject{
		&c.ConfigSet,
		&c.ConfigAdd,
		&c.ConfigNormalize,
	}
}
