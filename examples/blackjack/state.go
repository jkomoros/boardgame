package blackjack

import (
	"errors"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/behaviors"
	"github.com/jkomoros/boardgame/components/playingcards"
	"github.com/jkomoros/boardgame/enum"
)

//boardgame:codegen
const (
	phaseInitialDeal = iota
	phaseNormalPlay
)

func concreteStates(state boardgame.ImmutableState) (*gameState, []*playerState) {
	game := state.ImmutableGameState().(*gameState)

	players := make([]*playerState, len(state.ImmutablePlayerStates()))

	for i, player := range state.ImmutablePlayerStates() {
		players[i] = player.(*playerState)
	}

	return game, players
}

//boardgame:codegen
type gameState struct {
	base.SubState
	behaviors.RoundRobin
	Phase         enum.Val        `enum:"phase"`
	DiscardStack  boardgame.Stack `stack:"cards" sanitize:"len"`
	DrawStack     boardgame.Stack `stack:"cards" sanitize:"len"`
	UnusedCards   boardgame.Stack `stack:"cards"`
	CurrentPlayer boardgame.PlayerIndex
}

//boardgame:codegen
type playerState struct {
	base.SubState
	HiddenHand  boardgame.Stack       `stack:"cards,1" sanitize:"len"`
	VisibleHand boardgame.Stack       `stack:"cards"`
	Hand        boardgame.MergedStack `concatenate:"HiddenHand,VisibleHand"`
	Busted      bool
	Stood       bool
}

func (g *gameState) SetCurrentPlayer(currentPlayer boardgame.PlayerIndex) {
	g.CurrentPlayer = currentPlayer
}

func (g *gameState) SetCurrentPhase(phase int) {
	g.Phase.SetValue(phase)
}

func (p *playerState) TurnDone() error {
	if !p.Busted && !p.Stood {
		return errors.New("they have neither busted nor decided to stand")
	}
	return nil
}

func (p *playerState) ResetForTurnStart() error {
	p.Stood = false
	return nil
}

func (p *playerState) ResetForTurnEnd() error {
	return nil
}

func handValue(components []boardgame.ImmutableComponentInstance) int {
	var numUnconvertedAces int
	var currentValue int

	for _, c := range components {
		card := c.Values().(*playingcards.Card)
		switch card.Rank.Value() {
		case playingcards.RankAce:
			numUnconvertedAces++
			//We count the ace as 1 now. Later we'll check to see if we can
			//expand any aces.
			currentValue++
		case playingcards.RankJack, playingcards.RankQueen, playingcards.RankKing:
			currentValue += 10
		default:
			currentValue += card.Rank.Value()
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

//HandValue returns the value of the player's hand.
func (p *playerState) HandValue() int {
	return handValue(p.Hand.ImmutableComponents())
}
