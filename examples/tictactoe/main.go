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

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) error {
	component := c.Values.(*playerToken)

	p := state.(*mainState)

	switch component.Value {
	case X:
		p.users[0].UnusedTokens.InsertFront(c)
	case O:
		p.users[1].UnusedTokens.InsertFront(c)
	}
	return nil
}

func (g *gameDelegate) CheckGameFinished(state boardgame.State) (finished bool, winners []int) {

	s := state.(*mainState)

	tokens := make([]string, DIM*DIM)

	for i := 0; i < DIM*DIM; i++ {
		tokens[i] = s.game.tokenValueAtIndex(i)
	}

	finished, winner := checkGameFinished(tokens)

	if finished {

		if winner == Empty {
			//Draw
			return true, nil
		}
		return true, []int{s.userFromTokenValue(winner).playerIndex}
	}

	return false, nil

}

//state should be a DIM * DIM length string, of the form "XXO XO  O". Winner
//will be of the form "X" or "O".
func checkGameFinished(state []string) (finished bool, winner string) {
	/*The following are win conditions:

	* 1) For each row, check to see if the entire row across is same token value
	* 2) For each col, check if the entire col down shares same token value
	* 3) If the top left cell has a diagonal down to the bottom right with all same token value
	* 4) If the top righ cell has a diagonal down to the bottom left with all the same token value
	* 5) If all slots are filled but none of the other win conditions are true it's a draw.

	 */

	if len(state) != DIM*DIM {
		return false, Empty
	}

	//Check condition 1 (rows)

	for r := 0; r < DIM; r++ {
		var run []string
		for c := 0; c < DIM; c++ {
			run = append(run, state[rowColToIndex(r, c)])
		}
		result := checkRunWon(run)
		if result != Empty {
			return true, result
		}
	}

	//Check condition 2 (cols)

	for c := 0; c < DIM; c++ {
		var run []string
		for r := 0; r < DIM; r++ {
			run = append(run, state[rowColToIndex(r, c)])
		}
		result := checkRunWon(run)
		if result != Empty {
			return true, result
		}
	}

	//Check condition 3 and 4

	var diagonalDown []string
	var diagonalUp []string

	for i := 0; i < DIM; i++ {
		diagonalDown = append(diagonalDown, state[rowColToIndex(i, i)])
		diagonalUp = append(diagonalUp, state[rowColToIndex(DIM-i-1, i)])
	}

	result := checkRunWon(diagonalDown)
	if result != Empty {
		return true, result
	}

	result = checkRunWon(diagonalUp)

	if result != Empty {
		return true, result
	}

	//Check condition 5 (draw)

	for _, token := range state {
		if token == Empty {
			//We found at least one slot that wasn't filled, so the game can't be a draw.
			return false, Empty
		}
	}
	//If we get to here, then every slot is filled but no one is winner, so it's a draw.
	return true, Empty
}

//runState should be a string of length DIM, where empty spaces are
//represented by "", which represents one "run" in the state. The winner will
//be "X", "O", or "" if no winner in this run.
func checkRunWon(runState []string) string {

	if len(runState) != DIM {
		return Empty
	}

	targetToken := runState[0]

	if targetToken == Empty {
		return Empty
	}

	for _, token := range runState {
		if token != targetToken {
			return Empty
		}
	}

	return targetToken
}

func (g *gameDelegate) ProposeFixUpMove(state boardgame.State) boardgame.Move {

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

	starterState := &mainState{
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

	for i, user := range starterState.users {
		user.playerIndex = i
	}

	game := &boardgame.Game{
		Name:         gameName,
		StateWrapper: boardgame.NewStarterStateWrapper(starterState),
		Delegate:     &gameDelegate{},
	}

	game.SetChest(chest)

	game.AddMove(&MovePlaceToken{})
	game.AddMove(&MoveAdvancePlayer{})

	if err := game.SetUp(); err != nil {
		panic("Game couldn't be set up")
	}

	return game

}
