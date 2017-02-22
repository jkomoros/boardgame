/*

A simple package that basically just wires together tictactoe into a debug
server version.

*/

package main

import (
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server/api"
)

func main() {
	storage := api.NewDefaultStorageManager()
	defer storage.Close()
	api.NewServer(storage, tictactoe.NewManager(storage)).Start()
}
