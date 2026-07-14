# Critique dispositions — declarative legality spec rev 2.1

Every Blocking/Important finding from the four spec critiques, with its
disposition in the revised spec. (Process record, not normative.)

## spec-critique-goapi.md (Go API & idiom)

| Finding | Disposition |
|---|---|
| B1 zero `Verdict` reads as legal | ✅ `Outcome` zero value is `outcomeInvalid`; fails closed (§1) |
| B2 5-method interface vs repo func-type precedent | ✅ `Predicate` is a struct with an `Evaluate` func field (§1) |
| B3 `.WithMessage` doesn't type-check | ✅ `Spec.Message` field + `WithMessage` returns `Spec` by value (§1) |
| I1 `Bindings map[string]any` JSON lossiness | ✅ `BindingValue` tri-typed union (§1) |
| I2 template keys never validated | ✅ boot-time validation of every referenced key (§1, §6) |
| I3 `Verdict.Error()` nil-interface trap + Friendly mapping | ✅ rendering specified at the `Legal()` boundary (§3); implementation detail: concrete error type, never a typed-nil interface |
| I4 nil `Move` panic burden | ✅ runtime guard converts undeclared move reads to `Unknown`, never a panic (§1) |
| I5 opaque-override detection not mechanizable | ✅ boot-time behavioral probe (§ prime guarantee rule 4) |

## spec-critique-sugar.md (purely-sugar guarantee)

| Finding | Disposition |
|---|---|
| Embed+override+super-call breaks silently (double proposer check, Cost reorder) | ✅ imperative chain FROZEN (rule 1); plan evaluation only via opt-in; **Cost-sort removed from v1** — declaration-order evaluation (§4) |
| Opaque/no-inPhase moves vanish from fixup/move-forms | ✅ phase-agnostic bucket + superset property test (§5, §9) |
| `ConfigurePredicateConstructors` as GameDelegate member = compile break | ✅ optional interface via type-assertion (§1) |
| LegalCustom + wholesale Legal ambiguity | ✅ probe detects orphaned declarations; super-calling overrides legitimized (rule 4) |
| LegalForPlayerError string drift for legacy clients | ✅ un-migrated moves byte-identical (rule 1); migrated moves keep historical first-failure via declaration order (§8) |
| `Move: nil` nil-derefs on under-declared reads | ✅ Unknown-not-panic guard (§1) |

## spec-critique-contact.md (codebase contact)

| Finding | Disposition |
|---|---|
| Composition seam assumes 4-type chain; ~24 move types carry real logic | ✅ v1 seam = Default + CurrentPlayer only; all other framework types opaque, opt-in fails at boot naming the base type (§2) |
| ForceFinishTurn inherit-nothing pattern | ✅ `WithoutPrecondition` per stable name (§2) |
| ApplyUntil/RoundRobin negated semantics vs anti-`not` rule | ✅ stay opaque in v1 (§1 anti-tarpit) |
| "2 genuinely custom" wrong (4 hard + ~11 quantifier) | ✅ corrected survey (§8); `AllActivePlayers` quantifier in catalog v1 |
| `spaceIsBlack` has no registry path | ✅ game-registered predicates through the same registry (§8) |
| No phase-agnostic bucket; proposer check is field-dependent | ✅ both corrected (§5, §4) |

## spec-critique-client.md (client/TS future)

| Finding | Disposition |
|---|---|
| Template table has no home | ✅ `TemplateConfigurer` on delegate, shipped in chest JSON beside Enums (§6) |
| Go↔TS conformance corpus + catalog skew | ✅ shared JSON corpus as a v1 deliverable + `catalogVersion` stamp with graceful degradation (§6) |
| "needs Visible" contradicts memory-is-evaluable | ✅ facet-based evaluability (`Read.Facet`: values/count/occupancy/order vs sanitization policy) (§1, §6) |
| `evaluable:false` bindings leak hidden state (#693) | ✅ no state-derived bindings on inevaluable predicates (§6) |
| Field-dependent verdicts presented as definitive | ✅ `provisional: true` marking (§6) |
| `any` with mixed-evaluability children | ✅ Kleene rule: evaluable iff all children are (§6) |

## Cross-cutting decisions made while responding

- **Cost-sort removed from v1** (evaluation = declaration order): least-surprise
  for authors, no message churn on migration; `Cost` retained as lint/docs
  metadata and future opt-in.
- **Dirty-tracking stays deferred** (was already in rev 2): correctness risk
  priced by design B itself and both panel critics; v1 invalidates per version.
- The user's three goals are the acceptance frame: (1) developer ergonomics,
  (2) client-side proactive errors/disables cheaply, (3) correctness/no
  footguns with tooling.
