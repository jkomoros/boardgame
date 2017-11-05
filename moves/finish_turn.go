package moves

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//CurrentPlayerSetter should be implemented by gameStates that use FinishTurn.
type CurrentPlayerSetter interface {
	SetCurrentPlayer(currentPlayer boardgame.PlayerIndex)
}

//PlayerTurnFinisher is the interface your playerState is expected to adhere
//to when you use FinishTurn.
type PlayerTurnFinisher interface {
	//TurnDone should return nil when the turn is done, or a descriptive error
	//if the turn is not done.
	TurnDone() error
	//ResetForTurnStart will be called when this player begins their turn.
	ResetForTurnStart() error
	//ResetForTurnEnd will be called right before the CurrentPlayer is
	//advanced to the next player.
	ResetForTurnEnd() error
}

/*

FinishTurn is designed to be used as a FixUp move that advances the
CurrentPlayer to the next player when the current player's turn is done. Your
game's playerStates should implement the PlayerTurnFinisher interface, and
your gameState should implement CurrentPlayerSetter.

*/
type FinishTurn struct {
	Base
}

func (f *FinishTurn) ValidConfiguration(exampleState boardgame.MutableState) error {

	if _, ok := exampleState.GameState().(CurrentPlayerSetter); !ok {
		return errors.New("GameState does not implement CurrentPlayerSetter")
	}

	if _, ok := exampleState.PlayerStates()[0].(PlayerTurnFinisher); !ok {
		return errors.New("PlayerState does not implement PlayerTurnFinisher")
	}

	return nil
}

//Legal checks if the game's CurrentPlayer's TurnDone() returns true.
func (f *FinishTurn) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {

	if err := f.Base.Legal(state, proposer); err != nil {
		return err
	}

	currentPlayerIndex := state.Game().CurrentPlayerIndex()

	if !currentPlayerIndex.Valid(state) {
		return errors.New("Current player is not valid")
	}

	if currentPlayerIndex < 0 {
		return errors.New("Current player is not valid")
	}

	currentPlayer := state.PlayerStates()[currentPlayerIndex]

	currentPlayerTurnFinisher, ok := currentPlayer.(PlayerTurnFinisher)

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
	currentPlayer := state.PlayerStates()[state.Game().CurrentPlayerIndex()]

	currentPlayerTurnFinisher, ok := currentPlayer.(PlayerTurnFinisher)

	if !ok {
		return errors.New("The current player interface did not implement PlayerTurnFinisher")
	}

	if err := currentPlayerTurnFinisher.ResetForTurnEnd(); err != nil {
		return errors.New("Couldn't reset for turn end: " + err.Error())
	}

	newPlayerIndex := state.Game().CurrentPlayerIndex().Next(state)

	playerSetter, ok := state.GameState().(CurrentPlayerSetter)

	if !ok {
		return errors.New("Gamestate did not implement CurrentPlayerSetter")
	}

	playerSetter.SetCurrentPlayer(newPlayerIndex)

	currentPlayer = state.PlayerStates()[newPlayerIndex]

	currentPlayerTurnFinisher, ok = currentPlayer.(PlayerTurnFinisher)

	if !ok {
		return errors.New("The current player interface did not implement PlayerTurnFinisher")
	}

	if err := currentPlayerTurnFinisher.ResetForTurnStart(); err != nil {
		return errors.New("Couldn't reset for turn start: " + err.Error())
	}

	return nil

}
