# Declarative Move Legality — Design

**Date:** 2026-07-10
**Issues:** #761, #189 (the core thread); folds in #790, #644, #65; constrains #693, #44; enables #640, #213, #295
**Branch:** `declarative-legality-design`
**Status:** Design for review (implementation plan follows approval)

## Provenance

Synthesized from a three-lens design panel (composability / engine / DX) plus two
adversarial critiques, run 2026-07-10. Panel verdict: engine-lens design as the
spine, wearing the DX lens's error model and the composability lens's wire
format. Panel artifacts (not normative, kept for the record):
`.superpowers/design/design-{A-composability,B-engine,C-delight}.md`,
`critique-{acid,systems}.md`, `legality-brief.md`.

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
from its config bag — phase membership (`WithLegalPhases`), move-tape progression
(`WithLegalMoveProgression`), and stack constraints
(`WithSourceProperty`/`WithDestinationProperty` → `MayMoveTo`). This design
promotes that buried pattern to a first-class, inspectable, serializable system.

## The prime guarantee: purely sugar

**`Legal(state, proposer) error` remains the ground-truth contract, unchanged.**

- A game author can ignore this entire system and write imperative `Legal()`
  exactly as today, with zero new required concepts. The engine treats such a
  move as opaque and behaves exactly as it does now.
- When preconditions are declared, `moves.Default.Legal()` *is* their evaluator:
  declaring is implementing. An author never writes both a declaration and the
  code enforcing it.
- Every engine capability — phase bucketing, caching, the client ledger,
  structured errors — is opportunistic introspection through one optional
  interface, degrading gracefully to "call Legal()" when declarations are
  absent.
- A move that overrides `Legal()` wholesale (not via `LegalCustom`) opts out of
  the plan entirely; the engine falls back to today's behavior for that move.

## Design decisions locked before the panel

- **Design for client evaluation, ship server-first.** The representation is
  serializable and sanitization-aware from day one; the TypeScript evaluator is
  a designed-for follow-up, not in this campaign.
- **Break the Go API freely** (in the additive-sugar sense above): all in-repo
  example games and the three `../games` clients are migrated to declarations
  where they benefit; imperative Legal() keeps working everywhere.

---

## 1. Representation

One new package, `legal`, sitting beside `constraints` and deliberately rhyming
with it (name + string args, constructor registry, struct-tag-friendly).

```go
// package legal

// Outcome is a three-valued verdict. Unknown is load-bearing: it is how a
// predicate that cannot decide (hidden state, imperative escape hatch) stays
// honest instead of guessing.
type Outcome int

const (
    Pass Outcome = iota
    Fail
    Unknown
)

// Message is a template key plus named bindings — never a pre-baked string —
// so failures are localizable, greppable, and re-renderable on server or
// client. (Adapts to the existing errors.Friendly at the API boundary.)
type Message struct {
    Template string         // "reveal.no_cards_left"
    Bindings map[string]any // {"left": 0}
}

// Verdict is the result of evaluating one Predicate.
type Verdict struct {
    Outcome Outcome
    Message *Message // set on Fail (optionally on Unknown); nil on Pass
    Reason  string   // on Unknown: why ("reads hidden property HiddenCards")
}

// Context is the entire vocabulary a predicate may reference. The line before
// the Turing tarpit is drawn here, by construction: no I/O, no mutation, no
// access beyond these four values.
type Context struct {
    State    boardgame.ImmutableState
    Move     boardgame.Move // nil during field-independent evaluation
    Proposer boardgame.PlayerIndex
    Chest    *boardgame.ComponentChest
}

// Predicate is one legality question.
type Predicate interface {
    // Name + Args round-trip through the constructor registry and are the
    // serialized identity.
    Name() string
    Args() []string
    // Reads declares the property paths this predicate touches. Drives the
    // field-independent/dependent split, caching, and per-viewer client
    // evaluability. Must be a conservative over-approximation.
    Reads() []PropPath
    // Cost orders evaluation (cheap gates run first, short-circuit).
    Cost() Cost
    // Evaluate is pure: same Context in, same Verdict out.
    Evaluate(ctx Context) Verdict
}

type Cost int

const (
    CostTrivial  Cost = iota // int/bool compare, proposer check
    CostCheap                // single stack read, phase lookup
    CostModerate             // iterate players or a stack, walk the move tape
    CostExpensive            // opaque custom residue
)
```

### Wire format: leaf ≡ node

A compositor is just a predicate whose children are predicates. One serialized
shape covers leaves and composites with no duplication:

```jsonc
{"name": "playerPropAtLeast", "args": ["CardsLeftToReveal", "1"]}

{"name": "any", "sub": [
  {"name": "playerBool", "args": ["Eliminated"]},
  {"name": "playerBool", "args": ["Stood"]}
]}
```

```go
// Spec is the serializable form; resolved against the registry at
// NewGameManager time (fail-fast on unknown names / bad args).
type Spec struct {
    Name string   `json:"name"`
    Args []string `json:"args,omitempty"`
    Sub  []Spec   `json:"sub,omitempty"`
}

// PredicateConstructor mirrors constraints.StackConstraintConstructor.
type PredicateConstructor struct {
    Name        string
    Constructor func(spec Spec, chest *boardgame.ComponentChest,
        resolve func(Spec) (Predicate, error)) (Predicate, error)
}
```

Registry wiring mirrors constraints exactly:
`GameDelegate.ConfigurePredicateConstructors() []*legal.PredicateConstructor`,
with `base.GameDelegate` returning `legal.DefaultConstructors()`.

### Anti-tarpit rules (normative)

Findings from the adversarial critiques, adopted as hard rules:

1. **`any` is the only compositor in v1**, and the registry rejects a compositor
   nested inside a compositor (depth 1). No `all` (the plan's ordered list IS
   the conjunction); no first-class `not` (a Kleene-`Not` is a TS-conformance
   liability — client and server must agree exactly on `Unknown` semantics, and
   `not` doubles the surface where they can disagree).
2. **Branchy logic becomes a purpose-built named predicate with hand-written Go
   Eval** — not combinator surface. Proven on the acid test: memory's "no card
   at that index" vs "that card has already been revealed" disambiguation is a
   12-line `revealableCardAt` predicate, not an `ElseWhen`/`_if()` DSL invention
   (both of which failed critique).
3. **The governing rule for catalog growth:** if you can't say it as a relation
   over a path, push the computation into a computed state property (which a
   `propCompare` predicate can then read) or drop to the escape hatch. No user
   arithmetic, no loops, no lambdas in serialized form, ever.

### Path grammar (net-new build, honestly priced)

The panel designs all claimed to "reuse" `constraints/prop_path.go`; critique
ground-truthing showed that resolver handles a single component instance only.
The state-path resolver is a **new build** with this grammar:

| Path | Resolves to |
|---|---|
| `game.X` | game state property X |
| `player.X` | property X of the *current* player |
| `players[*].X` | property X across all players (quantified predicates only) |
| `move.X` | move field X |

Paths are validated against the reader hierarchy at `NewGameManager` — a typo'd
path fails at boot with the move name and path in the error, never mid-game.
`player.` is treated as `players[*]` for any future invalidation purposes
(coarse but sound).

---

## 2. Attachment & composition

### Authoring surface

One option family in `moves`, rhyming with the existing `With*` idiom:

```go
moves.WithPreconditions(specs ...legal.Spec)   // append, in order
moves.WithoutPrecondition(name string)          // suppress an inherited one
```

The catalog exposes typed builders that produce `Spec`s, so authoring is
readable and typo-resistant:

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

Struct tags are **not** in v1 (the DX panel proposed them; its own migrations
never used them — cut per YAGNI).

### The composition seam (designed here; the panel left it open)

Both critiques identified the same gap: the config bag is a *flat*
`PropertyCollection`, so "each embedding layer appends" is not free. The
mechanism, using Go's idiomatic embedding dispatch to return **data instead of
doing work**:

```go
// package moves

// ContributedPreconditions returns the preconditions this move type's
// embedding chain contributes, base-first. Embeddable types chain explicitly —
// the same pattern as today's Legal() super-calls, but written ONCE in the
// framework and returning declarations instead of verdicts.
func (d *Default) ContributedPreconditions() []legal.Spec {
    // phase, progression, stack-constraint specs derived from the config bag
    // (WithLegalPhases / WithLegalMoveProgression / WithSourceProperty keep
    // working — they are now thin shims that produce these specs).
    return d.specsFromConfig()
}

func (c *CurrentPlayer) ContributedPreconditions() []legal.Spec {
    return append(c.Default.ContributedPreconditions(),
        legal.ProposerIsCurrentPlayer())
}
```

The plan-builder (at NewGameManager) assembles, per move type:

```
plan = ContributedPreconditions()   // base-first, deterministic
     + configured WithPreconditions // author's, in declaration order
     - WithoutPrecondition names    // suppressions, matched by stable name
```

Every built-in has a stable name (`"inPhase"`, `"inProgression"`,
`"stackConstraints"`, `"proposerIsCurrentPlayer"`), so suppression and client
display are addressable. This is strictly more capable than today, where the
buried checks are unremovable. Game authors never write the chain; they inherit
it by embedding, exactly as now — minus the
`if err := m.CurrentPlayer.Legal(...)` boilerplate, which disappears.

---

## 3. Layering

```
boardgame (core)
    Verdict, Outcome, Message, PropPath, Spec, Cost   (value types)
    PreconditionsProvider (optional interface, below)
    the evaluation engine: plan build, buckets, phase index, caches,
    move-form ledger assembly
    — engine lives in core because game.go's apply/fixup loop calls it —
        ▲
moves
    ContributedPreconditions chain on Default/CurrentPlayer/FixUp/StartPhase
    WithPreconditions / WithoutPrecondition
    Default.Legal() = "evaluate my plan, then LegalCustom"
        ▲
legal (new; peer of constraints)
    the predicate catalog + DefaultConstructors()
    Errorf (template-errors from imperative code)
    template rendering
```

Dependency arrows point downward only. Core holds **types + engine, zero game
semantics** — the catalog lives in `legal` exactly as constraint implementations
live in `constraints`. Critically (a fatal flaw in one panel design, avoided
here): the structured `Verdict` — bindings and all — is what crosses the core
boundary. Nothing flattens to a rendered string until the last moment
(`Verdict.Error()` adapts to `error`/`errors.Friendly` for the Legal() return),
so the move-forms assembler in core has full explainability data.

```go
// package boardgame

// PreconditionsProvider is the single optional interface the engine introspects.
// moves.Default implements it; hand-rolled moves may; absence = opaque move,
// today's behavior.
type PreconditionsProvider interface {
    PreconditionPlan() *PreconditionPlan
}
```

---

## 4. Evaluation semantics

Per move *type*, built once at NewGameManager:

```go
type PreconditionPlan struct {
    fieldIndependent []legal.Predicate // Reads() has no move.* paths
    fieldDependent   []legal.Predicate
    custom           legal.Predicate   // LegalCustom wrapper, or nil
    allReadPaths     []legal.PropPath  // union, for future invalidation
    opaque           bool              // move overrides Legal() wholesale
}
```

- Buckets are stable-sorted by `Cost` (Trivial → Expensive); `custom` always
  last. Deterministic order ⇒ a given state always reports the same failure (no
  message flapping).
- Evaluation order: field-independent bucket → field-dependent bucket (only
  with a bound move; this is #761's split — phase/turn checks run before
  `DefaultsForState`/field-binding, `CardIndex`-in-range after) → custom.
- **Hot paths short-circuit** on first Fail (fixup loop, ProposeMove
  validation). **Move-forms assembly evaluates the full ledger** (once per
  request — richness is worth it there, and only there).
- `moves.Default.Legal()` becomes exactly: evaluate plan; return first
  failure's `Verdict.Error()`, or nil. One code path for every declarative
  move. (Purely sugar: this is the same signature and observable behavior
  contract as today.)

### The escape hatch

```go
// package boardgame
type CustomLegaler interface {
    // LegalCustom runs after all declarative preconditions pass. It is the
    // imperative residue (checkers capture graph, blackjack hand value).
    LegalCustom(state ImmutableState, proposer PlayerIndex) error
}
```

The engine wraps it as an opaque predicate: `Reads()` unknown, `CostExpensive`,
no serialized form. Consequences fall out of the metadata, not special cases:
runs last, never cached, client sees `"unknown"`. An imperative body may return
`legal.Errorf("checkers.illegal_dest", bindings)` to keep even residue failures
structured; a plain `error` is wrapped as an opaque single-use template.

---

## 5. Engine wins (honest table)

| Mechanism | v1 | Effect |
|---|---|---|
| **Phase bucketing** — `phaseIndex map[phase][]moveType` built from each plan's `inPhase` spec (∪ TreeEnum ancestors) | ✅ | Fixup loop and move-forms iterate `phaseIndex[currentPhase]` instead of all moves: candidate filtering is an index lookup with zero evaluations (#640) |
| **Cost-ordered short-circuit** | ✅ | The common rejections (wrong phase, not your turn) are Trivial/Cheap and fire before any player-loop or graph search |
| **Field-independent memo**, keyed `(moveType, stateVersion, proposer)` | ✅ | The move-forms double pass (player + admin) computes the stable half once; client field-editing re-runs only the field-dependent bucket |
| **Tape memoization** — `historicalMovesSincePhaseTransition` memoized per version | ✅ | Every `inProgression` predicate in a version shares one tape walk (retires the default.go:475 TODO) |
| **Dirty-tracking** — invalidate cached verdicts only when a move's Apply wrote paths intersecting `allReadPaths` | ❌ deferred | Both critiques converged: write-set capture must be *complete* or the legality cache is stale-and-wrong (a correctness bug, not a slow path). `Reads()` metadata makes this addable later behind an audit; v1 invalidates everything each version. |

What stays O(Legal): opaque moves and `LegalCustom` residue — by design, and
now gated behind cheap declarative checks so they run far less often.

---

## 6. Explainability

### Server

Every declarative failure is a `Verdict` with `Message{Template, Bindings}`.
The fixup loop, on rejecting a candidate, logs at debug level:
`fixup rejected move=MoveMoveToken predicate=proposerIsCurrentPlayer msg="it's not your turn"`
— #65 resolved with no exceptions, since even imperative residue can emit
templates via `legal.Errorf`.

### Client contract (shipped in v1; consumed richly by a later TS evaluator)

Move forms gain a per-predicate ledger alongside the preserved
`LegalForPlayer`/`LegalForAnyone`/`LegalForPlayerError`:

```jsonc
"Preconditions": [
  {"name": "proposerIsCurrentPlayer", "verdict": "pass", "evaluable": true},
  {"name": "playerPropAtLeast", "args": ["player.CardsLeftToReveal", "1"],
   "verdict": "fail",
   "message": {"template": "reveal.no_cards_left", "bindings": {"left": 0}},
   "evaluable": true},
  {"name": "custom", "verdict": "unknown", "evaluable": false}
]
```

`evaluable` is computed server-side, per predicate, per viewer:

```
evaluable = has a serialized form (not the escape hatch)
          ∧ every Reads() path is Visible under the sanitization
            transformation applied for this viewer
```

That is the honest three-valued story: the future TS evaluator re-runs
evaluable predicates locally against the sanitized state (live graying, zero
round trips, #189/#213) and displays the server's last verdict for everything
else. It never guesses whether a zero is a real zero or a hidden seven.

(Ground truth from critique: memory's stacks are `sanitize:"order"` — slot
*occupancy* is visible — so memory's entire plan is client-evaluable. The
common case is better than the panel assumed.)

---

## 7. Progression & #644

Move-tape matching becomes the `inProgression` predicate wrapping the existing
`matchTape` machinery — same plan, cache, and explain path as everything else
(`Reads: [game.moveHistory]`, `CostModerate`, so it runs after the cheap
gates). `MoveProgressionGroup.Satisfied` gains access to `legal.Context`
(plumbing change, named explicitly since the panel found it hand-waved
elsewhere), which unlocks #644: `moves.RepeatFromProp("game.RoundsThisTurn")`
resolves its count against live state at match time, and the backing path joins
the predicate's read-set mechanically.

---

## 8. Migrations (the acid tests, in full)

### memory/moveRevealCard — fully declarative, Legal() deleted

```go
//boardgame:codegen
type moveRevealCard struct {
    moves.CurrentPlayer
    CardIndex int
}

auto.Config(new(moveRevealCard),
    moves.WithPreconditions(
        legal.PropAtLeast("player.CardsLeftToReveal", 1).
            WithMessage("reveal.no_cards_left"), // "You have no cards left to reveal this turn"
        legal.RevealableCardAt("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
        legal.MayMoveToSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
    ),
)
```

`RevealableCardAt` is a purpose-built catalog predicate (the anti-tarpit rule
in action) whose Eval mirrors the original branch structure exactly:

```go
func (p *revealableCardAt) Evaluate(ctx legal.Context) legal.Verdict {
    idx := intField(ctx.Move, p.field)
    if stackAt(ctx, p.hidden).ImmutableComponentAt(idx) != nil {
        return legal.PassVerdict()
    }
    if stackAt(ctx, p.visible).ImmutableComponentAt(idx) == nil {
        return legal.FailT("reveal.no_card_here")        // "there is no card at that index"
    }
    return legal.FailT("reveal.already_revealed")        // "that card has already been revealed"
}
```

All three error strings preserved verbatim as templates; every predicate
client-evaluable under memory's sanitization.

### blackjack/moveStartRoundCleanup — fully declarative

```go
auto.Config(new(moveStartRoundCleanup),
    moves.WithPreconditions(
        // StartPhase contributes its phase/progression preconditions.
        legal.AllActivePlayers(
            legal.Any(legal.PlayerBool("Eliminated"), legal.PlayerBool("Stood")),
        ).WithMessage("cleanup.players_unfinished"), // "not all active players have finished their turn"
    ),
)
```

`AllActivePlayers` reuses `behaviors.PlayerIsInactive` to skip inactive seats;
`Reads: [players[*].Eliminated, players[*].Stood]`, `CostModerate` — the player
loop no longer runs in the common wrong-phase case.

### checkers/moveMoveToken — declarative gates + imperative residue

```go
auto.Config(new(moveMoveToken),
    moves.WithPreconditions(
        // CurrentPlayer contributes proposerIsCurrentPlayer.
        legal.ComponentPresentAtKey("game.Spaces", "move.TokenIndexToMove").
            WithMessage("checkers.no_token_there"),
        legal.ComponentPropEqualsCurrentPlayer("game.Spaces", "move.TokenIndexToMove", "Color").
            WithMessage("checkers.not_your_token"), // "that token isn't your token to move"
        legal.SpacePredicate("move.SpaceIndex", "spaceIsBlack").
            WithMessage("checkers.black_spaces_only"), // "you can only move to spaces that are black"
    ),
)

// The capture-graph search stays imperative — and honest:
func (m *moveMoveToken) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    // ... FreeNextSpaces / LegalCaptureSpaces walk, verbatim from today ...
    return legal.Errorf("checkers.illegal_dest", nil) // structured even in the residue
}
```

Client outcome: the four cheap gates are `evaluable:true` (≈80% of illegal
clicks rejected locally, later, by the TS evaluator); the graph walk shows
`verdict:"unknown"` — honestly.

### Migration scope

All `examples/*` moves move to declarations where the catalog covers them
(survey: ~5 phase, ~8 current-player, ~6 stack-size/presence, ~4 property
comparisons, ~3 MayMoveTo — all covered; 2 genuinely custom stay in
`LegalCustom`). `../games` (murdermrmonroe, pass, valentine) migrated the same
way, committed on a matching branch. Migration tests snapshot the acid-test
error messages, since Cost-reordering can legitimately change *which* failure
is reported first.

---

## 9. Testing

- **Unit (Go):** every catalog predicate — Pass/Fail/Unknown cases, Reads()
  conservativeness, registry round-trip (Spec → Predicate → Name/Args → Spec).
- **Plan tests:** contributed-chain assembly order; WithoutPrecondition
  suppression; opaque fallback for wholesale Legal() overrides (the purely-sugar
  guarantee, asserted).
- **Golden equivalence:** for every migrated move, a table test asserting the
  new plan and the old imperative Legal() agree (legal/illegal + message) across
  recorded game states — the migration is provably behavior-preserving.
- **Engine:** phase-index correctness incl. TreeEnum ancestors; memo hit/miss
  across the move-forms double pass; determinism of reported failure.
- **Ledger:** server e2e asserting the Preconditions array shape and
  per-viewer `evaluable` under each sanitization policy.

## 10. Risks & open questions

- **The path resolver is the biggest net-new component** (mis-claimed as reuse
  by all three panel designs). Boot-time validation contains the blast radius.
- **`Reads()` conservativeness for custom predicate authors** is by-convention;
  built-ins are verified by construction, and a lint/test helper ships with the
  catalog.
- **Cost-reordering changes error precedence** vs today's hand-ordered chains —
  handled via golden equivalence tests; a few messages may legitimately improve.
- **Catalog growth pressure** is permanent; the governing rule (relation over a
  path, or push to computed property / escape hatch) is normative and enforced
  in review.
- **Deferred dirty-tracking** is designed-for (Reads() exists) but requires a
  write-set audit of every mutation path in core before it can ever ship; the
  conservative default is correct-but-uncached.
