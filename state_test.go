package boardgame

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestState(t *testing.T) {

	game := testGame()

	game.SetUp(0)

	state, err := game.Manager().Storage().State(game, game.Version())

	if err != nil {
		t.Error("Unexpected error", err)
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

	reconstitutedState, err := game.Manager().StateFromBlob(blob)

	if err != nil {
		t.Error("StateFromBlob returned unexpected err", err)
	}

	gameState, _ := concreteStates(reconstitutedState)

	if !gameState.DrawDeck.Inflated() {
		t.Error("The stack was not inflated when it came back from StateFromBlob")
	}

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

	if deserializedIn == nil {
		t.Error("In didn't deserialize", message)
	}

	if deserializedGolden == nil {
		t.Error("Golden didn't deserialize", message)
	}

	if !reflect.DeepEqual(deserializedIn, deserializedGolden) {
		t.Error("Got wrong json.", message, "Got", string(in), "wanted", string(golden))
	}
}

func goldenJSON(fileName string, t *testing.T) []byte {
	contents, err := ioutil.ReadFile("./test/" + fileName)
	if err != nil {
		t.Fatal("Couldn't load golden JSON at " + fileName)
	}

	return contents

}
