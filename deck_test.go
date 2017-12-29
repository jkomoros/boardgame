package boardgame

import (
	"testing"
)

type testShadowValues struct {
	BaseComponentValues
	Message string
}

func (t *testShadowValues) Reader() PropertyReader {
	return getDefaultReader(t)
}

func TestDeckShadowComponent(t *testing.T) {

	manager := newTestGameManger(t)

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

	c = deck.ComponentAt(0)

	if c != deck.Components()[0] {
		t.Error("Deck.componenAt didn't return correct component for normal component")
	}

	c = deck.GenericComponent()

	if c == nil {
		t.Error("Generic Component returned nil")
	}

	altC := deck.GenericComponent()

	if c != altC {
		t.Error("Repated calls to generic component didn't return the same thign.")
	}

}
