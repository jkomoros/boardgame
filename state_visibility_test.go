package boardgame

import (
	"encoding/json"
	"testing"
)

func TestViewerVisibilityMatchesLegalityFacets(t *testing.T) {
	game := testDefaultGame(t, false)
	for _, policy := range []string{"visible", "order", "len", "nonempty", "hidden"} {
		t.Run(policy, func(t *testing.T) {
			(&sanitizationTestConfig{Game: map[string]string{"DrawDeck": policy}}).Install(game.Manager())
			view, err := game.JSONForPlayer(0, nil)
			if err != nil {
				t.Fatal(err)
			}
			blob, err := json.Marshal(view)
			if err != nil {
				t.Fatal(err)
			}
			var wire struct {
				CurrentState struct{ Visibility StateVisibility }
			}
			if err := json.Unmarshal(blob, &wire); err != nil {
				t.Fatal(err)
			}
			for _, facet := range stateVisibilityFacets {
				want := LegalReadEvaluable(game.CurrentState(), 0, LegalRead{Path: "game.DrawDeck", Facet: facet})
				got := containsVisibilityFacet(wire.CurrentState.Visibility.Game["DrawDeck"], visibilityFacetName(facet))
				if got != want {
					t.Fatalf("facet %v: visibility=%v legality=%v", facet, got, want)
				}
			}
		})
	}
}

func TestViewerVisibilityDistinguishesDefaultsAndStaysRestricted(t *testing.T) {
	game := testDefaultGame(t, false)
	(&sanitizationTestConfig{Player: map[string]string{"EnumVal": "hidden"}}).Install(game.Manager())
	view, err := game.CurrentState().SanitizedForPlayer(0)
	if err != nil {
		t.Fatal(err)
	}
	if !PropertyFacetAvailable(view.ImmutablePlayerStates()[0], "EnumVal", LegalFacetValues) {
		t.Fatal("owner's real default enum must remain known")
	}
	if PropertyFacetAvailable(view.ImmutablePlayerStates()[1], "EnumVal", LegalFacetValues) {
		t.Fatal("another player's hidden enum must not be known")
	}
	if PropertyFacetAvailable(view.ImmutablePlayerStates()[0], "Typo", LegalFacetValues) {
		t.Fatal("unknown property must fail closed")
	}
	copy, err := view.Copy(true)
	if err != nil {
		t.Fatal(err)
	}
	// Re-sanitizing for the other owner cannot recover that owner's old value.
	copy, err = copy.SanitizedForPlayer(1)
	if err != nil {
		t.Fatal(err)
	}
	if PropertyFacetAvailable(copy.ImmutablePlayerStates()[1], "EnumVal", LegalFacetValues) {
		t.Fatal("resanitization restored knowledge without restoring a value")
	}
	for _, st := range []ImmutableState{game.CurrentState(), view} {
		var stored map[string]json.RawMessage
		if err := json.Unmarshal(st.(*state).StorageRecord(), &stored); err != nil {
			t.Fatal(err)
		}
		if _, exists := stored["Visibility"]; exists {
			t.Fatal("viewer metadata entered durable state")
		}
	}
}
