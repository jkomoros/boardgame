package boardgame

import (
	"testing"

	"github.com/workfit/tester/assert"
)

/*
This file tests GameManager.LegalEvaluateLedger and GameManager.LegalRenderVerdict
(legal_plan.go) — the two pieces of engine-internal plumbing Task 10's server
ledger is built on. It follows legal_plan_test.go's own low-level pattern
(hand-built *legalPlan / legalPlanTestPred fixtures on a bare &GameManager{},
no real delegate/move needed) rather than standing up a full game, since
these methods only touch g.legalPlans/g.legalTemplateTable and
legalPlan.evaluate — already covered end-to-end by
TestLegalPlanEvaluateFullLedger.
*/

func TestLegalEvaluateLedgerOptedIn(t *testing.T) {
	g := &GameManager{
		legalPlans: map[string]*legalPlan{
			"Move": {
				fieldIndependent: []*LegalPredicate{legalPlanTestPred("indepFail", LegalFail)},
				fieldDependent:   []*LegalPredicate{legalPlanTestPred("depPass", LegalPass, moveReadOf("move.X"))},
			},
		},
		legalTemplateTable: map[string]string{"indepFail": "custom body for indepFail"},
	}

	verdict, entries, opted := g.LegalEvaluateLedger("Move", nil, nil, 0)

	assert.For(t, "opted").ThatActual(opted).Equals(true)
	assert.For(t, "verdict outcome").ThatActual(verdict.Outcome).Equals(LegalFail)
	assert.For(t, "entries len").ThatActual(len(entries)).Equals(2)

	assert.For(t, "entry0 name").ThatActual(entries[0].Name).Equals("indepFail")
	assert.For(t, "entry0 fieldDependent").ThatActual(entries[0].FieldDependent).Equals(false)
	assert.For(t, "entry0 verdict").ThatActual(entries[0].Verdict.Outcome).Equals(LegalFail)

	assert.For(t, "entry1 name").ThatActual(entries[1].Name).Equals("depPass")
	// FieldDependent marks the provisional flag the server ledger surfaces
	// (design spec §6: "provisional: true marks field-dependent verdicts").
	assert.For(t, "entry1 fieldDependent").ThatActual(entries[1].FieldDependent).Equals(true)
	assert.For(t, "entry1 serializable").ThatActual(entries[1].Serializable).Equals(true)
	assert.For(t, "entry1 reads").ThatActual(len(entries[1].Reads)).Equals(1)

	rendered := g.LegalRenderVerdict(verdict)
	assert.For(t, "rendered").ThatActual(rendered).Equals("custom body for indepFail")
}

func TestLegalEvaluateLedgerNotOptedIn(t *testing.T) {
	// No plans at all.
	g := &GameManager{}
	verdict, entries, opted := g.LegalEvaluateLedger("Anything", nil, nil, 0)
	assert.For(t, "opted").ThatActual(opted).Equals(false)
	assert.For(t, "verdict").ThatActual(verdict).Equals(LegalVerdict{})
	assert.For(t, "entries").ThatActual(len(entries)).Equals(0)

	// Plans exist, but not for this move name -- the pure-sugar seam: the
	// caller must fall back to the move's own frozen imperative Legal().
	g.legalPlans = map[string]*legalPlan{"Other": {}}
	verdict, entries, opted = g.LegalEvaluateLedger("Anything", nil, nil, 0)
	assert.For(t, "opted (wrong name)").ThatActual(opted).Equals(false)
	assert.For(t, "verdict (wrong name)").ThatActual(verdict).Equals(LegalVerdict{})
	assert.For(t, "entries (wrong name)").ThatActual(len(entries)).Equals(0)
}

func TestLegalRenderVerdictPassIsEmptyString(t *testing.T) {
	g := &GameManager{}
	assert.For(t, "pass").ThatActual(g.LegalRenderVerdict(LegalVerdict{Outcome: LegalPass})).Equals("")
}

func TestLegalRenderVerdictMatchesLegalEvaluatePlan(t *testing.T) {
	// Proves LegalRenderVerdict(ledgerVerdict) reproduces EXACTLY the text
	// LegalEvaluatePlan's hot-path short-circuit evaluation already attaches
	// via AttachTable -- this is what makes the server ledger's
	// LegalForPlayerError byte-identical to move.Legal()'s own error text
	// for an opted-in move (Task 10's requirement).
	plan := &legalPlan{
		fieldIndependent: []*LegalPredicate{
			legalPlanTestPred("first", LegalPass),
			legalPlanTestPred("boom", LegalFail),
		},
	}
	g := &GameManager{
		legalPlans:         map[string]*legalPlan{"Move": plan},
		legalTemplateTable: map[string]string{"boom": "boom rendered"},
	}

	_, hotErr := g.LegalEvaluatePlan("Move", nil, nil, 0)
	verdict, _, opted := g.LegalEvaluateLedger("Move", nil, nil, 0)
	ledgerErr := g.LegalRenderVerdict(verdict)

	assert.For(t, "opted").ThatActual(opted).Equals(true)
	assert.For(t, "hot path error").ThatActual(hotErr).IsNotNil()
	if hotErr != nil {
		assert.For(t, "hot path text").ThatActual(hotErr.Error()).Equals("boom rendered")
	}
	assert.For(t, "ledger text").ThatActual(ledgerErr).Equals("boom rendered")
}

func TestLegalRenderVerdictFallsBackToBareKeyWithoutTable(t *testing.T) {
	// No legalTemplateTable set (nil map): RenderLegalMessage's own
	// documented fallback is to render the bare template key.
	g := &GameManager{}
	v := LegalVerdict{Outcome: LegalFail, Message: &LegalMessage{Template: "some.key"}}
	assert.For(t, "bare key fallback").ThatActual(g.LegalRenderVerdict(v)).Equals("some.key")
}
