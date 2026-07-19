package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/internal/fileutil"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/moveargs"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

type emitMoveArgs struct {
	baseSubCommand
	Check bool
}

type generatedMoveArgsFile struct {
	path     string
	contents []byte
	gameName string
	moves    int
}

func (e *emitMoveArgs) Run(p writ.Path, positional []string) {

	c := e.Base().GetConfig(false)

	mode := c.Dev

	pkgs, err := mode.AllGamePackages()

	if err != nil {
		e.Base().errAndQuit("Not all game packages were valid: " + err.Error())
	}

	if err := emitMoveArgsForPackages(e.Base(), pkgs, e.Check); err != nil {
		e.Base().errAndQuit("Couldn't emit move args: " + err.Error())
	}

	if e.Check {
		fmt.Println("Generated _move_args.ts files are current")
	} else {
		fmt.Println("Successfully generated _move_args.ts files")
	}
}

func (e *emitMoveArgs) WritOptions() []*writ.Option {
	return []*writ.Option{{
		Names:       []string{"check"},
		Description: "Verify generated move-input contracts are current without writing files.",
		Decoder:     writ.NewFlagDecoder(&e.Check),
		Flag:        true,
	}}
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
move().with() action API on BoardgameBaseGameRenderer.

The generated files follow the same convention as _move_names.ts and _types.ts:
they are regenerated each time but should be committed to source control.`
}

// emitMoveArgsForPackages builds a temporary binary to extract move field info
// from the given game packages and writes _move_args.ts files into each game's
// client/ directory.
func emitMoveArgsForPackages(base *boardgameUtil, pkgs []*gamepkg.Pkg, check bool) error {
	generated, err := generateMoveArgsForPackages(base, pkgs, check)
	if err != nil {
		return err
	}
	return installGeneratedMoveArgs(generated, check)
}

func generateMoveArgsForPackages(base *boardgameUtil, pkgs []*gamepkg.Pkg, includeReadOnly bool) ([]generatedMoveArgsFile, error) {

	dir, err := newSystemTempDir("temp_moveargs_")
	if err != nil {
		return nil, fmt.Errorf("couldn't create temp directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: couldn't clean up temp dir %s: %v\n", dir, removeErr)
		}
	}()

	fmt.Fprintln(os.Stderr, "Extracting move field info from game packages")
	results, err := moveargs.Build(dir, pkgs)

	if err != nil {
		return nil, fmt.Errorf("couldn't build move args: %w", err)
	}
	resultImports := make([]string, 0, len(results))
	for _, result := range results {
		resultImports = append(resultImports, result.ImportPath)
	}
	if err := validateClientExtractionResults(pkgs, resultImports, "move-input"); err != nil {
		return nil, err
	}

	// Build a map from import path to pkg for quick lookup
	pkgByImport := make(map[string]*gamepkg.Pkg)
	for _, pkg := range pkgs {
		pkgByImport[pkg.Import()] = pkg
	}

	var generated []generatedMoveArgsFile
	for _, result := range results {
		pkg, ok := pkgByImport[result.ImportPath]
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: no package found for import path %s, skipping\n", result.ImportPath)
			continue
		}

		if pkg.ClientFolder() == "" {
			continue
		}

		if pkg.ReadOnly() && !includeReadOnly {
			continue
		}
		if err := moveargs.ValidateTypeScriptSchema(result.Moves); err != nil {
			return nil, fmt.Errorf("invalid generated TypeScript contract for %s: %w", result.PackageName, err)
		}

		generated = append(generated, generatedMoveArgsFile{
			path:     filepath.Join(pkg.ClientFolder(), "_move_args.ts"),
			contents: []byte(moveargs.GenerateTypeScript(result.Moves)),
			gameName: result.PackageName,
			moves:    len(result.Moves),
		})
	}
	if err := validateGeneratedMoveArgsTypeScript(generated); err != nil {
		return nil, err
	}
	return generated, nil
}

func validateGeneratedMoveArgsTypeScript(generated []generatedMoveArgsFile) error {
	if len(generated) == 0 {
		return nil
	}
	compiler, err := moveArgsTypeScriptCompiler(generated)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "boardgame-move-args-typecheck-")
	if err != nil {
		return fmt.Errorf("couldn't create move-input TypeScript validation directory: %w", err)
	}
	defer os.RemoveAll(dir)

	args := []string{"--noEmit", "--strict", "--skipLibCheck", "--target", "ES2020", "--module", "ES2020"}
	for i, file := range generated {
		path := filepath.Join(dir, fmt.Sprintf("contract-%03d.ts", i))
		if err := os.WriteFile(path, file.contents, 0600); err != nil {
			return fmt.Errorf("couldn't stage %s for TypeScript validation: %w", file.gameName, err)
		}
		args = append(args, path)
	}
	cmd := exec.Command(compiler, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("generated move-input TypeScript failed strict validation: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func moveArgsTypeScriptCompiler(generated []generatedMoveArgsFile) (string, error) {
	if configured := os.Getenv("BOARDGAME_TSC"); configured != "" {
		return configured, nil
	}
	// Prefer the creator project's pinned compiler. Search upward from each
	// generated client contract so invoking boardgame-util from a Go module does
	// not require a global TypeScript installation.
	seen := make(map[string]bool)
	for _, file := range generated {
		for dir := filepath.Dir(file.path); dir != filepath.Dir(dir); dir = filepath.Dir(dir) {
			if seen[dir] {
				continue
			}
			seen[dir] = true
			if compiler := localTypeScriptCompiler(dir); compiler != "" {
				return compiler, nil
			}
		}
	}
	cmd := exec.Command("go", "list", "-f", "{{.Dir}}", "github.com/jkomoros/boardgame")
	output, err := cmd.CombinedOutput()
	if err == nil {
		if compiler := localTypeScriptCompiler(filepath.Join(strings.TrimSpace(string(output)), "server", "static")); compiler != "" {
			return compiler, nil
		}
	}
	if compiler, err := exec.LookPath("tsc"); err == nil {
		return compiler, nil
	}
	return "", fmt.Errorf("couldn't find a TypeScript compiler; install project client dependencies so node_modules/.bin/tsc exists, or set BOARDGAME_TSC to the pinned compiler")
}

func localTypeScriptCompiler(dir string) string {
	name := "tsc"
	if runtime.GOOS == "windows" {
		name = "tsc.cmd"
	}
	candidate := filepath.Join(dir, "node_modules", ".bin", name)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate
	}
	return ""
}

func installGeneratedMoveArgs(generated []generatedMoveArgsFile, check bool) error {
	sort.Slice(generated, func(i, j int) bool { return generated[i].path < generated[j].path })

	if check {
		var stale []string
		for _, file := range generated {
			current, err := generatedFileCurrent(file.path, file.contents)
			if err != nil {
				return err
			}
			if !current {
				stale = append(stale, file.path)
			}
		}
		if len(stale) > 0 {
			return staleGeneratedClientContracts(fmt.Sprintf("generated move-input contracts are stale: %s", strings.Join(stale, ", ")))
		}
		fmt.Fprintf(os.Stderr, "Verified %d generated move-input contracts\n", len(generated))
		return nil
	}

	files := make(map[string]fileutil.FileSpec, len(generated))
	for _, file := range generated {
		if _, exists := files[file.path]; exists {
			return fmt.Errorf("duplicate generated destination %s", file.path)
		}
		files[file.path] = fileutil.FileSpec{Contents: file.contents, Mode: 0o644, ForceMode: true}
	}
	if err := fileutil.WriteFileSetAtomicAbsolute(files, true); err != nil {
		return fmt.Errorf("install generated move-input contracts: %w", err)
	}
	for _, file := range generated {
		fmt.Fprintf(os.Stderr, "  Generated %s/client/_move_args.ts (%d moves)\n", file.gameName, file.moves)
	}

	return nil
}
