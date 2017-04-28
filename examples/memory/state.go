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
