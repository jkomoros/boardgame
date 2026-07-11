# Declarative Move Legality Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the declarative move-legality system per `docs/superpowers/specs/2026-07-10-declarative-legality-design.md` (rev 2.1, NORMATIVE — every task implicitly includes "read the relevant spec section first").

**Architecture:** Value types + evaluation engine in core `boardgame` (prefixed `Legal*`); a new `legal` package (peer of `constraints`) holds the predicate catalog, registry defaults, templates, and type aliases so authors write `legal.Spec`. `moves.Default.Legal()` keeps its frozen imperative chain and gains an opt-in plan path. Server ships a per-predicate ledger + chest templates. Purely-sugar and golden-equivalence tests fence everything.

**Tech Stack:** Go 1.x (core + moves + legal + server/api), table-driven tests, JSON conformance corpus under `legal/testdata/conformance/`.

## Global Constraints

- Branch: `declarative-legality-design` (already checked out; spec at rev 2.1 committed). Commit after every task, often mid-task. `../games` = `/Users/jkomoros/Code/go/src/github.com/jkomoros/games`, gets matching branch in Task 13.
- Every commit keeps `go build ./...` and `go vet ./...` green from repo root; run `go test ./legal/... ./moves/... .` for the packages a task touches, full `go test ./...` before each task's final commit.
- **The frozen chain is inviolable:** no observable behavior change for any move without `WithPreconditions` — same checks, same order, same error STRINGS. `moves/default.go:339-631`'s existing logic may be *called from* new code paths but its behavior for un-opted-in moves must be byte-identical (property-tested in Task 8).
- Naming: core types are `Legal`-prefixed (`boardgame.LegalSpec`, `LegalVerdict`, ...); the `legal` package aliases them (`type Spec = boardgame.LegalSpec`). Authors only ever see `legal.*`. Mirrors `StackConstraint` precedent.
- Path grammar (spec §1): `game.X`, `player.X` (current player), `players[*].X` (quantifiers only), `move.X`. Facets: `FacetValues`, `FacetCount`, `FacetOccupancy`, `FacetOrder`.
- Zero `LegalOutcome` is invalid/fails closed. Evaluation order is plan order (contributed-first, then declaration order), NO Cost sorting. `any` is the only compositor, depth 1.
- Boot-time (`NewGameManager`) validation: unknown predicate name, bad args, bad path, unknown template key, orphaned declarations (probe), unsupported base type — all fail fast with the move's name in the error.
- Stable contributed-atom names: `"inPhase"`, `"inProgression"`, `"stackConstraints"`, `"proposerIsCurrentPlayer"`.
- Key existing anchors (verified by exploration): `move.go:216` (Legal signature), `moves/default.go:339` (Default.Legal), `:359-388` (stack constraints), `:425-469` (legalInPhase), `:561-631` (progression + matchTape), `moves/with.go` (With* options + configProp keys), `moves/auto_config.go:154` (AutoConfigurer.Config), `game.go:887` (applyMove→Legal), `game.go:562,647` (fixup loop), `server/api/main.go:1590-1629` (generateFormsWithLegality), `:1631-1657` (formFields), `stack_constraint.go:33-46` (constraint registry precedent), `sanitization.go` (policies), `enum/tree.go` (TreeEnum ancestors).

## File Structure

| File | Fate |
|---|---|
| `legal_types.go` (core) | Create T1: LegalOutcome/LegalBindingValue/LegalMessage/LegalVerdict/LegalSpec/LegalRead/LegalFacet/LegalCost/LegalPropPath |
| `legal_path.go` (core) | Create T2: path parse/validate/resolve + facet-vs-policy survival |
| `legal_predicate.go` (core) | Create T3: LegalPredicate/LegalContext/LegalPredicateConstructor + registry resolution incl. `any` |
| `legal_plan.go` (core) | Create T7-T8: plan assembly, probe, opt-in detection, evaluation |
| `legal_index.go` (core) | Create T9: phase index + agnostic bucket + memos |
| `legal/aliases.go` | Create T1: type aliases + verdict constructors |
| `legal/catalog_*.go` | Create T4-T5: predicate builders + constructors |
| `legal/templates.go` | Create T6: TemplateConfigurer, DefaultTemplates, Errorf, rendering |
| `legal/testdata/conformance/*.json` | Create T4+: corpus |
| `moves/with.go` | Modify T7: WithPreconditions/WithoutPrecondition |
| `moves/default.go` | Modify T7-T8: ContributedPreconditions + plan path in Legal() |
| `moves/current_player.go` | Modify T7: ContributedPreconditions |
| `game_manager.go` / `game.go` | Modify T8-T9: boot validation/probe; fixup+forms use index |
| `server/api/main.go:1590-1657` | Modify T10: ledger assembly |
| `server/static/src/types/api.d.ts` + selectors | Modify T10: ledger types |
| `examples/{memory,blackjack,checkers,...}` | Modify T11-T12: migrations |
| `../games/{murdermrmonroe,pass,valentine}` | Modify T13 |

Tasks 1–6 = the `legal` foundations (pure, no engine). 7–9 = moves/engine. 10 = server/wire. 11–14 = migrations + close-out.

---

### Task 1: Core value types + legal package aliases

**Files:** Create `legal_types.go`, `legal_types_test.go` (core); `legal/aliases.go`, `legal/doc.go`.

**Interfaces (Produces — later tasks rely on these EXACT names):**

```go
// package boardgame
type LegalOutcome int
const (
    legalOutcomeInvalid LegalOutcome = iota // zero fails closed
    LegalPass
    LegalFail
    LegalUnknown
)
type LegalBindingValue struct{ S *string; I *int; B *bool } // exactly one set
type LegalMessage struct {
    Template string
    Bindings map[string]LegalBindingValue
}
type LegalVerdict struct {
    Outcome LegalOutcome
    Message *LegalMessage // set on Fail (optionally Unknown); nil on Pass
    Reason  string        // on Unknown
}
type LegalCost int
const ( LegalCostTrivial LegalCost = iota; LegalCostCheap; LegalCostModerate; LegalCostExpensive )
type LegalFacet int
const ( LegalFacetValues LegalFacet = iota; LegalFacetCount; LegalFacetOccupancy; LegalFacetOrder )
type LegalPropPath string
type LegalRead struct{ Path LegalPropPath; Facet LegalFacet }
type LegalSpec struct {
    Name    string      `json:"name"`
    Args    []string    `json:"args,omitempty"`
    Sub     []LegalSpec `json:"sub,omitempty"`
    Message string      `json:"message,omitempty"`
}
func (s LegalSpec) WithMessage(templateKey string) LegalSpec // sets Message, returns copy
```

```go
// package legal
type Spec = boardgame.LegalSpec
type Verdict = boardgame.LegalVerdict
// ... alias every Legal* type; plus:
func PassVerdict() Verdict
func FailT(template string, bindings ...map[string]boardgame.LegalBindingValue) Verdict
func UnknownVerdict(reason string) Verdict
func String(s string) boardgame.LegalBindingValue // + Int(int), Bool(bool) helpers
```

- [ ] **Step 1: failing tests** in `legal_types_test.go`: (a) `LegalVerdict{}.Outcome` is not Pass/Fail/Unknown (zero fails closed); (b) `LegalSpec` JSON round-trips leaf and nested (`any` with two subs) exactly, omitting empty fields; (c) `WithMessage` returns a copy (original unmutated); (d) `LegalBindingValue` marshals as its single set field.
- [ ] **Step 2:** `go test . -run TestLegal` → FAIL (types undefined).
- [ ] **Step 3:** implement `legal_types.go` per the interfaces block verbatim (JSON via struct tags; BindingValue needs custom MarshalJSON/UnmarshalJSON emitting the bare value — string/number/bool — and inferring on decode).
- [ ] **Step 4:** tests pass; add `legal/aliases.go` + verdict constructors + `legal/doc.go` (one-paragraph package doc pointing at the spec).
- [ ] **Step 5:** `go build ./... && go test . ./legal/...` → green. Commit: `legal: core value types + alias package (spec §1)`.

---

### Task 2: Path grammar + facet survival

**Files:** Create `legal_path.go`, `legal_path_test.go` (core).

**Interfaces (Produces):**

```go
// package boardgame
type legalPathKind int
const ( pathGame legalPathKind = iota; pathPlayer; pathPlayersAll; pathMove )
type parsedLegalPath struct { kind legalPathKind; prop string }
func parseLegalPath(p LegalPropPath) (parsedLegalPath, error) // grammar per spec §1
// Validate against a manager's example state + example move readers; returns
// error naming the path and the missing property. moveReader may be nil for
// non-move paths.
func validateLegalPath(p LegalPropPath, exampleState ImmutableState, moveReader PropertyReader) error
// Resolve at eval time. For pathPlayer resolves against state.CurrentPlayer...
// wait — CurrentPlayer index comes from delegate.CurrentPlayerIndex(state).
func resolveLegalPath(p LegalPropPath, ctx LegalContext) (interface{}, PropertyType, error)
// facetSurvives reports whether reading `facet` from a property sanitized
// with `policy` yields trustworthy data (spec §6 table):
//   FacetCount survives Visible, Len, Order;  FacetOccupancy survives Visible, Order;
//   FacetOrder survives Visible, Order;      FacetValues survives Visible only.
func facetSurvives(policy Policy, facet LegalFacet) bool
```

- [ ] **Step 1: failing tests**: parse table (valid: `game.DrawStack`, `player.CardsLeftToReveal`, `players[*].Stood`, `move.CardIndex`; invalid: `Game.X` (case), `players[0].X`, `foo.X`, empty prop); validate against a test game's real readers (use the existing test game in `main_test.go` / `test_game.go` — grep for the canonical test game setup used by other core tests and reuse it) — unknown prop errors name path; facetSurvives full truth table.
- [ ] **Step 2:** run → FAIL. **Step 3:** implement. `resolveLegalPath` returns the raw property value + its PropertyType via the reader (`state.ImmutableGameState().Reader()`, `state.ImmutablePlayerStates()[i].Reader()`, `ctx.Move.ReadSetter()`); `pathPlayer` uses `ctx.State.CurrentPlayerIndex()`-equivalent — check how core exposes current player on ImmutableState (grep `CurrentPlayerIndex` in state.go/game_delegate.go; it's on the delegate: resolve via `state.Game().Manager().Delegate().CurrentPlayerIndex(state)` — find the exact accessor used elsewhere in core and use that).
- [ ] **Step 4:** green. **Step 5:** commit `legal: path grammar, boot validation, facet survival (spec §1, §6)`.

---

### Task 3: Predicate, Context, registry, `any`

**Files:** Create `legal_predicate.go`, `legal_predicate_test.go` (core); extend `legal/aliases.go`.

**Interfaces (Produces):**

```go
// package boardgame
type LegalContext struct {
    State    ImmutableState
    Move     Move // nil during field-independent evaluation
    Proposer PlayerIndex
    Chest    *ComponentChest
}
type LegalPredicate struct {
    Name     string
    Args     []string
    Reads    []LegalRead
    Cost     LegalCost
    Evaluate func(ctx LegalContext) LegalVerdict // pure
    Sub      []*LegalPredicate // only for compositors (any)
    opaque   bool              // escape-hatch wrapper: no serialized form
}
func (p *LegalPredicate) Serializable() bool // !opaque && all Sub serializable
type LegalPredicateConstructor struct {
    Name        string
    Constructor func(spec LegalSpec, chest *ComponentChest,
        resolve func(LegalSpec) (*LegalPredicate, error)) (*LegalPredicate, error)
}
// resolveLegalSpecs resolves specs against a registry map; enforces: unknown
// name = error naming spec; `any` only compositor; `any` depth 1 (a sub may
// not itself be `any`); `any` needs >=2 subs.
func resolveLegalSpecs(specs []LegalSpec, registry map[string]*LegalPredicateConstructor,
    chest *ComponentChest) ([]*LegalPredicate, error)
```

`any` evaluation (Kleene, spec §6): any child Pass → Pass; else if any child Unknown → Unknown; else Fail (message = spec's `.Message` key if set, else template `legal.any_failed`). Its `Reads` = union of children's. Cost = max of children's.

- [ ] **Step 1: failing tests**: registry resolution round-trip (register a fake `alwaysPass` constructor, resolve leaf + `any` of two); depth-2 `any` rejected with error naming the spec; unknown name rejected; Kleene truth table for `any` (P/U/F combinations); `Serializable()` false when a sub is opaque; zero-Verdict from an Evaluate treated as engine error: `evaluatePredicate` helper wraps Evaluate and converts `legalOutcomeInvalid` to `LegalUnknown` with reason `"predicate returned invalid verdict"` (fail closed, never legal).
- [ ] **Step 2:** FAIL. **Step 3:** implement incl. the wrapping helper `func evalLegalPredicate(p *LegalPredicate, ctx LegalContext) LegalVerdict` (also the home of the nil-Move guard: if a predicate's Reads include a `move.` path — or panics on nil Move, recovered — return `UnknownVerdict("undeclared move read")`; use `defer recover()` around Evaluate, converting panic to Unknown with the predicate name in Reason).
- [ ] **Step 4:** green. **Step 5:** commit `legal: predicate model, registry, any-compositor, fail-closed eval (spec §1)`.

---

### Task 4: Catalog part 1 — comparisons, presence, stack moves + conformance corpus

**Files:** Create `legal/catalog_compare.go`, `legal/catalog_stack.go`, `legal/catalog_test.go`, `legal/conformance_test.go`, `legal/testdata/conformance/README.md` + per-predicate `.json`.

**Interfaces (Produces — builders return `legal.Spec`):**

```go
// package legal — names are the registry names too (camelCase)
func PropAtLeast(path string, n int) Spec        // "propAtLeast"; works for any int prop
func PropCompare(path, op string, n int) Spec    // "propCompare"; op in ==,!=,<,<=,>,>=
func PlayerBool(prop string) Spec                // "playerBool": player.<prop> is true; in players[*] context, that player's
func ComponentPresentAt(stackPath, idxField string) Spec    // "componentPresentAt": occupancy facet
func ComponentPresentAtKey(stackPath, keyField string) Spec // enum-keyed variant (checkers Spaces)
func MayMoveTo(srcPath, dstPath, idxField string) Spec      // wraps ImmutableComponentInstance.MayMoveTo
func MayMoveToSlot(srcPath, dstPath, idxField string) Spec  // component.go:127,138 precedent
func DefaultConstructors() []*boardgame.LegalPredicateConstructor
func ExtendDefaults(extra ...*boardgame.LegalPredicateConstructor) []*boardgame.LegalPredicateConstructor
```

Reads facets: presence checks declare `FacetOccupancy` on the stack path; `PropAtLeast/Compare` declare `FacetValues`; `MayMoveTo*` declares occupancy on both stacks + values on move field. Every Fail uses a default template key (`legal.prop_at_least`, `legal.component_missing`, ...) overridable via `Spec.Message`.

**Conformance corpus** (spec §6): each catalog predicate gets `legal/testdata/conformance/<name>.json`:

```jsonc
{"predicate": "propAtLeast",
 "cases": [{"spec": {"name":"propAtLeast","args":["player.CardsLeftToReveal","1"]},
            "fixture": "memoryMidGame",   // named state fixtures built in conformance_test.go
            "proposer": 0, "expect": "pass"}, ...]}
```

`conformance_test.go` loads every file, builds the named fixtures (use the core test game + a memory-like fixture with two stacks), resolves the spec through `DefaultConstructors()`, asserts the verdict. This file format IS the future TS contract — keep it dumb JSON.

- [ ] **Step 1:** write `catalog_test.go` unit tests (per predicate: pass/fail/unknown paths, Reads correctness, template key on fail) + 2 corpus files → FAIL.
- [ ] **Step 2:** implement `catalog_compare.go` + `catalog_stack.go` + `DefaultConstructors` wiring.
- [ ] **Step 3:** green; add corpus files for ALL predicates in this task; conformance runner green.
- [ ] **Step 4:** commit `legal: catalog part 1 (compare/presence/stack) + conformance corpus (spec §1, §6)`.

---

### Task 5: Catalog part 2 — quantifier, any-builder, proposer, purpose-built

**Files:** Create `legal/catalog_players.go`, `legal/catalog_purpose.go`; extend tests + corpus.

**Interfaces (Produces):**

```go
func Any(subs ...Spec) Spec                       // "any" compositor builder
func AllActivePlayers(inner Spec) Spec            // "allActivePlayers": inner holds for every
                                                  // active player (skips behaviors-inactive);
                                                  // Reads: players[*].<inner reads>, FacetValues
func ProposerIsCurrentPlayer() Spec               // "proposerIsCurrentPlayer" — reads move.TargetPlayerIndex
                                                  // (FIELD-DEPENDENT, spec §4) + game current player;
                                                  // replicates moves/current_player.go:37-65 semantics
                                                  // and its exact error strings as templates
func RevealableCardAt(hiddenPath, visiblePath, idxField string) Spec // spec §8 two-branch, occupancy only
func ComponentPropEqualsCurrentPlayer(stackPath, keyField, prop string) Spec // checkers color check
```

`AllActivePlayers` evaluation: iterate `ctx.State.ImmutablePlayerStates()`; skip inactive (find the exported check — `behaviors.PlayerIsInactive` per `examples/blackjack/moves.go:44`; if behaviors can't be imported from `legal` without a cycle, replicate the check via the behaviors interface type-assertion the way blackjack does); evaluate `inner` with a per-player context (define how: the inner spec's `player.` paths resolve against the iterated player — implement via a context override field `playerOverride *PlayerIndex` on `LegalContext`... NO new core field mid-plan: instead resolve by constructing inner predicates with `players[*]` semantics — the quantifier resolves inner's `player.X` reads itself by reading each player's Reader directly. Keep inner limited in v1 to `playerBool` and `propCompare` on player paths; document + boot-error otherwise).

- [ ] **Step 1:** failing tests incl. current_player.go string-parity test (evaluate `ProposerIsCurrentPlayer` against states where it's not your turn; assert rendered message == the string at `moves/current_player.go` — read that file for the exact text) and the memory two-branch strings verbatim (`"there is no card at that index"`, `"that card has already been revealed"` via templates `legal.no_card_here` / `legal.already_revealed`).
- [ ] **Step 2:** implement; corpus files for each. **Step 3:** green. **Step 4:** commit `legal: quantifier, proposer, purpose-built predicates (spec §1, §8)`.

---

### Task 6: Templates — table, rendering, Errorf, boot validation

**Files:** Create `legal/templates.go`, `legal/templates_test.go`; extend `legal_types.go` (core) with rendering hook.

**Interfaces (Produces):**

```go
// package legal
type TemplateConfigurer interface { ConfigureLegalTemplates() map[string]string }
func DefaultTemplates() map[string]string // every legal.* key the catalog emits
// Errorf lets IMPERATIVE code (LegalCustom bodies, overrides) return
// structured failures: implements error, carries LegalMessage.
func Errorf(templateKey string, bindings map[string]boardgame.LegalBindingValue) error
// RenderMessage fills {placeholders} from bindings; missing binding renders
// the placeholder name (never panics).
func RenderMessage(m *boardgame.LegalMessage, table map[string]string) string
```

```go
// package boardgame — LegalVerdict.Error():
// returns nil for Pass; for Fail/Unknown returns a concrete *LegalError
// (never a typed-nil interface). LegalError.Error() renders via the game's
// template table when attached, else falls back to the template key.
type LegalError struct { Verdict LegalVerdict; table map[string]string }
```

Boot validation contract (consumed by Task 8): `func validateLegalTemplates(specs []LegalSpec, predicates []*LegalPredicate, table map[string]string) error` — every `Spec.Message` key and every template a catalog predicate can emit must exist in `DefaultTemplates() ∪ delegate table`; error names move + key.

- [ ] **Step 1:** failing tests: render with bindings; missing-binding renders placeholder; Errorf round-trips through `errors.As`; DefaultTemplates covers every key emitted anywhere in `legal/catalog_*.go` (test greps the catalog source? No — test evaluates every corpus fail-case and asserts its template key ∈ DefaultTemplates); Verdict.Error() nil-interface test (`var err error = verdict.Error(); err == nil` for Pass).
- [ ] **Step 2-3:** implement, green. **Step 4:** commit `legal: template table, rendering, Errorf, fail-case key coverage (spec §6)`.

---

### Task 7: moves wiring — options, contributions, wrapper predicates

**Files:** Modify `moves/with.go`, `moves/default.go`, `moves/current_player.go`; create `moves/preconditions.go`, `moves/preconditions_test.go`. Create `legal/catalog_framework.go` (inPhase/inProgression/stackConstraints predicate wrappers).

**Interfaces:**
- Consumes: everything from T1-T6.
- Produces:

```go
// moves/with.go — same idiom as existing options (with.go:38-266)
const configPropPreconditions = "github.com/jkomoros/boardgame/moves.Preconditions"        // []legal.Spec
const configPropSuppressedPreconditions = "github.com/jkomoros/boardgame/moves.SuppressedPreconditions" // []string
func WithPreconditions(specs ...legal.Spec) interfaces.CustomConfigurationOption
func WithoutPrecondition(name string) interfaces.CustomConfigurationOption

// moves/preconditions.go
// PreconditionsProvider is the optional interface core consults (spec §3).
// Default implements it; CurrentPlayer overrides to append proposer atom.
type PreconditionsProvider interface {
    ContributedPreconditions() []legal.Spec
}
func (d *Default) ContributedPreconditions() []legal.Spec   // derives inPhase spec from
    // configPropLegalPhases, inProgression from configPropLegalMoveProgression,
    // stackConstraints from configPropSourceProperty/DestinationProperty — only
    // for keys actually present in the config bag (moves/with.go:8-31)
func (c *CurrentPlayer) ContributedPreconditions() []legal.Spec // Default's + ProposerIsCurrentPlayer()
// DeclaredPreconditions returns authored specs (nil = not opted in) — read by core.
func (d *Default) DeclaredPreconditions() ([]legal.Spec, []string) // specs, suppressions
```

Wrapper predicates (in `legal/catalog_framework.go`) reuse the EXISTING logic — do not reimplement:
- `"inPhase"`: constructor closes over the phase list; Evaluate calls the same
  delegate-CurrentPhase + TreeEnum-ancestor logic as `moves/default.go:425-469`.
  To avoid moves↔legal import cycle, the shared helper moves to core as an
  exported-in-package function (`boardgame.legalInPhaseCheck(state, phases) *LegalVerdict`-style —
  put the extracted helper in `legal_predicate.go` or a small `legal_framework.go`
  in core, and have BOTH moves/default.go's frozen chain and the predicate call it;
  byte-identical error string preserved: `"Move is not legal in phase " + name`).
- `"inProgression"`: same treatment for `default.go:561-631` (matchTape); Reads `game.moveHistory` FacetValues + phase.
- `"stackConstraints"`: wraps the `default.go:359-388` MayMoveTo pre-check.
- `"proposerIsCurrentPlayer"` already exists (T5).

CRITICAL frozen-chain rule: extraction refactors must keep `moves/default.go`'s
imperative path calling the SAME extracted helpers with the SAME order and
strings. Add a string-freeze test in `moves/preconditions_test.go`: table of
(state, move) → expected error string, captured from the CURRENT
implementation BEFORE refactoring (write the test first against today's code,
then refactor, test stays green).

- [ ] **Step 1:** string-freeze test against today's chain (memory + blackjack test fixtures already exist in moves tests — grep `moves/*_test.go` for existing game fixtures to reuse) → PASS (baseline).
- [ ] **Step 2:** failing tests for WithPreconditions round-trip through auto.Config; ContributedPreconditions contents for Default (empty config → empty; phases configured → inPhase spec) and CurrentPlayer.
- [ ] **Step 3:** implement options + providers + extract-and-share helpers + framework wrapper predicates. String-freeze test must stay green throughout.
- [ ] **Step 4:** full `go test ./moves/... ./legal/... .` green. Commit `moves: precondition options, contributions, framework predicate wrappers (spec §2)`.

---

### Task 8: Plan assembly, probe, opt-in Legal path, purely-sugar property tests

**Files:** Create `legal_plan.go`, `legal_plan_test.go` (core); modify `moves/default.go` (Legal opt-in branch), `game_manager.go` (boot validation call site — find where moves are installed/validated at NewGameManager, grep `installMoves\|moveConfigs` in game_manager.go).

**Interfaces (Produces):**

```go
// package boardgame
type legalPlan struct {
    fieldIndependent []*LegalPredicate
    fieldDependent   []*LegalPredicate // includes proposerIsCurrentPlayer
    custom           *LegalPredicate   // CustomLegaler wrapper or nil
    allReads         []LegalRead
    specs            []LegalSpec       // serializable subset, for the ledger
}
type CustomLegaler interface {
    LegalCustom(state ImmutableState, proposer PlayerIndex) error
}
// buildLegalPlan: contributions (base-first) + authored − suppressions;
// resolve via registry; split buckets by "has move.* read"; wrap CustomLegaler.
// Boot errors: unsupported base type (declared specs but no PreconditionsProvider
// in the embed chain → name the base type), path/template validation (T2/T6),
// orphaned declarations (probe below).
func buildLegalPlan(m *GameManager, config MoveConfig) (*legalPlan, error)
// evaluateLegalPlan: order per spec §4; short-circuit on Fail for hot paths;
// full-ledger mode for move-forms (evaluates everything, returns []LegalVerdictEntry).
func (p *legalPlan) evaluate(ctx LegalContext, fullLedger bool) (LegalVerdict, []LegalVerdictEntry)
type LegalVerdictEntry struct {
    Name string; Args []string; Verdict LegalVerdict
    Serializable bool; FieldDependent bool; Reads []LegalRead
}
```

**The probe (spec, prime guarantee rule 4):** at boot, for each move type with declared preconditions, call `exampleMove.Legal(probeState, ObserverPlayerIndex)` where `probeState` is a sentinel the engine recognizes (`type legalProbeState struct{ ImmutableState }` with a marker method, or a package-private context flag on the manager set during probing — pick the implementation that requires NO change to the public Legal signature: a manager-scoped `probing bool` + `probeReached bool` pair that `Default.Legal()` checks-and-sets first thing works and is race-free at boot). If `probeReached` is false after the call → boot error: `move %q declares preconditions but its Legal() override never reaches moves.Default.Legal — declarations would be dead`.

**Default.Legal() opt-in branch** (`moves/default.go:339`):

```go
func (d *Default) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if probeHandled(state, d) { return nil }        // probe short-circuit (new)
    if plan := d.legalPlan(); plan != nil {          // opted-in: plan evaluation
        verdict, _ := plan.evaluate(...)             // short-circuit mode
        return verdict.Error()                       // rendered via game's table
    }
    // FROZEN CHAIN, untouched:
    ...
}
```

**Purely-sugar property tests** (spec §9, the campaign's conscience):
(a) run the full existing `go test ./...` — zero failures IS the frozen-chain
test for framework games; additionally a table test asserting for N recorded
(state, move, proposer) triples from the memory/blackjack fixtures that
`Legal()` output (nil-ness AND string) is identical before/after — capture the
expected values in the test from the string-freeze table of T7.
(b) probe tests: wholesale override + declarations → boot error naming move;
super-calling override + declarations → boots fine AND plan actually evaluates
(assert by declaring an always-fail predicate and seeing its template in the
error).
(c) declared-but-unsupported base type (embed `moves.DealCountComponents` +
WithPreconditions) → boot error naming the base type.

- [ ] Steps: failing tests (probe cases, plan bucket split incl. proposer-is-field-dependent, zero-verdict fail-closed through plan) → implement → green → full `go test ./...` → commit `core: legality plan, boot probe, opt-in Legal path + purely-sugar property tests (spec §4, prime guarantee)`.

---

### Task 9: Engine wins — phase index, agnostic bucket, memos, #65 logging

**Files:** Create `legal_index.go`, `legal_index_test.go` (core); modify `game.go` fixup loop (`:562,647`) and `server/api` move-forms caller to iterate the index (T10 wires the server; here expose the API).

**Interfaces (Produces):**

```go
// package boardgame
// candidateMoves returns moves possibly legal in state's current phase:
// phaseIndex[currentPhase ∪ TreeEnum ancestors] ∪ phaseAgnostic.
// phaseAgnostic contains: every opaque move (no plan) and every opted-in move
// with no inPhase atom. SUPERSET PROPERTY (tested): result ⊇ any move whose
// Legal() could return nil.
func (g *GameManager) candidateMoves(state ImmutableState) []Move
// fieldIndependentMemo keyed (moveTypeName, stateVersion, proposer) — bounded
// to the current head version (evict on version advance).
// tapeMemo: historicalMovesSincePhaseTransition memoized per version (retires
// moves/default.go:475 TODO — expose a hook so the inProgression predicate and
// the frozen chain share it).
```

Fixup loop change (`game.go:647` area): iterate `candidateMoves` instead of all moves — but ONLY as a filter around the existing delegate `ProposeFixUpMove` flow; ground-truth the current flow first: the delegate proposes (base/game_delegate.go:86 iterates); the safe v1 integration is exposing `candidateMoves` for `base.GameDelegate.ProposeFixUpMove` to iterate (modify `base/game_delegate.go:86`) — core loop untouched. #65: in `applyMove`'s fixup-rejection path (game.go:887 vicinity), log rejections at debug: `manager.Logger().Debugln("fixup rejected", move, predicateName, renderedMsg)` — the structured verdict is available for plan moves; opaque moves log the plain error string.

- [ ] Steps: failing tests (superset property with a mixed opaque/opted-in/no-phase move set; TreeEnum ancestor lookup; memo hit across double evaluation; memo eviction on version bump; tape shared between two inProgression moves) → implement → green → `go test ./...` → commit `core: phase index + agnostic bucket, memos, fixup rejection logging (spec §5, #65, #640)`.

---

### Task 10: Server ledger + chest templates + TS types

**Files:** Modify `server/api/main.go:1590-1657`; find chest JSON marshal (grep `MarshalJSON` in component_chest.go) for template injection; modify `server/static/src/types/api.d.ts`, `server/static/src/selectors.ts`.

**Interfaces (Produces — wire format, spec §6):**

```go
type preconditionEntry struct {
    Name        string                    `json:"name"`
    Args        []string                  `json:"args,omitempty"`
    Verdict     string                    `json:"verdict"` // "pass"|"fail"|"unknown"
    Message     *legalMessageJSON         `json:"message,omitempty"`
    Evaluable   bool                      `json:"evaluable"`
    Provisional bool                      `json:"provisional,omitempty"`
}
// moveForm gains: Preconditions []preconditionEntry `json:",omitempty"`
//                 (nil for opaque moves — legacy shape untouched)
// info response gains: LegalCatalogVersion int; chest JSON gains "LegalTemplates": {...}
```

Rules to implement exactly (spec §6): `evaluable = entry.Serializable ∧ every Read's facet survives this viewer's sanitization policy` (use `facetSurvives` from T2 against the policy the sanitization transformation computes for that viewer — grep `sanitizationTransformation` in sanitization.go for the per-property policy lookup); **#693 guard**: when `evaluable==false`, strip `Message.Bindings` (ship template key only); `provisional = entry.FieldDependent`; full-ledger evaluation ONCE per request replaces the second admin `Legal()` call for opted-in moves (`main.go:1611-1623`: LegalForPlayer/LegalForAnyone derived from the ledger where a plan exists; frozen path for opaque moves — strings byte-identical).

TS: add `PreconditionEntry` + `MoveForm.Preconditions?: PreconditionEntry[]` to `api.d.ts`; extend `selectMoveLegality` to pass the ledger through (`MoveLegalityInfo.preconditions?`). No UI feature in this task; `cd server/static && npm run type-check` green.

- [ ] Steps: Go failing tests in `server/api` (ledger shape for an opted-in move; bindings stripped when inevaluable; legacy fields byte-identical for opaque moves — assert against a recorded pre-change JSON fixture) → implement → green → TS types + type-check → full `go test ./...` → commit `server: precondition ledger, chest templates, catalog version (spec §6, #693 guard)`.

---

### Task 11: Golden harness + migrate memory & blackjack

**Files:** Create `moves/golden_legal_test.go` harness (or extend existing golden machinery — grep `golden` package usage in examples); modify `examples/memory/moves.go` + `examples/memory/main.go` (delegate: ConfigureLegalTemplates), `examples/blackjack/{moves.go,main.go}`.

Golden-equivalence harness: for a migrated move, table of recorded states (reuse each game's existing golden JSON under `examples/<game>/testdata` — grep for how golden tests load states); assert old-Legal (kept temporarily as `legacyLegal` private copy in the test file) and plan-Legal agree on nil-ness and message for every (state, proposer) pair.

Migrations per spec §8 exactly: memory `moveRevealCard` (delete Legal(), WithPreconditions(PropAtLeast + RevealableCardAt + MayMoveToSlot), templates `reveal.*` in delegate table); memory's two card-type-compare moves → `LegalCustom` with `legal.Errorf`. Blackjack `moveStartRoundCleanup` (AllActivePlayers(Any(PlayerBool,PlayerBool))); hand-value moves → `LegalCustom`.

- [ ] Steps: harness + failing goldens → migrate → goldens green → `go test ./examples/...` green → commit per game (`memory: migrate to declarative legality`, `blackjack: ...`).

---

### Task 12: Migrate checkers + survey-driven remaining examples

**Files:** `examples/checkers/{moves.go,main.go,components.go}`; then survey `examples/{tictactoe,pig,werewolf,debuganimations}/moves.go`.

Checkers per spec §8: game-registered `checkers.spaceIsBlack` via `ConfigurePredicateConstructors` (ExtendDefaults), declarative gates (ComponentPresentAtKey, ComponentPropEqualsCurrentPlayer, spaceIsBlack), capture-graph walk → `LegalCustom` returning `legal.Errorf("checkers.illegal_dest", nil)`. Then survey each remaining example: migrate moves the catalog covers (goldens each); leave hard-custom in LegalCustom; moves embedding unsupported base types stay untouched (document per game in the commit message).

- [ ] Steps: checkers first (failing goldens → migrate → green → commit), then one commit per surveyed game.

---

### Task 13: ../games migration

**Files:** `/Users/jkomoros/Code/go/src/github.com/jkomoros/games/{murdermrmonroe,pass,valentine,darwin,metaltrader}/moves.go` etc. — survey ALL five (darwin/metaltrader have Go moves even without clients).

- [ ] `git -C ../games checkout -b declarative-legality-design`.
- [ ] Survey every game's Legal() overrides; migrate catalog-covered moves with goldens (the games repo has its own tests — `go test ./...` there must stay green; check its go.mod replaces/points at the boardgame repo — if it pins a released version, add a `replace github.com/jkomoros/boardgame => ../boardgame` for the branch and note it in the commit).
- [ ] Commit per game in ../games; note un-migratable moves + why in commit messages.

---

### Task 14: Conformance finalization, docs, close-out

- [ ] Corpus completeness test: every registry name in `DefaultConstructors()` has a conformance file with ≥3 cases (pass/fail/unknown where expressible) — write the meta-test.
- [ ] Docs: `legal/doc.go` full package doc (authoring guide: the three rules of the catalog, purely-sugar guarantee, escape hatch, template tables); update `moves/doc.go` Legal section; `TUTORIAL.md` — grep for the Legal() teaching section and add the declarative alternative.
- [ ] Full verification: `go build ./... && go vet ./... && go test ./...` (boardgame) + same in ../games + `cd server/static && npm run type-check` + one Playwright smoke (`npx playwright test tests/animations/waapi-gate.spec.ts -g blackjack --reporter=line` with dev server, per OFFLINE_DEV_MODE.md, to prove the server changes didn't break move-forms in the real client).
- [ ] Close-out commit + summary: per-issue outcomes (#761 #189 #790 #644 #65 #640 partial #213-enabling), migration coverage stats (N moves declarative / M LegalCustom / K untouched), deferred items (TS evaluator, dirty-tracking, seam expansion).

---

## Self-review (run at planning time)

- **Spec coverage:** §1 types/paths/registry/anti-tarpit → T1-T5; §2 options/seam → T7-T8; §3 layering (Legal* core + aliases) → T1/Global; §4 plan/order/escape-hatch/probe → T8; §5 index/memos/honest-table → T9; §6 templates/ledger/facets/#693/corpus → T6/T10/T4; §7 progression+#644 → T7 wrapper + RepeatFromProp… **gap: #644 RepeatFromProp not tasked** — add to T7 Step 3: `moves.RepeatFromProp(path string)` ValidCounter variant + `Satisfied` gaining context plumbing per spec §7, with its own test. (Folded into T7.)
- **Placeholder scan:** T2 Step 3 directs a grep for the current-player accessor rather than naming it — acceptable directed-discovery, not a TBD; no TODO/TBD elsewhere.
- **Type consistency:** `LegalSpec`/`legal.Spec` alias used consistently; `evaluate(ctx, fullLedger)` signature consumed by T10 as written; `facetSurvives` (T2) consumed by T10; `LegalVerdictEntry` fields match T10's `preconditionEntry` mapping; stable atom names match T7 wrappers and Global Constraints.
