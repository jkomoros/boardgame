package playingcards

import (
	"github.com/jkomoros/boardgame"
	"testing"
)

func TestNewDeck(t *testing.T) {

	chest := boardgame.NewComponentChest()

	deck := NewDeck(false)

	chest.AddDeck("cards", deck)

	if len(deck.Components()) != 52 {
		t.Error("We asked for no jokers but got wrong number of cards", len(deck.Components()))
	}

	withJokers := NewDeck(true)

	chest.AddDeck("jokers", withJokers)

	if len(withJokers.Components()) != 54 {
		t.Error("Deck with jokers had wrong number of cards:", len(withJokers.Components()))
	}

}
