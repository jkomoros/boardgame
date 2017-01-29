package boardgame

import (
	"reflect"
	"sort"
	"testing"
)

func TestComponentChest(t *testing.T) {

	chest := &ComponentChest{}

	if chest.DeckNames() != nil {
		t.Error("We got a deck names array before we'd added anything")
	}

	deckOne := &Deck{}

	deckTwo := &Deck{}

	chest.AddDeck("test", deckOne)

	if chest.DeckNames() != nil {
		t.Error("We got decknames before we called freeze")
	}

	if chest.Deck("test") != nil {
		t.Error("We got a deck back before freeze was called")
	}

	chest.AddDeck("other", deckTwo)

	chest.Finish()

	chest.AddDeck("shoulfail", deckOne)

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

}
