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
