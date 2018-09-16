/*

Package tictactoe is an exceedingly simple game based on boardgame. It serves
as an example, and also helps verify that the design and implementation of
boardgame are useful for real games.

*/
package tictactoe

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/moves"
	"strings"
)

//go:generate boardgame-util codegen

const gameDisplayname = "Tic Tac Toe"
const gameName = "tictactoe"

const DIM = 3
const TOTAL_DIM = DIM * DIM

type gameDelegate struct {
	boardgame.DefaultGameDelegate
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

func (g *gameDelegate) Name() string {
	return gameName
}

func (g *gameDelegate) DisplayName() string {
	return gameDisplayname
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

func (g *gameDelegate) CheckGameFinished(state boardgame.ImmutableState) (finished bool, winners []boardgame.PlayerIndex) {

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

func (g *gameDelegate) ConfigureAgents() []boardgame.Agent {
	return []boardgame.Agent{
		&Agent{},
	}
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Add(
		auto.MustConfig(
			new(MovePlaceToken),
			moves.WithHelpText("Place a player's token in a specific space."),
		),
		auto.MustConfig(
			new(moves.FinishTurn),
		),
	)
}

func (g *gameDelegate) ConfigureConstants() boardgame.PropertyCollection {
	return boardgame.PropertyCollection{
		"TOTAL_DIM": TOTAL_DIM,
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

func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
