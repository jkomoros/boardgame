/*

Package tictactoe is an exceedingly simple game based on boardgame. It serves
as an example, and also helps verify that the design and implementation of
boardgame are useful for real games.

*/
package tictactoe

import (
	"github.com/jkomoros/boardgame"
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

func (g *gameDelegate) CheckGameFinished(state boardgame.StatePayload) (finished bool, winners []int) {

	//TODO: test this!

	/*The following are win conditions:

	* 1) For each row, check to see if the entire row across is same token value
	* 2) For each col, check if the entire col down shares same token value
	* 3) If the top left cell has a diagonal down to the bottom right with all same token value
	* 4) If the top righ cell has a diagonal down to the bottom left with all the same token value
	* 5) If all slots are filled but none of the other win conditions are true it's a draw.

	 */

	s := state.(*statePayload)

	//TODO: all of this feels like it could be totally refactored...

	//Check condition 1

	for r := 0; r < DIM; r++ {
		ok := true
		tokenValue := s.game.tokenValue(r, 0)
		if tokenValue == " " {
			continue
		}
		for c := 1; c < DIM; c++ {
			if s.game.tokenValue(r, c) != tokenValue {
				ok = false
				break
			}
		}
		if ok {
			//Found it!
			return true, []int{s.userFromTokenValue(tokenValue).playerIndex}
		}
	}

	//Check condition 2
	for c := 0; c < DIM; c++ {
		ok := true
		tokenValue := s.game.tokenValue(0, c)

		if tokenValue == " " {
			continue
		}

		for r := 1; r < DIM; r++ {
			if s.game.tokenValue(r, c) != tokenValue {
				ok = false
				break
			}
		}
		if ok {
			//Found it!
			return true, []int{s.userFromTokenValue(tokenValue).playerIndex}
		}
	}

	//Check condition 3

	ok := true
	tokenValue := s.game.tokenValue(0, 0)
	if tokenValue != " " {
		for i := 1; i < DIM; i++ {
			if s.game.tokenValue(i, i) != tokenValue {
				ok = false
				break
			}
		}
		if ok {
			return true, []int{s.userFromTokenValue(tokenValue).playerIndex}
		}
	}

	//Check condition 4

	ok = true
	tokenValue = s.game.tokenValue(DIM, 0)

	if tokenValue != " " {
		for i := 1; i < DIM; i++ {
			if s.game.tokenValue(DIM-i, i) != tokenValue {
				ok = false
				break
			}
		}
		if ok {
			return true, []int{s.userFromTokenValue(tokenValue).playerIndex}
		}
	}

	unfilledCells := 0
	for _, component := range s.game.Slots.ComponentValues() {
		if component == nil {
			unfilledCells++
		}
	}

	if unfilledCells == 0 {
		//All cells are filled and no one is a winner--draw!
		return true, nil
	}

	return false, nil

}

func (g *gameDelegate) ProposeFixUpMove(state boardgame.StatePayload) boardgame.Move {

	//TODO: when there's a concept of FixUp moves, the default delegate will
	//probably do what I want.

	move := g.Game.MoveByName("Advance Player")

	if move == nil {
		panic("Couldn't find advance player move")
	}

	moveToMake := move.Copy()

	//Advance Player only returns Legal if it makes sense to apply right now

	if err := moveToMake.Legal(state); err == nil {
		return moveToMake
	}

	return nil

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
