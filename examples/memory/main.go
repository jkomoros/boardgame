/*

memory is a simple example game based on memory--where players take turn
flipping over two cards, and keeping them if they match.

*/
package memory

import (
	"github.com/jkomoros/boardgame"
)

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "memory"
}

func (g *gameDelegate) DisplayName() string {
	return "Memory"
}

func (g *gameDelegate) DefaultNumPlayeres() int {
	return 2
}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers < 4 && numPlayers > 1
}

func EmptyGameState() boardgame.MutableGameState {
	return &gameState{}
}

func EmptyPlayerState() boardgame.MutablePlayerState {
	return nil
}
