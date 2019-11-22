package moves

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/interfaces"
)

/*

FinishTurn is designed to be used as a FixUp move that advances the
CurrentPlayer to the next player when the current player's turn is done. Your
game's playerStates should implement the PlayerTurnFinisher interface, and your
gameState should implement CurrentPlayerSetter (which you can implement easily
by using behaviors.CurrentPlayer). In practice because most of the logic is
contained on your game and playerStates, you can often use this move directly in
your game, with no configuration override or embedding, like so:

    auto.MustConfig(
        new(moves.FinishTurn),
    )

boardgame:codegen
*/
type FinishTurn struct {
	FixUp
}

//ValidConfiguration verfies that GameState implements
//interfaces.CurrentPlayerSetter and the PlayerState implements
//PlayerTurnFinisher.
func (f *FinishTurn) ValidConfiguration(exampleState boardgame.State) error {

	if _, ok := exampleState.GameState().(interfaces.CurrentPlayerSetter); !ok {
		return errors.New("GameState does not implement CurrentPlayerSetter")
	}

	if _, ok := exampleState.PlayerStates()[0].(interfaces.PlayerTurnFinisher); !ok {
		return errors.New("PlayerState does not implement PlayerTurnFinisher")
	}

	return f.FixUp.ValidConfiguration(exampleState)
}

//Legal checks if the game's CurrentPlayer's TurnDone() returns true.
func (f *FinishTurn) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {

	if err := f.Default.Legal(state, proposer); err != nil {
		return err
	}

	currentPlayerIndex := state.CurrentPlayerIndex()

	if !currentPlayerIndex.Valid(state) {
		return errors.New("Current player is not valid")
	}

	if currentPlayerIndex < 0 {
		return errors.New("Current player is not valid")
	}

	currentPlayer := state.ImmutablePlayerStates()[currentPlayerIndex]

	currentPlayerTurnFinisher, ok := currentPlayer.(interfaces.PlayerTurnFinisher)

	if !ok {
		return errors.New("The current player interface did not implement PlayerTurnFinisher")
	}

	if err := currentPlayerTurnFinisher.TurnDone(); err != nil {
		return errors.New("The current player is not done with their turn because " + err.Error())
	}

	return nil

}

//Apply resets the current player via ResetForTurnEnd, then advances to the
//next player (using game.SetCurrentPlayer), then calls ResetForTurnStart on
//the new player.
func (f *FinishTurn) Apply(state boardgame.State) error {
	currentPlayer := state.PlayerStates()[state.CurrentPlayerIndex()]

	currentPlayerTurnFinisher, ok := currentPlayer.(interfaces.PlayerTurnFinisher)

	if !ok {
		return errors.New("The current player interface did not implement PlayerTurnFinisher")
	}

	if err := currentPlayerTurnFinisher.ResetForTurnEnd(); err != nil {
		return errors.New("Couldn't reset for turn end: " + err.Error())
	}

	newPlayerIndex := state.CurrentPlayerIndex().Next(state)

	playerSetter, ok := state.GameState().(interfaces.CurrentPlayerSetter)

	if !ok {
		return errors.New("Gamestate did not implement CurrentPlayerSetter")
	}

	playerSetter.SetCurrentPlayer(newPlayerIndex)

	currentPlayer = state.PlayerStates()[newPlayerIndex]

	currentPlayerTurnFinisher, ok = currentPlayer.(interfaces.PlayerTurnFinisher)

	if !ok {
		return errors.New("The current player interface did not implement PlayerTurnFinisher")
	}

	if err := currentPlayerTurnFinisher.ResetForTurnStart(); err != nil {
		return errors.New("Couldn't reset for turn start: " + err.Error())
	}

	return nil

}

//FallbackName returns "Finish Turn". In many cases you only have one
//FinishTurn move in a game, so this name does not need to be overriden.
func (f *FinishTurn) FallbackName(m *boardgame.GameManager) string {
	return "Finish Turn"
}

//FallbackHelpText returns "Advances to the next player when the
//current player's turn is done."
func (f *FinishTurn) FallbackHelpText() string {
	return "Advances to the next player when the current player's turn is done."
}
