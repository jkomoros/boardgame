package checkers

import (
	"errors"

	"github.com/jkomoros/boardgame"
)

//boardgame:codegen
type playerState struct {
	boardgame.BaseSubState
	playerIndex          boardgame.PlayerIndex
	Hand                 boardgame.Stack `stack:"examplecards" sanitize:"len"`
	HasDrawnCardThisTurn bool
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

func (p *playerState) GameScore() int {
	//DefaultGameDelegate's PlayerScore will use the GameScore() method on
	//playerState automatically if it exists.

	//This method is exported as a computed property which means this method
	//will be called on created states, including ones that are sanitized.
	//Because Hand, as configured in the struct tag, will be sanitized 'len',
	//sometimes the values we need to sum will be generic placeholder
	//components. However, because newExampleCardDeck used SetGenericValues,
	//we'll always have a *exampleCard, never nil, to cast to.

	var sum int
	for _, c := range p.Hand.Components() {
		card := c.Values().(*exampleCard)
		sum += card.Value
	}
	return sum
}

//TurnDone returns true when this player's turn is done. moves.FinishTurn expects it.
func (p *playerState) TurnDone() error {
	if p.HasDrawnCardThisTurn {
		return nil
	}
	return errors.New("Player hasn't drawn their card yet.")
}

//ResetForTurnStart is called by moves.FinishTurn when the player's turn has
//just started,
func (p *playerState) ResetForTurnStart() error {
	p.HasDrawnCardThisTurn = false
	return nil
}

//ResetForTurnEnd is called by moves.FinishTurn when the player's turn has
//just finished,
func (p *playerState) ResetForTurnEnd() error {
	return nil
}
