package boardgame

import (
	"testing"
)

type testGameDelegate struct {
	DefaultGameDelegate
}

func (t *testGameDelegate) DistributeComponentToStarterStack(state State, c *Component) (Stack, error) {
	game, _ := concreteStates(state)
	return game.DrawDeck, nil
}

func (t *testGameDelegate) Name() string {
	return testGameName
}

func (t *testGameDelegate) ComputedGlobalProperties(state State) PropertyCollection {
	_, playerStates := concreteStates(state)

	allScores := 0

	for _, player := range playerStates {

		allScores += player.Score
	}

	return PropertyCollection{
		"SumAllScores": allScores,
	}
}

func (t *testGameDelegate) ComputedPlayerProperties(player PlayerState) PropertyCollection {

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
			Enum:  testColorEnum.NewMutableVal(),
		}
	}
	return nil
}

func (t *testGameDelegate) CheckGameFinished(state State) (bool, []PlayerIndex) {
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

func (t *testGameDelegate) LegalNumPlayers(numPlayers int) bool {
	//We don't do a lower bound check specifically to test that SetUp rejects
	//numbers of players less than 1.
	return numPlayers <= 5
}

func (t *testGameDelegate) Configs() map[string][]string {
	return map[string][]string{
		"color": {"blue", "red"},
	}
}

func (t *testGameDelegate) BeginSetUp(state MutableState, config GameConfig) {
	game, players := concreteStates(state)

	if len(players) != 3 {
		return
	}

	game.MyEnumValue.SetValue(colorGreen)

	players[0].MovesLeftThisTurn = 1
	players[2].IsFoo = true
	players[1].EnumVal.SetValue(colorGreen)
}

func (t *testGameDelegate) FinishSetUp(state MutableState) {

	//Set all IntVar's to 1 for dynamic values for all hands. This will help
	//us verify when they are being sanitized.

	game, _ := concreteStates(state)

	for _, c := range game.DrawDeck.Components() {
		values := c.DynamicValues(state).(*testingComponentDynamic)

		values.IntVar = 1
		values.Enum.SetValue(colorBlue)
	}
}

func (t *testGameDelegate) CurrentPlayerIndex(state State) PlayerIndex {
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
		MyEnumValue:        testColorEnum.NewMutableVal(),
		MyEnumConst:        testColorEnum.MustNewVal(colorBlue),
	}
}

func (t *testGameDelegate) PlayerStateConstructor(player PlayerIndex) ConfigurablePlayerState {
	return &testPlayerState{
		playerIndex: player,
	}
}

func TestTestGameDelegate(t *testing.T) {
	manager := newTestGameManger()

	if manager.Delegate().Name() != testGameName {
		t.Error("Manager.Name() was not overridden")
	}
}
