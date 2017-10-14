package playingcards

import (
	"github.com/jkomoros/boardgame"
	"testing"
)

func TestNewDeck(t *testing.T) {

	chest := boardgame.NewComponentChest(nil)

	deck := NewDeck(false)

	chest.AddDeck("cards", deck)

	if len(deck.Components()) != 52 {
		t.Error("We asked for no jokers but got wrong number of cards", len(deck.Components()))
	}

	g := deck.GenericComponent()

	if g.Values == nil {
		t.Error("Generic components had no values")
	}

	r, err := g.Values.Reader().EnumValProp("Rank")

	if err != nil {
		t.Error("Values on Reader had no Rank property")
	}

	if r.Value() != RankUnknown {
		t.Error("generic rank was not RankUnknown")
	}

	s, err := g.Values.Reader().EnumValProp("Suit")

	if err != nil {
		t.Error("Values on reader had no Suit property")
	}

	if s.Value() != SuitUnknown {
		t.Error("Generic suit was not suitunknown")
	}

	checkExpectedRun(deck, 0, t)

	withJokers := NewDeck(true)

	chest.AddDeck("jokers", withJokers)

	if len(withJokers.Components()) != 54 {
		t.Error("Deck with jokers had wrong number of cards:", len(withJokers.Components()))
	}

	multiDeck := NewDeckMulti(2, false)

	chest.AddDeck("multideck", multiDeck)

	if len(multiDeck.Components()) != 52*2 {
		t.Error("Got wrong number of components. Expected 52 *2, got", len(multiDeck.Components()))
	}

	checkExpectedRun(multiDeck, 0, t)
	checkExpectedRun(multiDeck, 52, t)

}

//Checks that the deck, at starting Index, has the 52 main cards in canonical order.
func checkExpectedRun(deck *boardgame.Deck, startingIndex int, t *testing.T) {

	if len(deck.Components()) < 52+startingIndex {
		t.Error("Deck didn't have enough items")
	}

	suits := []int{SuitSpades, SuitHearts, SuitClubs, SuitDiamonds}

	expectedRank := RankAce
	expectedSuitIndex := 0
	expectedSuit := suits[expectedSuitIndex]

	components := deck.Components()

	for i := startingIndex; i < (startingIndex + 52); i++ {
		card := components[i].Values.(*Card)

		if card.Rank.Value() != expectedRank {
			t.Error("Card", i, "had wrong rank. Wanted", expectedRank, "Got", card.Rank.Value())
		}

		if card.Suit.Value() != expectedSuit {
			t.Error("Card", i, "had wrong suit. Wanted", expectedSuit, "Got", card.Suit.Value())
		}

		expectedRank++
		if expectedRank > RankKing {
			expectedRank = RankAce
			expectedSuitIndex++
			if expectedSuitIndex < len(suits) {
				expectedSuit = suits[expectedSuitIndex]
			}
		}
	}

}
