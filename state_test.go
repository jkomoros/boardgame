package boardgame

import (
	"encoding/json"
	"github.com/workfit/tester/assert"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestState(t *testing.T) {

	game := testGame()

	game.SetUp(0)

	record, err := game.Manager().Storage().State(game.Id(), game.Version())

	if err != nil {
		t.Error("Unexpected error", err)
	}

	state, err := game.Manager().stateFromRecord(record)

	if err != nil {
		t.Error("StateFromBlob err", err)
	}

	if state == nil {
		t.Error("State could not be created")
	}

	currentJson, _ := json.Marshal(state)
	golden := goldenJSON("basic_state.json", t)

	compareJSONObjects(currentJson, golden, "Basic state", t)

	stateCopy := state.Copy(false)

	copyJson, _ := DefaultMarshalJSON(stateCopy)

	compareJSONObjects(copyJson, currentJson, "Copy was not same", t)

	_, playerStatesCopy := concreteStates(stateCopy)

	playerStatesCopy[0].MovesLeftThisTurn = 10

	_, playerStates := concreteStates(state)

	if playerStates[0].MovesLeftThisTurn == 10 {
		t.Error("Modifying a copy change the original")
	}

	if state.Sanitized() {
		t.Error("State reported being sanitized even when it wasn't")
	}

	sanitizedStateCopy := stateCopy.Copy(true)

	if !sanitizedStateCopy.Sanitized() {
		t.Error("A copy that was told it was sanitized did not report being sanitized.")
	}

	//TODO: test that GAmeState and UserStates are also copies
}

func TestStateSerialization(t *testing.T) {

	game := testGame()

	game.SetUp(0)

	if err := <-game.ProposeMove(&testMove{
		AString:           "bam",
		ScoreIncrement:    3,
		TargetPlayerIndex: 0,
		ABool:             true,
	}); err != nil {
		t.Fatal("Couldn't make move", err)
	}

	blob, err := json.Marshal(game.CurrentState())

	if err != nil {
		t.Fatal("Couldn't serialize state:", err)
	}

	reconstitutedState, err := game.Manager().stateFromRecord(blob)

	if err != nil {
		t.Error("StateFromBlob returned unexpected err", err)
	}

	gameState, _ := concreteStates(reconstitutedState)

	if !gameState.DrawDeck.Inflated() {
		t.Error("The stack was not inflated when it came back from StateFromBlob")
	}

	if !gameState.DrawDeck.ComponentAt(0).DynamicValues(reconstitutedState).(*testingComponentDynamic).Stack.Inflated() {
		t.Error("The stack on a component's dynamic value was not inflated coming back from storage.")
	}

	//This is lame, but when you create json for a State, it touches Computed,
	//which will make it non-nil, so if you're doing direct comparison they
	//won't compare equal even though they basically are. At this point
	//CurrentState has already been touched above by creating a json blob. So
	//just touch reconstitutedState, too. ¯\_(ツ)_/¯

	_, _ = json.Marshal(reconstitutedState)

	if !reflect.DeepEqual(reconstitutedState, game.CurrentState()) {

		rStateBlob, _ := json.Marshal(reconstitutedState)
		oStateBlob, _ := json.Marshal(game.CurrentState())

		t.Error("Reconstituted state and original state were not the same. Got", string(rStateBlob), "wanted", string(oStateBlob))
	}
}

func compareJSONObjects(in []byte, golden []byte, message string, t *testing.T) {

	//recreated in server/internal/teststoragemanager

	var deserializedIn interface{}
	var deserializedGolden interface{}

	json.Unmarshal(in, &deserializedIn)
	json.Unmarshal(golden, &deserializedGolden)

	assert.For(t).ThatActual(deserializedIn).IsNotNil()

	assert.For(t).ThatActual(deserializedGolden).IsNotNil()

	assert.For(t, message).ThatActual(deserializedGolden).Equals(deserializedIn).ThenDiffOnFail()

}

func goldenJSON(fileName string, t *testing.T) []byte {
	contents, err := ioutil.ReadFile("./test/" + fileName)

	if !assert.For(t, fileName).ThatActual(err).IsNil().Passed() {
		t.FailNow()
	}

	return contents

}
