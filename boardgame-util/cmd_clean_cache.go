package main

import (
	"fmt"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/static"
)

type cleanCache struct {
	baseSubCommand
}

func (c *cleanCache) Run(p writ.Path, positional []string) {

	if len(positional) > 0 {
		c.Base().errAndQuit("This command accepts no parameters")
	}

	err := static.CleanCache()

	if err != nil {
		c.Base().errAndQuit(err.Error())
	}

	fmt.Println("Cleaned cache")

}

func (c *cleanCache) Name() string {
	return "cache"
}

func (c *cleanCache) Description() string {
	return "Cleans up the central caches created by `build static`"
}

func (c *cleanCache) HelpText() string {

	return c.Name() + ` removes the caches created implicitly by "build static", specifically the central node_modules cache.`

}
