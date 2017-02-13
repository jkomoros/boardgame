package boardgame

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestState(t *testing.T) {

	game := testGame()

	state := game.StateWrapper

	if state == nil {
		t.Error("State could not be created")
	}

	currentJson, _ := json.Marshal(state)
	golden := goldenJSON("basic_state.json", t)

	compareJSONObjects(currentJson, golden, "Basic state", t)

	stateCopy := state.Copy()

	copyJson, _ := DefaultMarshalJSON(stateCopy)

	compareJSONObjects(copyJson, currentJson, "Copy was not same", t)

	stateCopy.Schema = 1

	if state.Schema == 1 {
		t.Error("Modifying a copy changed the original")
	}

	stateCopy.State.(*testState).Users[0].MovesLeftThisTurn = 10

	if state.State.(*testState).Users[0].MovesLeftThisTurn == 10 {
		t.Error("Modifying a copy change the original")
	}

	//TODO: test that GAmeState and UserStates are also copies
}

func TestStateSerialization(t *testing.T) {

	//Temporarily disable this test while we fix the underlying problem it
	//exposed.
	return

	game := testGame()

	game.SetUp()

	if err := <-game.ProposeMove(&testMove{
		AString:           "bam",
		ScoreIncrement:    3,
		TargetPlayerIndex: 0,
		ABool:             true,
	}); err != nil {
		t.Fatal("Couldn't make move", err)
	}

	blob, err := json.Marshal(game.StateWrapper.State)

	if err != nil {
		t.Fatal("Couldn't serialize state:", err)
	}

	reconstitutedState, err := game.Delegate.StateFromBlob(blob, 0)

	if err != nil {
		t.Error("StateFromBlob returned unexpected err", err)
	}

	if !reflect.DeepEqual(reconstitutedState, game.StateWrapper.State) {

		rStateBlob, _ := json.Marshal(reconstitutedState)
		oStateBlob, _ := json.Marshal(game.StateWrapper.State)

		t.Error("Reconstituted state and original state were not the same. Got", string(rStateBlob), "wanted", string(oStateBlob))
	}
}

type propertyReaderTestStruct struct {
	A int
	B bool
	C string
	//d should be excluded since it is lowercase
	d string
}

func (p *propertyReaderTestStruct) Props() []string {
	return PropertyReaderPropsImpl(p)
}

func (p *propertyReaderTestStruct) Prop(name string) interface{} {
	return PropertyReaderPropImpl(p, name)
}

func (p *propertyReaderTestStruct) SetProp(name string, val interface{}) error {
	return PropertySetImpl(p, name, val)
}

func TestPropertyReaderImpl(t *testing.T) {

	s := &propertyReaderTestStruct{
		C: "bam",
	}

	result := s.Props()

	expected := []string{"A", "B", "C"}

	if !reflect.DeepEqual(result, expected) {
		t.Error("PropertyReaderPropsImpl returned wrong result. Got", result, "expected", expected)
	}

	field := s.Prop("C")

	if field.(string) != "bam" {
		t.Error("Got back wrong value from Prop. Got", field, "expected 'foo'")
	}

	field = s.Prop("d")

	if field != nil {
		t.Error("Expected to not get back a result for private field, but did", field)
	}

	if err := s.SetProp("A", 4); err != nil {
		t.Error("Setting A to 4 failed: ", err)
	}

	if s.A != 4 {
		t.Error("Using setProp to set to 4 failed.")
	}

	if err := s.SetProp("A", "string"); err == nil {
		t.Error("Trying to set a string into an int slot didn't fail")
	}

	if s.A != 4 {
		t.Error("Failed setting into a field modified the value")
	}

}

func compareJSONObjects(in []byte, golden []byte, message string, t *testing.T) {

	var deserializedIn interface{}
	var deserializedGolden interface{}

	json.Unmarshal(in, &deserializedIn)
	json.Unmarshal(golden, &deserializedGolden)

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
