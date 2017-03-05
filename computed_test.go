package boardgame

import (
	"testing"
)

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
