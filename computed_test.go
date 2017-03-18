package boardgame

import (
	"testing"
)

type stateComputeDelegate struct {
	testGameDelegate
	config                  *ComputedPropertiesConfig
	returnDefaultCollection bool
}

func (s *stateComputeDelegate) ComputedPropertiesConfig() *ComputedPropertiesConfig {
	return s.config
}

func (s *stateComputeDelegate) EmptyComputedGlobalPropertyCollection() ComputedPropertyCollection {
	if s.returnDefaultCollection {
		return s.testGameDelegate.EmptyComputedGlobalPropertyCollection()
	}
	return nil
}

func (s *stateComputeDelegate) EmptyComputedPlayerPropertyCollection() ComputedPropertyCollection {
	if s.returnDefaultCollection {
		return s.testGameDelegate.EmptyComputedPlayerPropertyCollection()
	}
	return nil
}

func TestComputedPropertyDefinitionCompute(t *testing.T) {

	game := testGame()

	if err := game.SetUp(0); err != nil {
		t.Fatal("Game failed to set up", err)
	}

	var passedState *State

	definition := &ComputedGlobalPropertyDefinition{
		Dependencies: []StatePropertyRef{
			{
				Group:    StateGroupGame,
				PropName: "CurrentPlayer",
			},
			{
				Group:    StateGroupPlayer,
				PropName: "Score",
			},
		},
		PropType: TypeInt,
		Compute: func(state *State) (interface{}, error) {
			//For now we'll just pass it out for inspection
			passedState = state
			return nil, nil
		},
	}

	state := game.CurrentState()

	gameState, playerStates := concreteStates(state)

	gameState.CurrentPlayer = 5

	for i, playerState := range playerStates {
		playerState.Score = i + 1
	}

	definition.compute(state)

	if passedState == nil {
		t.Error("Calling compute on the rigged definition didn't set passedState")
	}

	if val, err := passedState.game.Reader().IntProp("CurrentPlayer"); err != nil {
		t.Error("Unexpected error reading CurrentPlayer prop", err)
	} else if val != gameState.CurrentPlayer {
		t.Error("The shadow current player was not the real value. Got", val, "wanted", gameState.CurrentPlayer)
	}

	for i, playerState := range playerStates {
		playerShadow := passedState.players[i]

		if val, err := playerShadow.Reader().IntProp("Score"); err != nil {
			t.Error("Unexpected error reading Score prop", err)
		} else if val != playerState.Score {
			t.Error("Unexpected score was not real value. Got", val, "wanted", playerState.Score)
		}
	}

}

func TestStateComputed(t *testing.T) {

	delegate := &stateComputeDelegate{}

	manager := NewGameManager(delegate, newTestGameChest(), newTestStorageManager())

	manager.SetUp()

	game := NewGame(manager)

	game.SetUp(0)

	state := game.CurrentState()

	gameState, playerStates := concreteStates(state)

	gameState.CurrentPlayer = 4

	playerStates[0].Score = 10
	playerStates[1].Score = 5

	config := &ComputedPropertiesConfig{
		Global: map[string]ComputedGlobalPropertyDefinition{
			"CurrentPlayerPlusFive": ComputedGlobalPropertyDefinition{
				Dependencies: []StatePropertyRef{
					{
						Group:    StateGroupGame,
						PropName: "CurrentPlayer",
					},
				},
				PropType: TypeInt,
				Compute: func(state *State) (interface{}, error) {

					game, _ := concreteStates(state)

					return game.CurrentPlayer + 5, nil
				},
			},
			"SumAllScores": ComputedGlobalPropertyDefinition{
				Dependencies: []StatePropertyRef{
					{
						Group:    StateGroupPlayer,
						PropName: "Score",
					},
				},
				PropType: TypeInt,
				Compute: func(state *State) (interface{}, error) {
					result := 0

					_, playerStates := concreteStates(state)

					for _, player := range playerStates {

						result += player.Score
					}
					return result, nil
				},
			},
		},
		Player: map[string]ComputedPlayerPropertyDefinition{
			"EffectiveScore": ComputedPlayerPropertyDefinition{
				Dependencies: []StatePropertyRef{
					{
						Group:    StateGroupPlayer,
						PropName: "Score",
					},
					{
						Group:    StateGroupPlayer,
						PropName: "Hand",
					},
				},
				PropType: TypeInt,
				Compute: func(state PlayerState) (interface{}, error) {

					playerState := state.(*testPlayerState)

					return playerState.Score + playerState.Hand.Len(), nil

				},
			},
		},
	}

	delegate.config = config

	computed := state.Computed()

	if val, err := computed.Global().Reader().IntProp("CurrentPlayerPlusFive"); err != nil {
		t.Error("Unexpected error retrieving CurrentPlayerPlusFive", err)
	} else {
		if val != 4+5 {
			t.Error("CurrentPlayerPlusFive was unexpected value. Wanted", 4+5, "got", val)
		}
	}

	if val, err := computed.Global().Reader().IntProp("SumAllScores"); err != nil {
		t.Error("Unexpected error retrieving SumAllScores", err)
	} else if val != 15 {
		t.Error("Unexpected result for SumAllScores. Got", val, "wanted", 15)
	}

	if _, err := computed.Global().Reader().BoolProp("Foo"); err == nil {
		t.Error("Didn't get an error reading an unexpected bool prop")
	}

	if val, err := computed.Player(0).Reader().IntProp("EffectiveScore"); err != nil {
		t.Error("Got error for EffectiveScore", err)
	} else if val != 12 {
		//We set player 0 score to 10 a the top of this test, and there are two items in hand.
		t.Error("Got wrong value for EffectiveScore. Got", val, "wanted 12")
	}

}
