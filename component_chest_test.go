package boardgame

import (
	"github.com/workfit/tester/assert"
	"io/ioutil"
	"reflect"
	"sort"
	"testing"
)

func TestComponentChestMarshal(t *testing.T) {
	game := testDefaultGame(t, false)

	chest := game.Manager().Chest()

	in, err := DefaultMarshalJSON(chest)

	assert.For(t).ThatActual(err).IsNil()

	golden, err := ioutil.ReadFile("testdata/component_chest.json")

	assert.For(t).ThatActual(err).IsNil()

	compareJSONObjects(in, golden, "Component chest did not match golden", t)

}

func TestComponentChestConstant(t *testing.T) {
	chest := newComponentChest(nil)

	err := chest.addConstant("int", 1)

	assert.For(t).ThatActual(err).IsNil()

	//Fails because already set
	err = chest.addConstant("int", 2)

	assert.For(t).ThatActual(err).IsNotNil()

	err = chest.addConstant("string", "foo")

	assert.For(t).ThatActual(err).IsNil()

	err = chest.addConstant("bool", true)

	assert.For(t).ThatActual(err).IsNil()

	//Illegal type
	err = chest.addConstant("float", 3.4)

	assert.For(t).ThatActual(err).IsNotNil()

	//chest not finishee so should be nil
	val := chest.Constant("int")

	assert.For(t).ThatActual(val).IsNil()

	chest.finish()

	assert.For(t).ThatActual(chest.ConstantNames()).Equals([]string{
		"bool",
		"int",
		"string",
	})

	assert.For(t).ThatActual(chest.Constant("int")).Equals(1)

	assert.For(t).ThatActual(chest.Constant("bool")).Equals(true)

	assert.For(t).ThatActual(chest.Constant("string")).Equals("foo")

	assert.For(t).ThatActual(chest.Constant("float")).IsNil()

}

func TestComponentInstanceIdentity(t *testing.T) {
	game := testDefaultGame(t, false)

	c := game.Manager().Chest().Deck("test").ComponentAt(0)

	one := c.ImmutableInstance(game.CurrentState())

	two := c.ImmutableInstance(game.CurrentState())

	if one != two {
		t.Error("Two equivalent components didn't match.")
	}
}

func TestComponentChest(t *testing.T) {

	chest := newComponentChest(nil)

	if chest.DeckNames() != nil {
		t.Error("We got a deck names array before we'd added anything")
	}

	deckOne := NewDeck()

	componentOne := &testingComponent{
		String:  "foo",
		Integer: 1,
	}

	deckOne.AddComponent(componentOne)

	componentTwo := &testingComponent{
		String:  "bar",
		Integer: 2,
	}

	deckOne.AddComponent(componentTwo)

	if deckOne.Components() == nil {
		t.Error("We got nil components before it was added to the chest, but now we're supposed to get them even before they're finished")
	}

	chest.addDeck("test", deckOne)

	componentValues := make([]ComponentValues, 2)

	for i, component := range deckOne.Components() {
		componentValues[i] = component.Values()
	}

	if !reflect.DeepEqual(componentValues, []ComponentValues{componentOne, componentTwo}) {
		t.Error("Deck gave back wrong items after being added to chest")
	}

	deckOne.AddComponent(&testingComponent{
		String:  "illegal",
		Integer: -1,
	})

	componentValues = make([]ComponentValues, 2)

	for i, component := range deckOne.Components() {
		componentValues[i] = component.Values()
	}

	if !reflect.DeepEqual(componentValues, []ComponentValues{componentOne, componentTwo}) {
		t.Error("Deck allowed itself to be mutated after it was added to chest")
	}

	if chest.DeckNames() != nil {
		t.Error("We got decknames before we called freeze")
	}

	if chest.Deck("test") != nil {
		t.Error("We got a deck back before freeze was called")
	}

	deckTwo := NewDeck()

	deckTwo.AddComponent(&testingComponent{
		String:  "another",
		Integer: 3,
	})

	chest.addDeck("other", deckTwo)

	chest.finish()

	c := deckTwo.ComponentAt(0)

	if c.Values().ContainingComponent() != c {
		t.Error("c.Values didn't have its containing component set")
	}

	chest.addDeck("shouldfail", deckOne)

	if chest.decks["shouldfail"] != nil {
		t.Fatal("We were able to add a deck after freezing")
	}

	sortedDeckNames := chest.DeckNames()

	sort.Strings(sortedDeckNames)

	expectedDeckNames := []string{"other", "test"}

	if !reflect.DeepEqual(sortedDeckNames, expectedDeckNames) {
		t.Error("Got unexpected decknames. got", sortedDeckNames, "wanted", expectedDeckNames)
	}

	if chest.Deck("test") != deckOne {
		t.Error("Got wrong value for deck one. Got", chest.Deck("test"), "wanted", deckOne)
	}

	if chest.Deck("other") != deckTwo {
		t.Error("Got wrong value for deck two. Got", chest.Deck("other"), "wanted", deckTwo)
	}

	if deckOne.name != "test" {
		t.Error("DeckOne didn't have its name set when added to the chest. Got", deckOne.name, "wanted test")
	}

	if deckTwo.name != "other" {
		t.Error("DeckTwo didn't have its name set when added to the chest. Got", deckTwo.name, "wanted other")
	}

	for i, c := range deckOne.Components() {
		if c.Deck() != deckOne {
			t.Error("At position", i, "deck name was not set correctly in component. Got", c.Deck(), "wanted", deckOne)
		}
		if c.DeckIndex() != i {
			t.Error("At position", i, "index was not set correctly in component. Got", c.DeckIndex(), "wanted", i)
		}
	}

}
