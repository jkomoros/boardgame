package static

import (
	"fmt"
	"os"
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
