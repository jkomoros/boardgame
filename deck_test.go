package boardgame

import (
	"testing"
)

type testShadowValues struct {
	Message string
}

func (t *testShadowValues) Reader() PropertyReader {
	return getDefaultReader(t)
}

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

	v := c.Values

	if v == nil {
		t.Error("Values was nil")
	}

	if _, err := v.Reader().StringProp("Message"); err != nil {
		t.Error("Values didn't have Message")
	}

	altC := deck.ComponentAt(-2)

	if c != altC {
		t.Error("Two calls to same shadow gave different components")
	}

	c = deck.ComponentAt(0)

	if c != deck.Components()[0] {
		t.Error("Deck.componenAt didn't return correct component for normal component")
	}

	c = deck.GenericComponent()

	altC = deck.ComponentAt(genericComponentSentinel)

	if c != altC {
		t.Error("GEneric component wasn't the expected component index")
	}

	altC = deck.GenericComponent()

	if c != altC {
		t.Error("Repated calls to generic component didn't return the same thign.")
	}

}
