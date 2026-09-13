package legal_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/jkomoros/boardgame"
	"github.com/jkomoros/boardgame/legal"
)

/*
boardgame.LegalCatalogVersion is the stamp a client reads to decide whether its
own bundled catalog covers this game's ledger. Its whole documented job is
graceful skew: "a client with an older catalog treats unknown predicate names as
evaluable: false and defers to server verdicts". Its own version history cites
"new predicate names" as the reason for the v1 -> v2 bump.

So a predicate name added without a bump is the stamp claiming a compatibility
that does not hold, and that is exactly what happened: mayMoveFirstTo and
mayMoveFirstToSlot shipped on top of a v4 stamp. Nothing caught it, because the
catalog's completeness test (TestDefaultConstructors) and the version's pin
(boardgame's TestLegalCatalogVersionIncludesAdminPolicy) each knew only their
own half.

catalogNamesByVersion is the missing link: it freezes the exact predicate
vocabulary each catalog version stamps. Adding a predicate now fails this test
until a version entry is added for it, and bumping the version fails it until
the entry exists -- which is the conscious decision the stamp is for.
*/
var catalogNamesByVersion = map[int][]string{
	5: {
		"allActivePlayers",
		"componentAbsentAt",
		"componentPresentAt",
		"componentPresentAtKey",
		"componentPropEqualsCurrentPlayer",
		"inPhase",
		"mayMoveAllTo",
		"mayMoveCountTo",
		"mayMoveFirstTo",
		"mayMoveFirstToSlot",
		"mayMoveFixedCountTo",
		"mayMoveTo",
		"mayMoveToSlot",
		"maySwapComponents",
		"maySwapComponentsByKey",
		"playerBool",
		"playerBoolAt",
		"propAtLeast",
		"propCompare",
		"propEquals",
		"propNotEquals",
		"proposerIsCurrentPlayer",
		"proposerIsPlayerFromMove",
		"revealableCardAt",
		"stackConstraints",
		"stackCount",
		"stackEmpty",
		"stackNotEmpty",
	},
	6: {
		"allActivePlayers",
		"componentAbsentAt",
		"componentPresentAt",
		"componentPresentAtKey",
		"componentPropEquals",
		"componentPropEqualsCurrentPlayer",
		"componentPropNotEquals",
		"componentPropsEqual",
		"componentPropsNotEqual",
		"inPhase",
		"mayMoveAllTo",
		"mayMoveCountTo",
		"mayMoveFirstTo",
		"mayMoveFirstToSlot",
		"mayMoveFixedCountTo",
		"mayMoveTo",
		"mayMoveToSlot",
		"maySwapComponents",
		"maySwapComponentsByKey",
		"playerBool",
		"playerBoolAt",
		"propAtLeast",
		"propCompare",
		"propEquals",
		"propNotEquals",
		"proposerIsCurrentPlayer",
		"proposerIsPlayerFromMove",
		"revealableCardAt",
		"stackConstraints",
		"stackCount",
		"stackEmpty",
		"stackNotEmpty",
	},
}

func TestCatalogVocabularyMatchesCatalogVersion(t *testing.T) {

	want, ok := catalogNamesByVersion[boardgame.LegalCatalogVersion]
	if !ok {
		t.Fatalf("boardgame.LegalCatalogVersion is %d, but catalogNamesByVersion records no vocabulary for it. "+
			"A version bump must record the predicate names that version stamps; see legal_types.go's history.",
			boardgame.LegalCatalogVersion)
	}

	var got []string
	for _, constructor := range legal.DefaultConstructors() {
		got = append(got, constructor.Name)
	}
	sort.Strings(got)

	sorted := append([]string(nil), want...)
	sort.Strings(sorted)

	if strings.Join(got, ",") != strings.Join(sorted, ",") {
		t.Errorf("the catalog vocabulary no longer matches what LegalCatalogVersion %d stamps.\n got: %v\nwant: %v\n"+
			"A new or removed predicate name is a catalog change an older client's evaluator cannot safely interpret: "+
			"bump LegalCatalogVersion, add a history entry in legal_types.go, update the exact pin in "+
			"legal_catalog_version_test.go, and record the new vocabulary here.",
			boardgame.LegalCatalogVersion, got, sorted)
	}
}
