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

type statePayload struct {
	game *gameState
	//we have no user state because all state is public.
}

func (s *statePayload) Game() boardgame.GameState {
	return s.game
}

func (s *statePayload) Users() []boardgame.UserState {
	return nil
}

func (s *statePayload) JSON() boardgame.JSONObject {
	return boardgame.JSONMap{
		"Game": s.game.JSON(),
	}
}

func (s *statePayload) Copy() boardgame.StatePayload {
	return &statePayload{
		game: s.game.Copy().(*gameState),
	}
}
