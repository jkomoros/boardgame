package tictactoe

import (
	"github.com/jkomoros/boardgame"
	"strings"
)

type gameState struct {
	CurrentPlayer int
	Slots         *boardgame.SizedStack
}

func (g *gameState) tokenValue(row, col int) string {
	return g.tokenValueAtIndex(rowColToIndex(row, col))
}

func (g *gameState) tokenValueAtIndex(index int) string {
	c := g.Slots.ComponentAt(index)
	if c == nil {
		return ""
	}
	return c.Values.(*playerToken).Value
}

func rowColToIndex(row, col int) int {
	return row*DIM + col
}

func (g *gameState) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader(g)
}

func (g *gameState) Copy() boardgame.GameState {
	var result gameState
	result = *g
	result.Slots = g.Slots.Copy()
	return &result
}

type playerState struct {
	playerIndex  int
	TokenValue   string
	UnusedTokens *boardgame.GrowableStack
	//How many tokens they have left to place this turn.
	TokensToPlaceThisTurn int
}

func (p *playerState) Reader() boardgame.PropertyReader {
	return boardgame.DefaultReader(p)
}

func (p *playerState) Copy() boardgame.PlayerState {
	var result playerState
	result = *p
	result.UnusedTokens = p.UnusedTokens.Copy()
	return &result
}

func (p *playerState) PlayerIndex() int {
	return p.playerIndex
}

type mainState struct {
	Game    *gameState
	Players []*playerState
}

func (s *mainState) userFromTokenValue(value string) *playerState {
	for _, player := range s.Players {
		if player.TokenValue == value {
			return player
		}
	}
	return nil
}

func (s *mainState) Diagram() string {

	//Get an array of *playerTokenValues corresponding to tokens currently in
	//the stack.
	tokens := playerTokenValues(s.Game.Slots.ComponentValues())

	tokenValues := make([]string, len(tokens))

	for i, token := range tokens {
		if token == nil {
			tokenValues[i] = " "
			continue
		}
		tokenValues[i] = token.Value
	}

	result := make([]string, 7)

	//TODO: loop thorugh this instead of unrolling the loop by hand
	result[0] = tokenValues[0] + "|" + tokenValues[1] + "|" + tokenValues[2]
	result[1] = strings.Repeat("-", len(result[0]))
	result[2] = tokenValues[3] + "|" + tokenValues[4] + "|" + tokenValues[5]
	result[3] = result[1]
	result[4] = tokenValues[6] + "|" + tokenValues[7] + "|" + tokenValues[8]
	result[5] = ""
	result[6] = "Next player: " + s.Players[s.Game.CurrentPlayer].TokenValue

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
