# Declarative Legality — Completeness Round (Sub-project A) — Design

**Date:** 2026-07-12
**Base:** stacks on `declarative-legality-design` (v1 spec: `2026-07-10-declarative-legality-design.md`, normative for everything not restated here)
**Scope:** sub-project A of the deferred-work campaign (A: catalog/core completeness → B: TS client evaluator + live wiring → C: dirty-tracking audit). B and C get their own specs after A ships.

## Provenance

Initial design adversarially critiqued against the live branch code before spec
(verdict: rethink — two items structurally unsound, one mis-scoped). Critique
checked in: `docs/superpowers/design/2026-07-12-legality-completeness/critique.md`.
This spec is the amended design the critique's verdict prescribes.

## Goals (unchanged): 1 developer ergonomics, 2 cheap client proactivity, 3 correctness/no-footguns.

---

## 1. Count predicates + a new facet

- `legal.StackCount(path string, op string, n int) Spec` — registry name
  `"stackCount"`; component count of the stack at `path` compared via the
  existing op vocabulary (`==,!=,<,<=,>,>=`); `Reads: {path, FacetCount}`;
  `CostCheap`; template `legal.stack_count` with bindings `{count, op, n}`.
- **New facet `LegalFacetNonEmpty`**, added to core: survives
  `PolicyVisible, PolicyOrder, PolicyLen, PolicyNonEmpty` (everything except
  Hidden). Rationale (critique finding 4): `PolicyNonEmpty` reveals exactly
  emptiness; an emptiness predicate must stay client-evaluable under it.
  `facetSurvives` truth table extends to 5 facets × 5 policies, all cells
  tested.
- `legal.StackEmpty(path)` / `legal.StackNotEmpty(path)` — real registry
  predicates (`"stackEmpty"`/`"stackNotEmpty"`, not sugar over stackCount)
  declaring `FacetNonEmpty`; templates `legal.stack_empty`/`legal.stack_not_empty`.
- Honest unblock list: debuganimations' four single-threshold moves
  (`FirstShortStack<1`, `DiscardStack<3`, `FanStack>1` ×2) and pass's
  stack-count gate. The two disjunction-of-conjunctions moves in
  debuganimations (`(6∧3)∨(5∧4)` shapes) STAY in `LegalCustom` — no `all`
  compositor exists and `any` is depth-1; noted, not solved, this round.

## 2. Typed equality (the real valentine blocker)

- `legal.PropEquals(path string, value string) Spec` — registry
  `"propEquals"`; dispatches on the RESOLVED property type:
  - int: `value` parses as int (boot-validated),
  - bool: `"true"`/`"false"`,
  - enum: `value` is an enum value NAME, resolved against the property's enum
    via the chest at boot (unknown name = boot error naming move+path+value),
  - PlayerIndex: `value` is an int, or the specials `"observer"`/`"admin"`
    (valentine compares `RevealedCardOwner == AdminPlayerIndex`).
- `Reads: {path, FacetValues}`; `CostTrivial`; template `legal.prop_equals`
  with `{value, want}` bindings. `PropNotEquals` mirror (`"propNotEquals"`).
- Type dispatch happens at boot (constructor resolves the path's
  PropertyType from the example state and bakes the right comparator) — a
  runtime type surprise is impossible.

## 3. Move-field-indexed player paths (the real darwin unblock)

- Grammar gains a fifth path kind: `players[move.<Field>].<Prop>` — the
  playerState of the player whose index is the move's `<Field>` value.
- Parse/validate in core `legal_path.go` beside the existing kinds: boot
  validates `<Field>` exists on the move reader with PropertyType
  PlayerIndex (or int), and `<Prop>` exists on the player reader.
- Resolve at eval: read the field → `EnsureValid`-guard the index → that
  player's reader. Out-of-range/Observer/Admin field value → the predicate
  returns `Unknown` (never a panic; mirrors invalid-current-player handling).
- Predicates using this path are field-DEPENDENT by construction (they read
  a move field). Under the server's `LegalForAnyone` admin evaluation the
  path reads the move's own field — identical to the old imperative
  `players[m.TargetPlayerIndex]` semantics (critique finding 1's fix: this
  replaces the dropped `proposer.X` design).
- All player-path-shaped catalog predicates accept the new form wherever
  they accept `player.X` (the path resolver, not each predicate, owns this).

## 4. Negation leaves (scoped honestly)

- `legal.PlayerBoolIs(prop string, want bool) Spec` (registry `"playerBool"`
  gains an optional second arg; absent = today's `true` — backward
  compatible, corpus extended).
- `legal.ComponentAbsentAt(stackPath, idxField)` — occupancy facet, inverse
  of `ComponentPresentAt`, template `legal.component_present_unexpected`.
- These are catalog completeness, not headline unblocks: darwin's
  `DoneWithPhase==false` checks additionally need §3's paths to be
  reachable; both together re-enable those migrations.

## 5. Seam widening — invariant-enforced

- `FixUp`, `FixUpMulti`, `StartPhase` become opt-in-capable: verified (this
  round, against source) to declare NO `Legal()` override — their legality
  IS `Default.Legal`, so plan evaluation composes exactly as for `Default`.
- Enforcement is structural, not a bare name list: a `moves`-package test
  parses `moves/*.go` with `go/parser` and asserts no allowlisted type
  declares a `Legal` method — a future override flips the test red and
  forces a conscious seam decision.
- `FixUpMulti` precondition: before allowlisting, a test proves the
  contributed `inProgression` atom honors `AllowMultipleInProgression()`
  (repeated same-move tapes legal exactly when the imperative chain says
  so). If divergent, FixUpMulti is excluded and the divergence documented.
- The seam boot-error message and probe error text gain one sentence each
  pointing at the allowlist rule.
- `FinishTurn` / `DealCountComponents` / the rest: STAY blocked (critique
  finding 2 — partial contribution desyncs the ledger; round-robin
  predicates are future work).

## 6. Re-migrations unblocked by 1–5 (each golden-fenced per the v1 harness)

- boardgame: debuganimations ×4 (counts), blackjack `moveStartRoundCleanup`
  restored to the spec-§8 `moves.StartPhase` spelling (seam now allows it;
  the hand-rolled Apply workaround is removed), plus any Task-12 survey move
  the new predicates now cover (re-check tictactoe/werewolf notes).
- ../games: pass (counts); darwin's reverted moves via §3 paths + §4 leaves
  (population/DynamicComponentValues residue stays `LegalCustom` — no
  DCV paths this round, documented); valentine via §2 typed equality
  (card-value/DCV residue stays `LegalCustom`).

## Rails (carried from v1, restated as binding)

- Purely-sugar: frozen chain untouched; `TestLegalChainStringFreeze` green
  throughout; goldens for every (re)migrated move, divergence cells narrow
  and named.
- Every new predicate: minimal-facet Reads, template keys in
  DefaultTemplates, EmittedTemplates populated, conformance corpus file with
  ≥3 cases (+ templates pinned on fails), corpus completeness meta-test
  extended automatically (it enumerates the registry).
- `LegalCatalogVersion` → 2. No design argument may rely on client-side
  skew handling (doesn't exist until sub-project B).
- Docs ride along: tutorial's catalog table + limits section, `legal/doc.go`,
  spec impl-notes appendix updated for anything adjudicated mid-flight.

## Non-goals this round (deferred, with owners)

- `proposer.X` proposer-relative legality — requires a `LegalForAnyone`
  redesign (per-player existential evaluation); own future design.
- Round-robin/progression predicates → DealCountComponents/FinishTurn
  contributions.
- `MayMoveTo` facet narrowing — unsound while `Stack.AddConstraint` is
  public runtime API; revisit with constraint immutability.
- `all` compositor / nested compositors (disjunction-of-conjunctions).
- DynamicComponentValues paths.
- Sub-projects B (TS evaluator + live wiring) and C (dirty-tracking audit).

## Testing summary

Facet truth table 5×5; per-predicate unit + corpus; path-kind parse/validate/
resolve incl. Unknown guards; seam source-parse test; FixUpMulti progression
equivalence; goldens for all re-migrations (both repos); purely-sugar suite;
full gates both repos (3 known boardgame sandbox exclusions + valentine's
pre-existing TestGolden).
