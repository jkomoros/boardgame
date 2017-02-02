/*

Package tictactoe is an exceedingly simple game based on boardgame. It serves
as an example, and also helps verify that the design and implementation of
boardgame are useful for real games.

*/
package tictactoe

import (
	"github.com/jkomoros/boardgame"
	"strings"
)

const gameName = "Tic Tac Toe"

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

func Renderer(state boardgame.StatePayload) string {
	p := state.(*statePayload)

	//Get an array of *playerTokenValues corresponding to tokens currently in
	//the stack.
	tokens := playerTokenValues(p.game.Slots.ComponentValues())

	tokenValues := make([]string, len(tokens))

	for i, token := range tokens {
		if token == nil {
			tokenValues[i] = "  "
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
	result[6] = "Next player: " + p.users[p.game.CurrentPlayer].TokenValue

	return strings.Join(result, "\n")

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
				TokensToPlaceThisTurn: 1,
				TokenValue:            X,
				UnusedTokens:          boardgame.NewGrowableStack(tokens, 0),
			},
			&userState{
				TokensToPlaceThisTurn: 0,
				TokenValue:            O,
				UnusedTokens:          boardgame.NewGrowableStack(tokens, 0),
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

	game.AddMove(&MovePlaceToken{})
	game.AddMove(&MoveAdvancePlayer{})

	if err := game.SetUp(); err != nil {
		panic("Game couldn't be set up")
	}

	return game

}
