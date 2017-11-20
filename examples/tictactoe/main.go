/*

Package tictactoe is an exceedingly simple game based on boardgame. It serves
as an example, and also helps verify that the design and implementation of
boardgame are useful for real games.

*/
package tictactoe

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"strings"
)

//go:generate autoreader

const gameDisplayname = "Tic Tac Toe"
const gameName = "tictactoe"

const DIM = 3

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {
	component := c.Values.(*playerToken)

	_, players := concreteStates(state)

	switch component.Value {
	case X:
		return players[0].UnusedTokens, nil
	case O:
		return players[1].UnusedTokens, nil
	}
	return nil, errors.New("Component with unexpected value")
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.State) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	return game.CurrentPlayer
}

func (g *gameDelegate) Name() string {
	return gameName
}

func (g *gameDelegate) DisplayName() string {
	return gameDisplayname
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 2
}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers == 2
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	tokens := g.Manager().Chest().Deck("tokens")

	if tokens == nil {
		return nil
	}

	//We want to set the sized stack to a certain value imperatively, so we'll
	//do it ourselves and not rely on tag-based auto-inflation.
	return &gameState{
		Slots: tokens.NewSizedStack(DIM * DIM),
	}
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurablePlayerState {

	value := X

	if playerIndex == 1 {
		value = O
	}

	return &playerState{
		TokensToPlaceThisTurn: 1,
		TokenValue:            value,
		playerIndex:           playerIndex,
	}
}

func (g *gameDelegate) CheckGameFinished(state boardgame.State) (finished bool, winners []boardgame.PlayerIndex) {

	game, players := concreteStates(state)

	tokens := make([]string, DIM*DIM)

	for i := 0; i < DIM*DIM; i++ {
		tokens[i] = game.tokenValueAtIndex(i)
	}

	finished, winner := checkGameFinished(tokens)

	if finished {

		if winner == Empty {
			//Draw
			return true, nil
		}

		var winningPlayer int

		for i, player := range players {
			if player.TokenValue == winner {
				winningPlayer = i
			}
		}

		return true, []boardgame.PlayerIndex{boardgame.PlayerIndex(winningPlayer)}
	}

	return false, nil

}

func (g *gameDelegate) Diagram(state boardgame.State) string {

	game, players := concreteStates(state)

	//Get an array of *playerTokenValues corresponding to tokens currently in
	//the stack.
	tokens := playerTokenValues(game.Slots.ComponentValues())

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
	result[6] = "Next player: " + players[game.CurrentPlayer].TokenValue

	return strings.Join(result, "\n")
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

func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) {
	chest := boardgame.NewComponentChest(Enums)

	tokens := boardgame.NewDeck()

	//How many tokens of each of x's and o's do we need so that no matter who
	//goes first we always have enough?
	numTokens := 5

	tokens.AddComponentMulti(&playerToken{
		Value: X,
	}, numTokens)

	tokens.AddComponentMulti(&playerToken{
		Value: O,
	}, numTokens)

	if err := chest.AddDeck("tokens", tokens); err != nil {
		return nil, errors.New("couldn't add deck: " + err.Error())
	}

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, storage)

	if manager == nil {
		return nil, errors.New("No manager returned")
	}

	bulkMoveTypeConfigs := []*boardgame.MoveTypeConfig{
		&movePlayTokenConfig,
		&moveFinishTurnConfig,
	}

	if err := manager.AddMoves(bulkMoveTypeConfigs...); err != nil {
		return nil, errors.New("Couldn't add moves: " + err.Error())
	}

	manager.AddAgent(&Agent{})

	if err := manager.SetUp(); err != nil {
		return nil, errors.New("Couldn't set up manager: " + err.Error())
	}

	return manager, nil
}

func NewGame(manager *boardgame.GameManager) *boardgame.Game {
	game := manager.NewGame()

	if err := game.SetUp(0, nil, nil); err != nil {
		panic(err)
	}

	return game
}
