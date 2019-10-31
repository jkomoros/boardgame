/*

Package checkers is a simple example of the classic checkers game. It exercises
a grid-like board.

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

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Combine(
		moves.AddOrderedForPhase(PhaseSetup,
			auto.MustConfig(
				new(movePlaceToken),
				moves.WithHelpText("Places one token at a time on the board."),
			),
			auto.MustConfig(
				new(moves.StartPhase),
				moves.WithPhaseToStart(PhasePlaying, PhaseEnum),
			),
		),
		moves.AddForPhase(PhasePlaying,
			auto.MustConfig(
				new(moveCrownToken),
				moves.WithHelpText("Crowns tokens that make it to the other end of the board."),
				moves.WithSourceProperty("Spaces"),
			),
			auto.MustConfig(
				new(moves.FinishTurn),
			),
			auto.MustConfig(
				new(moveMoveToken),
				moves.WithHelpText("Moves a token from one place to another"),
			),
		),
	)
}

func (g *gameDelegate) ConfigureConstants() boardgame.PropertyCollection {
	return boardgame.PropertyCollection{
		"BOARD_SIZE": boardSize,
	}
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {
	return &playerState{
		playerIndex: index,
	}
}

func (g *gameDelegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	if deck.Name() != tokenDeckName {
		return nil
	}
	return new(tokenDynamic)
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game := state.ImmutableGameState().(*gameState)
	if c.Deck().Name() == tokenDeckName {
		return game.UnusedTokens, nil
	}
	return nil, errors.New("Unknown deck")
}

func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	_, players := concreteStates(state)
	for _, p := range players {
		if p.CapturedTokens.NumComponents() >= numTokens {
			return true
		}
	}

	return false
}

func (g *gameDelegate) PlayerScore(pState boardgame.ImmutablePlayerState) int {
	p := pState.(*playerState)
	return p.CapturedTokens.NumComponents()
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	return map[string]*boardgame.Deck{
		tokenDeckName: newTokenDeck(),
	}
}

//NewDelegate is the primary entrypoint of the package, returning a new delegate
//that configures a game of checkers.
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
