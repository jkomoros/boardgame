package boardgame

import (
	"strings"
	"testing"

	"github.com/workfit/tester/assert"
)

func TestParseLegalPath(t *testing.T) {
	tests := []struct {
		name     string
		path     LegalPropPath
		wantErr  bool
		wantKind legalPathKind
		wantProp string
	}{
		{"game path", "game.DrawStack", false, pathGame, "DrawStack"},
		{"player path", "player.CardsLeftToReveal", false, pathPlayer, "CardsLeftToReveal"},
		{"players[*] path", "players[*].Stood", false, pathPlayersAll, "Stood"},
		{"move path", "move.CardIndex", false, pathMove, "CardIndex"},
		{"nested prop is not itself rejected by parse", "game.Sub.Field", false, pathGame, "Sub.Field"},

		{"wrong case kind", "Game.X", true, 0, ""},
		{"concrete player index", "players[0].X", true, 0, ""},
		{"unknown kind", "foo.X", true, 0, ""},
		{"empty prop with trailing dot", "game.", true, 0, ""},
		{"no dot at all", "game", true, 0, ""},
		{"totally empty", "", true, 0, ""},
		{"players wildcard wrong bracket contents", "players[1].X", true, 0, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseLegalPath(tc.path)
			if tc.wantErr {
				assert.For(t).ThatActual(err).IsNotNil()
				return
			}
			assert.For(t).ThatActual(err).IsNil()
			assert.For(t).ThatActual(got.kind).Equals(tc.wantKind)
			assert.For(t).ThatActual(got.prop).Equals(tc.wantProp)
		})
	}
}

func TestValidateLegalPath(t *testing.T) {
	manager := newTestGameManger(t)
	exampleState := manager.ExampleState()
	assert.For(t).ThatActual(exampleState).IsNotNil()

	moveReader := manager.ExampleMoveByName("Test").Reader()
	assert.For(t).ThatActual(moveReader).IsNotNil()

	t.Run("valid game path", func(t *testing.T) {
		err := validateLegalPath("game.DrawDeck", exampleState, nil)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("valid player path", func(t *testing.T) {
		err := validateLegalPath("player.Score", exampleState, nil)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("valid players[*] path", func(t *testing.T) {
		err := validateLegalPath("players[*].Score", exampleState, nil)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("valid move path", func(t *testing.T) {
		err := validateLegalPath("move.AString", exampleState, moveReader)
		assert.For(t).ThatActual(err).IsNil()
	})

	t.Run("unknown game property names path and property", func(t *testing.T) {
		err := validateLegalPath("game.TotallyNotAProp", exampleState, nil)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "game.TotallyNotAProp")).Equals(true)
		assert.For(t).ThatActual(strings.Contains(err.Error(), "TotallyNotAProp")).Equals(true)
	})

	t.Run("unknown player property names path and property", func(t *testing.T) {
		err := validateLegalPath("player.NopeNotReal", exampleState, nil)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "player.NopeNotReal")).Equals(true)
		assert.For(t).ThatActual(strings.Contains(err.Error(), "NopeNotReal")).Equals(true)
	})

	t.Run("unknown move property names path and property", func(t *testing.T) {
		err := validateLegalPath("move.NopeNotReal", exampleState, moveReader)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "move.NopeNotReal")).Equals(true)
		assert.For(t).ThatActual(strings.Contains(err.Error(), "NopeNotReal")).Equals(true)
	})

	t.Run("move path with nil moveReader is an error", func(t *testing.T) {
		err := validateLegalPath("move.AString", exampleState, nil)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("malformed path surfaces parse error", func(t *testing.T) {
		err := validateLegalPath("Game.X", exampleState, nil)
		assert.For(t).ThatActual(err).IsNotNil()
	})
}

func TestResolveLegalPath(t *testing.T) {
	game := testDefaultGame(t, false)
	state := game.CurrentState()
	assert.For(t).ThatActual(state).IsNotNil()

	manager := game.Manager()
	move := manager.ExampleMoveByName("Test")
	assert.For(t).ThatActual(move).IsNotNil()

	rs := move.ReadSetter()
	assert.For(t).ThatActual(rs.SetStringProp("AString", "hello")).IsNil()
	assert.For(t).ThatActual(rs.SetIntProp("ScoreIncrement", 7)).IsNil()
	assert.For(t).ThatActual(rs.SetBoolProp("ABool", true)).IsNil()
	assert.For(t).ThatActual(rs.SetPlayerIndexProp("TargetPlayerIndex", PlayerIndex(1))).IsNil()

	t.Run("game.X resolves a stack", func(t *testing.T) {
		val, propType, err := resolveLegalPath("game.DrawDeck", state, nil)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypeStack)
		_, ok := val.(ImmutableStack)
		assert.For(t).ThatActual(ok).Equals(true)
	})

	t.Run("game.X resolves an enum", func(t *testing.T) {
		val, propType, err := resolveLegalPath("game.Phase", state, nil)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypeEnum)
		assert.For(t).ThatActual(val).IsNotNil()
	})

	t.Run("player.X resolves current player's int prop", func(t *testing.T) {
		// testGameDelegate.BeginSetUp sets players[0].MovesLeftThisTurn = 1,
		// and the zero-value game.CurrentPlayer is 0, so the current player
		// is player 0.
		val, propType, err := resolveLegalPath("player.MovesLeftThisTurn", state, nil)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypeInt)
		assert.For(t).ThatActual(val).Equals(1)
	})

	t.Run("player.X resolves current player's bool prop", func(t *testing.T) {
		// testGameDelegate.BeginSetUp only sets players[2].IsFoo = true, so
		// current player (0) should read false here.
		val, propType, err := resolveLegalPath("player.IsFoo", state, nil)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypeBool)
		assert.For(t).ThatActual(val).Equals(false)
	})

	t.Run("move.X round-trips string", func(t *testing.T) {
		val, propType, err := resolveLegalPath("move.AString", state, move)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypeString)
		assert.For(t).ThatActual(val).Equals("hello")
	})

	t.Run("move.X round-trips int", func(t *testing.T) {
		val, propType, err := resolveLegalPath("move.ScoreIncrement", state, move)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypeInt)
		assert.For(t).ThatActual(val).Equals(7)
	})

	t.Run("move.X round-trips bool", func(t *testing.T) {
		val, propType, err := resolveLegalPath("move.ABool", state, move)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypeBool)
		assert.For(t).ThatActual(val).Equals(true)
	})

	t.Run("move.X round-trips PlayerIndex", func(t *testing.T) {
		val, propType, err := resolveLegalPath("move.TargetPlayerIndex", state, move)
		assert.For(t).ThatActual(err).IsNil()
		assert.For(t).ThatActual(propType).Equals(TypePlayerIndex)
		assert.For(t).ThatActual(val).Equals(PlayerIndex(1))
	})

	t.Run("move.X with nil move errors, not panics", func(t *testing.T) {
		_, _, err := resolveLegalPath("move.AString", state, nil)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("players[*] path cannot be resolved directly", func(t *testing.T) {
		_, _, err := resolveLegalPath("players[*].Score", state, nil)
		assert.For(t).ThatActual(err).IsNotNil()
	})

	t.Run("unknown property errors naming path and property", func(t *testing.T) {
		_, _, err := resolveLegalPath("game.NotAProp", state, nil)
		assert.For(t).ThatActual(err).IsNotNil()
		assert.For(t).ThatActual(strings.Contains(err.Error(), "game.NotAProp")).Equals(true)
		assert.For(t).ThatActual(strings.Contains(err.Error(), "NotAProp")).Equals(true)
	})

	t.Run("malformed path surfaces parse error, not panic", func(t *testing.T) {
		_, _, err := resolveLegalPath("Game.X", state, nil)
		assert.For(t).ThatActual(err).IsNotNil()
	})
}

// TestResolveLegalPathInvalidCurrentPlayer verifies that player.X resolution
// against an invalid, Observer, Admin, or Any current player returns an
// error rather than panicking. testGameDelegate.CurrentPlayerIndex just
// returns the concrete testGameState.CurrentPlayer field, so we can drive
// every case directly by setting that field to each special/invalid value.
func TestResolveLegalPathInvalidCurrentPlayer(t *testing.T) {
	tests := []struct {
		name  string
		value PlayerIndex
	}{
		{"one past last valid index", PlayerIndex(3)},
		{"observer", ObserverPlayerIndex},
		{"admin", AdminPlayerIndex},
		{"any", AnyPlayerIndex},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			game := testDefaultGame(t, false)
			state := game.CurrentState()

			gameState, _ := concreteStates(state)
			assert.For(t).ThatActual(gameState).IsNotNil()
			gameState.CurrentPlayer = tc.value

			_, _, err := resolveLegalPath("player.Score", state, nil)
			assert.For(t).ThatActual(err).IsNotNil()
		})
	}
}

func TestFacetSurvives(t *testing.T) {
	// Full 5 policy x 4 facet truth table (spec §6, task-2 brief):
	//   FacetValues:    Visible only
	//   FacetCount:     Visible, Order, Len
	//   FacetOccupancy: Visible, Order
	//   FacetOrder:     Visible, Order
	//   Nothing survives NonEmpty or Hidden.
	want := map[LegalFacet]map[Policy]bool{
		LegalFacetValues: {
			PolicyVisible:  true,
			PolicyOrder:    false,
			PolicyLen:      false,
			PolicyNonEmpty: false,
			PolicyHidden:   false,
		},
		LegalFacetCount: {
			PolicyVisible:  true,
			PolicyOrder:    true,
			PolicyLen:      true,
			PolicyNonEmpty: false,
			PolicyHidden:   false,
		},
		LegalFacetOccupancy: {
			PolicyVisible:  true,
			PolicyOrder:    true,
			PolicyLen:      false,
			PolicyNonEmpty: false,
			PolicyHidden:   false,
		},
		LegalFacetOrder: {
			PolicyVisible:  true,
			PolicyOrder:    true,
			PolicyLen:      false,
			PolicyNonEmpty: false,
			PolicyHidden:   false,
		},
	}

	facetNames := map[LegalFacet]string{
		LegalFacetValues:    "FacetValues",
		LegalFacetCount:     "FacetCount",
		LegalFacetOccupancy: "FacetOccupancy",
		LegalFacetOrder:     "FacetOrder",
	}
	policyNames := map[Policy]string{
		PolicyVisible:  "Visible",
		PolicyOrder:    "Order",
		PolicyLen:      "Len",
		PolicyNonEmpty: "NonEmpty",
		PolicyHidden:   "Hidden",
	}

	facets := []LegalFacet{LegalFacetValues, LegalFacetCount, LegalFacetOccupancy, LegalFacetOrder}
	policies := []Policy{PolicyVisible, PolicyOrder, PolicyLen, PolicyNonEmpty, PolicyHidden}

	for _, facet := range facets {
		for _, policy := range policies {
			facet, policy := facet, policy
			t.Run(facetNames[facet]+"/"+policyNames[policy], func(t *testing.T) {
				got := facetSurvives(policy, facet)
				assert.For(t).ThatActual(got).Equals(want[facet][policy])
			})
		}
	}
}
