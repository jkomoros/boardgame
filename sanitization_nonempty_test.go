package boardgame

import (
	"encoding/json"
	"testing"
)

// NonEmpty may disclose emptiness, but neither membership, count, nor history.
// Compare the entire wire object, including the animator's identity metadata.
func TestNonEmptyStackViewsAreIndistinguishable(t *testing.T) {
	game := testDefaultGame(t, false)
	st := game.CurrentState().(*state)
	deck := game.Manager().Chest().Deck("test")
	for _, sized := range []bool{false, true} {
		var previous string
		for count := 1; count <= 3; count++ {
			var stack Stack
			if sized {
				stack = deck.NewSizedStack(5)
			} else {
				stack = deck.NewStack(5)
			}
			stack.setState(st)
			for i := 0; i < count; i++ {
				stack.insertComponentAt(i, deck.ComponentAt(i).Instance(st))
			}
			stack.applySanitizationPolicy(PolicyNonEmpty)
			if len(stack.IDsLastSeen()) != 0 {
				t.Fatalf("sized=%v count=%d retained identity history", sized, count)
			}
			if stack.NumComponents() != 1 {
				t.Fatalf("nonempty stack did not collapse to one generic component")
			}
			blob, err := json.Marshal(stack)
			if err != nil {
				t.Fatal(err)
			}
			if count > 1 && string(blob) != previous {
				t.Fatalf("sized=%v disclosed count or membership: %s != %s", sized, blob, previous)
			}
			previous = string(blob)
		}
	}
}
