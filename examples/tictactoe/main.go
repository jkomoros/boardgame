/*

Package tictactoe is an exceedingly simple game based on boardgame. It serves
as an example, and also helps verify that the design and implementation of
boardgame are useful for real games.

*/
package tictactoe

import (
	"encoding/json"
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
		p.Players[0].UnusedTokens.InsertFront(c)
	case O:
		p.Players[1].UnusedTokens.InsertFront(c)
	}
	return nil
}

func (g *gameDelegate) Name() string {
	return gameName
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 2
}

func (g *gameDelegate) StateFromBlob(blob []byte, schema int) (boardgame.State, error) {
	result := &mainState{}
	if err := json.Unmarshal(blob, result); err != nil {
		return nil, err
	}

	result.Game.Slots.Inflate(g.Manager().Chest())

	for i, player := range result.Players {
		player.playerIndex = i
		player.UnusedTokens.Inflate(g.Manager().Chest())
	}

	return result, nil
}

func (g *gameDelegate) StartingState(numUsers int) boardgame.State {

	if numUsers != 2 {
		return nil
	}

	tokens := g.Manager().Chest().Deck("tokens")

	if tokens == nil {
		return nil
	}

	result := &mainState{
		Game: &gameState{
			Slots: boardgame.NewSizedStack(tokens, DIM*DIM),
		},
		Players: []*playerState{
			&playerState{
				TokensToPlaceThisTurn: 1,
				TokenValue:            X,
				UnusedTokens:          boardgame.NewGrowableStack(tokens, 0),
			},
			&playerState{
				TokensToPlaceThisTurn: 0,
				TokenValue:            O,
				UnusedTokens:          boardgame.NewGrowableStack(tokens, 0),
			},
		},
	}

	for i, player := range result.Players {
		player.playerIndex = i
	}

	return result
}

func (g *gameDelegate) CheckGameFinished(state boardgame.State) (finished bool, winners []int) {

	s := state.(*mainState)

	tokens := make([]string, DIM*DIM)

	for i := 0; i < DIM*DIM; i++ {
		tokens[i] = s.Game.tokenValueAtIndex(i)
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

func NewManager(optionalStorage boardgame.StorageManager) *boardgame.GameManager {
	chest := boardgame.NewComponentChest()

	tokens := &boardgame.Deck{}

	//How many tokens of each of x's and o's do we need so that no matter who
	//goes first we always have enough?
	numTokens := 5

	tokens.AddComponentMulti(&playerToken{
		Value: X,
	}, numTokens)

	tokens.AddComponentMulti(&playerToken{
		Value: O,
	}, numTokens)

	chest.AddDeck("tokens", tokens)

	if optionalStorage == nil {
		optionalStorage = boardgame.NewInMemoryStorageManager()
	}

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, optionalStorage)

	manager.AddPlayerMove(&MovePlaceToken{})
	manager.AddFixUpMove(&MoveAdvancePlayer{})

	manager.SetUp()

	return manager
}

func NewGame(manager *boardgame.GameManager) *boardgame.Game {
	game := boardgame.NewGame(manager)

	if err := game.SetUp(0); err != nil {
		panic(err)
	}

	return game
}
