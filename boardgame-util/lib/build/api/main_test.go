package api

import (
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

var apiExpected = `/*
A server binary generated automatically by 'boardgame-util/lib/build/api/Build()'
*/
package main

import (
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api"
	"github.com/jkomoros/boardgame/storage/bolt"
)

// companionCapableGames is the list of game names that ship the Table+Hand
// renderer pair (boardgame-render-game-<name>-table.ts AND -hand.ts) as of
// this build. Computed by boardgame-util at build time via a filesystem
// walk (see boardgame-util/lib/build/static.CompanionCapableGames). The
// server uses this to populate managerInfo.supportsTableHandMode and surface
// it in doListManager so the create-game form can show the
// "Use shared projector + phones" toggle for supporting games. (Spec §5.3.)
var companionCapableGames = []string{}

func main() {

	storage := api.NewServerStorageManager(bolt.NewStorageManager(".database"))
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
	"github.com/jkomoros/boardgame/storage/bolt"
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
var companionCapableGames = []string{}

func main() {

	storage := api.NewServerStorageManager(bolt.NewStorageManager(".database"))
	defer storage.Close()
	api.NewServer(storage,
		blackjack.NewDelegate(),
		checkers.NewDelegate(),
		tictactoe.NewDelegate(),
	).AddOverrides(overrides).WithCompanionCapableGames(companionCapableGames).Start()
}
`
