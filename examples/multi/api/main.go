/*

multi is a server that loads up multiple games on one server to demonstrate how that works.

*/
package main

import (
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api"
)

func main() {
	storage := api.NewDefaultStorageManager()
	defer storage.Close()
	api.NewServer(storage, blackjack.NewManager(storage), tictactoe.NewManager(storage)).Start()
}
