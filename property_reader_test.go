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

}
