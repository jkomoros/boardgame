package boardgame

import (
	"errors"
	"github.com/jkomoros/boardgame/enum"
	"testing"
)

type testGameDelegate struct {
	DefaultGameDelegate
	//if this is higher than 0, then will craete this many extra comoponents
	extraComponentsToCreate int
	moveInstaller           func(manager *GameManager) []MoveConfig
}

func (t *testGameDelegate) ConfigureAgents() []Agent {
	return []Agent{
		&testAgent{},
	}
}

func (t *testGameDelegate) ConfigureEnums() *enum.Set {
	return testEnums
}

func (t *testGameDelegate) ConfigureConstants() PropertyCollection {
	return PropertyCollection{
		"ConstantStackSize": 4,
		"MyBool":            false,
	}
}

func (t *testGameDelegate) PhaseEnum() enum.Enum {
	return testPhaseEnum
}

func (t *testGameDelegate) ConfigureDecks() map[string]*Deck {
	deck := NewDeck()

	deck.AddComponent(&testingComponent{
		String:  "foo",
		Integer: 1,
	})

	deck.AddComponent(&testingComponent{
		String:  "bar",
		Integer: 2,
	})

	deck.AddComponent(&testingComponent{
		String:  "baz",
		Integer: 5,
	})

	deck.AddComponent(&testingComponent{
		String:  "slam",
		Integer: 10,
	})

	deck.AddComponent(&testingComponent{
		String:  "basic",
		Integer: 8,
	})

	for i := 0; i < t.extraComponentsToCreate; i++ {
		deck.AddComponent(&testingComponent{
			String:  "Extra",
			Integer: 8 + i,
		})
	}

	deck.SetGenericValues(&testShadowValues{
		Message: "Foo",
	})

	return map[string]*Deck{
		"test": deck,
	}

}

func (t *testGameDelegate) ConfigureMoves() []MoveConfig {
	return t.moveInstaller(t.Manager())
}

func (t *testGameDelegate) DistributeComponentToStarterStack(state ImmutableState, c Component) (ImmutableStack, error) {
	game, _ := concreteStates(state)
	return game.DrawDeck, nil
}

func (t *testGameDelegate) Name() string {
	return testGameName
}

func (t *testGameDelegate) ComputedGlobalProperties(state ImmutableState) PropertyCollection {
	_, playerStates := concreteStates(state)

	allScores := 0

	for _, player := range playerStates {

		allScores += player.Score
	}

	return PropertyCollection{
		"SumAllScores": allScores,
	}
}

func (t *testGameDelegate) ComputedPlayerProperties(player ImmutablePlayerState) PropertyCollection {

	playerState := player.(*testPlayerState)

	effectiveMovesLeftThisTurn := playerState.MovesLeftThisTurn

	//Players with Isfoo get a bonus.
	if playerState.IsFoo {
		effectiveMovesLeftThisTurn += 5
	}

	return PropertyCollection{
		"EffectiveMovesLeftThisTurn": effectiveMovesLeftThisTurn,
	}
}

func (t *testGameDelegate) DynamicComponentValuesConstructor(deck *Deck) ConfigurableSubState {
	if deck.Name() == "test" {
		return &testingComponentDynamic{
			Stack: deck.NewSizedStack(1),
			Enum:  testColorEnum.NewVal(),
		}
	}
	return nil
}

func (t *testGameDelegate) CheckGameFinished(state ImmutableState) (bool, []PlayerIndex) {
	_, players := concreteStates(state)

	var winners []PlayerIndex

	for i, player := range players {
		if player.Score >= 5 {
			//This user won!
			winners = append(winners, PlayerIndex(i))

			//Keep going through to see if anyone else won at the same time
		}
	}

	if len(winners) > 0 {
		return true, winners
	}

	return false, nil
}

func (t *testGameDelegate) DefaultNumPlayers() int {
	return 3
}

func (t *testGameDelegate) MinNumPlayers() int {
	return 1
}

func (t *testGameDelegate) MaxNumPlayers() int {
	return 5
}

func (t *testGameDelegate) Variants() map[string][]string {
	return map[string][]string{
		"color": {"blue", "red"},
	}
}

func (t *testGameDelegate) BeginSetUp(state State, variant Variant) error {
	game, players := concreteStates(state)

	if len(players) != 3 {
		return errors.New("Only three players are supported")
	}

	game.MyEnumValue.SetValue(colorGreen)

	players[0].MovesLeftThisTurn = 1
	players[2].IsFoo = true
	players[1].EnumVal.SetValue(colorGreen)
	return nil
}

func (t *testGameDelegate) FinishSetUp(state State) error {

	//Set all IntVar's to 1 for dynamic values for all hands. This will help
	//us verify when they are being sanitized.

	game, _ := concreteStates(state)

	for _, c := range game.DrawDeck.Components() {
		values := c.DynamicValues().(*testingComponentDynamic)

		values.IntVar = 1
		values.Enum.SetValue(colorBlue)
	}

	return game.DrawDeck.ComponentAt(game.DrawDeck.Len() - 1).MoveToFirstSlot(game.MyBoard.SpaceAt(1))
}

func (t *testGameDelegate) CurrentPlayerIndex(state ImmutableState) PlayerIndex {
	game, _ := concreteStates(state)

	return game.CurrentPlayer
}

func (t *testGameDelegate) GameStateConstructor() ConfigurableSubState {
	chest := t.Manager().Chest()

	deck := chest.Deck("test")
	return &testGameState{
		CurrentPlayer:      0,
		DrawDeck:           deck.NewStack(0),
		Timer:              NewTimer(),
		MyIntSlice:         make([]int, 0),
		MyBoolSlice:        make([]bool, 0),
		MyStringSlice:      make([]string, 0),
		MyPlayerIndexSlice: make([]PlayerIndex, 0),
		MyEnumValue:        testColorEnum.NewVal(),
		MyEnumConst:        testColorEnum.MustNewImmutableVal(colorBlue),
	}
}

func (t *testGameDelegate) PlayerStateConstructor(player PlayerIndex) ConfigurablePlayerState {
	return &testPlayerState{
		playerIndex: player,
	}
}

func TestTestGameDelegate(t *testing.T) {
	manager := newTestGameManger(t)

	if manager.Delegate().Name() != testGameName {
		t.Error("Manager.Name() was not overridden")
	}
}
