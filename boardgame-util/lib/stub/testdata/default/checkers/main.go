/*

Package checkers implements a game that is a classic game for two players where you advance across the board, capturing the other player's pawns

*/
package checkers

import (
	"errors"
	"reflect"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/moves"
)

/*

Call the code generation for readers and enums here, so "go generate" will generate code correctly.

*/
//go:generate boardgame-util codegen

type gameDelegate struct {
	base.GameDelegate
}

var memoizedDelegateName string

func (g *gameDelegate) Name() string {

	//If our package name and delegate.Name() don't match, NewGameManager will
	//fail with an error. Given they have to be the same, we might as well
	//just ensure they are actually the same, via a one-time reflection.

	if memoizedDelegateName == "" {
		pkgPath := reflect.ValueOf(g).Elem().Type().PkgPath()
		pathPieces := strings.Split(pkgPath, "/")
		memoizedDelegateName = pathPieces[len(pathPieces)-1]
	}
	return memoizedDelegateName
}

func (g *gameDelegate) Description() string {
	return "A classic game for two players where you advance across the board, capturing the other player's pawns"
}
func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		moves.AddOrderedForPhase(
			phaseSetUp,

			//Because we used AddOrderedForPhase, this next move won't apply
			//until the move before it is done applying.
			auto.MustConfig(new(moves.StartPhase),
				moves.WithPhaseToStart(phaseNormal, phaseEnum),
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

//NewDelegate is the primary entrypoint to the package. It implements a
//boardgame.GameDelegate that configures a game of checkers
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
