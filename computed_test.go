package boardgame

import (
	"testing"
)

type stateComputeDelegate struct {
	testGameDelegate
	config *ComputedPropertiesConfig
}

func (s *stateComputeDelegate) ComputedPropertiesConfig() *ComputedPropertiesConfig {
	return s.config
}

func TestComputedPropertyDefinitionCompute(t *testing.T) {

	game := testGame()

	game.SetUp(0)

	var passedShadow *ShadowState

	definition := &ComputedPropertyDefinition{
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
		Compute: func(shadow *ShadowState) (interface{}, error) {
			//For now we'll just pass it out for inspection
			passedShadow = shadow
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

	if passedShadow == nil {
		t.Error("Calling compute on the rigged definition didn't set passedState")
	}

	if val, err := passedShadow.Game.IntProp("CurrentPlayer"); err != nil {
		t.Error("Unexpected error reading CurrentPlayer prop", err)
	} else if val != gameState.CurrentPlayer {
		t.Error("The shadow current player was not the real value. Got", val, "wanted", gameState.CurrentPlayer)
	}

	for i, playerState := range playerStates {
		playerShadow := passedShadow.Players[i]

		if val, err := playerShadow.IntProp("Score"); err != nil {
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
		Properties: map[string]ComputedPropertyDefinition{
			"CurrentPlayerPlusFive": ComputedPropertyDefinition{
				Dependencies: []StatePropertyRef{
					{
						Group:    StateGroupGame,
						PropName: "CurrentPlayer",
					},
				},
				PropType: TypeInt,
				Compute: func(shadow *ShadowState) (interface{}, error) {
					val, err := shadow.Game.IntProp("CurrentPlayer")
					if err != nil {
						return nil, err
					}
					return val + 5, nil
				},
			},
			"SumAllScores": ComputedPropertyDefinition{
				Dependencies: []StatePropertyRef{
					{
						Group:    StateGroupPlayer,
						PropName: "Score",
					},
				},
				PropType: TypeInt,
				Compute: func(shadow *ShadowState) (interface{}, error) {
					result := 0
					for _, player := range shadow.Players {
						val, err := player.IntProp("Score")

						if err != nil {
							return nil, err
						}

						result += val
					}
					return result, nil
				},
			},
		},
	}

	delegate.config = config

	computed := state.Computed()

	if val, err := computed.IntProp("CurrentPlayerPlusFive"); err != nil {
		t.Error("Unexpected error retrieving CurrentPlayerPlusFive", err)
	} else {
		if val != 4+5 {
			t.Error("CurrentPlayerPlusFive was unexpected value. Wanted", 4+5, "got", val)
		}
	}

	if val, err := computed.IntProp("SumAllScores"); err != nil {
		t.Error("Unexpected error retrieving SumAllScores", err)
	} else if val != 15 {
		t.Error("Unexpected result for SumAllScores. Got", val, "wanted", 15)
	}

	if _, err := computed.BoolProp("Foo"); err == nil {
		t.Error("Didn't get an error reading an unexpected bool prop")
	}

}
