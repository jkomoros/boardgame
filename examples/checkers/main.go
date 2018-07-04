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

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveTypeConfig {
	return moves.Combine(
		moves.AddOrderedForPhase(PhaseSetup,
			auto.MustConfig(
				new(MovePlaceToken),
				moves.WithHelpText("Places one token at a time on the board."),
			),
			auto.MustConfig(
				new(moves.StartPhase),
				moves.WithPhaseToStart(PhasePlaying, PhaseEnum),
			),
		),
		moves.AddForPhase(PhasePlaying,
			auto.MustConfig(
				new(MoveCrownToken),
				moves.WithHelpText("Crowns tokens that make it to the other end of the board."),
				moves.WithSourceStack("Spaces"),
			),
			auto.MustConfig(
				new(moves.FinishTurn),
			),
			auto.MustConfig(
				new(MoveMoveToken),
				moves.WithHelpText("Moves a token from one place to another"),
			),
		),
	)
}

func (g *gameDelegate) ConfigureConstants() map[string]interface{} {
	return map[string]interface{}{
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

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
