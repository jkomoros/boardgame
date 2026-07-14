# Declarative Move Legality — Design

**Date:** 2026-07-10 (rev 2, post-critique)
**Issues:** #761, #189 (the core thread); folds in #790, #644, #65; constrains #693, #44; enables #640, #213, #295
**Branch:** `declarative-legality-design`
**Status:** Design for review (implementation plan follows approval)

## Provenance

Synthesized from a three-lens design panel (composability / engine / DX) plus
two adversarial critiques of the panel outputs, then revised against four
adversarial critiques of the spec itself (Go API/idiom, codebase contact,
purely-sugar guarantee, client/TS future). Panel and critique artifacts (not
normative, checked in for the record):
`docs/superpowers/design/2026-07-10-declarative-legality/` — the design brief,
three lens designs (A-composability, B-engine, C-delight), two panel critiques
(acid, systems), and four spec critiques (goapi, client, sugar, contact).
Finding-by-finding dispositions:
`docs/superpowers/specs/2026-07-10-declarative-legality-critique-responses.md`.

## Problem

`Move.Legal(state ImmutableState, proposer PlayerIndex) error` is imperative Go.
Three consequences:

1. **The engine can't optimize it.** The fixup loop polls candidate moves'
   Legal() after every state change (up to 256 recursions), and the server runs
   every non-fixup move's Legal() twice (player + admin) per move-forms request,
   per state version, uncached (#640).
2. **The client can't evaluate it.** Moves can't be grayed out or rejected
   without a round trip (#189/#213); the client gets only post-hoc booleans.
3. **Nothing can explain it.** Error strings are ad-hoc `errors.New`; fixup
   rejections are silently discarded (#65).

Meanwhile `moves.Default.Legal()` already secretly runs three declarative checks
from its config bag — phase membership, move-tape progression, and stack
constraints. This design promotes that buried pattern to a first-class,
inspectable, serializable system.

## The prime guarantee: purely sugar

**`Legal(state, proposer) error` remains the ground-truth contract, unchanged —
and the existing imperative chain is FROZEN.**

The critique process showed the naive version of this guarantee ("Default.Legal
becomes 'evaluate plan'") silently breaks the dominant real-world pattern —
every in-repo and `../games` move embeds a framework type, overrides `Legal()`,
and super-calls the chain (`if err := m.CurrentPlayer.Legal(state, proposer);
err != nil {...}`). The revised, normative rules:

1. **The imperative chain is frozen.** `Default.Legal()`,
   `CurrentPlayer.Legal()`, and every other framework move type's `Legal()`
   keep their current implementations, byte-for-byte observable behavior:
   same checks, same order, same error strings. Un-migrated games — including
   games that never migrate — behave identically to today, including the exact
   `LegalForPlayerError` strings legacy clients see.
2. **Plan-based evaluation is opt-in per move type.** A move opts in by
   declaring `moves.WithPreconditions(...)` in its `auto.Config` AND not
   overriding `Legal()` (it may implement `LegalCustom`, §4). For opted-in
   moves, `Default.Legal()` detects the declarations and evaluates the plan
   instead of the frozen chain — declaring is implementing.
3. **Everything engine-side is opportunistic.** Moves without declarations are
   *opaque*: the engine calls their `Legal()` exactly as today, places them in
   every phase bucket (§5), caches nothing about them, and reports them to the
   client exactly as today. No capability of this design ever requires a game
   to adopt it.
4. **Dead declarations are a boot error, detected behaviorally.** Go cannot
   see method overrides statically, so `NewGameManager` runs a one-time
   **probe**: for each move type with declared preconditions, it calls the
   example instance's `Legal()` against a sentinel probe state that
   `Default.Legal()` recognizes and records before doing anything else. If the
   probe never reaches `Default.Legal()`, the declarations can never execute —
   boot fails naming the move and explaining that its wholesale `Legal()`
   override orphans its declarations. A `Legal()` override that super-calls
   into the chain passes the probe and is fully supported: the super-call
   evaluates the plan, and the override's own imperative checks compose around
   it — exactly today's embedding pattern. (`LegalCustom` remains the
   preferred way to add imperative residue; override-plus-super-call is the
   compatible legacy spelling.)

## Design decisions locked before the panel

- **Design for client evaluation, ship server-first.** The representation is
  serializable and sanitization-aware from day one; the TypeScript evaluator is
  a designed-for follow-up.
- **Migrate freely.** In-repo example games and the three `../games` clients
  are migrated to declarations where the catalog covers them; imperative
  Legal() keeps working everywhere, forever.

---

## 1. Representation

One new package, `legal`, sitting beside `constraints` and deliberately rhyming
with it (name + string args, constructor registry).

```go
// package legal

// Outcome is a three-valued verdict. The zero value is deliberately INVALID so
// a forgotten Verdict fails closed (an accidentally-zero Verdict must never
// read as "legal").
type Outcome int

const (
    outcomeInvalid Outcome = iota // zero value: fails closed, reported as engine error
    Pass
    Fail
    Unknown
)

// BindingValue keeps Bindings JSON-round-trippable and TS-conformant: no
// arbitrary `any` in the wire format.
type BindingValue struct { // exactly one field set
    S *string
    I *int
    B *bool
}

// Message is a template KEY plus named bindings — never a pre-baked string —
// so failures are localizable, greppable, and re-renderable on server or
// client. Template keys resolve through the game's template table (§6).
type Message struct {
    Template string
    Bindings map[string]BindingValue
}

// Verdict is the result of evaluating one predicate.
type Verdict struct {
    Outcome Outcome
    Message *Message // set on Fail (optionally on Unknown); nil on Pass
    Reason  string   // on Unknown: why ("reads hidden property HiddenCards")
}

func PassVerdict() Verdict
func FailT(template string, bindings ...map[string]BindingValue) Verdict
func UnknownVerdict(reason string) Verdict
```

### Predicates are structs-with-a-func, not interfaces

Following the repo's own precedent (`StackConstraint` is a func type, not an
interface — four of the five proposed interface methods were immutable
getters):

```go
// Context is the entire vocabulary a predicate may reference — the wall
// before the Turing tarpit: no I/O, no mutation, nothing beyond these four.
type Context struct {
    State    boardgame.ImmutableState
    Move     boardgame.Move // nil during field-independent evaluation (§4)
    Proposer boardgame.PlayerIndex
    Chest    *boardgame.ComponentChest
}

// Predicate is one resolved legality question.
type Predicate struct {
    Name     string     // registry identity, e.g. "playerPropAtLeast"
    Args     []string   // with Name, round-trips the registry
    Reads    []Read     // declared read-set (conservative over-approximation)
    Cost     Cost       // Trivial | Cheap | Moderate | Expensive
    Evaluate func(ctx Context) Verdict // pure
}

// Read is a property path plus the FACET the predicate needs from it — this
// is what makes client evaluability precise under sanitization (§6): a
// stack-size check needs only the count facet, which PolicyLen preserves.
type Read struct {
    Path  PropPath
    Facet Facet // FacetValues | FacetCount | FacetOccupancy | FacetOrder
}
```

### Wire format: leaf ≡ node, with message

A compositor is a spec whose children are specs. One serialized shape:

```go
// Spec is the serializable, registry-resolvable form.
type Spec struct {
    Name    string `json:"name"`
    Args    []string `json:"args,omitempty"`
    Sub     []Spec `json:"sub,omitempty"`
    Message string `json:"message,omitempty"` // template-key override
}

// Builders return Spec by value; WithMessage sets Spec.Message.
func PropAtLeast(path string, n int) Spec
func (s Spec) WithMessage(templateKey string) Spec
```

```jsonc
{"name": "playerPropAtLeast", "args": ["player.CardsLeftToReveal", "1"],
 "message": "reveal.no_cards_left"}

{"name": "any", "sub": [
  {"name": "playerBool", "args": ["Eliminated"]},
  {"name": "playerBool", "args": ["Stood"]}
]}
```

Constructors mirror `constraints.StackConstraintConstructor`; **games register
their own predicates through the same registry** (this is how checkers'
board-geometry check stays serializable, §8):

```go
type PredicateConstructor struct {
    Name        string
    Constructor func(spec Spec, chest *boardgame.ComponentChest,
        resolve func(Spec) (*Predicate, error)) (*Predicate, error)
}
```

The registry is consumed from the delegate via **type-assertion on an optional
interface** (never a new `GameDelegate` method — that would be a compile break
for every existing delegate):

```go
// package legal — implemented optionally by delegates; base.GameDelegate
// does NOT need changes. Absence = DefaultConstructors().
type ConstructorConfigurer interface {
    ConfigurePredicateConstructors() []*PredicateConstructor
}
```

### Anti-tarpit rules (normative)

1. **`any` is the only compositor in v1**, registry-enforced to depth 1. No
   `all` (the plan's ordered list IS the conjunction); no first-class `not`
   (a Kleene-`Not` doubles the Go↔TS conformance surface). Framework move
   types whose semantics are negated/conditional (`ApplyUntil`,
   `ApplyUntilCount`, `RoundRobin`) stay opaque in v1 rather than bending this
   rule.
2. **Branchy logic becomes a purpose-built named predicate with hand-written
   Go Eval** — proven on the acid test: memory's two-branch disambiguation is
   a 12-line `revealableCardAt`, not DSL surface.
3. **Catalog growth rule:** if you can't say it as a relation over a path,
   push the computation into a computed state property or the escape hatch.
   No user arithmetic, loops, or lambdas in serialized form, ever.

### Path grammar (net-new build, honestly priced)

`constraints/prop_path.go` resolves a single component instance; the state
resolver is a **new build**: `game.X`, `player.X` (current player),
`players[*].X` (quantified predicates only), `move.X`. All paths — and all
template keys (§6) — are validated at `NewGameManager`: a typo fails at boot
naming the move and path, never mid-game. Runtime guard: a predicate whose
declared reads omit `move.*` but which touches `ctx.Move` when nil returns
`UnknownVerdict("undeclared move read")` rather than panicking.

---

## 2. Attachment & composition

### Authoring surface

```go
moves.WithPreconditions(specs ...legal.Spec) // opt in + append, in order
moves.WithoutPrecondition(name string)        // suppress an inherited one
```

Struct tags are not in v1 (cut per YAGNI — the panel's own migrations never
used them).

### The composition seam — v1 scope: Default and CurrentPlayer only

The critique ground-truthed the `moves/` package: beyond
`Default`/`CurrentPlayer`, ~24 move types carry real legality logic
(`DealComponents` counting rounds, `FinishTurn` readiness, seat-management
checks, `ForceFinishTurn` which deliberately super-calls *nothing*, negated
`ApplyUntil` family). Modeling all of them declaratively in v1 is neither
necessary nor honest. Normative v1 scope:

- **`Default` and `CurrentPlayer` contribute declaratively** via a
  data-returning chain that mirrors today's super-call pattern, written once
  in the framework:

```go
func (d *Default) ContributedPreconditions() []legal.Spec {
    // inPhase / inProgression / stackConstraints — derived from the config
    // bag (WithLegalPhases / WithLegalMoveProgression / WithSourceProperty
    // keep working; they are now also readable as specs).
    return d.specsFromConfig()
}
func (c *CurrentPlayer) ContributedPreconditions() []legal.Spec {
    return append(c.Default.ContributedPreconditions(),
        legal.ProposerIsCurrentPlayer())
}
```

- **Every other framework move type is opaque in v1**: its frozen `Legal()`
  runs as today. A game move embedding, say, `DealCountComponents` cannot opt
  in to plans in v1 (fail-fast at boot if it tries, with a message naming the
  unsupported base type). Extending contribution to more base types is
  follow-up work, one type at a time, each with golden-equivalence tests.
- A move may opt out of inherited contributions entirely with
  `WithoutPrecondition` per name (stable names: `"inPhase"`,
  `"inProgression"`, `"stackConstraints"`, `"proposerIsCurrentPlayer"`) — the
  `ForceFinishTurn` inherit-nothing pattern, now expressible.

Plan assembly, per opted-in move type, at `NewGameManager`:
`ContributedPreconditions()` (base-first, deterministic) + authored
`WithPreconditions` (declaration order) − `WithoutPrecondition` suppressions.

---

## 3. Layering

```
boardgame (core)
    Verdict, Outcome, Message, PropPath, Read, Facet, Spec, Cost (value types)
    the evaluation engine: plan build, buckets, phase index, memo,
    move-form ledger assembly   (lives here: game.go's loops call it)
    optional interfaces consumed by type-assertion:
        PreconditionsProvider  (moves.Default implements)
        CustomLegaler
        legal.ConstructorConfigurer / legal.TemplateConfigurer (on delegates)
        ▲
moves
    ContributedPreconditions chain on Default/CurrentPlayer
    WithPreconditions / WithoutPrecondition
    Default.Legal(): frozen chain, OR plan evaluation for opted-in moves
        ▲
legal (new; peer of constraints)
    predicate catalog + DefaultConstructors()
    Errorf (template-errors from imperative code) + template rendering
```

Dependency arrows point downward only; core holds types + engine, zero game
semantics. The structured `Verdict` (bindings and all) is what crosses the
core boundary; nothing flattens to a rendered string until `Verdict.Error()`
adapts it at the `Legal()` return (rendering against the template table, §6,
with the template key as fallback text if unregistered — but unregistered keys
are a boot error anyway).

---

## 4. Evaluation semantics

Per opted-in move type, built once at `NewGameManager`:

```go
type PreconditionPlan struct {
    fieldIndependent []*legal.Predicate // no move.* reads
    fieldDependent   []*legal.Predicate // includes proposerIsCurrentPlayer:
                                        // it reads move.TargetPlayerIndex
    custom           *legal.Predicate   // LegalCustom wrapper, or nil
    allReads         []legal.Read
}
```

- **Evaluation order is plan order: contributed atoms first (base-first), then
  authored atoms in declaration order; `custom` always last. No Cost-sorting
  in v1** — what you declare is what runs, in the order you wrote it (least
  surprise; migrated moves keep their historical first-failure messages
  without snapshot churn). Field-independent predicates are memoized
  individually, so caching never changes this observable order. `Cost` stays on every predicate as metadata: docs
  and lints use it ("expensive predicate declared before cheap ones"), and a
  future opt-in reordering can use it without a representation change.
  Deterministic order ⇒ the same state always reports the same failure.
- Classification: field-independent predicates may be memoized; field-dependent
  predicates require a bound move. Both remain in plan order, followed by
  custom. Note the critique correction: the proposer check is
  field-**dependent** (it reads `move.TargetPlayerIndex`), so the plan
  preserves today's error-precedence for turn violations.
- **Hot paths short-circuit** on first Fail (fixup loop, ProposeMove).
  **Move-forms assembly evaluates the full ledger** (once per request).
- `Default.Legal()` for opted-in moves: evaluate plan; return first failure's
  `Verdict.Error()` or nil. For everything else: the frozen chain (§ prime
  guarantee).

### The escape hatch

```go
type CustomLegaler interface {
    // Runs after all declarative preconditions pass; the imperative residue.
    LegalCustom(state ImmutableState, proposer PlayerIndex) error
}
```

Wrapped as an opaque predicate (`Reads` unknown, `CostExpensive`, no
serialized form): runs last, never cached, client sees `"unknown"`. Imperative
bodies may return `legal.Errorf("checkers.illegal_dest", nil)` to stay
structured; plain errors are wrapped as opaque one-off templates.
`LegalCustom` + wholesale `Legal()` override on the same type = boot error.

---

## 5. Engine wins (honest table)

| Mechanism | v1 | Effect |
|---|---|---|
| **Phase bucketing** — `phaseIndex[phase] = opted-in moves whose inPhase admits it (∪ TreeEnum ancestors)`, **plus a phase-agnostic bucket containing every opaque move and every opted-in move with no inPhase spec**; lookups take `phaseIndex[current] ∪ phaseAgnostic` | ✅ | Fixup loop and move-forms skip declaratively-impossible moves with zero evaluations (#640); opaque moves are never skipped (superset property, tested §9) |
| **Declaration-order short-circuit** (opted-in moves) | ✅ | Contributed cheap gates (phase, proposer) run before authored checks; residue always last. Cost metadata powers a lint, not a reorder (v1) |
| **Field-independent memo** keyed `(moveType, stateVersion, proposer)` | ✅ | Move-forms' player+admin double pass computes the stable half once |
| **Tape memoization** per version | ✅ | All `inProgression` predicates share one tape walk (retires the default.go:475 TODO) |
| **Dirty-tracking** by read/write path intersection | ❌ deferred | Write-set capture must be complete or the legality cache is stale-and-wrong (correctness, not perf). `Reads` metadata makes it addable behind a future audit; v1 invalidates per version. |

Honest framing from critique: with v1's seam scope (opaque framework types in
every bucket), the bucketing win grows as games migrate — it is proportional
to adoption, not automatic. What stays O(Legal): opaque moves and residue, by
design, now gated behind cheap declarative checks where adopted.

---

## 6. Explainability

### The template table (critique: this had no home; now it does)

Template keys resolve through a per-game table configured on the delegate via
an optional interface and **shipped to the client inside the chest JSON**,
exactly the channel enums already ride (`expandMoveForms` precedent):

```go
// package legal — optional on delegates; validated at NewGameManager:
// every template key referenced by any Spec/FailT/Errorf must exist.
type TemplateConfigurer interface {
    ConfigureLegalTemplates() map[string]string
    // {"reveal.no_cards_left": "You have no cards left to reveal this turn"}
}
```

The catalog ships defaults for built-in predicate failures
(`legal.DefaultTemplates()`); games extend/override. Server-side rendering
(`Verdict.Error()`, fixup logs) and the future client renderer read the same
table.

### Server

Every declarative failure carries `Message{Template, Bindings}`. The fixup
loop logs rejections at debug level (`fixup rejected move=X predicate=Y
msg=...`) — #65, no exceptions.

### Client contract

Move forms gain a per-predicate ledger alongside the preserved (and for
un-migrated moves, byte-identical) `LegalForPlayer` /
`LegalForPlayerError` / `LegalForAnyone`:

```jsonc
"Preconditions": [
  {"name": "proposerIsCurrentPlayer", "verdict": "pass", "evaluable": true,
   "provisional": true},   // field-dependent: verdict used server-set defaults
  {"name": "playerPropAtLeast", "args": ["player.CardsLeftToReveal", "1"],
   "verdict": "fail",
   "message": {"template": "reveal.no_cards_left", "bindings": {"left": 0}},
   "evaluable": true},
  {"name": "custom", "verdict": "unknown", "evaluable": false}
]
```

- `provisional: true` marks field-dependent verdicts (computed against
  `DefaultsForState` bindings — a different field choice could differ; the
  `LegalForAnyone` analog at predicate granularity).
- `evaluable` is per predicate, per viewer, per **facet**:
  `evaluable = has serialized form ∧ every Read's Facet survives this viewer's
  sanitization` — `FacetCount` survives `PolicyLen`; `FacetOccupancy` survives
  `PolicyOrder` (which is why memory's whole plan is client-evaluable —
  ground-truthed in critique); `FacetValues` requires `PolicyVisible`.
  An `any` compositor is evaluable iff all children are (Kleene-honest).
- **#693 guard:** when `evaluable: false`, the ledger ships verdict + reason
  only — never bindings derived from state the viewer can't see.
- A `catalogVersion` stamp ships with the ledger; a client with an older
  catalog treats unknown predicate names as `evaluable: false` and defers to
  server verdicts (graceful skew).

### Go↔TS conformance (designed-for deliverable)

The representation deliverable includes a **shared JSON conformance corpus**:
for every catalog predicate, a table of `(spec, context-fixture, expected
verdict)` cases checked by the Go tests in this campaign and, later, by the TS
evaluator's test suite verbatim. Divergence = failing test on either side.
This is the mechanism that keeps two evaluators honest, and it's why the
catalog stays small and `not` stays out.

---

## 7. Progression & #644

Move-tape matching becomes the `inProgression` predicate wrapping the existing
`matchTape` machinery — same plan/cache/explain path (`Reads:
[game.moveHistory/FacetValues]`, `CostModerate`).
`MoveProgressionGroup.Satisfied` gains access to `legal.Context` (named
plumbing change), unlocking #644: `moves.RepeatFromProp("game.RoundsThisTurn")`
resolves its count against live state at match time; the backing path joins
the predicate's read-set mechanically.

---

## 8. Migrations (the acid tests)

Corrected survey (critique ground-truthing): across the example games, the
catalog as specified covers phase / current-player / stack-size / presence /
property-compare checks; **4 moves are hard-custom** (memory's two card-type
comparisons, blackjack's hand-value arithmetic, checkers' capture graph) and
stay in `LegalCustom`; **~11 player-quantifier checks** (blackjack, werewolf)
are covered by the `AllActivePlayers` quantifier primitive. Framework move
types beyond Default/CurrentPlayer stay opaque in v1 (§2).

### memory/moveRevealCard — fully declarative, Legal() deleted

```go
auto.Config(new(moveRevealCard),
    moves.WithPreconditions(
        legal.PropAtLeast("player.CardsLeftToReveal", 1).
            WithMessage("reveal.no_cards_left"),
        legal.RevealableCardAt("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
        legal.MayMoveToSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
    ),
)
```

`RevealableCardAt` (catalog, purpose-built, `Reads` facets: occupancy only):

```go
Evaluate: func(ctx legal.Context) legal.Verdict {
    idx := intField(ctx.Move, p.field)
    if stackAt(ctx, p.hidden).ImmutableComponentAt(idx) != nil {
        return legal.PassVerdict()
    }
    if stackAt(ctx, p.visible).ImmutableComponentAt(idx) == nil {
        return legal.FailT("reveal.no_card_here")       // "there is no card at that index"
    }
    return legal.FailT("reveal.already_revealed")       // "that card has already been revealed"
},
```

All three strings preserved verbatim in the game's template table; every
predicate client-evaluable (occupancy facets survive memory's
`sanitize:"order"`).

### blackjack/moveStartRoundCleanup — fully declarative

```go
auto.Config(new(moveStartRoundCleanup),
    moves.WithPreconditions(
        legal.AllActivePlayers(
            legal.Any(legal.PlayerBool("Eliminated"), legal.PlayerBool("Stood")),
        ).WithMessage("cleanup.players_unfinished"),
    ),
)
```

(Blackjack's `moveCurrentPlayerHit`/hand-value moves keep their arithmetic in
`LegalCustom` — hard-custom per the corrected survey.)

### checkers/moveMoveToken — declarative gates + game-registered predicate + residue

`spaceIsBlack` is an unexported free function today; it becomes a
**game-registered predicate** — the registry is open to games, same as
constraints:

```go
// in checkers' delegate:
func (g *gameDelegate) ConfigurePredicateConstructors() []*legal.PredicateConstructor {
    return []*legal.PredicateConstructor{{
        Name: "checkers.spaceIsBlack",
        Constructor: func(spec legal.Spec, _ *boardgame.ComponentChest,
            _ func(legal.Spec) (*legal.Predicate, error)) (*legal.Predicate, error) {
            field := spec.Args[0] // "move.SpaceIndex"
            return &legal.Predicate{
                Name: "checkers.spaceIsBlack", Args: spec.Args,
                Reads: []legal.Read{{Path: legal.PropPath(field), Facet: legal.FacetValues}},
                Cost:  legal.CostTrivial,
                Evaluate: func(ctx legal.Context) legal.Verdict {
                    if spaceIsBlack(intField(ctx.Move, field)) {
                        return legal.PassVerdict()
                    }
                    return legal.FailT("checkers.black_spaces_only")
                },
            }, nil
        },
    }}
}

auto.Config(new(moveMoveToken),
    moves.WithPreconditions(
        legal.ComponentPresentAtKey("game.Spaces", "move.TokenIndexToMove").
            WithMessage("checkers.no_token_there"),
        legal.ComponentPropEqualsCurrentPlayer("game.Spaces", "move.TokenIndexToMove", "Color").
            WithMessage("checkers.not_your_token"),
        legal.Predicate1("checkers.spaceIsBlack", "move.SpaceIndex"),
    ),
)

func (m *moveMoveToken) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    // capture-graph walk, verbatim from today
    return legal.Errorf("checkers.illegal_dest", nil)
}
```

Client outcome: cheap gates `evaluable:true`; graph walk honestly `unknown`.
Game-registered predicates are client-evaluable only if the game also ships a
TS implementation (future); otherwise they degrade to server verdicts —
`checkers.spaceIsBlack` is a natural first test of that extension path.

### Migration scope

All example-game moves that the catalog covers; `../games` clients likewise on
a matching branch. Un-migrated moves are untouched and behavior-frozen.
Golden-equivalence tests fence every migration (§9); with declaration-order
evaluation, migrated moves keep their historical first-failure messages.

---

## 9. Testing

- **Unit:** every catalog predicate — Pass/Fail/Unknown, facet-level `Reads`
  conservativeness, registry round-trip, template-key existence.
- **Purely-sugar property tests:** (a) a game using only frozen-chain moves
  produces byte-identical Legal() results and LegalForPlayerError strings
  before/after this change; (b) **bucket superset property** — every opaque
  move appears in the candidate set for every phase; (c) the
  orphaned-declarations probe — boot fails for a wholesale `Legal()` override
  with declarations, passes for a super-calling override (whose super-call
  path evaluates the plan); unsupported-base-type opt-in fails at boot.
- **Golden equivalence:** for every migrated move, table tests asserting plan
  vs. old imperative Legal() agree (legal/illegal + message) across recorded
  states.
- **Engine:** phase-index correctness incl. TreeEnum ancestors and the
  phase-agnostic bucket; memo hit/miss across the move-forms double pass;
  deterministic failure reporting.
- **Ledger:** server e2e asserting Preconditions shape, `provisional`
  marking, per-viewer facet-based `evaluable`, and no-bindings-on-inevaluable
  (#693 guard).
- **Conformance corpus:** generated and checked in Go now; consumed by TS
  later.

## 10. Risks & open questions

- **The path/facet resolver is the largest net-new component**; boot-time
  validation contains the blast radius.
- **`Reads` conservativeness for game-registered predicates** is
  by-convention; a lint helper ships with the catalog, and the nil-Move
  runtime guard converts the worst case (undeclared move read) into `Unknown`,
  never a panic or a wrong verdict.
- **Seam expansion pressure**: games embedding the other ~24 framework move
  types can't opt in until those types get contribution support — each
  expansion is mechanical but needs golden tests. Sequencing risk, not design
  risk.
- **Catalog growth pressure** is permanent; the governing rule (§1) is
  normative and enforced in review.
- **Deferred dirty-tracking** stays deferred until a complete write-set audit
  exists; the conservative default is correct-but-uncached.

---

## Implementation notes (2026-07-11)

Recorded during Task 14 close-out, from the accumulated execution ledger
(`.superpowers/sdd/progress-legality.md`). Each note is a divergence between
this spec's normative text/samples and what actually shipped, adjudicated
during implementation and left standing (not reverted) because the
divergence was itself correct or unavoidable given an earlier, binding
design decision.

- **StartPhase embed is seam-blocked; §8's blackjack sample diverges.**
  §8's `moveStartRoundCleanup` sample still shows the move embedding
  `moves.StartPhase`. It cannot: §2's v1 seam is Default/CurrentPlayer only,
  and any other `moves` package embed — including `StartPhase`, even though
  it has no `Legal()` override of its own — is treated as an unsupported
  base type at boot. The shipped move embeds `moves.Default` directly and
  hand-rolls the one behavior it used from `StartPhase.Apply` (setting the
  game's current phase); verified behaviorally equivalent for blackjack,
  since its `gameState` implements neither `BeforeLeavePhaser` nor
  `BeforeEnterPhaser`.

- **Bucket-reordering was removed in the API-polish follow-up (§4).** The
  original implementation evaluated the field-independent bucket before the
  field-dependent bucket, changing first-failure messages. The follow-up
  memoizes field-independent predicates individually while traversing the
  original plan order; golden tests no longer whitelist those divergences.

- **`legal.Predicate1` (§8's checkers sample) does not exist.** The
  literal spelling in the shipped checkers migration is a bare
  `legal.Spec{Name: "checkers.spaceIsBlack", Args: [...]}` value, not a
  `legal.Predicate1(...)` builder function — no such convenience
  constructor was ever built for single-arg game-registered predicates.
  Game authors write the `Spec` literal directly today.

- **`player.X` cannot express "the proposing player" in simultaneous-move
  phases (§1's path grammar, hit in Task 13).** `player.X` resolves against
  the game's `CurrentPlayerIndex`. Darwin (`../games`) has a
  simultaneous-move phase where every player proposes at once and the
  game's own "current player" is Admin/none; `player.X` there returns an
  error or `Unknown`, never the actual proposer's properties. Darwin's
  attempted migration was reverted by its golden-equivalence fence rather
  than shipped with degraded behavior — zero regression shipped, but the gap
  is real and durable, not a fixture artifact.

- **Catalog gaps found and left open (Tasks 11-13 surveys), tracked for
  future growth per §1's rule of growth:** no count/stack-size threshold
  predicate exists, despite `Read`'s `FacetCount` facet already being
  designed for exactly this purpose and sitting unused by every catalog
  predicate; no negation compositor (`any` is the only one, matching §1's
  anti-tarpit rule, but it leaves genuinely-negated checks stuck in
  `LegalCustom`); `MayMoveTo`/`MayMoveToSlot` take a single `idxField`,
  so there is no variant expressing a source index distinct from a
  destination index. Each blocked at least one real migration (memory's
  timer-start check, several `../games` stack-count checks, blackjack's
  hand-arithmetic residue) and is recorded in the affected games'
  migration-survey commit messages and `legal/doc.go`'s "v1 limits."

- **`LegalForAnyone` is plan evaluation under `AdminPlayerIndex` (§4/§6),
  not a separate exemption path.** The pre-existing "is this legal for
  ANY player" computation used to run its own ad hoc exemption logic; Task
  10 deleted that logic and defined `LegalForAnyone` for an opted-in move as
  literally "evaluate the assembled plan with `proposer =
  AdminPlayerIndex`" — the old semantics fall out by construction (Admin
  bypasses the proposer-identity check) rather than needing a parallel code
  path. A parity invariant is asserted across every opted-in fixture.

- **The frozen-wire test is a differential reimplementation, not a
  recorded pre-change fixture (§9, accepted with a note in the Task 10
  ledger).** The ideal test would replay a byte-for-byte JSON payload
  captured before this campaign began; what shipped instead independently
  reconstructs the pre-change move-form JSON shape in the test itself and
  diffs against it live. This is weaker (a bug shared between the
  reimplementation and the real code would go undetected) but was judged
  sufficient given the shape's simplicity; flagged here rather than silently
  accepted as equivalent to a recorded fixture.
