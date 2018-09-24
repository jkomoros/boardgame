package main

import (
	"fmt"
	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
	"github.com/jkomoros/boardgame/boardgame-util/lib/golden"
)

type Golden struct {
	Serve
}

func (g *Golden) Run(p writ.Path, positional []string) {
	g.Prod = false
	g.Storage = "filesystem"

	pkg, err := gamepkg.NewFromPath(".", "")

	if err != nil {
		g.Base().errAndQuit("Current directory is not a valid package. You must run this command sitting in the root of a valid package. " + err.Error())
	}

	if err := golden.MakeGoldenTest(pkg); err != nil {
		g.Base().errAndQuit("Couldn't create golden directory: " + err.Error())
	}

	g.doServe(p, positional, []*gamepkg.Pkg{pkg}, `"`+golden.GameRecordsFolder+`"`)

	fmt.Println("Cleaning golden folder...")
	if err := golden.CleanGoldenTest(pkg); err != nil {
		g.Base().errAndQuit("Couldn't clean golden: " + err.Error())
	}
}

func (g *Golden) Name() string {
	return "golden"
}

func (g *Golden) Description() string {
	return "Helps create golden test files for the current package"
}

func (g *Golden) HelpText() string {
	return g.Name() + ` helps create golden example games to test the current game package.

You run it sittig in the root of a game package, and it will create a stub server (with similiar behavior to what you'd get with 'boardgame-util serve'), but with only one game, and the games will all be persisted to a testdata folder, with a golden_test.go created.

This is useful for saving runs of games that are known good so that you can ensure you don't mess with the game logic later.
`
}

func (s *Golden) WritOptions() []*writ.Option {
	//Skip the first two, which are not valid for us.
	return s.Serve.WritOptions()[2:]
}
