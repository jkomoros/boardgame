package static

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
)

// CompanionCapableGames returns the names of game packages that opt into
// Table+Hand companion mode by shipping BOTH a boardgame-render-game-
// <name>-table.ts AND a boardgame-render-game-<name>-hand.ts file in their
// client renderer folder (see Pkg.ClientFolder).
//
// This is the build-time half of the spec §5.1 / §5.2 capability detection:
// the result gets injected into ClientConfig.TableHandSupportedGames (and a
// matching field on config.json for the api binary) so the create-game form
// can show the toggle for supporting games. The api binary does no
// filesystem walking; it just consumes the precomputed list.
//
// Result is sorted for stable output.
func CompanionCapableGames(pkgs []*gamepkg.Pkg) []string {
	var capable []string
	for _, p := range pkgs {
		if pkgIsCompanionCapable(p) {
			capable = append(capable, p.Name())
		}
	}
	sort.Strings(capable)
	return capable
}

func pkgIsCompanionCapable(p *gamepkg.Pkg) bool {
	if p == nil {
		return false
	}
	clientDir := p.ClientFolder()
	if clientDir == "" {
		return false
	}
	tableFile := filepath.Join(clientDir, "boardgame-render-game-"+p.Name()+"-table.ts")
	handFile := filepath.Join(clientDir, "boardgame-render-game-"+p.Name()+"-hand.ts")
	return fileExists(tableFile) && fileExists(handFile)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
