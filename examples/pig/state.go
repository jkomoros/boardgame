package pig

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
)

//boardgame:codegen
type gameState struct {
	base.SubState
	behaviors.CurrentPlayerBehavior
	Die         boardgame.SizedStack `sizedstack:"dice"`
	TargetScore int
}

//boardgame:codegen
type playerState struct {
	base.SubState
	Busted     bool
	Done       bool
	DieCounted bool
	RoundScore int
	TotalScore int
}

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

func (p *playerState) TurnDone() error {
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

func (p *playerState) ResetForTurnStart() error {
	p.ResetForTurn()
	return nil
}

func (p *playerState) ResetForTurnEnd() error {
	if p.Done {
		p.TotalScore += p.RoundScore
	}
	p.ResetForTurn()
	return nil
}
