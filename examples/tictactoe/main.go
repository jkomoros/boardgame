/*

Package tictactoe is an exceedingly simple game based on boardgame. It serves
as an example, and also helps verify that the design and implementation of
boardgame are useful for real games.

*/
package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

const (
	X rune = 'X'
	O rune = 'O'
)

const gameName = "tictactoe"

type playerToken struct {
	Value rune
}

func (p *playerToken) Props() []string {
	return boardgame.PropertyReaderPropsImpl(p)
}

func (p *playerToken) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(p, name)
}

type gameState struct {
	Slots boardgame.SizedStack
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

func ticTacToeGame() *boardgame.Game {

	chest := boardgame.NewComponentChest(gameName)

	xes := &boardgame.Deck{}

	//How many tokens of each of x's and o's do we need so that no matter who
	//goes first we always have enough?
	numTokens := 5

	//TODO: use deck.AddComponentMulti when that exists
	for i := 0; i < numTokens; i++ {
		xes.AddComponent(&boardgame.Component{
			Values: &playerToken{
				Value: X,
			},
		})
	}

	chest.AddDeck("xes", xes)

	oes := &boardgame.Deck{}

	for i := 0; i < numTokens; i++ {
		oes.AddComponent(&boardgame.Component{
			Values: &playerToken{
				Value: O,
			},
		})
	}

	chest.AddDeck("oes", oes)

	game := &boardgame.Game{
		Name: gameName,
		//TODO: state
	}

	game.SetChest(chest)

	return game

}
