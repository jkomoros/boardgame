package boardgame

import (
	"encoding/json"
	"github.com/workfit/tester/assert"
	"io/ioutil"
	"log"
	"strconv"
	"testing"
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

	manager.gameValidator.sanitizationPolicy = s.policyForSubObject(manager.Delegate().GameStateConstructor().Reader(), s.Game, false)
	manager.playerValidator.sanitizationPolicy = s.policyForSubObject(manager.Delegate().PlayerStateConstructor(0).Reader(), s.Player, true)

	for _, deckName := range manager.Chest().DeckNames() {
		deck := manager.Chest().Deck(deckName)
		manager.dynamicComponentValidator[deckName].sanitizationPolicy = s.policyForSubObject(manager.Delegate().DynamicComponentValuesConstructor(deck).Reader(), s.DynamicComponentValues[deckName], false)
	}

}

func (s *sanitizationTestConfig) policyForSubObject(reader PropertyReader, config map[string]string, isPlayer bool) map[string]map[int]Policy {

	result := make(map[string]map[int]Policy)

	defaultGroup := "all"

	if isPlayer {
		defaultGroup = "other"
	}

	for propName, _ := range reader.Props() {
		result[propName] = policyFromStructTag(config[propName], defaultGroup)
	}

	return result

}

func TestSanitization(t *testing.T) {

	tests := []struct {
		policy           *sanitizationTestConfig
		playerIndex      PlayerIndex
		inputFileName    string
		expectedFileName string
	}{
		{
			&sanitizationTestConfig{},
			AdminPlayerIndex,
			"sanitization_basic_in.json",
			"sanitization_basic_in.json",
		},
		{

			&sanitizationTestConfig{
				Game: map[string]string{
					"DrawDeck":           "len",
					"MyIntSlice":         "len",
					"MyBoolSlice":        "len",
					"MyStringSlice":      "len",
					"MyPlayerIndexSlice": "len",
				},
			},
			0,
			"sanitization_basic_in.json",
			"sanitization_basic_len.json",
		},
		{
			&sanitizationTestConfig{
				Player: map[string]string{
					"Hand": "len",
				},
			},
			0,
			"sanitization_basic_in.json",
			"sanitization_basic_len_player.json",
		},
		{
			&sanitizationTestConfig{
				Game: map[string]string{
					"DrawDeck":           "order",
					"MyIntSlice":         "order",
					"MyBoolSlice":        "order",
					"MyStringSlice":      "order",
					"MyPlayerIndexSlice": "order",
				},
			},
			0,
			"sanitization_basic_in.json",
			"sanitization_basic_order.json",
		},
		{
			&sanitizationTestConfig{
				Player: map[string]string{
					"Hand": "order",
				},
			},
			0,
			"sanitization_basic_in.json",
			"sanitization_basic_order_player.json",
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
				},
				Player: map[string]string{
					"Hand":              "all:hidden,self:visible",
					"MovesLeftThisTurn": "all:hidden",
					"IsFoo":             "all:hidden",
					"EnumVal":           "all:hidden",
				},
			},
			0,
			"sanitization_basic_in.json",
			"sanitization_basic_hidden.json",
		},
		{
			&sanitizationTestConfig{
				Game: map[string]string{
					"DrawDeck":           "nonempty",
					"MyIntSlice":         "nonempty",
					"MyBoolSlice":        "nonempty",
					"MyStringSlice":      "nonempty",
					"MyPlayerIndexSlice": "nonempty",
				},
				Player: map[string]string{
					"Hand": "all:nonempty",
				},
			},
			0,
			"sanitization_basic_in.json",
			"sanitization_basic_nonempty.json",
		},
		{
			&sanitizationTestConfig{
				DynamicComponentValues: map[string]map[string]string{
					"test": map[string]string{
						"IntVar": "hidden",
					},
				},
			},
			0,
			"basic_state_after_dynamic_component_move.json",
			"sanitization_with_dynamic_state.json",
		},
		{
			&sanitizationTestConfig{
				DynamicComponentValues: map[string]map[string]string{
					"test": map[string]string{
						"IntVar": "hidden",
					},
				},
				Player: map[string]string{
					"Hand": "hidden",
				},
			},
			1,
			"basic_state_after_dynamic_component_move.json",
			"sanitization_with_dynamic_state_sanitized.json",
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
			"basic_state_after_dynamic_component_move.json",
			"sanitization_with_dynamic_state_transitive.json",
		},
	}

	game := testGame(t)

	makeTestGameIdsStable(game)

	for i, test := range tests {

		inputBlob, err := ioutil.ReadFile("test/" + test.inputFileName)

		assert.For(t).ThatActual(err).IsNil()

		state, err := game.manager.stateFromRecord(inputBlob)

		if !assert.For(t).ThatActual(err).IsNil().Passed() {
			log.Println(test.inputFileName)
		}

		//This is hacky, but we don't really need the game for much more anyway
		state.game = game

		assert.For(t).ThatActual(err).IsNil()

		test.policy.Install(game.Manager())

		sanitizedState := state.SanitizedForPlayer(test.playerIndex)

		assert.For(t).ThatActual(sanitizedState).IsNotNil()

		sanitizedBlob, err := json.MarshalIndent(sanitizedState, "", "\t")

		assert.For(t).ThatActual(err).IsNil()

		goldenBlob, err := ioutil.ReadFile("test/" + test.expectedFileName)

		assert.For(t, i, test.expectedFileName).ThatActual(err).IsNil()

		compareJSONObjects(sanitizedBlob, goldenBlob, "Test Sanitization "+strconv.Itoa(i)+" "+test.inputFileName+" "+test.expectedFileName, t)

	}

}
