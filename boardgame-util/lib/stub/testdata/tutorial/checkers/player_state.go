package checkers

import (
	"github.com/jkomoros/boardgame"
)

//boardgame:codegen
type playerState struct {
	boardgame.BaseSubState
	playerIndex boardgame.PlayerIndex
	Hand        boardgame.Stack `stack:"examplecards" sanitize:"len"`
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

func (p *playerState) GameScore() int {
	//DefaultGameDelegate's PlayerScore will use the GameScore() method on
	//playerState automatically if it exists.
	var sum int
	for _, c := range p.Hand.Components() {
		card := c.Values().(*exampleCard)
		sum += card.Value
	}
	return sum
}
