package static

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

// badFrameworkImport is the import-path shape that resolves from a game's
// real on-disk location (examples/<name>/client/) but NOT through the
// symlinked static dir that `boardgame-util serve` and `build static`
// assemble — renderers using it 500 in the dev server with an opaque vite
// overlay. The working convention resolves inside the served root:
//
//	import '../../src/components/boardgame-card.js'   // correct
//	import '../../../server/static/src/components/…'  // resolves only in-repo
const badFrameworkImport = "../../../server/static/src/"

// goodFrameworkImport is what badFrameworkImport should be instead.
const goodFrameworkImport = "../../src/"

// LintGameClientImports scans every .ts file in each game's client folder
// for the broken framework-import convention and returns one human-readable
// warning line per offending file. It never fails the build: existing games
// with the bad pattern still transpile in prod builds (vite resolves via
// realpath there), so a hard error would break `serve` for everyone over a
// game they may not even be testing. Callers print the warnings.
func LintGameClientImports(pkgs []*gamepkg.Pkg) []string {
	var warnings []string
	for _, p := range pkgs {
		if p == nil {
			continue
		}
		clientDir := p.ClientFolder()
		if clientDir == "" {
			continue
		}
		entries, err := os.ReadDir(clientDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".ts") {
				continue
			}
			path := filepath.Join(clientDir, e.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			if strings.Contains(string(content), badFrameworkImport) {
				warnings = append(warnings, fmt.Sprintf(
					"WARNING: %s imports framework code via '%s', which will fail to load (500) in the dev server. Use '%s' instead (see examples/memory/client for the convention).",
					filepath.Join(p.Name(), e.Name()), badFrameworkImport, goodFrameworkImport))
			}
		}
	}
	return warnings
}

// gameSrcTypeCheckTSConfig is the type-check-only tsconfig written into the
// assembled static dir. It extends the real tsconfig (so compiler options
// match production) but widens rootDir + include to cover the symlinked
// game-src renderers — which vite/esbuild transpile WITHOUT type-checking,
// so authoring mistakes (wrong move-arg types, missing fields) otherwise
// ship silently. noEmit + a scratch outDir keep it side-effect-free.
const gameSrcTypeCheckTSConfig = `{
  "extends": "./tsconfig.json",
  "compilerOptions": { "rootDir": ".", "noEmit": true, "declaration": false, "composite": false },
  "include": ["src/**/*", "game-src/**/*.ts"],
  "exclude": ["node_modules", "dist", "src/**/*.test.ts"]
}`

// TypeCheckGameSrc runs `tsc --noEmit` over the assembled static dir,
// including the symlinked game-src renderers, and returns tsc's diagnostic
// lines. staticDir must be the assembled dir (game-src symlinks + a
// node_modules symlink already in place — i.e. call AFTER LinkGameClientFolders).
//
// Production callers treat returned diagnostics as fatal: Vite transpiles
// TypeScript without checking it, so continuing would ship a renderer whose
// generated contract is already known to be violated. Returns (nil, err) only
// for infrastructure failures; type errors are returned as diagnostic lines.
func TypeCheckGameSrc(staticDir string) ([]string, error) {
	// A game-src type-check only makes sense once renderers are symlinked in.
	if _, err := os.Stat(filepath.Join(staticDir, "game-src")); err != nil {
		return nil, nil
	}

	configPath := filepath.Join(staticDir, "tsconfig.gamesrc.json")
	if err := os.WriteFile(configPath, []byte(gameSrcTypeCheckTSConfig), 0644); err != nil {
		return nil, fmt.Errorf("couldn't write game-src tsconfig: %w", err)
	}
	defer os.Remove(configPath)

	cmd := exec.Command("npx", "tsc", "--noEmit", "-p", "tsconfig.gamesrc.json")
	cmd.Dir = staticDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil, nil // clean type-check
	}
	// tsc exits non-zero when it reports diagnostics; surface only the
	// lines that name a game-src file (the src/** lines are the framework's
	// own concern, checked by the normal build).
	var diagnostics []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "game-src/") {
			diagnostics = append(diagnostics, "WARNING: game renderer type error: "+line)
		}
	}
	if len(diagnostics) == 0 && len(out) > 0 {
		// tsc failed for a non-diagnostic reason (e.g. couldn't start).
		return nil, fmt.Errorf("game-src type-check could not run: %s", strings.TrimSpace(string(out)))
	}
	return diagnostics, nil
}
