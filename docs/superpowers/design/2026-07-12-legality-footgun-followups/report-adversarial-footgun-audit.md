# Exploration report: adversarial footgun audit of the declarative-legality system

> Provenance: sub-agent adversarial audit, 2026-07-12, session on branch
> `declarative-legality-design`. Checked in verbatim (with post-fix status
> annotations in [BATCH: ...] brackets) as the source findings for
> `docs/superpowers/plans/2026-07-12-legality-footgun-batch.md`, which
> implemented the fixes. File:line citations reference the branch state at
> commit d8716722 (pre-batch).

Scope: things that **compile, boot cleanly, and then do the wrong thing or confuse**. Every finding below was verified against the code on this branch. Ranked by likelihood x confusion cost.

---

## F1 — HIGH: Super-calling `Legal()` override residue is invisible to the server ledger (button-enabled / move-rejected desync)

[BATCH: FIXED in 1d82dfba — booleans now derive from ground-truth `move.Legal()` calls; ledger is advisory. Memo-cost comment corrected in 63519c73.]

**Scenario (the "forgot to delete Legal()" super-calling variant):** a move opts in via `WithPreconditions` but keeps a `Legal()` override that super-calls and then runs leftover imperative checks. This is explicitly blessed: the boot probe passes it (`probeLegalReachable`, `legal_plan.go:612-639`), spec rule 4 calls it "fully supported", and `legal/doc.go:153-158` calls it "fully compatible".

**The verified hole:** for an opted-in move, the server never calls `move.Legal()` at all when building move forms. `generateFormsWithLegality` (`server/api/main.go:1662-1688`) dispatches to `legalFormFromLedger` (`server/api/main.go:1751-1768`), which computes `LegalForPlayer`, `LegalForPlayerError`, `LegalForAnyone`, and the `Preconditions` ledger **from the plan alone** (`LegalEvaluateLedger` + a hot-path admin `LegalEvaluatePlan`). But `ProposeMove` runs the real `move.Legal()`, which is plan **plus** the override's residue. If the residue rejects while the plan passes: ledger says legal, all precondition entries `pass`, button enabled — proposal rejected on click. No boot check, no ledger entry (unlike `LegalCustom`, which appears as a `custom`/`unknown` entry).

The codebase demonstrably treats exactly this failure class as boot-error-worthy — the `WithoutPrecondition("proposerIsCurrentPlayer")` + CurrentPlayer-embed guard exists *solely* to prevent "the ledger says legal while Legal() still rejects" (`legal_plan.go:422-440`) — but only fences the framework's own residue, never the author's.

**Related sub-case:** an override with a conditional early `return nil` *before* the super-call that doesn't fire on the example state passes the probe, then at runtime that path skips the plan entirely (declarations silently not enforced on that path). The probe error text (`legal_plan.go:636`) warns only about the false-positive direction.

**Matrix (verified):**
- Legal() present, never super-calls → **boot error** (probe). Caught.
- Legal() present, super-calls → boots; plan runs once at super-call; residue runs after — but the ledger ignores residue. **Latent (this finding).**
- Legal() deleted → intended path. Fine.

**Fix proposal:** (a) cheapest correct: in `legalFormFromLedger`, when the plan verdict is Pass, additionally call `move.Legal(state, playerIndex)` (and the admin analog) and demote `LegalForPlayer` on error. (b) Stronger: make opt-in + own-`Legal()`-override a boot error (contradicts the spec's blessing). [BATCH: implemented a variant stronger than (a) — booleans always come from `move.Legal()`, fixing BOTH desync directions.]

---

## F2 — HIGH: `WithoutPrecondition` names are unvalidated strings; typos and no-op suppressions boot cleanly

[BATCH: FIXED in 72b981b4 (validation) + 3586f278 (exported constants) + 77d36a45 (all-names error listing, prose fix).]

Verified: `assembleLegalSpecList` (`legal_plan.go:507-522`) drops contributed specs whose `Name` is in the suppression set and **silently ignores every suppression that matches nothing**. `WithoutPrecondition` (`moves/with.go:289-294`) just appends strings. The only validated case is `"proposerIsCurrentPlayer"` on a CurrentPlayer-embedding move (`legal_plan.go:434-440`).

Three latent flavors, all boot-clean:
1. **Typo/case:** `WithoutPrecondition("inphase")` / `"InPhase"` — suppresses nothing; the phase gate stays; author debugs a "why is my move still phase-blocked" mystery with zero signal.
2. **Suppressing a check the move doesn't contribute** (e.g. `"inProgression"` on a move with no `WithLegalMoveProgression`): silent no-op. Harmless today, but masks flavor 1.
3. **`WithoutPrecondition` without any `WithPreconditions`:** the move is not opted in at all (`legal_plan.go:408-413`: empty authored → skip), so the suppression is dead and the frozen chain still enforces the check. Only an internal code comment documents this (`legal_plan.go:53-56`).

**Fix:** at plan assembly, (i) reject suppression names that match no contributed spec, (ii) boot-error when suppressions exist with no authored specs, (iii) export the four names as constants.

---

## F3 — HIGH severity / lower likelihood: `Reads` is pure honor system, and under-declaring a `move.*` read poisons the server-side memo (wrong verdicts, not just wrong UI)

[BATCH: docs FIXED + boot smoke probe SHIPPED in 3f11d57c/5f1ab043 — sentinel-panic detection for field-independent game-registered predicates. Documented blind spots remain: delegate overriding a universal name; conditional reads the example state doesn't trigger; access via Info().ConcreteMove()/concrete assertions. Optional debug-mode read-tracking proxies remain FUTURE work.]

Verified: there is no runtime read-tracking. The only guard is in `evalLegalPredicate` (`legal_predicate.go:397-419`): a **panic** while `ctx.Move == nil` and no declared move path becomes `Unknown`. But every production path passes a **non-nil** Move, so the guard is effectively dormant.

Consequences of an under-declared move read on a game-registered predicate, in increasing badness:
- Client evaluability lies (spec-acknowledged, `legal/doc.go:213-218`): button state wrong, bindings possibly shipped for state the viewer can't see (a #693-adjacent leak, since `LegalReadEvaluable` trusts the declared read-set — `legal_evaluable.go:51`).
- **Memo poisoning (previously undocumented):** the predicate lands in the field-*independent* bucket (`buildLegalPlanFromPredicates`, `legal_plan.go:543-557`, bucketed purely by declared Reads) and its bucket verdict is memoized keyed `(moveName, version, proposer)` (`legal_plan.go:232-247`, `legal_memo.go:52-90`) — move fields are **not** in the key. Concrete sequence: player proposes a move with bad field values at version V → rejected, bucket verdict `Fail` memoized; player fixes the field and re-proposes at the same V → memo hit → correctly-filled move rejected with a stale message. The inverse (stale `Pass`) also works. Same mechanism bites any impure predicate (time/randomness).

**Fix:** (i) document the memo consequence where game-registered predicates are taught; (ii) boot smoke probe with a sentinel Move whose reader panics; (iii) optional debug mode wrapping ctx.State/ctx.Move in tracking proxies asserting touches are a subset of declared Reads.

---

## F4 — MEDIUM: `WithMessage` key is boot-validated, but template *placeholders* vs. predicate *bindings* are not

[BATCH: FIXED in b96e13b6 — LegalPredicate.EmittedBindings metadata + validateLegalEmittedBindings boot check + DefaultTemplates pinning tests. Review minors (intersection-rule docs, malformed-metadata boot error, negative game-registered test) applied in the polish commit.]

Verified: `validateLegalTemplates`/`validateLegalEmittedTemplates` (`legal_error.go:145-264`) check only that keys exist in the merged table. Nothing checks that the template body's `{placeholders}` match the bindings the predicate actually emits. `PropAtLeast(...).WithMessage("reveal.no_cards_left")` with a game template body `"You have {left} cards"` boots cleanly and renders `"You have left cards"` (missing binding renders as the bare placeholder name — `RenderLegalMessage`, `legal_error.go:47-72`). There's no rendered-string fallback on the wire either — a future client rendering an `evaluable:false` entry (bindings stripped by the #693 guard) hits the same bare-placeholder garbage by construction.

**Fix:** the binding names per template are already documented prose on every `Template*` constant. Promote them to metadata (`EmittedBindings map[templateKey][]string` alongside `EmittedTemplates`), extract placeholders with the existing `legalPlaceholderPattern`, and boot-error on placeholders not a subset of bindings.

---

## F5 — MEDIUM-HIGH: `LegalCustom` implemented without `WithPreconditions` is silently never called — and fails *open*

[BATCH: annotation corrected — this was NOT fixed in 72b981b4; that commit implemented F2 (WithoutPrecondition-without-opt-in), a sibling "X-without-opt-in" check that was conflated with F5. F5 remained documentation-only (legal/doc.go) — a fail-open honor system — until it was genuinely FIXED in 42102032: assembleLegalPlans now boot-errors any CustomLegaler move that gets no plan, on BOTH no-plan paths (a legalDeclarer with no authored specs, and a move that is not a legalDeclarer at all, e.g. a core base.Move). Path-tailored message + TDD coverage on both paths; adversarially reviewed (round 1 caught a single-site fix missing the non-declarer path). Full gates green in both repos ⇒ all real LegalCustom implementers already opt in (zero false positives).]

Verified: `assembleLegalPlans` returns early on `len(authored) == 0` (`legal_plan.go:408-413`) before ever type-asserting `CustomLegaler`; the wrapper is only built inside `buildLegalPlanFromPredicates` (`legal_plan.go:552-554`). So a move that implements `LegalCustom` but declares no specs gets no plan, no probe, no boot error — and `LegalCustom` never runs. If an author migrates by moving their `Legal()` body into `LegalCustom` and deleting `Legal()` without adding specs, every check in that body silently stops being enforced (previously-illegal moves become legal). The rule is documented (`legal/doc.go:127-130`) but implementing an interface method that is then never invoked, with zero boot signal, is exactly the "compiles, boots, wrong" shape.

**Fix:** boot check: if `move.(CustomLegaler)` and no plan was assembled → boot error naming the move.

---

## F6 — MEDIUM-LOW: `any(Fail, Unknown)` blocks the move with a vague message, and the author's `WithMessage` is ignored on the Unknown path

[BATCH: FIXED in e195c947 — Message attached on the whole sawUnknown branch (mixed AND all-Unknown); Reason names the first unknown child.]

Verified: `evalLegalAnyKleene` (`legal_predicate.go:309-327`) returns `Unknown{Reason: "a sub-predicate of \"any\" was unknown"}` with **no Message** — the spec's `Message`/`legal.any_failed` template is used only when *all* children Fail. Hot path is fail-closed (`legal_plan.go:139-146`), so the user-visible rejection is the raw Reason string via `LegalError.Error()` (`legal_error.go:101-107`). Unknown children arise legitimately server-side: `player.X` when the current player is Observer/Admin (simultaneous phases — `legal_path.go:236-239`), unknown enum names in `propEquals` (deferred to evaluate time by design), panicking game predicates.

**Fix:** attach the template as a Message on the Unknown verdict too (LegalVerdict explicitly permits Message on Unknown — `legal_types.go:130-132`), and fold the failing sub's identity into Reason.

---

## F7 — LOW-MEDIUM, mostly mitigated: bucket reordering silently changes first-failure messages across the field-dependent split

[BATCH: docs FIXED in 5f1ab043 — moves/doc.go caveat, spec §4 correction, legal_plan.go comment correction.]

Verified mechanics: evaluation is field-independent bucket → field-dependent bucket → custom (`legal_plan.go:192-216`), buckets assigned by declared Reads (`legal_plan.go:543-551`). A field-independent authored check declared *after* a field-dependent one (including `proposerIsCurrentPlayer`, which reads `move.TargetPlayerIndex`) reports first when both fail. **Mitigated:** honestly documented (`legal/doc.go:300-313`) and pinned by migrated games' `knownMessageOrderingDivergence` golden tests. **Residual:** `legal_plan.go:83-84` ("what you declare is what runs, in the order you wrote it") and spec §4's identical claim were contradicted by the split, and `moves/doc.go`'s WithPreconditions section omitted the caveat.

---

## F8 — LOW: doc overclaims boot validation of `Errorf`/`FailT` keys

[BATCH: docs FIXED in 5f1ab043.]

`legal/doc.go:181-184` said every key referenced by "any declared Spec, FailT call, or Errorf call" is boot-validated. Verified false for: `legal.Errorf` keys inside `LegalCustom` bodies (closures can't be introspected) and any game-registered constructor that forgets to populate `EmittedTemplates` (honor system, acknowledged at `legal_predicate.go:50-64`). Both degrade to bare-key rendering at runtime.

---

## F9 — LOW: `players[move.Field]` with special/out-of-range indices — one small trap

[BATCH: FIXED in 06419dbf — boot validation restricted to TypePlayerIndex (no in-repo TypeInt usage existed). The runtime int tolerance in resolveLegalPath was initially kept for ad-hoc undeclared ResolvePath calls, but was subsequently DELETED in 1e2dfce4 (Task-4 review minor, reviewer preferred deletion): resolveLegalPath now returns an error for a TypeInt move field too, so the residual ad-hoc case fails closed (Unknown) rather than silently reading a wrong-but-valid player.]

At evaluation, admin/observer/out-of-range field values return a resolution error (`legal_path.go:265-268`) which catalog predicates convert to fail-closed `Unknown` with a descriptive reason. Boot validated the field exists and is `TypePlayerIndex` **or `TypeInt`** (`legal_path.go:168-174`). The `TypeInt` allowance was the trap: `players[move.CardIndex].Score` boots and silently reads a *wrong but valid* player whenever the int is in range.

---

## F10 — LOW / deferred: ledger `provisional`/`evaluable` misrender risk is prospective

[BATCH: NOT addressed — belongs to sub-project B (TS evaluator/client renderer). Carried in the roadmap.]

No client renderer exists yet (`server/static/src/selectors.ts:174-193` stores `Preconditions` verbatim; nothing renders it). Two documented-but-confusable semantics for sub-project B to watch: (i) `provisional: true` + `verdict: fail` means "might pass with different field values" — a client that disables on any fail will wrongly disable; (ii) `omitempty` on `provisional` makes false indistinguishable from absent, and `evaluable:false` entries ship a template key with no bindings and no pre-rendered string.

---

## Checked, fine (verified non-issues)

- **Typo'd predicate names, paths, args** → boot errors: unknown name (`legal_predicate.go:186-189`), path shape + property existence against real example readers (`legal_path.go:138-199`), constructor arg parsing.
- **Wholesale `Legal()` override + declarations** → boot error via behavioral probe (`legal_plan.go:612-639`); probe flags are single-threaded-boot-only, race-free.
- **Nested `any` / fewer-than-2 subs / constructor-mediated `any` smuggling** → construction errors, belt-and-suspenders (`legal_predicate.go:160-256`).
- **Zero-value verdicts, panics, nil Evaluate** → fail-closed `Unknown` via `evalLegalPredicate`; `Unknown` treated as illegal on the hot path (`legal_plan.go:139-146`).
- **Memo concurrency and lifetime** → mutex-guarded, bounded to head version, compute-outside-lock (`legal_memo.go`); proposer and moveName are in the key.
- **Unsupported base type + opt-in** → boot error naming the type; allowlist structurally enforced by `moves/seam_source_test.go`.
- **Phase index conservativeness** → `inPhase` inside `any` and multi-`inPhase` union both resolve to superset-safe bucketing (`legal_index.go:91-136`); opaque moves always candidates.
- **`WithMessage` key existence** and catalog `EmittedTemplates` keys → boot-validated (`legal_error.go:145-264`); `legal.DefaultTemplates()` coverage test-pinned.
- **`LegalCustom` returning a plain error** → wrapped as one-off template rendering as its own text (`legal_plan.go:604-607`).
- **`inPhase`'s by-convention `game.Phase` Read** on a delegate with unconventional `CurrentPhase` → loud boot path-validation error, documented (`legal/catalog_framework.go:101-117`).
