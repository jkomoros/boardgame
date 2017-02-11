package chestbuilder

import (
	"github.com/jkomoros/boardgame"
	"io/ioutil"
	"testing"
)

type CardComponent struct {
	Suit int
	Name string
}

type TestChest struct {
	Cards []*CardComponent
}

func (c *CardComponent) Props() []string {
	return boardgame.PropertyReaderPropsImpl(c)
}

func (c *CardComponent) Prop(name string) interface{} {
	return boardgame.PropertyReaderPropImpl(c, name)
}

func TestChestBuilder(t *testing.T) {

	blob, err := ioutil.ReadFile("test/input.json")
	if err != nil {
		t.Fatal("Couldn't load file: ", err)
	}

	if blob == nil {
		t.Fatal("Didn't load blob")
	}

	container := &TestChest{}

	chest, err := FromConfig(blob, container)

	if err != nil {
		t.Fatal(err)
	}

	if chest == nil {
		t.Error("No chest returned")
	}

	deck := chest.Deck("Cards")

	if deck == nil {
		t.Fatal("Chest had no deck named cards")
	}

	if len(deck.Components()) != 2 {
		t.Error("Got wrong length of deck")
	}

	card := deck.Components()[0]

	v := card.Values.(*CardComponent)

	if v.Name != "Bob" {
		t.Error("Got wrong component in first position. Expected 'bob' got", v.Name)
	}

	if v.Suit != 3 {
		t.Error("Got wrong component in first positon. Expected '3', got", v.Suit)
	}
}
