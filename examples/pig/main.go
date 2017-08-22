/*

pig is a very simple game involving dice rolls.

*/
package pig

import (
	"errors"
	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/components/dice"
	"math/rand"
	"strconv"
	"strings"
)

//go:generate autoreader

const DefaultTargetScore = 100
const diceDeckName = "dice"

type gameDelegate struct {
	boardgame.DefaultGameDelegate
}

func (g *gameDelegate) Name() string {
	return "pig"
}

func (g *gameDelegate) DisplayName() string {
	return "Pig"
}

func (g *gameDelegate) LegalNumPlayers(numPlayers int) bool {
	return numPlayers > 0 && numPlayers < 6
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 2
}

func (g *gameDelegate) CurrentPlayerIndex(state boardgame.State) boardgame.PlayerIndex {
	game, _ := concreteStates(state)
	return game.CurrentPlayer
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.State, c *boardgame.Component) (boardgame.Stack, error) {
	game, _ := concreteStates(state)
	return game.Die, nil
}

func (g *gameDelegate) FinishSetUp(state boardgame.MutableState) {

	game, _ := concreteStates(state)

	//Pick a player to start randomly.
	startingPlayer := boardgame.PlayerIndex(rand.Intn(len(state.PlayerStates())))

	game.CurrentPlayer = startingPlayer
	game.TargetScore = DefaultTargetScore

}

func (g *gameDelegate) CheckGameFinished(state boardgame.State) (finished bool, winners []boardgame.PlayerIndex) {
	game, players := concreteStates(state)

	for i, player := range players {
		if player.TotalScore >= game.TargetScore {
			winners = append(winners, boardgame.PlayerIndex(i))
		}
	}

	if len(winners) > 0 {
		return true, winners
	}

	return false, nil
}

func (g *gameDelegate) Diagram(state boardgame.State) string {
	var parts []string

	game, players := concreteStates(state)

	dieValue := game.Die.ComponentAt(0).DynamicValues(state).(*dice.DynamicValue).Value

	parts = append(parts, "Die: "+strconv.Itoa(dieValue))

	parts = append(parts, "\nPlayers")

	for i, player := range players {
		parts = append(parts, "Player "+strconv.Itoa(i)+": "+strconv.Itoa(player.RoundScore)+", "+strconv.Itoa(player.TotalScore))
	}

	return strings.Join(parts, "\n")
}

func (g *gameDelegate) GameStateConstructor() boardgame.MutableSubState {
	return &gameState{
		CurrentPlayer: 0,
		TargetScore:   DefaultTargetScore,
	}
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.MutablePlayerState {
	return &playerState{
		playerIndex: index,
	}
}

func (g *gameDelegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.MutableSubState {
	if deck.Name() == diceDeckName {
		return &dice.DynamicValue{
			Value: 1,
		}
	}
	return nil
}

func MustNewManager(storage boardgame.StorageManager) *boardgame.GameManager {

	manager, err := NewManager(storage)

	if err != nil {
		panic("Couldn't create manager: " + err.Error())
	}

	return manager

}

func NewManager(storage boardgame.StorageManager) (*boardgame.GameManager, error) {
	chest := boardgame.NewComponentChest(nil)

	diceDeck := boardgame.NewDeck()

	diceDeck.AddComponent(dice.DefaultDie())

	if err := chest.AddDeck(diceDeckName, diceDeck); err != nil {
		return nil, errors.New("Couldn't add deck: " + err.Error())
	}

	manager := boardgame.NewGameManager(&gameDelegate{}, chest, storage)

	if manager == nil {
		return nil, errors.New("No manager returned")
	}

	moveTypeConfigs := []*boardgame.MoveTypeConfig{
		&moveRollDiceConfig,
		&moveDoneTurnConfig,
		&moveCountDieConfig,
		&moveFinishTurnConfig,
	}

	if err := manager.BulkAddMoveTypes(moveTypeConfigs); err != nil {
		return nil, errors.New("couldnt add move types: " + err.Error())
	}

	if err := manager.SetUp(); err != nil {
		return nil, errors.New("Couldn't set up manager: " + err.Error())
	}

	return manager, nil
}
