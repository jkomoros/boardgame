package boardgame

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestState(t *testing.T) {

	game := testGame()

	state := game.State

	if state == nil {
		t.Error("State could not be created")
	}

	json := state.JSON()
	golden := goldenJSON("basic_state.json", t)

	compareJSONObjects(json, golden, "Basic state", t)

	stateCopy := state.Copy()

	compareJSONObjects(stateCopy.JSON(), state.JSON(), "Copy was not same", t)

	stateCopy.Schema = 1

	if state.Schema == 1 {
		t.Error("Modifying a copy changed the original")
	}

	//TODO: test that GAmeState and UserStates are also copies
}

func compareJSONObjects(in JSONObject, golden JSONObject, message string, t *testing.T) {
	serializedIn := string(Serialize(in))
	serializedGolden := string(Serialize(golden))
	if serializedIn != serializedGolden {
		t.Error("Got wrong json.", message, "Got", serializedIn, "wanted", serializedGolden)
	}
}

func goldenJSON(fileName string, t *testing.T) JSONObject {
	contents, err := ioutil.ReadFile("./test/" + fileName)
	if err != nil {
		t.Fatal("Couldn't load golden JSON at " + fileName)
	}

	result := make(JSONMap)

	if err := json.Unmarshal(contents, &result); err != nil {
		t.Fatal("Couldn't parse golden json at " + fileName + err.Error())
	}

	return result

}
