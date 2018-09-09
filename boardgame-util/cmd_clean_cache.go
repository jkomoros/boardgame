package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build"
)

type CleanCache struct {
	baseSubCommand
}

func (c *CleanCache) Run(p writ.Path, positional []string) {

	if len(positional) > 0 {
		c.Base().errAndQuit("This command accepts no parameters")
	}

	err := build.CleanCache()

	if err != nil {
		c.Base().errAndQuit(err.Error())
	}

	fmt.Println("Cleaned cache")

}

func (c *CleanCache) Name() string {
	return "cache"
}

func (c *CleanCache) Description() string {
	return "Cleans up the central caches created by `build static`"
}

func (c *CleanCache) HelpText() string {

	return c.Name() + ` removes the caches created implicitly by "build static", specifically the central node_modules cache.`

}
