/*

Package tictactoe is an exceedingly simple game based on boardgame. It serves
as an example, and also helps verify that the design and implementation of
boardgame are useful for real games.

*/
package tictactoe

import (
	"github.com/jkomoros/boardgame"
)

const gameName = "tictactoe"

const DIM = 3

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) DistributeComponentToStarterStack(payload boardgame.StatePayload, c *boardgame.Component) error {
	component := c.Values.(*playerToken)

	p := payload.(*statePayload)

	switch component.Value {
	case X:
		p.users[0].UnusedTokens.InsertFront(c)
	case O:
		p.users[1].UnusedTokens.InsertFront(c)
	}
	return nil
}

func Renderer(state boardgame.StatePayload) []string {
	return []string{"TODO: actually do something here"}
}

func NewGame() *boardgame.Game {

	chest := boardgame.NewComponentChest(gameName)

	tokens := &boardgame.Deck{}

	//How many tokens of each of x's and o's do we need so that no matter who
	//goes first we always have enough?
	numTokens := 5

	//TODO: use deck.AddComponentMulti when that exists
	for i := 0; i < numTokens; i++ {
		tokens.AddComponent(&boardgame.Component{
			Values: &playerToken{
				Value: X,
			},
		})
	}

	for i := 0; i < numTokens; i++ {
		tokens.AddComponent(&boardgame.Component{
			Values: &playerToken{
				Value: O,
			},
		})
	}

	chest.AddDeck("tokens", tokens)

	starterPayload := &statePayload{
		game: &gameState{
			Slots: boardgame.NewSizedStack(tokens, DIM*DIM),
		},
		users: []*userState{
			&userState{
				TokenValue:   X,
				UnusedTokens: boardgame.NewGrowableStack(tokens, 0),
			},
			&userState{
				TokenValue:   O,
				UnusedTokens: boardgame.NewGrowableStack(tokens, 0),
			},
		},
	}

	for i, user := range starterPayload.users {
		user.playerIndex = i
	}

	game := &boardgame.Game{
		Name:     gameName,
		State:    boardgame.NewStarterState(starterPayload),
		Delegate: &gameDelegate{},
	}

	game.SetChest(chest)

	if err := game.SetUp(); err != nil {
		panic("Game couldn't be set up")
	}

	return game

}
