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

	stateCopy := state.Copy()

	compareJSONObjects(stateCopy.JSON(), state.JSON(), "Copy was not same", t)

	stateCopy.Schema = 1

	if state.Schema == 1 {
		t.Error("Modifying a copy changed the original")
	}

	//TODO: test that GAmeState and UserStates are also copies
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

//TODO: move testingComponent to component_test.go

//testingComponent is a very basic thing that fufills the Component interface.
type testingComponent struct {
	deckName  string
	deckIndex int
	String    string
	Integer   int
}

const testGameName = "testgame"

func (t *testingComponent) Props() []string {
	return []string{"String", "Integer"}
}

func (t *testingComponent) Prop(name string) interface{} {
	switch name {
	case "String":
		return t.String
	case "Integer":
		return t.Integer
	default:
		return nil
	}
}

func (t *testingComponent) Deck() string {
	return t.deckName
}

func (t *testingComponent) DeckIndex() int {
	return t.deckIndex
}

func (t *testingComponent) GameName() string {
	return testGameName
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
		testGameName,
		chest,
		nil,
	}

	return game
}
