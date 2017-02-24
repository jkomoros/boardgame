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

	stateCopy := state.Copy()

	copyJson, _ := DefaultMarshalJSON(stateCopy)

	compareJSONObjects(copyJson, currentJson, "Copy was not same", t)

	stateCopy.(*testState).Players[0].MovesLeftThisTurn = 10

	if state.(*testState).Players[0].MovesLeftThisTurn == 10 {
		t.Error("Modifying a copy change the original")
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

	reconstitutedState, err := game.Manager().Delegate().StateFromBlob(blob)

	if err != nil {
		t.Error("StateFromBlob returned unexpected err", err)
	}

	if !reconstitutedState.(*testState).Game.DrawDeck.Inflated() {
		t.Error("The stack was not inflated when it came back from StateFromBlob")
	}

	if !reflect.DeepEqual(reconstitutedState, game.CurrentState()) {

		rStateBlob, _ := json.Marshal(reconstitutedState)
		oStateBlob, _ := json.Marshal(game.CurrentState())

		t.Error("Reconstituted state and original state were not the same. Got", string(rStateBlob), "wanted", string(oStateBlob))
	}
}

type propertyReaderTestStruct struct {
	A int
	B bool
	C string
	G *GrowableStack
	S *SizedStack
	//d should be excluded since it is lowercase
	d string
}

func (p *propertyReaderTestStruct) ReadSetter() PropertyReadSetter {
	return DefaultReadSetter(p)
}

func TestPropertyReaderImpl(t *testing.T) {

	deck := &Deck{}

	p := &propertyReaderTestStruct{
		C: "bam",
		G: NewGrowableStack(deck, 3),
		S: NewSizedStack(deck, 3),
	}

	s := p.ReadSetter()

	result := s.Props()

	expected := map[string]PropertyType{"A": TypeInt, "B": TypeBool, "C": TypeString, "G": TypeGrowableStack, "S": TypeSizedStack}

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

	if p.A != 4 {
		t.Error("Using setProp to set to 4 failed.")
	}

	if err := s.SetProp("A", "string"); err == nil {
		t.Error("Trying to set a string into an int slot didn't fail")
	}

	if p.A != 4 {
		t.Error("Failed setting into a field modified the value")
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
