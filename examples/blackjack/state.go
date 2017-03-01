package blackjack

import (
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/playingcards"
)

func concreteStates(state *boardgame.State) (*gameState, []*playerState) {
	game := state.Props.Game.(*gameState)

	players := make([]*playerState, len(state.Props.Players))

	for i, player := range state.Props.Players {
		players[i] = player.(*playerState)
	}

	return game, players
}

type gameState struct {
	DiscardStack  *boardgame.GrowableStack
	DrawStack     *boardgame.GrowableStack
	UnusedCards   *boardgame.GrowableStack
	CurrentPlayer int
}

type playerState struct {
	playerIndex    int
	GotInitialDeal bool
	HiddenHand     *boardgame.GrowableStack
	VisibleHand    *boardgame.GrowableStack
	Busted         bool
	Stood          bool
}

func (g *gameState) Reader() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(g)
}

func (g *gameState) Copy() boardgame.GameState {
	var result gameState
	result = *g
	result.DiscardStack = g.DiscardStack.Copy()
	result.DrawStack = g.DrawStack.Copy()
	result.UnusedCards = g.UnusedCards.Copy()
	return &result
}

func (p *playerState) Reader() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(p)
}

func (p *playerState) Copy() boardgame.PlayerState {
	var result playerState
	result = *p
	result.VisibleHand = p.VisibleHand.Copy()
	result.HiddenHand = p.HiddenHand.Copy()
	return &result
}

func (p *playerState) PlayerIndex() int {
	return p.playerIndex
}

func (p *playerState) EffectiveHand() []*playingcards.Card {
	return append(playingcards.ValuesToCards(p.HiddenHand.ComponentValues()), playingcards.ValuesToCards(p.VisibleHand.ComponentValues())...)
}

//HandValue returns the value of the player's hand.
func (p *playerState) HandValue() int {

	var numUnconvertedAces int
	var currentValue int

	for _, card := range p.EffectiveHand() {
		switch card.Rank {
		case playingcards.RankAce:
			numUnconvertedAces++
			//We count the ace as 1 now. Later we'll check to see if we can
			//expand any aces.
			currentValue += 1
		case playingcards.RankJack, playingcards.RankQueen, playingcards.RankKing:
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
