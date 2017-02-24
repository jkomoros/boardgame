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

type TokenComponent struct {
	Color int
}

type RepeatTokenComponent struct {
	Repeat    int
	Component *TokenComponent
}

type TestChest struct {
	Cards  []*CardComponent
	Tokens []*RepeatTokenComponent
}

func (c *CardComponent) Reader() boardgame.PropertyReader {
	return boardgame.NewDefaultReader(c)
}

func (t *TokenComponent) Reader() boardgame.PropertyReader {
	return boardgame.NewDefaultReader(t)
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

	deck = chest.Deck("Tokens")

	if deck == nil {
		t.Fatal("Chest had no deck named tokens")
	}

	if len(deck.Components()) != 5 {
		t.Error("Tokens had wrong length. Wanted 5, got", len(deck.Components()))
	}

	//TODO: verify that it's three of the first and two of the rest.

	for i, component := range deck.Components() {
		c := component.Values.(*TokenComponent)
		if i < 3 {
			if c.Color != 4 {
				t.Error("Expected component", i, "to have color 4. Got", c.Color)
			}
		} else {
			if c.Color != 2 {
				t.Error("Expected component", i, "to have color 2. Got", c.Color)
			}
		}
	}
}
