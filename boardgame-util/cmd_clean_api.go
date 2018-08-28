package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
)

type CleanApi struct {
	baseSubCommand
}

func (c *CleanApi) Run(p writ.Path, positional []string) {

	if len(positional) > 1 {
		c.Base().errAndQuit(c.Name() + " called with more than one positional argument")
	}

	dir := "."

	err := build.CleanApi(dir)

	if err != nil {
		c.Base().errAndQuit(err.Error())
	}

	fmt.Println("Cleaned api folder")

}

func (c *CleanApi) Name() string {
	return "api"
}

func (c *CleanApi) Description() string {
	return "Cleans up an api server folder created by `build api`"
}

func (c *CleanApi) Usage() string {
	return "DIR"
}

func (c *CleanApi) HelpText() string {

	return c.Name() + ` removes the api server folder (binary and code) within
DIR that was created by 'build api'.

If DIR is not provided, defaults to "."

`
}
