package boardgame

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/jkomoros/boardgame/internal/patchtree"
	"github.com/workfit/tester/assert"
)

func TestPolicyFromString(t *testing.T) {
	tests := []struct {
		in       string
		expected Policy
	}{
		{
			"VisibLE",
			PolicyVisible,
		},
		{
			"order ",
			PolicyOrder,
		},
		{
			"len",
			PolicyLen,
		},
		{
			"nonempty",
			PolicyNonEmpty,
		},
		{
			"Hidden",
			PolicyHidden,
		},
		{
			"visibl",
			PolicyInvalid,
		},
	}

	for i, test := range tests {
		result := policyFromString(test.in)

		if result != test.expected {
			t.Error("For test ", i, "wanted", test.expected, "got", result)
		}
	}
}

//Basically has the information that WOULD have been provided by sruct tags
type sanitizationTestConfig struct {
	Game                   map[string]string
	Player                 map[string]string
	DynamicComponentValues map[string]map[string]string
}

//Install sets up the given manager's iternal methods as though the config was
//read in from the struct tags.
func (s *sanitizationTestConfig) Install(manager *GameManager) {

	//Reset memoization of special group names, since sanitization machinery
	//relies on cached values for them.
	manager.memoizedSpecialGroupNames = nil

	manager.gameValidator.sanitizationPolicy = s.policyForSubObject(manager.Delegate().GameStateConstructor().Reader(), s.Game, false)
	manager.playerValidator.sanitizationPolicy = s.policyForSubObject(manager.Delegate().PlayerStateConstructor(0).Reader(), s.Player, true)

	for _, deckName := range manager.Chest().DeckNames() {
		deck := manager.Chest().Deck(deckName)
		manager.dynamicComponentValidator[deckName].sanitizationPolicy = s.policyForSubObject(manager.Delegate().DynamicComponentValuesConstructor(deck).Reader(), s.DynamicComponentValues[deckName], false)
	}

}

func (s *sanitizationTestConfig) policyForSubObject(reader PropertyReader, config map[string]string, isPlayer bool) map[string]map[string]Policy {

	result := make(map[string]map[string]Policy)

	defaultGroup := "all"

	if isPlayer {
		defaultGroup = "other"
	}

	for propName := range reader.Props() {
		result[propName] = policyFromStructTag(config[propName], defaultGroup)
	}

	return result

}

func TestSanitization(t *testing.T) {

	tests := []struct {
		policy            *sanitizationTestConfig
		playerIndex       PlayerIndex
		inputPatchTree    string
		expectedPatchTree string
	}{
		{
			&sanitizationTestConfig{},
			AdminPlayerIndex,
			"sanitize",
			"sanitize",
		},
		{

			&sanitizationTestConfig{
				Game: map[string]string{
					"DrawDeck":           "len",
					"MyIntSlice":         "len",
					"MyBoolSlice":        "len",
					"MyStringSlice":      "len",
					"MyPlayerIndexSlice": "len",
					"MyBoard":            "len",
				},
			},
			0,
			"sanitize",
			"sanitize/len",
		},
		{
			&sanitizationTestConfig{
				Player: map[string]string{
					"Hand": "len",
				},
			},
			0,
			"sanitize",
			"sanitize/len_player",
		},
		{
			&sanitizationTestConfig{
				Game: map[string]string{
					"DrawDeck":           "order",
					"MyIntSlice":         "order",
					"MyBoolSlice":        "order",
					"MyStringSlice":      "order",
					"MyPlayerIndexSlice": "order",
					"MyBoard":            "order",
				},
			},
			0,
			"sanitize",
			"sanitize/order",
		},
		{
			&sanitizationTestConfig{
				Player: map[string]string{
					"Hand": "order",
				},
			},
			0,
			"sanitize",
			"sanitize/order_player",
		},
		{
			&sanitizationTestConfig{
				Game: map[string]string{
					"DrawDeck":           "hidden",
					"MyIntSlice":         "hidden",
					"MyBoolSlice":        "hidden",
					"MyStringSlice":      "hidden",
					"MyPlayerIndexSlice": "hidden",
					"MyEnumValue":        "hidden",
					"MyBoard":            "hidden",
				},
				Player: map[string]string{
					"Hand":              "all:hidden,self:visible",
					"MovesLeftThisTurn": "all:hidden",
					"IsFoo":             "all:hidden",
					"EnumVal":           "all:hidden",
				},
			},
			0,
			"sanitize",
			"sanitize/hidden",
		},
		{
			&sanitizationTestConfig{
				Game: map[string]string{
					"DrawDeck":           "nonempty",
					"MyIntSlice":         "nonempty",
					"MyBoolSlice":        "nonempty",
					"MyStringSlice":      "nonempty",
					"MyPlayerIndexSlice": "nonempty",
					"MyBoard":            "nonempty",
				},
				Player: map[string]string{
					"Hand": "all:nonempty",
				},
			},
			0,
			"sanitize",
			"sanitize/nonempty",
		},
		{
			&sanitizationTestConfig{
				DynamicComponentValues: map[string]map[string]string{
					"test": {
						"IntVar": "hidden",
					},
				},
			},
			0,
			"after_dynamic_component_move",
			"after_dynamic_component_move/sanitization_with_dynamic_state",
		},
		{
			&sanitizationTestConfig{
				DynamicComponentValues: map[string]map[string]string{
					"test": {
						"IntVar": "hidden",
					},
				},
				Player: map[string]string{
					"Hand": "hidden",
				},
			},
			1,
			"after_dynamic_component_move",
			"after_dynamic_component_move/sanitization_with_dynamic_state_sanitized",
		},
		{
			&sanitizationTestConfig{
				Game: map[string]string{
					"DrawDeck": "len",
				},
				Player: map[string]string{
					"Hand": "len",
				},
			},
			0,
			"after_dynamic_component_move",
			"after_dynamic_component_move/sanitization_with_dynamic_state_transitive",
		},
	}

	game := testDefaultGame(t, true)

	for i, test := range tests {

		inputBlob, err := patchtree.JSON("testdata/" + test.inputPatchTree)

		if err != nil {
			t.Fatal("patchtree failure: " + err.Error())
		}

		//version doesn't matter in this context
		state, err := game.manager.stateFromRecord(inputBlob, 0)

		//This is hacky, but we don't really need the game for much more anyway
		state.game = game

		assert.For(t).ThatActual(err).IsNil()

		test.policy.Install(game.Manager())

		sanitizedState, err := state.SanitizedForPlayer(test.playerIndex)

		assert.For(t).ThatActual(err).IsNil()

		assert.For(t).ThatActual(sanitizedState).IsNotNil()

		sanitizedBlob, err := json.MarshalIndent(sanitizedState, "", "\t")

		assert.For(t).ThatActual(err).IsNil()

		expectedBlob, err := patchtree.JSON("testdata/" + test.expectedPatchTree)

		if err != nil {
			t.Fatal("patchetree failure for expected: " + err.Error())
		}

		compareJSONObjects(sanitizedBlob, expectedBlob, "Test Sanitization "+strconv.Itoa(i)+" "+test.inputPatchTree+" "+test.expectedPatchTree, t)

	}

}
