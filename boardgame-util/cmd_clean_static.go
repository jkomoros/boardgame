package main

import (
	"fmt"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/static"
)

type cleanStatic struct {
	baseSubCommand
}

func (c *cleanStatic) Run(p writ.Path, positional []string) {

	dir := dirPositionalOrDefault(c.Base(), positional, false)

	err := static.Clean(dir)

	if err != nil {
		c.Base().errAndQuit(err.Error())
	}

	fmt.Println("Cleaned static folder")

}

func (c *cleanStatic) Name() string {
	return "static"
}

func (c *cleanStatic) Description() string {
	return "Cleans up a static server assets folder created by `build static`"
}

func (c *cleanStatic) Usage() string {
	return "DIR"
}

func (c *cleanStatic) HelpText() string {

	return c.Name() + ` removes the static server folder within
DIR that was created by 'build static'.

If DIR is not provided, defaults to "."

`
}
