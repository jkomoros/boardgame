package boardgame

import (
	"encoding/json"
	"io/ioutil"
	"testing"
)

func TestState(t *testing.T) {
	state := &State{
		0,
		0,
		nil,
		nil,
	}

	if state == nil {
		t.Error("State could not be created")
	}

	json := state.JSON()
	golden := goldenJSON("empty_state.json", t)

	compareJSONObjects(json, golden, "Empty state", t)
}

func compareJSONObjects(in JSONObject, golden JSONObject, message string, t *testing.T) {
	if string(in.Serialize()) != string(golden.Serialize()) {
		t.Error("Got wrong json.", message, "Got", in, "wanted", golden)
	}
}

func goldenJSON(fileName string, t *testing.T) JSONObject {
	contents, err := ioutil.ReadFile("./test/" + fileName)
	if err != nil {
		t.Fatal("Couldn't load golden JSON at " + fileName)
	}

	result := make(JSONObject)

	if err := json.Unmarshal(contents, &result); err != nil {
		t.Fatal("Couldn't parse golden json at " + fileName + err.Error())
	}

	return result

}

func testGame() *Game {
	//TODO: some kind of way to set the deckName/Index automatically at insertion?
	chest := ComponentChest{
		"test": &Deck{
			Name: "test",
			Components: []Component{
				&testingComponent{
					"test",
					0,
					"foo",
					1,
				},
				&testingComponent{
					"test",
					1,
					"bar",
					2,
				},
			},
		},
	}

	game := &Game{
		chest,
		nil,
	}

	return game
}
