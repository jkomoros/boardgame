package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves/moveinterfaces"
)

/*

FinishTurn is designed to be used as a FixUp move that advances the
CurrentPlayer to the next player when the current player's turn is done. Your
game's playerStates should implement the PlayerTurnFinisher interface, and
your gameState should implement CurrentPlayerSetter.

+autoreader
*/
type FinishTurn struct {
	Base
}

func (f *FinishTurn) ValidConfiguration(exampleState boardgame.MutableState) error {

	if _, ok := exampleState.GameState().(moveinterfaces.CurrentPlayerSetter); !ok {
		return errors.New("GameState does not implement CurrentPlayerSetter")
	}

	if _, ok := exampleState.PlayerStates()[0].(moveinterfaces.PlayerTurnFinisher); !ok {
		return errors.New("PlayerState does not implement PlayerTurnFinisher")
	}

	return nil
}

//Legal checks if the game's CurrentPlayer's TurnDone() returns true.
func (f *FinishTurn) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := f.Base.Legal(state, proposer); err != nil {
		return err
	}

	currentPlayerIndex := state.CurrentPlayerIndex()

	if !currentPlayerIndex.Valid(state) {
		return errors.New("Current player is not valid")
	}

	if currentPlayerIndex < 0 {
		return errors.New("Current player is not valid")
	}

	currentPlayer := state.PlayerStates()[currentPlayerIndex]

	currentPlayerTurnFinisher, ok := currentPlayer.(moveinterfaces.PlayerTurnFinisher)

	if !ok {
		return errors.New("The current player interface did not implement PlayerTurnFinisher")
	}

	if err := currentPlayerTurnFinisher.TurnDone(); err != nil {
		return errors.New("The current player is not done with their turn because " + err.Error())
	}

	return nil

}

//Aoply resets the current player via ResetForTurnEnd, then advances to the
//next player (using game.SetCurrentPlayer), then calls ResetForTurnStart on
//the new player.
func (f *FinishTurn) Apply(state boardgame.MutableState) error {
	currentPlayer := state.PlayerStates()[state.CurrentPlayerIndex()]

	currentPlayerTurnFinisher, ok := currentPlayer.(moveinterfaces.PlayerTurnFinisher)

	if !ok {
		return errors.New("The current player interface did not implement PlayerTurnFinisher")
	}

	if err := currentPlayerTurnFinisher.ResetForTurnEnd(); err != nil {
		return errors.New("Couldn't reset for turn end: " + err.Error())
	}

	newPlayerIndex := state.CurrentPlayerIndex().Next(state)

	playerSetter, ok := state.GameState().(moveinterfaces.CurrentPlayerSetter)

	if !ok {
		return errors.New("Gamestate did not implement CurrentPlayerSetter")
	}

	playerSetter.SetCurrentPlayer(newPlayerIndex)

	currentPlayer = state.PlayerStates()[newPlayerIndex]

	currentPlayerTurnFinisher, ok = currentPlayer.(moveinterfaces.PlayerTurnFinisher)

	if !ok {
		return errors.New("The current player interface did not implement PlayerTurnFinisher")
	}

	if err := currentPlayerTurnFinisher.ResetForTurnStart(); err != nil {
		return errors.New("Couldn't reset for turn start: " + err.Error())
	}

	return nil

}

//MoveTypeFallbackName returns "Finish Turn"
func (f *FinishTurn) MoveTypeFallbackName() string {
	return "Finish Turn"
}

//MoveTypeFallbackHelpText returns "Finishes the player's turn".
func (f *FinishTurn) MoveTypeFallbackHelpText() string {
	return "Finishes the player's turn."
}

//MoveTypeFallbackIsFixUp returns false
func (f *FinishTurn) MoveTypeFallbackIsFixUp() bool {
	return false
}
