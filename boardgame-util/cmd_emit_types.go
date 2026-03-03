package main

import (
	"fmt"
	"os"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/gametypes"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

type emitTypes struct {
	baseSubCommand
}

func (e *emitTypes) Run(p writ.Path, positional []string) {

	c := e.Base().GetConfig(false)

	mode := c.Dev

	pkgs, err := mode.AllGamePackages()

	if err != nil {
		e.Base().errAndQuit("Not all game packages were valid: " + err.Error())
	}

	if err := emitTypesForPackages(e.Base(), pkgs); err != nil {
		e.Base().errAndQuit("Couldn't emit types: " + err.Error())
	}

	fmt.Println("Successfully generated _types.ts files")
}

func (e *emitTypes) Name() string {
	return "emit-types"
}

func (e *emitTypes) Description() string {
	return "Generates TypeScript type definitions for each game's client"
}

func (e *emitTypes) HelpText() string {
	return e.Name() + ` generates a _types.ts file in each game's client/ directory
containing typed interfaces for GameState, PlayerState, component values, and enums.
These types provide type safety and IDE autocomplete when accessing game state in
client rendering code.

The generated files follow the same convention as _move_names.ts:
they are regenerated each time but should be committed to source control.`
}

// emitTypesForPackages builds a temporary binary to extract type info from
// the given game packages and writes _types.ts files into each game's
// client/ directory. It is used by both the emit-types command and the
// serve command.
func emitTypesForPackages(base *boardgameUtil, pkgs []*gamepkg.Pkg) error {

	dir, err := os.MkdirTemp(".", "temp_gametypes_")
	if err != nil {
		return fmt.Errorf("couldn't create temp directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			fmt.Printf("Warning: couldn't clean up temp dir %s: %v\n", dir, removeErr)
		}
	}()

	fmt.Println("Extracting type information from game packages")
	results, err := gametypes.Build(dir, pkgs)

	if err != nil {
		return fmt.Errorf("couldn't build game types: %w", err)
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

		ts := gametypes.GenerateTypeScript(result)

		if err := pkg.WriteFile("client/_types.ts", []byte(ts), true); err != nil {
			return fmt.Errorf("couldn't write _types.ts for %s: %w", result.PackageName, err)
		}

		fmt.Printf("  Generated %s/client/_types.ts (%d game fields, %d player fields)\n",
			result.PackageName, len(result.GameFields), len(result.PlayerFields))
	}

	return nil
}
