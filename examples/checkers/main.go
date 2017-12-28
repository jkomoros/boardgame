/*

	checkers is a simple example of the classic checkers game. It exercises a
	grid-like board.

*/
package checkers

import (
	"github.com/jkomoros/boardgame"
)

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "checkers"
}

func (g *gameDelegate) DisplayName() string {
	return "Checkers"
}

func (g *gameDelegate) Description() string {
	return "Checkers is the classic game on a grid where players compete to capture opponents' pieces."
}

func (g *gameDelegate) MinNumPlayers() int {
	return 2
}

func (g *gameDelegate) MaxNumPlayers() int {
	return 2
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 2
}
