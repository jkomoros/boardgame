package boardgame

import (
	"encoding/json"
	"github.com/workfit/tester/assert"
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
		{
			&StatePolicy{
				DynamicComponentValues: map[string]SubStatePolicy{
					"test": SubStatePolicy{
						"IntVar": GroupPolicy{
							GroupAll: PolicyHidden,
						},
					},
				},
				Player: SubStatePolicy{
					"Hand": GroupPolicy{
						GroupOther: PolicyHidden,
					},
				},
			},
			1,
			"basic_state_after_dynamic_component_move.json",
			"sanitization_with_dynamic_state_sanitized.json",
		},
		{
			&StatePolicy{
				Game: SubStatePolicy{
					"DrawDeck": GroupPolicy{
						GroupAll: PolicyLen,
					},
				},
				Player: SubStatePolicy{
					"Hand": GroupPolicy{
						GroupOther: PolicyLen,
					},
				},
			},
			0,
			"basic_state_after_dynamic_component_move.json",
			"sanitization_with_dynamic_state_transitive.json",
		},
	}

	for i, test := range tests {

		inputBlob, err := ioutil.ReadFile("test/" + test.inputFileName)

		assert.For(t).ThatActual(err).IsNil()

		state, err := manager.StateFromBlob(inputBlob)

		assert.For(t).ThatActual(err).IsNil()

		manager.delegate.(*testSanitizationDelegate).policy = test.policy

		sanitizedState := state.SanitizedForPlayer(test.playerIndex)

		assert.For(t).ThatActual(sanitizedState).IsNotNil()

		sanitizedBlob, err := json.MarshalIndent(sanitizedState, "", "\t")

		assert.For(t).ThatActual(err).IsNil()

		goldenBlob, err := ioutil.ReadFile("test/" + test.expectedFileName)

		assert.For(t, i, test.expectedFileName).ThatActual(err).IsNil()

		compareJSONObjects(sanitizedBlob, goldenBlob, "Test Sanitization "+strconv.Itoa(i), t)

	}

}
