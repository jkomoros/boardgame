# Legality Completeness Round (Sub-project A) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `docs/superpowers/specs/2026-07-12-legality-completeness-design.md` (NORMATIVE): count predicates + NonEmpty facet, typed equality, move-field-indexed player paths, negation leaves, invariant-enforced seam widening, and the re-migrations they unblock.

**Architecture:** Pure extension of the shipped v1 system on the same branch — new facet + path kind in core `legal_path.go`, new predicates in `legal/catalog_*.go` following the established constructor/corpus/template pattern, seam allowlist in `legal_plan.go` guarded by a source-parse test, re-migrations using the Task 11/12 golden harness patterns verbatim.

**Tech Stack:** Go (core, legal, moves, examples, ../games). No TS work this round (that's sub-project B).

## Global Constraints

- Branch `declarative-legality-design` (stacking); ledger `.superpowers/sdd/progress-legality.md` continues; commit per task.
- All v1 runbook §0/§2 conventions apply verbatim (gates, 3 boardgame sandbox exclusions + ../games' valentine TestGolden pre-existing failure, frozen chain / `TestLegalChainStringFreeze` green always, minimal-facet Reads, template keys in DefaultTemplates + EmittedTemplates, corpus file per predicate ≥3 cases with fail templates pinned, no bare Pass/Fail/Unknown name collisions).
- Reference implementations (read before writing similar code): `legal/catalog_compare.go` + `legal/catalog_stack.go` (predicate + constructor + corpus pattern), `examples/memory/legal_golden_test.go` + `examples/pig/legal_golden_test.go` (golden oracle pattern incl. divergence maps — find exact filenames via `ls examples/*/legal_golden_test.go` / grep `knownMessageOrderingDivergence`), `../games/valentine/legal_golden_test.go` (fixture-builder pattern).
- Spec §-references below are to `2026-07-12-legality-completeness-design.md`.
- `LegalCatalogVersion` bumps to 2 in Task 8, not before (single bump).

## File Structure

| File | Fate |
|---|---|
| `legal_path.go` + test (core) | T1: `LegalFacetNonEmpty`; T3: `players[move.<Field>].<Prop>` path kind |
| `legal/catalog_stack.go` (or new `catalog_count.go`) + tests + corpus | T2: stackCount/stackEmpty/stackNotEmpty; T5: componentAbsentAt |
| `legal/catalog_compare.go` + tests + corpus | T4: propEquals/propNotEquals typed; T5: playerBool optional arg |
| `legal_plan.go` + `moves/seam_source_test.go` | T6: seam allowlist + go/parser invariant + FixUpMulti gate |
| `examples/debuganimations`, `examples/blackjack` | T7: re-migrations |
| `../games/{pass,darwin,valentine}` | T8 part: re-migrations |
| docs (tutorial, legal/doc.go, spec impl-notes) + `legal_types.go` version | T8: close-out |

---

### Task 1: LegalFacetNonEmpty

**Files:** Modify `legal_path.go` (facet consts + `facetSurvives`), `legal_path_test.go`.
**Interfaces:** Produces `boardgame.LegalFacetNonEmpty` (append AFTER existing consts — do not renumber; serialized specs don't carry facet ints but don't risk it) + alias `legal.FacetNonEmpty` in `legal/aliases.go`. Survival row (spec §1): Visible ✓, Order ✓, Len ✓, NonEmpty ✓, Hidden ✗.

- [ ] **Step 1:** extend the `facetSurvives` truth-table test to 5 facets × 5 policies (25 cells, every cell asserted; new rows exactly per spec §1) → RED.
- [ ] **Step 2:** add the const + doc comment (rationale: PolicyNonEmpty reveals exactly emptiness — critique finding 4) + `facetSurvives` case + alias → GREEN.
- [ ] **Step 3:** `go build ./... && go vet ./... && go test . ./legal/... -count=1` → commit `legal: LegalFacetNonEmpty (spec §1)`.

### Task 2: Count predicates

**Files:** Create `legal/catalog_count.go`, `legal/catalog_count_test.go`; corpus `legal/testdata/conformance/{stackCount,stackEmpty,stackNotEmpty}.json`; extend `defaultTemplateKeys` + `legal/templates.go` DefaultTemplates.
**Interfaces:** Produces (spec §1 exactly):
```go
func StackCount(path string, op string, n int) Spec // "stackCount"; Reads {path, FacetCount}; CostCheap; template legal.stack_count {count,op,n}
func StackEmpty(path string) Spec                   // "stackEmpty"; Reads {path, FacetNonEmpty}; template legal.stack_empty
func StackNotEmpty(path string) Spec                // "stackNotEmpty"; same facet; template legal.stack_not_empty
```
Ops reuse the existing comparator table (grep `legalCompareOps` in catalog_compare.go — share, don't duplicate). Count = the stack's `NumComponents()` (verify the ImmutableStack method name by reading stack.go). Unknown when path unresolvable (nil-move rule doesn't apply — no move fields — but nil/missing stack → Unknown).

- [ ] **Step 1:** unit tests per predicate (pass/fail/unknown, Reads facet asserted, template key) + 3 corpus files w/ pinned fail templates → RED. Fixtures: reuse the existing conformance fixture games.
- [ ] **Step 2:** implement + wire `DefaultConstructors()` + templates → GREEN incl. `TestDefaultTemplatesCoversCorpusFailCases` + completeness meta-test (auto-picks up the new names).
- [ ] **Step 3:** gates → commit `legal: count predicates + emptiness on FacetNonEmpty (spec §1)`.

### Task 3: players[move.Field].X path kind

**Files:** Modify `legal_path.go` (+test).
**Interfaces:** Produces grammar `players[move.<Field>].<Prop>` (spec §3): parse (case-sensitive; only `move.`-prefixed index expressions; `players[0]`/`players[*].X` behavior unchanged); `validateLegalPath` checks `<Field>` on move reader is TypePlayerIndex or TypeInt AND `<Prop>` on player reader; `resolveLegalPath` reads field → guards index via the same valid-range logic `ImmutableCurrentPlayer` uses (out-of-range/Observer/Admin → error the caller maps to Unknown, matching the existing invalid-current-player pattern — read the pathPlayer branch and mirror). Predicates reading this path are field-dependent (path contains `move.` — verify `legalReadsIncludeMovePath` in legal_plan.go treats it so; extend if it string-matches only prefix `move.`).

- [ ] **Step 1:** parse table (valid: `players[move.TargetPlayerIndex].Hand`; invalid: `players[move.].X`, `players[game.X].Y`, `players[move.F]` w/o prop) + validate (missing field / wrong type / missing prop each named) + resolve round-trips + guard cases (Observer/Admin/oob field values → error not panic) + bucket-classification test (a predicate reading this path lands field-dependent) → RED.
- [ ] **Step 2:** implement → GREEN. **Step 3:** gates + freeze test → commit `legal: move-field-indexed player paths (spec §3)`.

### Task 4: Typed equality

**Files:** Modify `legal/catalog_compare.go` (+test); corpus `{propEquals,propNotEquals}.json`; templates.
**Interfaces (spec §2):**
```go
func PropEquals(path string, value string) Spec    // "propEquals"; Reads {path, FacetValues}; CostTrivial; template legal.prop_equals {value,want}
func PropNotEquals(path string, value string) Spec // "propNotEquals"; template legal.prop_not_equals
```
Boot-time type dispatch: constructor resolves the path's PropertyType against the example state (available via `resolveLegalSpecs`'s exampleState — check how existing constructors access it; if constructors don't receive it, the dispatch must happen lazily on FIRST Evaluate with the comparator cached... NO: read `legal_predicate.go` — if exampleState isn't reachable from constructors, do the type dispatch inside Evaluate via the resolved PropertyType (resolveLegalPath returns it) with the VALUE parsed/validated at construction for all four types (int parse, bool parse, enum name existence deferred to a boot-validation hook if chest-independent... enum names need the chest — constructors DO receive chest: validate enum names against every enum containing the name? No — the path's enum isn't known without the example state. Pragmatic v1: parse int/bool eagerly; enum/PlayerIndex values validated on first resolve with Unknown on mismatch AND a boot smoke: extend `validateLegalPath`-driven boot validation to also let predicates self-validate — check whether `resolveLegalSpecs` runs any per-predicate validation hook; if not, accept eval-time Unknown + document, noting the boot-error aspiration in the spec impl-notes). RECORD whichever branch reality forces in the report — this is the task's one judgment point.
PlayerIndex specials: `"observer"` → ObserverPlayerIndex, `"admin"` → AdminPlayerIndex, else int.

- [ ] **Step 1:** unit tests per type arm (int/bool/enum/PlayerIndex incl. specials; mismatched value → boot error or Unknown per the branch taken; wrong-type path) + corpus → RED.
- [ ] **Step 2:** implement (share comparators; do NOT duplicate resolveIntPath-style helpers — extend `legal/path_resolve.go` wrappers if a typed getter is missing) → GREEN.
- [ ] **Step 3:** gates → commit `legal: typed equality predicates (spec §2)`.

### Task 5: Negation leaves

**Files:** Modify `legal/catalog_players.go` (playerBool optional arg), `legal/catalog_stack.go` or `catalog_count.go` (componentAbsentAt); tests; corpus updates (`playerBool.json` gains want-false cases; new `componentAbsentAt.json`); templates.
**Interfaces (spec §4):**
```go
func PlayerBoolIs(prop string, want bool) Spec        // registry "playerBool", args [prop] or [prop, "false"|"true"]; absent 2nd arg == true (backward compat — existing corpus/specs unchanged)
func ComponentAbsentAt(stackPath, idxField string) Spec // "componentAbsentAt"; occupancy facet; template legal.component_present_unexpected
```
`AllActivePlayers`'s restricted inner-grammar must accept the 2-arg playerBool (read its inner-compilation in catalog_players.go and extend).

- [ ] **Step 1:** tests (2-arg parse, backward compat 1-arg, AllActivePlayers inner acceptance, absent-at pass/fail/unknown + facet) + corpus → RED. **Step 2:** implement → GREEN. **Step 3:** gates → commit `legal: negation leaves — playerBool want-arg, componentAbsentAt (spec §4)`.

### Task 6: Seam widening, invariant-enforced

**Files:** Modify `legal_plan.go` (`legalUnsupportedMovesBaseType` allowlist); create `moves/seam_source_test.go`; extend `moves/preconditions_test.go` (FixUpMulti gate); adjust the seam/probe boot-error texts; the existing `TestUnsupportedBaseTypeStartPhaseIsBootError` inverts to a boots-fine test.
**Interfaces (spec §5):** allowlist becomes {Default, CurrentPlayer, FixUp, FixUpMulti, StartPhase}. Source-parse test: `go/parser` over every `moves/*.go`, collect types declaring a `Legal` method, assert intersection with the allowlist minus {Default, CurrentPlayer} is EMPTY (those two legitimately declare it), with a failure message telling the future author to make a conscious seam decision.

- [ ] **Step 1:** FixUpMulti equivalence gate FIRST: a test with a FixUpMulti-embedding move in an ordered progression, repeated occurrences — assert plan-evaluated legality == frozen-chain legality across a tape exercising `AllowMultipleInProgression` (build on the moves package's existing progression fixtures; grep `AllowMultipleInProgression` usage in tests). If this test proves divergence: EXCLUDE FixUpMulti, document in the spec impl-notes, and proceed with {FixUp, StartPhase} only — report it, don't force it.
- [ ] **Step 2:** source-parse test (RED against a deliberately-wrong temp allowlist entry, then fixed) + allowlist change + boot tests: StartPhase-embedding move with WithPreconditions now BOOTS and its plan evaluates (assert an always-fail declared predicate surfaces); DealCountComponents still boot-errors.
- [ ] **Step 3:** error-text sentences (seam + probe) per spec §5 → gates + freeze test → commit `core: seam widened to no-Legal framework types, source-parse invariant (spec §5)`.

### Task 7: boardgame re-migrations

**Files:** `examples/debuganimations/{moves.go,main.go}` + `legal_golden_test.go` (new); `examples/blackjack/moves.go` (+ golden update); survey re-check of tictactoe/werewolf Task-12 notes.
Per spec §6: debuganimations' four single-threshold moves → StackCount/StackEmpty gates (goldens per the Task 11/12 pattern; the disjunction pair EXPLICITLY stays LegalCustom with the header comment updated to name what still blocks them); blackjack `moveStartRoundCleanup` restored to `moves.StartPhase` embed (delete the hand-rolled Apply; assert TestGolden full-game replay still green — the Task 11 report documents why the swap was made; reverse it cleanly). Re-check tictactoe/werewolf survey comments — migrate anything the new predicates now cover (with goldens) or update the comments to name the remaining blocker precisely.

- [ ] Steps: goldens-first → migrate → gates (`go test ./examples/... ./moves/... ./legal/... ./server/api/... . -count=1` + freeze) → commit per game.

### Task 8: ../games re-migrations + close-out

**Files:** `../games/{pass,darwin,valentine}` moves + goldens; `legal_types.go` (LegalCatalogVersion=2); TUTORIAL.md + `legal/doc.go` (new predicates/paths in the catalog table + limits section updated: disjunction/DCV/proposer-relative still out); spec impl-notes appendix extended for anything adjudicated mid-flight; corpus meta-test auto-covers new names (verify).
Per spec §6: pass (counts), darwin (players[move.X] paths + PlayerBoolIs; population/DCV residue stays LegalCustom, documented), valentine (PropEquals for RevealedCardOwner==admin + enum compares; card-value residue stays LegalCustom). Golden oracles per the ../games Task-13 fixture-builder pattern. Both repos' full gates; ledger updated; commits per game in ../games + one close-out commit in boardgame.

- [ ] Steps: goldens-first per game → migrate → both-repo gates (incl. `TestLegalChainStringFreeze`, exclusions per Global Constraints) → commits → close-out commit `legal: catalog v2 — completeness round close-out (docs, version, impl-notes)`.

---

## Self-review

- **Spec coverage:** §1→T1+T2; §2→T4; §3→T3; §4→T5; §5→T6; §6→T7+T8; rails (corpus/templates/version/docs)→woven into T2-T8 with the single version bump in T8. Non-goals need no tasks. ✓
- **Placeholder scan:** T4 contains a genuine either/or branch (constructor exampleState access) — it names both branches concretely with a decision rule and a reporting requirement; not a TBD. Reference-implementation reads are directed at named committed files. ✓
- **Type consistency:** `StackCount/StackEmpty/StackNotEmpty/PropEquals/PropNotEquals/PlayerBoolIs/ComponentAbsentAt` and `LegalFacetNonEmpty` names match spec exactly and are used consistently in T7/T8. ✓
