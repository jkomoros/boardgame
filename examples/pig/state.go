package pig

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
)

func init() {

	//Make sure that we get compile-time errors if our player and game state
	//don't adhere to the interfaces that moves.FinishTurn expects
	var playerTurnFinisher moves.PlayerTurnFinisher
	playerTurnFinisher = &playerState{}

	if playerTurnFinisher == nil {
		panic("Nil")
	}

	var currentPlayerSetter moves.CurrentPlayerSetter
	currentPlayerSetter = &gameState{}

	if currentPlayerSetter == nil {
		panic("Nil")
	}
}

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

func (g *gameState) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
	g.CurrentPlayer = currentPlayer
}

func (p *playerState) PlayerIndex() boardgame.PlayerIndex {
	return p.playerIndex
}

func (p *playerState) TurnDone(state boardgame.State) error {
	if !p.DieCounted {
		return errors.New("the most recent die roll has not been counted")
	}

	if !p.Busted && !p.Done {
		return errors.New("they have not either busted or signaled that they are done")
	}

	return nil
}

func (p *playerState) ResetForTurn() {
	p.Done = false
	p.Busted = false
	p.RoundScore = 0
	p.DieCounted = true
}

func (p *playerState) ResetForTurnStart(state boardgame.State) {
	p.ResetForTurn()
}

func (p *playerState) ResetForTurnEnd(state boardgame.State) {
	if p.Done {
		p.TotalScore += p.RoundScore
	}
	p.ResetForTurn()
}
