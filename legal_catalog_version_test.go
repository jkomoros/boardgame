package boardgame

import (
	"strings"
	"testing"

	"github.com/workfit/tester/assert"
)

// TestLegalCatalogVersionIsPositive pins LegalCatalogVersion (legal_types.go)
// as a real, non-zero version stamp — Task 10's info response ships it
// verbatim, and a client comparing against "no catalog at all" needs it to
// never collide with a zero/unset value.
func TestLegalCatalogVersionIsPositive(t *testing.T) {
	assert.For(t, "catalog version").ThatActual(LegalCatalogVersion > 0).Equals(true)
}

// TestLegalCatalogVersionIsCompletenessRoundValue pins the EXACT value
// (completeness-round design spec §6, Task 8's single version bump): 2,
// bumped from 1 because this round's new predicate names (stackCount,
// stackEmpty, stackNotEmpty, propEquals, propNotEquals, componentAbsentAt),
// playerBool's optional second arg, and the players[move.<Field>].<Prop>
// path grammar kind are all vocabulary an older (v1) client's evaluator
// could not safely interpret — see legal_types.go's doc comment for the
// full "v1 -> v2" rationale. Unlike TestLegalCatalogVersionIsPositive
// (which stays true forever), this test is DELIBERATELY exact: it must be
// updated by hand on any future bump, forcing a conscious decision rather
// than a silent drift.
func TestLegalCatalogVersionIsCompletenessRoundValue(t *testing.T) {
	assert.For(t, "catalog version").ThatActual(LegalCatalogVersion).Equals(2)
}

// TestComponentChestMarshalIncludesLegalTemplates pins that the chest JSON
// gains a "LegalTemplates" key (design spec §6: "shipped to the client
// inside the chest JSON, exactly the channel enums already ride") once the
// owning manager has a non-empty merged legal template table.
func TestComponentChestMarshalIncludesLegalTemplates(t *testing.T) {
	game := testDefaultGame(t, false)
	manager := game.Manager()
	manager.legalTemplateTable = map[string]string{"some.key": "some rendered body"}

	data, err := DefaultMarshalJSON(manager.Chest())
	assert.For(t, "marshal error").ThatActual(err).IsNil()

	json := string(data)
	assert.For(t, "has LegalTemplates key").ThatActual(strings.Contains(json, `"LegalTemplates"`)).Equals(true)
	assert.For(t, "has template body").ThatActual(strings.Contains(json, "some rendered body")).Equals(true)
}

// TestComponentChestMarshalOmitsLegalTemplatesWhenEmpty pins the omitempty
// side: a game type with no declarative-legality moves at all (an empty
// merged template table, this package's own test fixture's normal state —
// see TestComponentChestMarshal's golden fixture, which predates
// LegalTemplates and must stay byte-identical) must not gain the key.
func TestComponentChestMarshalOmitsLegalTemplatesWhenEmpty(t *testing.T) {
	game := testDefaultGame(t, false)
	manager := game.Manager()
	// This package's own test fixture never imports package moves (see
	// legal_evaluable_test.go's package doc), so legalTemplateTable is nil
	// here already; assert explicitly since that's the behavior under test.
	assert.For(t, "empty by default").ThatActual(len(manager.legalTemplateTable)).Equals(0)

	data, err := DefaultMarshalJSON(manager.Chest())
	assert.For(t, "marshal error").ThatActual(err).IsNil()

	assert.For(t, "omits LegalTemplates key").ThatActual(strings.Contains(string(data), "LegalTemplates")).Equals(false)
}
