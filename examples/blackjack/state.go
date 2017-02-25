package blackjack

import (
	"fmt"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/playingcards"
	"strings"
)

type mainState struct {
	sanitized bool
	Game      *gameState
	Players   []*playerState
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
	result.Hand = p.Hand.Copy()
	return &result
}

func (p *playerState) PlayerIndex() int {
	return p.playerIndex
}

//HandValue returns the value of the player's hand.
func (p *playerState) HandValue() int {

	var numUnconvertedAces int
	var currentValue int

	for _, card := range playingcards.ValuesToCards(p.Hand.ComponentValues()) {
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

func (m *mainState) Sanitized() bool {
	return m.sanitized
}

func (m *mainState) Diagram() string {
	var result []string

	result = append(result, fmt.Sprintf("Cards left in deck: %d", m.Game.DrawStack.NumComponents()))

	for i, player := range m.Players {

		playerLine := fmt.Sprintf("Player %d", i)

		if i == m.Game.CurrentPlayer {
			playerLine += "  *CURRENT*"
		}

		result = append(result, playerLine)

		statusLine := fmt.Sprintf("\tValue: %d", player.HandValue())

		if player.Busted {
			statusLine += " BUSTED"
		}

		if player.Stood {
			statusLine += " STOOD"
		}

		result = append(result, statusLine)

		result = append(result, "\tCards:")

		for _, card := range playingcards.ValuesToCards(player.Hand.ComponentValues()) {
			result = append(result, "\t\t"+card.String())
		}

		result = append(result, "")
	}

	return strings.Join(result, "\n")
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

func (s *mainState) Copy(sanitized bool) boardgame.State {
	array := make([]*playerState, len(s.Players))

	for i := 0; i < len(s.Players); i++ {
		array[i] = s.Players[i].Copy().(*playerState)
	}

	return &mainState{
		Game:      s.Game.Copy().(*gameState),
		Players:   array,
		sanitized: sanitized,
	}
}
