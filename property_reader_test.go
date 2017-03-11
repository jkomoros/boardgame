package boardgame

import (
	"reflect"
	"testing"
)

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

	deck := NewDeck()

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

	field, err := s.Prop("C")

	if err != nil {
		t.Error("Unexpected error fetching generic prop", err)
	}

	if field.(string) != "bam" {
		t.Error("Got back wrong value from Prop. Got", field, "expected 'foo'")
	}

	field, err = s.Prop("d")

	if err == nil {
		t.Error("Didn't get expected error when fetching invalid property")
	}

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

	intResult, err := s.IntProp("A")

	if err != nil {
		t.Error("Unexpected error fetching int prop", err)
	}

	if intResult != 4 {
		t.Error("Unexpected result from intprop", intResult)
	}

	intResult, err = s.IntProp("B")

	if err == nil {
		t.Error("Fetch on non-int prop with intprop did not fail")
	}

	boolResult, err := s.BoolProp("B")

	if err != nil {
		t.Error("Unexpected error fetching bool prop", err)
	}

	if boolResult != p.B {
		t.Error("Unexpected bool result")
	}

	boolResult, err = s.BoolProp("A")

	if err == nil {
		t.Error("Didn't get error fetching non-bool prop")
	}

	stringResult, err := s.StringProp("C")

	if err != nil {
		t.Error("Unexpected error fetching string", err)
	}

	if stringResult != p.C {
		t.Error("Unexpeted string result")
	}

	stringResult, err = s.StringProp("A")

	if err == nil {
		t.Error("Didn't error trying to string prop a non-string")
	}

	growableStackResult, err := s.GrowableStackProp("G")

	if err != nil {
		t.Error("Unexpted error fetching growable stack", err)
	}

	if growableStackResult != p.G {
		t.Error("Unexpected growable stack result")
	}

	growableStackResult, err = s.GrowableStackProp("A")

	if err == nil {
		t.Error("didn't get error for unreasonable growable stack fetch")
	}

	sizedStackResult, err := s.SizedStackProp("S")

	if err != nil {
		t.Error("Unexpted error fetching sized stack", err)
	}

	if sizedStackResult != p.S {
		t.Error("Unexpected sized stack result")
	}

	sizedStackResult, err = s.SizedStackProp("A")

	if err == nil {
		t.Error("didn't get error for unreasonable sized stack fetch")
	}

}

func TestGenericReader(t *testing.T) {
	reader := newGenericReader()

	if _, err := reader.Prop("test"); err == nil {
		t.Error("A read on an empty reader didn't fail")
	}

	if err := reader.SetIntProp("intProp", 1); err != nil {
		t.Error("Unexpected error setting int: ", err)
	}

	if intVal, err := reader.IntProp("intProp"); err != nil {
		t.Error("Unexpected error reading back vaid int", err)
	} else if intVal != 1 {
		t.Error("Reading back legit int was not expected val. Got", intVal, "wanted", 1)
	}

	if err := reader.SetBoolProp("boolProp", true); err != nil {
		t.Error("Unexpected error setting bool: ", err)
	}

	if boolVal, err := reader.BoolProp("boolProp"); err != nil {
		t.Error("Unexpected error reading back valid bool", err)
	} else if boolVal != true {
		t.Error("Reading back legit bool was not expected val. Got", boolVal, "wanted", true)
	}

	if err := reader.SetIntProp("boolProp", 2); err == nil {
		t.Error("Setting an int on a previously set bool prop didn't fail as expected")
	}

	if val, err := reader.Prop("intProp"); err != nil {
		t.Error("Got unexpected error reading back generic prop", err)
	} else if val.(int) != 1 {
		t.Error("Reading back generic value didn't get right int. WAnted", 1, "got", val)
	}

}
