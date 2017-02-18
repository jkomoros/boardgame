/*

A simple package that basically just wires together tictactoe into a debug
server version.

*/

package main

import (
	"github.com/jkomoros/boardgame/examples/tictactoe"
	"github.com/jkomoros/boardgame/server"
)

func main() {
	storage := server.NewDefaultStorageManager()
	defer storage.Close()
	server.NewServer(tictactoe.NewManager(storage), storage).Start()
}
