package boardgame

import (
	"errors"
	"testing"
)

func TestLegalResidueConstructor(t *testing.T) {
	ctor := NewLegalResidueConstructor("someBaseCheck", func(ctx LegalContext) error {
		return errors.New("the frozen text")
	})
	if !ctor.IsResidue() {
		t.Fatal("a residue constructor must report IsResidue")
	}
	if (&LegalPredicateConstructor{Name: "plain"}).IsResidue() {
		t.Fatal("an ordinary constructor must not report IsResidue")
	}

	if _, err := ctor.Constructor(LegalSpec{Name: "someBaseCheck", Args: []string{"x"}}, nil, nil); err == nil {
		t.Error("a framework check takes no args")
	}
	if _, err := ctor.Constructor(LegalSpec{Name: "someBaseCheck", Message: "key"}, nil, nil); err == nil {
		t.Error("a framework check renders its own text and cannot take WithMessage")
	}

	pred, err := ctor.Constructor(LegalSpec{Name: "someBaseCheck"}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !pred.residue || !legalPredicateOpaque(pred) || pred.Serializable() || pred.ClientEvaluable {
		t.Fatalf("a residue predicate is opaque, unserializable and not client-evaluable: %+v", pred)
	}
	verdict := pred.Evaluate(LegalContext{})
	if verdict.Outcome != LegalFail || verdict.Message == nil || verdict.Message.Template != "the frozen text" {
		t.Fatalf("the frozen error text must render as itself, got %+v", verdict)
	}

	// Opacity is inherited through a compositor: its own Reads are only the
	// union of its children's, so a nested opaque child must still keep the
	// whole predicate out of the field-independent memo.
	compositor := &LegalPredicate{Name: legalAnyCompositorName, Sub: []*LegalPredicate{{Name: "leaf"}, pred}}
	if !legalPredicateOpaque(compositor) {
		t.Fatal("a compositor with an opaque child must be treated as opaque")
	}
	plan := buildLegalPlanFromPredicates("m", []*LegalPredicate{compositor}, nil, nil)
	if len(plan.fieldIndependent) != 0 || !plan.ordered[0].fieldDependent {
		t.Fatal("an opaque predicate must never be bucketed field-independent (memoized)")
	}
}

func TestLegalValidateResidueAtoms(t *testing.T) {
	registry := map[string]*LegalPredicateConstructor{
		"baseCheck": NewLegalResidueConstructor("baseCheck", func(LegalContext) error { return nil }),
		"plain":     {Name: "plain"},
	}
	contributed := []LegalSpec{{Name: "plain"}, {Name: "baseCheck"}}

	cases := []struct {
		name         string
		authored     []LegalSpec
		suppressions []string
		wantErr      bool
	}{
		{"untouched", nil, nil, false},
		{"suppress plain atom", nil, []string{"plain"}, false},
		{"suppress residue without re-authoring", nil, []string{"baseCheck"}, true},
		{"suppress and re-author residue", []LegalSpec{{Name: "plain"}, {Name: "baseCheck"}}, []string{"baseCheck"}, false},
		{"author residue without suppressing", []LegalSpec{{Name: "baseCheck"}}, nil, true},
		{"nest residue in any", []LegalSpec{{Name: legalAnyCompositorName, Sub: []LegalSpec{{Name: "plain"}, {Name: "baseCheck"}}}}, nil, true},
	}
	for _, tc := range cases {
		err := legalValidateResidueAtoms("m", contributed, tc.authored, tc.suppressions, registry)
		if (err != nil) != tc.wantErr {
			t.Errorf("%v: err = %v, wantErr %v", tc.name, err, tc.wantErr)
		}
	}
}
