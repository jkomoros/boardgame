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
	//TurnDone should return true when this player's turn is done, false
	//otherwise.
	TurnDone() bool
	//ResetForTurnStart will be called when this player begins their turn.
	ResetForTurnStart()
	//ResetForTurnEnd will be called right before the CurrentPlayer is
	//advanced to the next player.
	ResetForTurnEnd()
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

//Legal checks if the game's CurrentPlayer's TurnDone() returns true.
func (f *FinishTurn) Legal(state boardgame.State, proposer boardgame.PlayerIndex) error {
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

	if !currentPlayerTurnFinisher.TurnDone() {
		return errors.New("The current player is not done with their turn")
	}

	return nil

}

//Aoply resets the current player via ResetForTurnEnd, then advances to the
//next player (using game.SetCurrentPlayer), then calls ResetForTurnStart on
//the new player.
func (f *FinishTurn) Apply(state boardgame.State) error {
	currentPlayer := state.PlayerStates()[state.Game().CurrentPlayerIndex()]

	currentPlayerTurnFinisher, ok := currentPlayer.(PlayerTurnFinisher)

	if !ok {
		return errors.New("The current player interface did not implement PlayerTurnFinisher")
	}

	currentPlayerTurnFinisher.ResetForTurnEnd()

	newPlayerIndex := state.Game().CurrentPlayerIndex().Next(state)

	playerSetter, ok := state.GameState().(CurrentPlayerSetter)

	if !ok {
		return errors.New("Gamestate did not implement CurrentPlayerSetter")
	}

	playerSetter.SetCurrentPlayer(newPlayerIndex)

	if state.Game().CurrentPlayerIndex() != newPlayerIndex {
		return errors.New("Calling SetCurrentPlayer did not set the CurrentPlayer to the  new index")
	}

	currentPlayer = state.PlayerStates()[state.Game().CurrentPlayerIndex()]

	currentPlayerTurnFinisher, ok = currentPlayer.(PlayerTurnFinisher)

	if !ok {
		return errors.New("The current player interface did not implement PlayerTurnFinisher")
	}

	currentPlayerTurnFinisher.ResetForTurnStart()

	return nil

}
