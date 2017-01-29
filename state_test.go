package boardgame

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
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

type propertyReaderTestStruct struct {
	A int
	B bool
	C string
}

func (p *propertyReaderTestStruct) Props() []string {
	return PropertyReaderPropsImpl(p)
}

func (p *propertyReaderTestStruct) Prop(name string) interface{} {
	return PropertyReaderPropImpl(p, name)
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

}

func compareJSONObjects(in JSONObject, golden JSONObject, message string, t *testing.T) {
	serializedIn := Serialize(in)
	serializedGolden := Serialize(golden)

	var deserializedIn interface{}
	var deserializedGolden interface{}

	json.Unmarshal(serializedIn, &deserializedIn)
	json.Unmarshal(serializedGolden, &deserializedGolden)

	if !reflect.DeepEqual(deserializedIn, deserializedGolden) {
		t.Error("Got wrong json.", message, "Got", string(serializedIn), "wanted", string(serializedGolden))
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
