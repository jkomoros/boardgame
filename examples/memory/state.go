package memory

import (
	"github.com/jkomoros/boardgame"
)

type gameState struct {
	CurrentPlayer  boardgame.PlayerIndex
	HiddenCards    *boardgame.SizedStack
	RevealedCards  *boardgame.SizedStack
	HideCardsTimer *boardgame.Timer
}

type playerState struct {
	playerIndex       boardgame.PlayerIndex
	CardsLeftToReveal int
	WonCards          *boardgame.GrowableStack
}

func concreteStates(state boardgame.State) (*gameState, []*playerState) {
	game := state.Game().(*gameState)

	players := make([]*playerState, len(state.Players()))

	for i, player := range state.Players() {
		players[i] = player.(*playerState)
	}

	return game, players
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

func (g *gameState) Copy() boardgame.GameState {
	return g.MutableCopy()
}

func (p *playerState) Copy() boardgame.PlayerState {
	return p.MutableCopy()
}

func (g *gameState) MutableCopy() boardgame.MutableGameState {
	var result gameState
	result = *g
	result.HiddenCards = g.HiddenCards.Copy()
	result.RevealedCards = g.RevealedCards.Copy()
	result.HideCardsTimer = g.HideCardsTimer.Copy()
	return &result
}

func (p *playerState) MutableCopy() boardgame.MutablePlayerState {
	var result playerState
	result = *p
	result.WonCards = p.WonCards.Copy()
	return &result
}

func (g *gameState) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader(g)
}

func (p *playerState) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader(p)
}

func (g *gameState) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(g)
}

func (p *playerState) ReadSetter() boardgame.PropertyReadSetter {
	return boardgame.DefaultReadSetter(p)
}
