package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/movenames"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

type emitMoveNames struct {
	baseSubCommand
	Check bool
}

type generatedMoveNamesFile struct {
	path, gameName string
	contents       []byte
	moves          int
}

func (e *emitMoveNames) Run(p writ.Path, positional []string) {

	c := e.Base().GetConfig(false)

	mode := c.Dev

	pkgs, err := mode.AllGamePackages()

	if err != nil {
		e.Base().errAndQuit("Not all game packages were valid: " + err.Error())
	}

	if err := emitMoveNamesForPackages(e.Base(), pkgs, e.Check); err != nil {
		e.Base().errAndQuit("Couldn't emit move names: " + err.Error())
	}

	if e.Check {
		fmt.Println("Generated _move_names.ts files are current")
	} else {
		fmt.Println("Successfully generated _move_names.ts files")
	}
}

func (e *emitMoveNames) WritOptions() []*writ.Option {
	return []*writ.Option{{
		Names: []string{"check"}, Description: "Verify generated move-name contracts are current without writing files.",
		Decoder: writ.NewFlagDecoder(&e.Check), Flag: true,
	}}
}

func (e *emitMoveNames) Name() string {
	return "emit-move-names"
}

func (e *emitMoveNames) Description() string {
	return "Generates TypeScript move name constants for each game's client"
}

func (e *emitMoveNames) HelpText() string {
	return e.Name() + ` generates a _move_names.ts file in each game's client/ directory
containing typed constants for all player-proposable move names. These constants
provide type safety and IDE autocomplete when referencing move names in client code.

The generated files follow the same convention as auto_reader.go and auto_enum.go:
they are regenerated each time but should be committed to source control.`
}

// emitMoveNamesForPackages builds a temporary binary to extract move names from
// the given game packages and writes _move_names.ts files into each game's
// client/ directory. It is used by both the emit-move-names command and the
// serve command.
func emitMoveNamesForPackages(base *boardgameUtil, pkgs []*gamepkg.Pkg, check ...bool) error {

	dir, err := ioutil.TempDir(".", "temp_movenames_")
	if err != nil {
		return fmt.Errorf("couldn't create temp directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			fmt.Printf("Warning: couldn't clean up temp dir %s: %v\n", dir, removeErr)
		}
	}()

	fmt.Println("Extracting move names from game packages")
	results, err := movenames.Build(dir, pkgs)

	if err != nil {
		return fmt.Errorf("couldn't build move names: %w", err)
	}

	// Build a map from import path to pkg for quick lookup
	pkgByImport := make(map[string]*gamepkg.Pkg)
	for _, pkg := range pkgs {
		pkgByImport[pkg.Import()] = pkg
	}

	var generated []generatedMoveNamesFile
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

		generated = append(generated, generatedMoveNamesFile{
			path: filepath.Join(pkg.ClientFolder(), "_move_names.ts"), contents: []byte(movenames.GenerateTypeScript(result.MoveNames)),
			gameName: result.PackageName, moves: len(result.MoveNames),
		})
	}
	return installGeneratedMoveNames(generated, len(check) > 0 && check[0])
}

func installGeneratedMoveNames(generated []generatedMoveNamesFile, check bool) error {
	sort.Slice(generated, func(i, j int) bool { return generated[i].path < generated[j].path })
	if check {
		var stale []string
		for _, file := range generated {
			current, err := os.ReadFile(file.path)
			if err != nil || !bytes.Equal(current, file.contents) {
				stale = append(stale, file.path)
			}
		}
		if len(stale) > 0 {
			return fmt.Errorf("generated move-name contracts are stale: %s", strings.Join(stale, ", "))
		}
		return nil
	}
	for _, file := range generated {
		if err := os.WriteFile(file.path, file.contents, 0644); err != nil {
			return fmt.Errorf("couldn't write _move_names.ts for %s: %w", file.gameName, err)
		}
		fmt.Printf("  Generated %s/client/_move_names.ts (%d moves)\n", file.gameName, file.moves)
	}
	return nil
}
