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

func (g *gameState) Props() []string {
	return boardgame.PropertyReaderPropsImpl(g)
}

func (g *gameState) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(g, name)
}

func (g *gameState) Copy() boardgame.GameState {
	var result gameState
	result = *g
	return &result
}

type userState struct {
	playerIndex  int
	TokenValue   string
	UnusedTokens *boardgame.GrowableStack
	//How many tokens they have left to place this turn.
	TokensToPlaceThisTurn int
}

func (u *userState) Props() []string {
	return boardgame.PropertyReaderPropsImpl(u)
}

func (u *userState) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(u, name)
}

func (u *userState) Copy() boardgame.UserState {
	var result userState
	result = *u
	return &result
}

func (u *userState) PlayerIndex() int {
	return u.playerIndex
}

type mainState struct {
	Game  *gameState
	Users []*userState
}

func (s *mainState) userFromTokenValue(value string) *userState {
	for _, user := range s.Users {
		if user.TokenValue == value {
			return user
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
	result[6] = "Next player: " + s.Users[s.Game.CurrentPlayer].TokenValue

	return strings.Join(result, "\n")

}

func (s *mainState) GameState() boardgame.GameState {
	return s.Game
}

func (s *mainState) UserStates() []boardgame.UserState {
	array := make([]boardgame.UserState, len(s.Users))

	for i := 0; i < len(s.Users); i++ {
		array[i] = s.Users[i]
	}

	return array
}

func (s *mainState) Copy() boardgame.State {
	array := make([]*userState, len(s.Users))

	for i := 0; i < len(s.Users); i++ {
		array[i] = s.Users[i].Copy().(*userState)
	}

	return &mainState{
		Game:  s.Game.Copy().(*gameState),
		Users: array,
	}
}
