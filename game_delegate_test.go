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

func (t *testGlobalPropertiesCollection) Reader() PropertyReadSetter {
	return DefaultReadSetter(t)
}

func (t *testPlayerPropertiesCollection) Reader() PropertyReadSetter {
	return DefaultReadSetter(t)
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
				Compute: func(state *State) (interface{}, error) {

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

func (t *testGameDelegate) DistributeComponentToStarterStack(state *State, c *Component) (Stack, error) {
	game, _ := concreteStates(state)
	return game.DrawDeck, nil
}

func (t *testGameDelegate) Name() string {
	return testGameName
}

func (t *testGameDelegate) ComputedPropertiesConfig() *ComputedPropertiesConfig {
	return testPropertiesConfig
}

func (t *testGameDelegate) EmptyComputedGlobalPropertyCollection() ComputedPropertyCollection {
	return &testGlobalPropertiesCollection{}
}

func (t *testGameDelegate) EmptyComputedPlayerPropertyCollection() ComputedPropertyCollection {
	return &testPlayerPropertiesCollection{}
}

func (t *testGameDelegate) CheckGameFinished(state *State) (bool, []int) {
	_, players := concreteStates(state)

	var winners []int

	for i, player := range players {
		if player.Score >= 5 {
			//This user won!
			winners = append(winners, i)

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

func (t *testGameDelegate) BeginSetUp(state *State) {
	_, players := concreteStates(state)

	if len(players) != 3 {
		return
	}

	players[0].MovesLeftThisTurn = 1
	players[2].IsFoo = true
}

func (t *testGameDelegate) EmptyGameState() GameState {
	chest := t.Manager().Chest()

	deck := chest.Deck("test")
	return &testGameState{
		CurrentPlayer: 0,
		DrawDeck:      NewGrowableStack(deck, 0),
	}
}

func (t *testGameDelegate) EmptyPlayerState(playerIndex int) PlayerState {
	chest := t.Manager().Chest()

	deck := chest.Deck("test")

	return &testPlayerState{
		playerIndex:       0,
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
