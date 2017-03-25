package boardgame

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"
)

type testSanitizationDelegate struct {
	testGameDelegate
	policy *StatePolicy
}

func (t *testSanitizationDelegate) StateSanitizationPolicy() *StatePolicy {
	return t.policy
}

func TestSanitization(t *testing.T) {
	manager := NewGameManager(&testSanitizationDelegate{}, newTestGameChest(), newTestStorageManager())

	manager.SetUp()

	tests := []struct {
		policy           *StatePolicy
		playerIndex      int
		inputFileName    string
		expectedFileName string
	}{
		{
			nil,
			-1,
			"sanitization_basic_in.json",
			"sanitization_basic_in.json",
		},
		{
			&StatePolicy{
				Game: map[string]GroupPolicy{
					"DrawDeck": GroupPolicy{
						GroupAll: PolicyLen,
					},
				},
			},
			0,
			"sanitization_basic_in.json",
			"sanitization_basic_len.json",
		},
		{
			&StatePolicy{
				Player: map[string]GroupPolicy{
					"Hand": GroupPolicy{
						GroupOther: PolicyLen,
					},
				},
			},
			0,
			"sanitization_basic_in.json",
			"sanitization_basic_len_player.json",
		},
		{
			&StatePolicy{
				Game: map[string]GroupPolicy{
					"DrawDeck": GroupPolicy{
						GroupAll: PolicyHidden,
					},
				},
				Player: map[string]GroupPolicy{
					"Hand": GroupPolicy{
						GroupAll:  PolicyHidden,
						GroupSelf: PolicyVisible,
					},
					"MovesLeftThisTurn": GroupPolicy{
						GroupAll: PolicyHidden,
					},
					"IsFoo": GroupPolicy{
						GroupAll: PolicyHidden,
					},
				},
			},
			0,
			"sanitization_basic_in.json",
			"sanitization_basic_hidden.json",
		},
		{
			&StatePolicy{
				Game: map[string]GroupPolicy{
					"DrawDeck": GroupPolicy{
						GroupAll: PolicyNonEmpty,
					},
				},
				Player: map[string]GroupPolicy{
					"Hand": GroupPolicy{
						GroupAll: PolicyNonEmpty,
					},
				},
			},
			0,
			"sanitization_basic_in.json",
			"sanitization_basic_nonempty.json",
		},
		{
			&StatePolicy{
				DynamicComponentValues: map[string]SubStatePolicy{
					"test": SubStatePolicy{
						"IntVar": GroupPolicy{
							GroupAll: PolicyHidden,
						},
					},
				},
			},
			0,
			"basic_state_after_dynamic_component_move.json",
			"sanitization_with_dynamic_state.json",
		},
	}

	for i, test := range tests {

		inputBlob, err := ioutil.ReadFile("test/" + test.inputFileName)

		if err != nil {
			t.Fatal("couldn't load input file", i, test.inputFileName, err)
		}

		state, err := manager.StateFromBlob(inputBlob)

		if err != nil {
			t.Fatal(i, "Failed to deserialize", err)
		}

		manager.delegate.(*testSanitizationDelegate).policy = test.policy

		sanitizedState := state.SanitizedForPlayer(test.playerIndex)

		if sanitizedState == nil {
			t.Fatal(i, "state sanitization came back nil")
		}

		sanitizedBlob, err := json.Marshal(sanitizedState)

		if err != nil {
			t.Fatal(i, "Sanitized serialize failed", err)
		}

		goldenBlob, err := ioutil.ReadFile("test/" + test.expectedFileName)

		if err != nil {
			t.Fatal("Couldn't load file", i, test.expectedFileName, err)
		}

		compareJSONObjects(sanitizedBlob, goldenBlob, "Test Sanitization "+strconv.Itoa(i), t)

	}

}
