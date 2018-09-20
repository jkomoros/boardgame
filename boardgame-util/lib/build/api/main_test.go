package api

import (
	"github.com/jkomoros/boardgame/boardgame-util/lib/gamepkg"
	"github.com/workfit/tester/assert"
	"testing"
)

func TestCode(t *testing.T) {

	managers := []string{
		"github.com/jkomoros/boardgame/examples/blackjack",
		"github.com/jkomoros/boardgame/examples/checkers",
		"github.com/jkomoros/boardgame/examples/tictactoe",
	}

	pkgs, err := gamepkg.AllPackages(managers, "")

	assert.For(t).ThatActual(err).IsNil()

	code, err := Code(pkgs, StorageBolt)

	assert.For(t).ThatActual(err).IsNil()

	assert.For(t).ThatActual(string(code)).Equals(apiExpected)
}

var apiExpected = `/*

A server binary generated automatically by 'boardgame-util/lib/build.Api()'

*/
package main

import (
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/checkers"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api"
	"github.com/jkomoros/boardgame/storage/bolt"
)

func main() {

	storage := api.NewServerStorageManager(bolt.NewStorageManager(".database"))
	defer storage.Close()
	api.NewServer(storage,
		blackjack.NewDelegate(),
		checkers.NewDelegate(),
		tictactoe.NewDelegate(),
	).Start()
}
`
