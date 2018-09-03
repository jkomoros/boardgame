package checkers

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

//TODO: define your move structs here. Don't forget the 'boardgame:codegen'
//magic comment, and don't forget to return them from
//delegate.ConfigureMoves().

//Typically you create a separate file for moves of each major phase, and put
//the moves for that phase in it.

//boardgame:codegen
type moveDrawCard struct {
	moves.CurrentPlayer
}

func (m *moveDrawCard) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
	//There's important logic up the chain of move types; it's always
	//important to call your parent's Legal. CurrentPlayer will ensure that
	//it's the proposer's turn.
	if err := m.CurrentPlayer.Legal(state, proposer); err != nil {
		return err
	}

	game := state.ImmutableGameState().(*gameState)

	if game.DrawStack.Len() < 1 {
		return errors.New("No cards left to draw!")
	}

	return nil

}

func (m *moveDrawCard) Apply(state boardgame.State) error {
	game := state.GameState().(*gameState)
	player := state.CurrentPlayer().(*playerState)

	if err := game.DrawStack.First().MoveToLastSlot(player.Hand); err != nil {
		return err
	}

	player.HasDrawnCardThisTurn = true
	return nil
}
