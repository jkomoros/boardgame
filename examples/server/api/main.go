/*

multi is a server that loads up multiple games on one server to demonstrate
how that works. It is also possible to have a server with only one game.

*/
package main

import (
	"github.com/jkomoros/boardgame/examples/blackjack"
	"github.com/jkomoros/boardgame/examples/debuganimations"
	"github.com/jkomoros/boardgame/examples/memory"
	"github.com/jkomoros/boardgame/examples/pig"
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api"
	"github.com/jkomoros/boardgame/storage/bolt"
)

func main() {

	//This example uses the bolt db backend because it's easier to get set up
	//quickly. Normall your server would use api.NewDefaultStorageManager
	//here, which would use the MySQL backend.
	storage := api.NewServerStorageManager(bolt.NewStorageManager(".database"))
	defer storage.Close()
	api.NewServer(storage,
		api.MustNewManager(blackjack.NewManager(storage)),
		api.MustNewManager(tictactoe.NewManager(storage)),
		api.MustNewManager(memory.NewManager(storage)),
		api.MustNewManager(debuganimations.NewManager(storage)),
		api.MustNewManager(pig.NewManager(storage)),
	).Start()
}
