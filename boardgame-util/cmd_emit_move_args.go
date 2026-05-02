package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/moveargs"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

type emitMoveArgs struct {
	baseSubCommand
}

func (e *emitMoveArgs) Run(p writ.Path, positional []string) {

	c := e.Base().GetConfig(false)

	mode := c.Dev

	pkgs, err := mode.AllGamePackages()

	if err != nil {
		e.Base().errAndQuit("Not all game packages were valid: " + err.Error())
	}

	if err := emitMoveArgsForPackages(e.Base(), pkgs); err != nil {
		e.Base().errAndQuit("Couldn't emit move args: " + err.Error())
	}

	fmt.Println("Successfully generated _move_args.ts files")
}

func (e *emitMoveArgs) Name() string {
	return "emit-move-args"
}

func (e *emitMoveArgs) Description() string {
	return "Generates TypeScript move argument type definitions for each game's client"
}

func (e *emitMoveArgs) HelpText() string {
	return e.Name() + ` generates a _move_args.ts file in each game's client/ directory
containing typed argument interfaces for all player-proposable moves. These types
provide compile-time safety when proposing moves from client code via the typed
proposeMove() helper on BoardgameBaseGameRenderer.

The generated files follow the same convention as _move_names.ts and _types.ts:
they are regenerated each time but should be committed to source control.`
}

// emitMoveArgsForPackages builds a temporary binary to extract move field info
// from the given game packages and writes _move_args.ts files into each game's
// client/ directory.
func emitMoveArgsForPackages(base *boardgameUtil, pkgs []*gamepkg.Pkg) error {

	dir, err := ioutil.TempDir(".", "temp_moveargs_")
	if err != nil {
		return fmt.Errorf("couldn't create temp directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			fmt.Printf("Warning: couldn't clean up temp dir %s: %v\n", dir, removeErr)
		}
	}()

	fmt.Println("Extracting move field info from game packages")
	results, err := moveargs.Build(dir, pkgs)

	if err != nil {
		return fmt.Errorf("couldn't build move args: %w", err)
	}

	// Build a map from import path to pkg for quick lookup
	pkgByImport := make(map[string]*gamepkg.Pkg)
	for _, pkg := range pkgs {
		pkgByImport[pkg.Import()] = pkg
	}

	for _, result := range results {
		pkg, ok := pkgByImport[result.ImportPath]
		if !ok {
			fmt.Printf("Warning: no package found for import path %s, skipping\n", result.ImportPath)
			continue
		}

		if pkg.ClientFolder() == "" {
			continue
		}

		if pkg.ReadOnly() {
			continue
		}

		ts := moveargs.GenerateTypeScript(result.Moves)

		if err := pkg.WriteFile("client/_move_args.ts", []byte(ts), true); err != nil {
			return fmt.Errorf("couldn't write _move_args.ts for %s: %w", result.PackageName, err)
		}

		fmt.Printf("  Generated %s/client/_move_args.ts (%d moves)\n", result.PackageName, len(result.Moves))
	}

	return nil
}
