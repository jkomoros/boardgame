package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

type gameState struct {
	Slots *boardgame.SizedStack
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
	UnusedTokens *boardgame.GrowableStack
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
