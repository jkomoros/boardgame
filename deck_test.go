package boardgame

import (
	"testing"
)

func TestDeckShadowComponent(t *testing.T) {

	manager := newTestGameManger()

	deck := manager.Chest().Deck("test")

	if deck == nil {
		t.Fatal("Couldn't find test deck")
	}

	c := deck.ComponentAt(emptyIndexSentinel)

	if c != nil {
		t.Error("ComponentAt didn't give nil for empty index sentitel", c)
	}

	c = deck.ComponentAt(-2)

	if c == nil {
		t.Error("Negative value gave nil")
	}

	altC := deck.ComponentAt(-2)

	if c != altC {
		t.Error("Two calls to same shadow gave different components")
	}

}
