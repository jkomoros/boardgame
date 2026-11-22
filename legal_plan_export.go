package boardgame

// LegalPlanSpecs exports, for definition tooling, the declarative-legality
// plan NewGameManager assembled for the move type named moveName
// (assembleLegalPlans in legal_plan.go): contributed base checks first, then
// authored WithLegalPreconditions specs, with WithoutLegalPrecondition
// suppressions already removed. It is additive and read-only: no engine path
// calls it.
//
// ok is false when no move type named moveName is installed (the name is
// matched without regard to case, as ExampleMoveByName does). declared is
// true when the move opted in and a plan was assembled; when it is false the
// move runs its frozen imperative Legal() chain and specs is nil. hasCustom
// is true when the plan ends with the opaque LegalCustom escape hatch, which
// is not a LegalSpec and so is not in specs (a ledger lists it last, as
// "custom").
//
// specs is a deep copy (Args and Sub included), so a caller cannot change the
// manager's plan through it.
func (g *GameManager) LegalPlanSpecs(moveName string) (specs []LegalSpec, declared bool, hasCustom bool, ok bool) {
	if g == nil {
		return nil, false, false, false
	}
	// moveTypeByName ignores case, as ExampleMoveByName does; the plan map is
	// keyed by the installed name.
	mType := g.moveTypeByName(moveName)
	if mType == nil {
		return nil, false, false, false
	}
	plan := g.legalPlans[mType.Name()]
	if plan == nil {
		return nil, false, false, true
	}
	return copyLegalSpecs(plan.specs), true, plan.custom != nil, true
}

func copyLegalSpecs(specs []LegalSpec) []LegalSpec {
	if specs == nil {
		return nil
	}
	result := make([]LegalSpec, len(specs))
	for i, spec := range specs {
		result[i] = spec
		if spec.Args != nil {
			result[i].Args = append([]string(nil), spec.Args...)
		}
		result[i].Sub = copyLegalSpecs(spec.Sub)
	}
	return result
}
