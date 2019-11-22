/*

Package pig is a very simple game involving dice rolls.

*/
package pig

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/base"
	"github.com/jkomoros/boardgame/components/dice"
	"github.com/jkomoros/boardgame/moves"
)

//go:generate boardgame-util codegen

const defaultTargetScore = 100
const diceDeckName = "dice"

type gameDelegate struct {
	base.GameDelegate
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

func (g *gameDelegate) Description() string {
	return "Players roll the dice, collecting points, but bust if they roll a one."
}

func (g *gameDelegate) MinNumPlayers() int {
	return 2
}

func (g *gameDelegate) MaxNumPlayers() int {
	return 6
}

func (g *gameDelegate) DefaultNumPlayers() int {
	return 2
}

func (g *gameDelegate) DistributeComponentToStarterStack(state boardgame.ImmutableState, c boardgame.Component) (boardgame.ImmutableStack, error) {
	game, _ := concreteStates(state)
	return game.Die, nil
}

func (g *gameDelegate) FinishSetUp(state boardgame.State) error {

	game, _ := concreteStates(state)

	//Pick a player to start randomly.
	startingPlayer := boardgame.PlayerIndex(state.Rand().Intn(len(state.PlayerStates())))

	game.CurrentPlayer = startingPlayer
	game.TargetScore = defaultTargetScore

	return nil

}

func (g *gameDelegate) GameEndConditionMet(state boardgame.ImmutableState) bool {
	game, players := concreteStates(state)

	for _, player := range players {
		if player.TotalScore >= game.TargetScore {
			return true
		}
	}

	return false
}

func (g *gameDelegate) PlayerScore(pState boardgame.ImmutableSubState) int {
	return pState.(*playerState).TotalScore
}

func (g *gameDelegate) Diagram(state boardgame.ImmutableState) string {
	var parts []string

	game, players := concreteStates(state)

	dieValue := game.Die.ComponentAt(0).DynamicValues().(*dice.DynamicValue).Value

	parts = append(parts, "Die: "+strconv.Itoa(dieValue))

	parts = append(parts, "\nPlayers")

	for i, player := range players {
		parts = append(parts, "Player "+strconv.Itoa(i)+": "+strconv.Itoa(player.RoundScore)+", "+strconv.Itoa(player.TotalScore))
	}

	return strings.Join(parts, "\n")
}

func (g *gameDelegate) GameStateConstructor() boardgame.ConfigurableSubState {
	return &gameState{
		TargetScore: defaultTargetScore,
	}
}

func (g *gameDelegate) PlayerStateConstructor(index boardgame.PlayerIndex) boardgame.ConfigurableSubState {
	return new(playerState)
}

func (g *gameDelegate) DynamicComponentValuesConstructor(deck *boardgame.Deck) boardgame.ConfigurableSubState {
	if deck.Name() == diceDeckName {
		return &dice.DynamicValue{
			Value: 1,
		}
	}
	return nil
}

func (g *gameDelegate) ConfigureMoves() []boardgame.MoveConfig {

	auto := moves.NewAutoConfigurer(g)

	return moves.Add(
		auto.MustConfig(
			new(moveRollDice),
			moves.WithHelpText("Rolls the dice for the current player"),
		),
		auto.MustConfig(
			new(moveDoneTurn),
			moves.WithHelpText("Played when a player is done with their turn and wants to keep their score."),
		),
		auto.MustConfig(
			new(moveCountDie),
			moves.WithHelpText("After a die has been rolled, tabulating its impact"),
			moves.WithIsFixUp(true),
		),
		auto.MustConfig(
			new(moves.FinishTurn),
			moves.WithHelpText("Advance to the next player when the current player has busted or said they are done."),
		),
	)
}

func (g *gameDelegate) ConfigureDecks() map[string]*boardgame.Deck {

	diceDeck := boardgame.NewDeck()
	diceDeck.AddComponent(dice.DefaultDie())

	return map[string]*boardgame.Deck{
		diceDeckName: diceDeck,
	}
}

//NewDelegate is the primary entrypoint of the package. It returns a delegate
//that configures a game of pig.
func NewDelegate() boardgame.GameDelegate {
	return &gameDelegate{}
}
