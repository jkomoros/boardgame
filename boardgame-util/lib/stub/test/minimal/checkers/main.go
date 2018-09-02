package checkers

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

/*

Call the code generation for readers and enums here, so "go generate" will generate code correctly.

*/
//go:generate boardgame-util codegen

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "checkers"
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return &playerState{
		playerIndex: index,
	}
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {

	return nil, errors.New("Not yet implemented")

}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	//TODO: this is where you should return your list of moves, often using
	//moves.Combine. See TUTORIAL.md for an example.

	return nil

}

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
