/*

	checkers is a simple example of the classic checkers game. It exercises a
	grid-like board.

*/
package checkers

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
	"github.com/jkomoros/boardgame/moves/auto"
)

//go:generate autoreader

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

func (g *gameDelegate) ConfigureMoves() *boardgame.MoveTypeConfigBundle {
	return boardgame.NewMoveTypeConfigBundle().AddOrderedMovesForPhase(PhaseSetup,
		auto.MustConfig(
			new(MovePlaceToken),
			moves.WithHelpText("Places one token at a time on the board."),
		),
		auto.MustConfig(
			new(moves.StartPhase),
			moves.WithPhaseToStart(PhasePlaying, PhaseEnum),
		),
	).AddMovesForPhase(PhasePlaying,
		auto.MustConfig(
			new(MoveMoveToken),
			moves.WithHelpText("Moves a token from one place to another"),
		),
		//TODO: a DoneTurn move, that sets turnDone to true (if they don't
		//want to move again after capturing).

		//TODO: a FinishTurn move.
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

func (g *gameDelegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	if deck.Name() != tokenDeckName {
		return nil
	}
	return new(tokenDynamic)
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {
	game := state.GameState().(*gameState)
	if c.Deck.Name() == tokenDeckName {
		return game.UnusedTokens, nil
	}
	return nil, errors.New("Unknown deck")
}

func (g *gameDelegate) GameEndConditionMet(state boardgame.State) bool {
	_, players := concreteStates(state)
	for _, p := range players {
		if p.CapturedTokens.NumComponents() >= numTokens {
			return true
		}
	}

	return false
}

func (g *gameDelegate) PlayerScore(pState boardgame.PlayerState) int {
	p := pState.(*playerState)
	return p.CapturedTokens.NumComponents()
}
