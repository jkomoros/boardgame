package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bobziuchkovski/writ"
	"github.com/jkomoros/boardgame/boardgame-util/lib/build/gametypes"
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

type emitTypes struct {
	baseSubCommand
	Check bool
}

func (e *emitTypes) Run(p writ.Path, positional []string) {

	c := e.Base().GetConfig(false)

	mode := c.Dev

	pkgs, err := mode.AllGamePackages()

	if err != nil {
		e.Base().errAndQuit("Not all game packages were valid: " + err.Error())
	}
	// Extract and schema-validate state types before refreshing dependencies. A
	// bad deck/enum name must not mutate an otherwise-valid generated client.
	generated, err := generateGameTypesForPackages(e.Base(), pkgs)
	if err != nil {
		e.Base().errAndQuit("Couldn't emit types: " + err.Error())
	}
	if err := emitMoveNamesForPackages(e.Base(), pkgs, e.Check); err != nil {
		e.Base().errAndQuit("Couldn't emit move names required by client contracts: " + err.Error())
	}
	if err := emitMoveArgsForPackages(e.Base(), pkgs, e.Check); err != nil {
		e.Base().errAndQuit("Couldn't emit move inputs required by client contracts: " + err.Error())
	}
	if err := emitBoardSpacesForPackages(pkgs, e.Check); err != nil {
		e.Base().errAndQuit("Couldn't emit authored board spaces: " + err.Error())
	}
	if err := validateGeneratedGameTypesTypeScript(generated); err != nil {
		e.Base().errAndQuit("Couldn't validate generated client contracts: " + err.Error())
	}
	if e.Check {
		if err := checkGeneratedGameTypes(generated); err != nil {
			e.Base().errAndQuit("Generated client contracts are stale: " + err.Error())
		}
	} else {
		if err := installGeneratedGameTypes(generated); err != nil {
			e.Base().errAndQuit("Couldn't install generated client contracts: " + err.Error())
		}
	}

	if e.Check {
		fmt.Println("Generated client contract files are current")
	} else {
		fmt.Println("Successfully generated client contract files")
	}
}

func (e *emitTypes) WritOptions() []*writ.Option {
	return []*writ.Option{{Names: []string{"check"}, Description: "Verify every generated client contract is current without writing files.", Decoder: writ.NewFlagDecoder(&e.Check), Flag: true}}
}

func (e *emitTypes) Name() string {
	return "emit-types"
}

func (e *emitTypes) Description() string {
	return "Generates complete TypeScript client contracts and bound renderer bases"
}

func (e *emitTypes) HelpText() string {
	return e.Name() + ` generates a _types.ts file in each game's client/ directory
containing typed interfaces for GameState, PlayerState, component values,
DynamicComponentValues, configured constants, and enums. It also refreshes
_move_names.ts and _move_args.ts, then generates _game_renderer.ts, the
zero-generic base class game renderers extend. The complete generated surface
is strictly validated before state/renderer outputs are installed. These
contracts provide type safety and IDE autocomplete across state and moves.

Enum fields are resolved at runtime, so even enums from imported packages
(e.g. playingcards Suit/Rank) are emitted as typed string literal unions.

The generated files follow the same convention as _move_names.ts:
they are regenerated each time but should be committed to source control.`
}

// emitTypesForPackages builds a temporary binary to extract type info from
// the given game packages and writes _types.ts files into each game's
// client/ directory. It is used by both the emit-types command and the
// serve command.
func emitTypesForPackages(base *boardgameUtil, pkgs []*gamepkg.Pkg) error {
	generated, err := generateGameTypesForPackages(base, pkgs)
	if err != nil {
		return err
	}
	if err := validateGeneratedGameTypesTypeScript(generated); err != nil {
		return err
	}
	return installGeneratedGameTypes(generated)
}

func generateGameTypesForPackages(base *boardgameUtil, pkgs []*gamepkg.Pkg, includeReadOnly ...bool) ([]generatedGameTypeFile, error) {

	dir, err := os.MkdirTemp(".", "temp_gametypes_")
	if err != nil {
		return nil, fmt.Errorf("couldn't create temp directory: %w", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(dir); removeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: couldn't clean up temp dir %s: %v\n", dir, removeErr)
		}
	}()

	fmt.Fprintln(os.Stderr, "Extracting type information from game packages")
	results, err := gametypes.Build(dir, pkgs)

	if err != nil {
		return nil, fmt.Errorf("couldn't build game types: %w", err)
	}
	resultImports := make([]string, 0, len(results))
	for _, result := range results {
		resultImports = append(resultImports, result.ImportPath)
	}
	if err := validateClientExtractionResults(pkgs, resultImports, "game-type"); err != nil {
		return nil, err
	}

	// Build a map from import path to pkg for quick lookup
	pkgByImport := make(map[string]*gamepkg.Pkg)
	for _, pkg := range pkgs {
		pkgByImport[pkg.Import()] = pkg
	}

	var generated []generatedGameTypeFile
	for _, result := range results {
		pkg, ok := pkgByImport[result.ImportPath]
		if !ok {
			fmt.Fprintf(os.Stderr, "Warning: no package found for import path %s, skipping\n", result.ImportPath)
			continue
		}

		if pkg.ClientFolder() == "" {
			continue
		}

		if pkg.ReadOnly() && !(len(includeReadOnly) > 0 && includeReadOnly[0]) {
			continue
		}

		if err := gametypes.ValidateTypeResult(result); err != nil {
			return nil, fmt.Errorf("invalid generated TypeScript contract for %s: %w", result.PackageName, err)
		}
		generated = append(generated,
			generatedGameTypeFile{path: filepath.Join(pkg.ClientFolder(), "_types.ts"), contents: []byte(gametypes.GenerateTypeScript(result)), gameName: result.PackageName, gameFields: len(result.GameFields), playerFields: len(result.PlayerFields)},
			generatedGameTypeFile{path: filepath.Join(pkg.ClientFolder(), "_game_renderer.ts"), contents: []byte(gametypes.GenerateRendererTypeScript(result.PackageName)), gameName: result.PackageName},
		)
	}
	return generated, nil
}

type generatedGameTypeFile struct {
	path, tempPath, backupPath, gameName string
	contents                             []byte
	gameFields, playerFields             int
	hadOriginal, installed               bool
}

func checkGeneratedGameTypes(generated []generatedGameTypeFile) error {
	var stale []string
	for _, file := range generated {
		current, err := os.ReadFile(file.path)
		if err != nil || !bytes.Equal(current, file.contents) {
			stale = append(stale, file.path)
		}
	}
	sort.Strings(stale)
	if len(stale) > 0 {
		return staleGeneratedClientContracts(strings.Join(stale, ", "))
	}
	return nil
}

func validateGeneratedGameTypesTypeScript(generated []generatedGameTypeFile) error {
	if len(generated) == 0 {
		return nil
	}
	moveArgFiles := make([]generatedMoveArgsFile, 0, len(generated))
	for _, file := range generated {
		moveArgFiles = append(moveArgFiles, generatedMoveArgsFile{path: file.path})
	}
	compiler, err := moveArgsTypeScriptCompiler(moveArgFiles)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "boardgame-game-types-typecheck-")
	if err != nil {
		return fmt.Errorf("couldn't create game-type TypeScript validation directory: %w", err)
	}
	defer os.RemoveAll(dir)

	// Generated game files live at game-src/<game>/client in the assembled
	// static tree, so their ../../src imports resolve to game-src/src.
	frameworkDir := filepath.Join(dir, "games", "src")
	if err := os.MkdirAll(filepath.Join(frameworkDir, "types"), 0700); err != nil {
		return fmt.Errorf("couldn't stage TypeScript framework declarations: %w", err)
	}
	frameworkTypes := `export type Board = unknown;
export type CatalogComponent<S = Readonly<Record<string, unknown>>> = { readonly Index: number; readonly Values: S };
export type ExpandedBoard<S = Readonly<Record<string, unknown>>, D = Readonly<Record<string, unknown>>> = unknown;
export type ExpandedStack<S = Readonly<Record<string, unknown>>, D = Readonly<Record<string, unknown>>> = unknown;
export type ExpandedTimer = unknown;
export type FullGameState<GS, PS, GC, PC, DC> = { readonly Game: GS; readonly Players: readonly PS[]; readonly Components?: DC; readonly Computed?: { readonly Global?: GC; readonly Players?: readonly PC[] } };
export type RawStack = unknown;
`
	frameworkClient := `export abstract class BoardgameBaseGameRenderer<S, C extends object, M extends string, A extends Readonly<Record<M, object>>, K extends object = object> extends HTMLElement {
  protected readonly moveInputSchema!: unknown;
  protected readonly moveInputSchemaFingerprint!: string;
  readonly state!: S;
  readonly chest!: { readonly Decks?: C; readonly Constants?: K };
}
export abstract class BoardgameBasePlayerInfoRenderer<S, PS> extends HTMLElement {
  state!: S | null;
  playerIndex!: number;
  playerState!: PS | null;
}
export abstract class BoardgameTableViewBase<S, C extends object, M extends string, A extends Readonly<Record<M, object>>, K extends object = object> extends BoardgameBaseGameRenderer<S, C, M, A, K> {}
export abstract class BoardgameHandViewBase<S, C extends object, M extends string, A extends Readonly<Record<M, object>>, K extends object = object> extends BoardgameBaseGameRenderer<S, C, M, A, K> {}
`
	if err := os.WriteFile(filepath.Join(frameworkDir, "types", "boardgame-types.ts"), []byte(frameworkTypes), 0600); err != nil {
		return fmt.Errorf("couldn't stage TypeScript type declarations: %w", err)
	}
	if err := os.WriteFile(filepath.Join(frameworkDir, "client.ts"), []byte(frameworkClient), 0600); err != nil {
		return fmt.Errorf("couldn't stage TypeScript renderer declarations: %w", err)
	}

	args := []string{"--noEmit", "--strict", "--skipLibCheck", "--target", "ES2020", "--module", "ES2020", "--moduleResolution", "bundler"}
	gameDirs := make(map[string]string)
	for _, file := range generated {
		gameDir, ok := gameDirs[file.gameName]
		if !ok {
			gameDir = filepath.Join(dir, "games", fmt.Sprintf("game-%03d", len(gameDirs)))
			gameDirs[file.gameName] = gameDir
			if err := os.MkdirAll(filepath.Join(gameDir, "client"), 0700); err != nil {
				return fmt.Errorf("couldn't stage generated contract for %s: %w", file.gameName, err)
			}
			for _, dependency := range []string{"_move_args.ts", "_move_names.ts"} {
				contents, err := os.ReadFile(filepath.Join(filepath.Dir(file.path), dependency))
				if err != nil {
					command := strings.ReplaceAll(strings.TrimSuffix(strings.TrimPrefix(dependency, "_"), ".ts"), "_", "-")
					return fmt.Errorf("couldn't validate %s for %s; generate it first with boardgame-util emit-%s: %w", dependency, file.gameName, command, err)
				}
				if err := os.WriteFile(filepath.Join(gameDir, "client", dependency), contents, 0600); err != nil {
					return fmt.Errorf("couldn't stage %s dependency for %s: %w", dependency, file.gameName, err)
				}
			}
		}
		path := filepath.Join(gameDir, "client", filepath.Base(file.path))
		if err := os.WriteFile(path, file.contents, 0600); err != nil {
			return fmt.Errorf("couldn't stage %s for strict TypeScript validation: %w", file.gameName, err)
		}
		args = append(args, path)
	}
	cmd := exec.Command(compiler, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("generated game-type TypeScript failed strict validation: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func installGeneratedGameTypes(generated []generatedGameTypeFile) error {
	sort.Slice(generated, func(i, j int) bool { return generated[i].path < generated[j].path })
	cleanup := func() {
		for _, file := range generated {
			if file.tempPath != "" {
				_ = os.Remove(file.tempPath)
			}
		}
	}
	defer cleanup()

	seen := make(map[string]bool, len(generated))
	for i := range generated {
		if seen[generated[i].path] {
			return fmt.Errorf("duplicate generated destination %s", generated[i].path)
		}
		seen[generated[i].path] = true
		info, err := os.Lstat(generated[i].path)
		if err == nil {
			if !info.Mode().IsRegular() {
				return fmt.Errorf("refusing to replace non-file generated destination %s", generated[i].path)
			}
			generated[i].hadOriginal = true
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("couldn't inspect generated destination %s: %w", generated[i].path, err)
		}
	}

	// Prepare every replacement on the destination filesystem before changing
	// any destination. Extraction, schema validation, or staging failures leave
	// the prior complete generation untouched.
	for i := range generated {
		file, err := os.CreateTemp(filepath.Dir(generated[i].path), ".game-types-*")
		if err != nil {
			return fmt.Errorf("couldn't stage %s for %s: %w", filepath.Base(generated[i].path), generated[i].gameName, err)
		}
		generated[i].tempPath = file.Name()
		if err := file.Chmod(0644); err != nil {
			_ = file.Close()
			return fmt.Errorf("couldn't set staged permissions for %s: %w", generated[i].gameName, err)
		}
		if _, err := file.Write(generated[i].contents); err != nil {
			_ = file.Close()
			return fmt.Errorf("couldn't stage generated types for %s: %w", generated[i].gameName, err)
		}
		if err := file.Sync(); err != nil {
			_ = file.Close()
			return fmt.Errorf("couldn't sync generated types for %s: %w", generated[i].gameName, err)
		}
		if err := file.Close(); err != nil {
			return fmt.Errorf("couldn't close generated types for %s: %w", generated[i].gameName, err)
		}
	}
	rollback := func(last int) error {
		var rollbackErr error
		for i := last; i >= 0; i-- {
			if generated[i].installed {
				if err := os.Remove(generated[i].path); err != nil && !os.IsNotExist(err) && rollbackErr == nil {
					rollbackErr = err
				}
			}
			if generated[i].backupPath != "" {
				if err := restoreGeneratedGameTypeFile(generated[i].backupPath, generated[i].path); err != nil {
					if rollbackErr == nil {
						rollbackErr = err
					}
				} else {
					generated[i].backupPath = ""
				}
			}
		}
		return rollbackErr
	}
	for i := range generated {
		if generated[i].hadOriginal {
			generated[i].backupPath = generated[i].tempPath + ".backup"
			if err := renameGeneratedGameTypeFile(generated[i].path, generated[i].backupPath); err != nil {
				rollbackErr := rollback(i - 1)
				return fmt.Errorf("couldn't preserve prior %s for %s: %w (rollback: %v)", filepath.Base(generated[i].path), generated[i].gameName, err, rollbackErr)
			}
		}
		if err := renameGeneratedGameTypeFile(generated[i].tempPath, generated[i].path); err != nil {
			var restoreErr error
			if generated[i].backupPath != "" {
				restoreErr = restoreGeneratedGameTypeFile(generated[i].backupPath, generated[i].path)
				if restoreErr == nil {
					generated[i].backupPath = ""
				}
			}
			rollbackErr := rollback(i - 1)
			return fmt.Errorf("couldn't atomically replace %s for %s: %w (current restore: %v; prior rollback: %v; backup: %s)", filepath.Base(generated[i].path), generated[i].gameName, err, restoreErr, rollbackErr, generated[i].backupPath)
		}
		generated[i].tempPath = ""
		generated[i].installed = true
		if generated[i].gameFields != 0 || generated[i].playerFields != 0 {
			fmt.Fprintf(os.Stderr, "  Generated %s/client/_types.ts and _game_renderer.ts (%d game fields, %d player fields)\n", generated[i].gameName, generated[i].gameFields, generated[i].playerFields)
		}
	}
	for i := range generated {
		if generated[i].backupPath != "" {
			if err := os.Remove(generated[i].backupPath); err != nil {
				return fmt.Errorf("installed generated contracts but couldn't remove backup for %s: %w", generated[i].gameName, err)
			}
			generated[i].backupPath = ""
		}
	}
	return nil
}

var renameGeneratedGameTypeFile = os.Rename
var restoreGeneratedGameTypeFile = os.Rename
