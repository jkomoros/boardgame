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

func (g *gameState) JSON() boardgame.JSONObject {
	return g
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

func (u *userState) JSON() boardgame.JSONObject {
	return u
}

func (u *userState) PlayerIndex() int {
	return u.playerIndex
}

type statePayload struct {
	game  *gameState
	users []*userState
}

func (s *statePayload) userFromTokenValue(value string) *userState {
	for _, user := range s.users {
		if user.TokenValue == value {
			return user
		}
	}
	return nil
}

func (s *statePayload) Diagram() string {

	//Get an array of *playerTokenValues corresponding to tokens currently in
	//the stack.
	tokens := playerTokenValues(s.game.Slots.ComponentValues())

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
	result[6] = "Next player: " + s.users[s.game.CurrentPlayer].TokenValue

	return strings.Join(result, "\n")

}

func (s *statePayload) Game() boardgame.GameState {
	return s.game
}

func (s *statePayload) Users() []boardgame.UserState {
	array := make([]boardgame.UserState, len(s.users))

	for i := 0; i < len(s.users); i++ {
		array[i] = s.users[i]
	}

	return array
}

func (s *statePayload) JSON() boardgame.JSONObject {

	array := make([]boardgame.JSONObject, len(s.users))

	for i, user := range s.users {
		array[i] = user.JSON()
	}

	return boardgame.JSONMap{
		"Game":  s.game.JSON(),
		"Users": array,
	}
}

func (s *statePayload) Copy() boardgame.StatePayload {
	array := make([]*userState, len(s.users))

	for i := 0; i < len(s.users); i++ {
		array[i] = s.users[i].Copy().(*userState)
	}

	return &statePayload{
		game:  s.game.Copy().(*gameState),
		users: array,
	}
}
