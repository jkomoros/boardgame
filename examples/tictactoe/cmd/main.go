/*

A simple package that basically just wires together tictactoe into a CLI
version.

*/

package main

import (
	"github.com/jkomoros/boardgame/cli"
	"github.com/jkomoros/boardgame/examples/tictactoe"
)

func main() {

	cli.NewController(tictactoe.NewGame(), tictactoe.Renderer).MainLoop()
}
