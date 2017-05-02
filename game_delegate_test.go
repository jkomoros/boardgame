package boardgame

import (
	"testing"
)

var testPropertiesConfig *ComputedPropertiesConfig

type testGlobalPropertiesCollection struct {
	SumAllScores int
}

type testPlayerPropertiesCollection struct {
	EffectiveMovesLeftThisTurn int
}

func (t *testGlobalPropertiesCollection) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
}

func (t *testGlobalPropertiesCollection) Reader() PropertyReader {
	return DefaultReader(t)
}

func (t *testPlayerPropertiesCollection) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(t)
}

func (t *testPlayerPropertiesCollection) Reader() PropertyReader {
	return DefaultReader(t)
}

func init() {
	testPropertiesConfig = &ComputedPropertiesConfig{
		Global: map[string]ComputedGlobalPropertyDefinition{
			"SumAllScores": ComputedGlobalPropertyDefinition{
				Dependencies: []StatePropertyRef{
					{
						Group:    StateGroupPlayer,
						PropName: "Score",
					},
				},
				PropType: TypeInt,
				Compute: func(state State) (interface{}, error) {

					_, playerStates := concreteStates(state)

					result := 0

					for _, player := range playerStates {

						result += player.Score
					}
					return result, nil
				},
			},
		},
		Player: map[string]ComputedPlayerPropertyDefinition{
			"EffectiveMovesLeftThisTurn": ComputedPlayerPropertyDefinition{
				Dependencies: []StatePropertyRef{
					{
						Group:    StateGroupPlayer,
						PropName: "MovesLeftThisTurn",
					},
					{
						Group:    StateGroupPlayer,
						PropName: "IsFoo",
					},
				},
				PropType: TypeInt,
				Compute: func(state PlayerState) (interface{}, error) {

					playerState := state.(*testPlayerState)

					effectiveMovesLeftThisTurn := playerState.MovesLeftThisTurn

					//Players with Isfoo get a bonus.
					if playerState.IsFoo {
						effectiveMovesLeftThisTurn += 5
					}

					return effectiveMovesLeftThisTurn, nil
				},
			},
		},
	}
}

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

func (t *testGameDelegate) ComputedPropertiesConfig() *ComputedPropertiesConfig {
	return testPropertiesConfig
}

func (t *testGameDelegate) EmptyDynamicComponentValues(deck *Deck) MutableSubState {
	if deck.Name() == "test" {
		return &testingComponentDynamic{
			Stack: NewSizedStack(deck, 1),
		}
	}
	return nil
}

func (t *testGameDelegate) EmptyComputedGlobalPropertyCollection() MutableSubState {
	return &testGlobalPropertiesCollection{}
}

func (t *testGameDelegate) EmptyComputedPlayerPropertyCollection() MutableSubState {
	return &testPlayerPropertiesCollection{}
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

func (t *testGameDelegate) BeginSetUp(state MutableState) {
	_, players := concreteStates(state)

	if len(players) != 3 {
		return
	}

	players[0].MovesLeftThisTurn = 1
	players[2].IsFoo = true
}

func (t *testGameDelegate) FinishSetUp(state MutableState) {

	//Set all IntVar's to 1 for dynamic values for all hands. This will help
	//us verify when they are being sanitized.

	game, _ := concreteStates(state)

	for _, c := range game.DrawDeck.Components() {
		values := c.DynamicValues(state).(*testingComponentDynamic)

		values.IntVar = 1
	}
}

func (t *testGameDelegate) CurrentPlayerIndex(state State) PlayerIndex {
	game, _ := concreteStates(state)

	return game.CurrentPlayer
}

func (t *testGameDelegate) EmptyGameState() MutableSubState {
	chest := t.Manager().Chest()

	deck := chest.Deck("test")
	return &testGameState{
		CurrentPlayer:      0,
		DrawDeck:           NewGrowableStack(deck, 0),
		Timer:              NewTimer(),
		MyIntSlice:         make([]int, 0),
		MyBoolSlice:        make([]bool, 0),
		MyStringSlice:      make([]string, 0),
		MyPlayerIndexSlice: make([]PlayerIndex, 0),
	}
}

func (t *testGameDelegate) EmptyPlayerState(player PlayerIndex) MutablePlayerState {
	chest := t.Manager().Chest()

	deck := chest.Deck("test")

	return &testPlayerState{
		playerIndex:       player,
		Score:             0,
		MovesLeftThisTurn: 0,
		Hand:              NewSizedStack(deck, 2),
		IsFoo:             false,
	}
}

func TestTestGameDelegate(t *testing.T) {
	manager := newTestGameManger()

	if manager.Delegate().Name() != testGameName {
		t.Error("Manager.Name() was not overridden")
	}
}
