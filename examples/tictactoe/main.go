/*

Package tictactoe is an exceedingly simple game based on boardgame. It serves
as an example, and also helps verify that the design and implementation of
boardgame are useful for real games.

*/
package tictactoe

import (
	"errors"
	"reflect"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

const dim = 3
const totalDim = dim * dim

type gameDelegate struct {
	base.GameDelegate
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	component := c.Values().(*playerToken)

	_, players := concreteStates(state)

	switch component.Value {
	case X:
		return players[0].UnusedTokens, nil
	case O:
		return players[1].UnusedTokens, nil
	}
	return nil, errors.New("Component with unexpected value")
}

var memoizedDelegateName string

func (g *gameDelegate) Name() string {

	//If our package name and delegate.Name() don't match, NewGameManager will
	//fail with an error. Given they have to be the same, we might as well
	//just ensure they are actually the same, via a one-time reflection.

	if memoizedDelegateName == "" {
		pkgPath := reflect.ValueOf(g).Elem().Type().PkgPath()
		pathPieces := strings.Split(pkgPath, "/")
		memoizedDelegateName = pathPieces[len(pathPieces)-1]
	}
	return memoizedDelegateName
}

func (g *gameDelegate) DisplayName() string {
	return "Tic Tac Toe"
}

func (g *gameDelegate) Description() string {
	return "A classic game where players place X's and O's and try to get three in a row"
}

func (g *gameDelegate) MinNumPlayers() int {
	return 2
}

func (g *gameDelegate) MaxNumPlayers() int {
	return 2
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 2
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return new(gameState)
}

func (g *gameDelegate) PlayerStateConstructor(playerIndex boardgame.PlayerIndex) boardgame.ConfigurableSubState {

	value := X

	if playerIndex == 1 {
		value = O
	}

	return &playerState{
		TokensToPlaceThisTurn: 1,
		TokenValue:            value,
	}
}

func (g *gameDelegate) CheckGameFinished(state boardgame.ImmutableState) (finished bool, winners []boardgame.PlayerIndex) {

	game, players := concreteStates(state)

	tokens := make([]string, totalDim)

	for i := 0; i < totalDim; i++ {
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

func (g *gameDelegate) ConfigureAgents() []boardgame.Agent {
	return []boardgame.Agent{
		&Agent{},
	}
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Add(
		auto.MustConfig(
			new(movePlaceToken),
			moves.WithHelpText("Place a player's token in a specific space."),
		),
		auto.MustConfig(
			new(moves.FinishTurn),
		),
	)
}

func (g *gameDelegate) ConfigureConstants() boardgame.PropertyCollection {
	return boardgame.PropertyCollection{
		"TOTAL_DIM": totalDim,
	}
}

func (g *gameDelegate) Diagram(state boardgame.ImmutableState) string {

	game, players := concreteStates(state)

	//Get an array of *playerTokenValues corresponding to tokens currently in
	//the stack.
	tokens := game.Slots.Components()

	tokenValues := make([]string, len(tokens))

	for i, token := range tokens {
		if token == nil {
			tokenValues[i] = " "
			continue
		}
		tokenValues[i] = token.Values().(*playerToken).Value
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

	if len(state) != totalDim {
		return false, Empty
	}

	//Check condition 1 (rows)

	for r := 0; r < dim; r++ {
		var run []string
		for c := 0; c < dim; c++ {
			run = append(run, state[rowColToIndex(r, c)])
		}
		result := checkRunWon(run)
		if result != Empty {
			return true, result
		}
	}

	//Check condition 2 (cols)

	for c := 0; c < dim; c++ {
		var run []string
		for r := 0; r < dim; r++ {
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

	for i := 0; i < dim; i++ {
		diagonalDown = append(diagonalDown, state[rowColToIndex(i, i)])
		diagonalUp = append(diagonalUp, state[rowColToIndex(dim-i-1, i)])
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

	if len(runState) != dim {
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

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {
	tokens := boardgame.NewDeck()

	//How many tokens of each of x's and o's do we need so that no matter who
	//goes first we always have enough?
	numTokens := 5

	for i := 0; i < numTokens; i++ {
		tokens.AddComponent(&playerToken{
			Value: X,
		})
	}

	for i := 0; i < numTokens; i++ {
		tokens.AddComponent(&playerToken{
			Value: O,
		})
	}
	return map[string]*boardgame.Deck{
		"tokens": tokens,
	}
}

//NewDelegate is the primary entrypoint for this package. It returns a
//GameDelegate that configures a game of pig.
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
