package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
)

type CleanStatic struct {
	baseSubCommand
}

func (c *CleanStatic) Run(p writ.Path, positional []string) {

	dir := dirPositionalOrDefault(positional, false)

	err := build.CleanStatic(dir)

	if err != nil {
		errAndQuit(err.Error())
	}

	fmt.Println("Cleaned static folder")

}

func (c *CleanStatic) Name() string {
	return "static"
}

func (c *CleanStatic) Description() string {
	return "Cleans up a static server assets folder created by `build static`"
}

func (c *CleanStatic) Usage() string {
	return "DIR"
}

func (c *CleanStatic) HelpText() string {

	return c.Name() + ` removes the static server folder within
DIR that was created by 'build static'.

If DIR is not provided, defaults to "."

`
}
