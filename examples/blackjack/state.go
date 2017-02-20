package blackjack

import (
	"github.com/jkomoros/boardgame"
)

type mainState struct {
	Game    *gameState
	Players []*playerState
}

type gameState struct {
	DiscardStack  *boardgame.GrowableStack
	DrawStack     *boardgame.GrowableStack
	UnusedCards   *boardgame.GrowableStack
	CurrentPlayer int
}

type playerState struct {
	playerIndex int
	Hand        *boardgame.GrowableStack
	Busted      bool
	Stood       bool
}

func (g *gameState) Props() []string {
	return boardgame.PropertyReaderPropsImpl(g)
}

func (g *gameState) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(g, name)
}

func (g *gameState) Copy() boardgame.GameState {
	var result gameState
	result = *g
	return &result
}

func (p *playerState) Props() []string {
	return boardgame.PropertyReaderPropsImpl(p)
}

func (p *playerState) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(p, name)
}

func (p *playerState) Copy() boardgame.PlayerState {
	var result playerState
	result = *p
	return &result
}

func (p *playerState) PlayerIndex() int {
	return p.playerIndex
}

//HandValue returns the value of the player's hand.
func (p *playerState) HandValue() int {

	var numUnconvertedAces int
	var currentValue int

	for _, card := range cards(p.Hand.ComponentValues()) {
		switch card.Rank {
		case RankAce:
			numUnconvertedAces++
			//We count the ace as 1 now. Later we'll check to see if we can
			//expand any aces.
			currentValue += 1
		case RankJack, RankQueen, RankKing:
			currentValue += 10
		default:
			currentValue += int(card.Rank)
		}
	}

	for numUnconvertedAces > 0 {

		if currentValue >= (targetScore - 10) {
			break
		}

		numUnconvertedAces--
		currentValue += 10
	}

	return currentValue

}

func (m *mainState) Diagram() string {
	return "TODO: implement this"
}

func (s *mainState) GameState() boardgame.GameState {
	return s.Game
}

func (s *mainState) PlayerStates() []boardgame.PlayerState {
	array := make([]boardgame.PlayerState, len(s.Players))

	for i := 0; i < len(s.Players); i++ {
		array[i] = s.Players[i]
	}

	return array
}

func (s *mainState) Copy() boardgame.State {
	array := make([]*playerState, len(s.Players))

	for i := 0; i < len(s.Players); i++ {
		array[i] = s.Players[i].Copy().(*playerState)
	}

	return &mainState{
		Game:    s.Game.Copy().(*gameState),
		Players: array,
	}
}
