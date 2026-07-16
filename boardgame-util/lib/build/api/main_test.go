package api

import (
	"strings"
	"testing"

	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
	"github.com/workfit/tester/assert"
)

func TestCode(t *testing.T) {

	managers := []string{
		"github.com/jkomoros/boardgame/examples/blackjack",
		"github.com/jkomoros/boardgame/examples/checkers",
		"github.com/jkomoros/boardgame/examples/tictactoe",
	}

	pkgs, err := gamepkg.AllPackages(managers, "")

	assert.For(t).ThatActual(err).IsNil()

	code, err := Code(pkgs, StorageBolt, nil)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(string(code)).Equals(apiExpected)
}

func TestOverrideCode(t *testing.T) {
	managers := []string{
		"github.com/jkomoros/boardgame/examples/blackjack",
		"github.com/jkomoros/boardgame/examples/checkers",
		"github.com/jkomoros/boardgame/examples/tictactoe",
	}

	pkgs, err := gamepkg.AllPackages(managers, "")

	assert.For(t).ThatActual(err).IsNil()

	code, err := Code(pkgs, StorageBolt, &Options{OverrideOfflineDevMode: true})

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(string(code)).Equals(apiOverrideExpected).ThenDiffOnFail()
}

func TestStorageImportCannotCollideWithGamePackage(t *testing.T) {
	pkgs, err := gamepkg.AllPackages([]string{
		"github.com/jkomoros/boardgame/examples/memory",
	}, "")
	assert.For(t).ThatActual(err).IsNil()

	code, err := Code(pkgs, StorageMemory, nil)
	assert.For(t).ThatActual(err).IsNil()

	text := string(code)
	if !strings.Contains(text, "\"github.com/jkomoros/boardgame/examples/memory\"") {
		t.Fatal("generated code did not import the Memory game")
	}
	if !strings.Contains(text, "boardgamestorage \"github.com/jkomoros/boardgame/storage/memory\"") {
		t.Fatal("generated code did not alias the memory storage import")
	}
	if !strings.Contains(text, "boardgamestorage.NewStorageManager()") {
		t.Fatal("generated code did not use the storage alias in its constructor")
	}
}

func TestAvailableStorageAliasAvoidsPackageNames(t *testing.T) {
	pkgs, err := gamepkg.AllPackages([]string{
		"github.com/jkomoros/boardgame/examples/memory",
	}, "")
	assert.For(t).ThatActual(err).IsNil()

	// The real package proves ordinary discovery. A small package-name seam
	// test proves aliases remain deterministic even when the preferred spelling
	// is occupied; constructing a synthetic gamepkg.Pkg would require reaching
	// through that package's intentionally private fields.
	if alias := availableImportAlias(map[string]bool{"boardgamestorage": true}); alias != "boardgamestorage1" {
		t.Fatalf("alias = %q, want boardgamestorage1", alias)
	}
	if alias := availableStorageAlias(pkgs); alias != "boardgamestorage" {
		t.Fatalf("alias = %q, want boardgamestorage", alias)
	}
}

func TestCodeIncludesAllowedOriginsOverrideWithoutOfflineMode(t *testing.T) {
	const origins = "http://localhost:49152,http://127.0.0.1:49152"
	code, err := Code(nil, StorageMemory, &Options{
		OverrideAllowedOrigins: origins,
	})
	if err != nil {
		t.Fatal(err)
	}

	generated := string(code)
	for _, want := range []string{
		`"github.com/jkomoros/boardgame/boardgame-util/lib/config"`,
		`config.OverrideAllowedOrigins("` + origins + `")`,
		`.AddOverrides(overrides)`,
	} {
		if !strings.Contains(generated, want) {
			t.Errorf("generated code did not contain %q:\n%s", want, generated)
		}
	}
	if strings.Contains(generated, "EnableOfflineDevMode") {
		t.Errorf("allowed-origins-only override unexpectedly enabled offline mode:\n%s", generated)
	}
}

func TestCodeOmitsConfigOverrideMachineryByDefault(t *testing.T) {
	code, err := Code(nil, StorageMemory, nil)
	if err != nil {
		t.Fatal(err)
	}

	generated := string(code)
	for _, unwanted := range []string{
		`boardgame-util/lib/config`,
		`var overrides`,
		`.AddOverrides(`,
	} {
		if strings.Contains(generated, unwanted) {
			t.Errorf("default generated code unexpectedly contained %q:\n%s", unwanted, generated)
		}
	}
}

var apiExpected = `/*
A server binary generated automatically by 'boardgame-util/lib/build/api/Build()'
*/
package main

import (
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api"
	boardgamestorage "github.com/jkomoros/boardgame/storage/bolt"
)

// companionCapableGames is the list of game names that ship the Table+Hand
// renderer pair (boardgame-render-game-<name>-table.ts AND -hand.ts) as of
// this build. Computed by boardgame-util at build time via a filesystem
// walk (see boardgame-util/lib/build/static.CompanionCapableGames). The
// server uses this to populate managerInfo.supportsTableHandMode and surface
// it in doListManager so the create-game form can show the
// "Use shared projector + phones" toggle for supporting games. (Spec §5.3.)
var companionCapableGames = []string{
	"blackjack",
}

func main() {

	storage := api.NewServerStorageManager(boardgamestorage.NewStorageManager(".database"))
	defer storage.Close()
	api.NewServer(storage,
		blackjack.NewDelegate(),
		checkers.NewDelegate(),
		tictactoe.NewDelegate(),
	).WithCompanionCapableGames(companionCapableGames).Start()
}
`

var apiOverrideExpected = `/*
A server binary generated automatically by 'boardgame-util/lib/build/api/Build()'
*/
package main

import (
	"github.com/jkomoros/boardgame/boardgame-util/lib/config"
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api"
	boardgamestorage "github.com/jkomoros/boardgame/storage/bolt"
)

var overrides []config.OptionOverrider

func init() {
	overrides = append(overrides, config.EnableOfflineDevMode())
}

// companionCapableGames is the list of game names that ship the Table+Hand
// renderer pair (boardgame-render-game-<name>-table.ts AND -hand.ts) as of
// this build. Computed by boardgame-util at build time via a filesystem
// walk (see boardgame-util/lib/build/static.CompanionCapableGames). The
// server uses this to populate managerInfo.supportsTableHandMode and surface
// it in doListManager so the create-game form can show the
// "Use shared projector + phones" toggle for supporting games. (Spec §5.3.)
var companionCapableGames = []string{
	"blackjack",
}

func main() {

	storage := api.NewServerStorageManager(boardgamestorage.NewStorageManager(".database"))
	defer storage.Close()
	api.NewServer(storage,
		blackjack.NewDelegate(),
		checkers.NewDelegate(),
		tictactoe.NewDelegate(),
	).AddOverrides(overrides).WithCompanionCapableGames(companionCapableGames).Start()
}
`
