package memory

import (
	"errors"
	"github.com/jkomoros/boardgame"
)

//+autoreader
type gameState struct {
	boardgame.BaseSubState
	CardSet        string
	NumCards       int
	CurrentPlayer  boardgame.PlayerIndex
	HiddenCards    boardgame.MutableStack `sizedstack:"cards,40" sanitize:"order"`
	RevealedCards  boardgame.MutableStack `sizedstack:"cards,40"`
	HideCardsTimer boardgame.MutableTimer
	//Where cards not in use reside most of the time
	UnusedCards boardgame.MutableStack `stack:"cards"`
}

//+autoreader
type playerState struct {
	boardgame.BaseSubState
	playerIndex       boardgame.PlayerIndex
	CardsLeftToReveal int
	WonCards          boardgame.MutableStack `stack:"cards"`
}

func concreteStates(state boardgame.State) (*gameState, []*playerState) {
	game := state.GameState().(*gameState)

	players := make([]*playerState, len(state.PlayerStates()))

	for i, player := range state.PlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

func (p *playerState) TurnDone() error {
	if p.CardsLeftToReveal > 0 {
		return errors.New("they still have cards left to reveal")
	}

	game, _ := concreteStates(p.State())

	if game.RevealedCards.NumComponents() > 0 {
		return errors.New("there are still some cards revealed, which they must hide")
	}

	return nil
}

func (p *playerState) ResetForTurnStart() error {
	p.CardsLeftToReveal = 2
	return nil
}

func (p *playerState) ResetForTurnEnd() error {
	return nil
}

func (g *gameState) CardsInGrid() int {
	return g.HiddenCards.NumComponents() + g.RevealedCards.NumComponents()
}

func (g *gameState) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
	g.CurrentPlayer = currentPlayer
}

func (g *gameState) CurrentPlayerHasCardsToReveal() bool {
	_, players := concreteStates(g.State())

	p := players[g.CurrentPlayer]

	return p.CardsLeftToReveal > 0
}
