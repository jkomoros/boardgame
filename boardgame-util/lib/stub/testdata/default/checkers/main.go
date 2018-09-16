/*

	checkers is a classic game for two players where you advance across the board, capturing the other player's pawns

*/
package checkers

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
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

func (g *gameDelegate) DisplayName() string {
	return "Checkers"
}

func (g *gameDelegate) Description() string {
	return "A classic game for two players where you advance across the board, capturing the other player's pawns"
}
func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		moves.AddOrderedForPhase(
			PhaseSetUp,

			//Because we used AddOrderedForPhase, this next move won't apply
			//until the move before it is done applying.
			auto.MustConfig(new(moves.StartPhase),
				moves.WithPhaseToStart(PhaseNormal, PhaseEnum),
				moves.WithHelpText("Move to the normal play phase."),
			),
		),
	)

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

func (g *gameDelegate) DefaultNumPlayers() int {
	return 2
}
func (g *gameDelegate) MinNumPlayers() int {
	return 2
}
func (g *gameDelegate) MaxNumPlayers() int {
	return 4
}

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
