package pig

import (
	"github.com/jkomoros/boardgame"
)

//+autoreader
type gameState struct {
	CurrentPlayer boardgame.PlayerIndex
	Die           *boardgame.SizedStack
	TargetScore   int
}

//+autoreader
type playerState struct {
	playerIndex boardgame.PlayerIndex
	Busted      bool
	Done        bool
	DieCounted  bool
	RoundScore  int
	TotalScore  int
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

func (p *playerState) ResetForTurn() {
	p.Done = false
	p.Busted = false
	p.RoundScore = 0
	p.DieCounted = true
}
