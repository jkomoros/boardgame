/*

A simple package that basically just wires together tictactoe into a debug
server version.

*/

package main

import (
	"github.com/jkomoros/boardgame/debugserver"
	"github.com/jkomoros/boardgame/examples/tictactoe"
)

func main() {
	debugserver.NewServer(tictactoe.NewGame()).Start()
}
