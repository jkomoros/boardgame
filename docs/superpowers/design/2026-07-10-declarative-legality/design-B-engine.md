# Design B — Declarative Move Legality (Engine & Performance lens)

## Executive summary

A precondition is a **named, serializable predicate value** — `constraints`-style
`Name + args`, resolved to a pure `Predicate` closure at manager-build time. Each
move carries an ordered `[]*Precondition` assembled from its embedding chain,
plus an optional imperative `LegalCustom()` escape hatch for genuinely-gnarly
residue (checkers capture graph). The whole point of declarative-ness, from this
lens, is that a predicate *declares what it reads and when*: each one exposes a
`Refs` set (state property paths, move-field names, proposer, phase) and a
`Stage` (FieldIndependent vs FieldDependent). That metadata is precomputed once
per move type into a **PreconditionPlan** and lets the engine do four things no
opaque `Legal()` allowed:

1. **Phase-bucket** candidate moves so the fixup loop never even *considers*
   moves whose phase precondition can't hold in the current phase (index lookup,
   not evaluation).
2. **Evaluate field-independent preconditions once per (move-type, state
   version)** and memoize — shared across the fixup poll, both admin/player
   move-forms passes, and every candidate field-binding.
3. **Dirty-track by referenced property path**: a per-state-version evaluation
   cache invalidated only for plans whose `Refs` intersect the paths a move's
   Apply actually mutated — so most cache entries survive a state transition.
4. **Order cheap→expensive and short-circuit**, deterministically, driven by a
   declared `Cost` on each predicate constructor.

Progression matching becomes *just another precondition* (`InProgression{...}`)
so it flows through the same plan, cache, and explain machinery. The imperative
residue coexists as a terminal `LegalCustom` predicate whose `Refs` is "unknown"
— it never caches and the client renders it as `unknown`.

Layering: core `boardgame` gets three tiny interfaces (`Predicate`,
`PredicateConstructor`, `PreconditionFailure`) and the evaluation engine; a new
`preconditions` package (rhyming with `constraints`) holds the built-in library;
`moves` wires them into `Default`/`CurrentPlayer` via `With*`.

---

## 1. Representation

A precondition is a value pair — a **serializable spec** and the **resolved
predicate** it constructs — exactly mirroring `StackConstraint` /
`StackConstraintConstructor`.

```go
// package boardgame (core — minimal interfaces only)

// PreconditionInput is everything a predicate may read. The engine constructs
// it; a predicate may ONLY touch these — this is the wall before the Turing
// tarpit.
type PreconditionInput struct {
    State    ImmutableState
    Proposer PlayerIndex
    Move     Move          // for field-dependent predicates; nil at FieldIndependent stage
    Chest    *ComponentChest
}

// Predicate is a resolved, pure legality check. Returns nil if satisfied, or a
// structured *PreconditionFailure. It must be a pure function of PreconditionInput.
type Predicate interface {
    // Eval is the check. MUST NOT mutate anything.
    Eval(in PreconditionInput) *PreconditionFailure
    // Refs declares exactly what this predicate reads. The engine uses it for
    // dirty-tracking and for the field-independent/dependent split. An empty
    // Refs with ReadsUnknown=true means "opaque" (the escape hatch).
    Refs() Refs
    // Cost is a static ordering hint (cheaper predicates run first).
    Cost() Cost
    // Spec returns the serializable form (name + args) or ok=false if this
    // predicate is opaque and cannot be shipped to a client.
    Spec() (name string, args []string, ok bool)
}

type Cost int
const (
    CostTrivial Cost = iota // integer/bool field compare, proposer==x
    CostCheap               // single stack read, phase lookup
    CostModerate            // iterate players / a stack
    CostExpensive           // graph search, opaque custom
)

// Refs is the read-set of a predicate: which parts of state/move it depends on.
type Refs struct {
    StatePaths   []string // e.g. "Game.HiddenCards", "Players[current].CardsLeftToReveal"
    MoveFields   []string // e.g. "CardIndex"
    ReadsPhase   bool
    ReadsProposer bool
    ReadsUnknown bool     // opaque escape hatch — defeats all caching
}
```

`Refs.StatePaths` reuses the existing property-path vocabulary the `constraints`
package already parses (`prop_path.go`), extended with two sentinels:
`Players[current]` (resolved against `CurrentPlayer`) and `Players[*]` (all
players). Paths are namespaced `Game.` / `Players[...].` so they align 1:1 with
what the reader hierarchy exposes and, crucially, with what a move's `Apply`
mutates (see §7).

**How constructed / serialized.** Same registry idiom as constraints:

```go
type PredicateConstructor struct {
    Name        string
    Constructor func(args []string, chest *ComponentChest) (Predicate, error)
}
```

A move stores specs (`{name, args}`) in its config bag; at manager build time the
engine resolves each spec against the registered constructors into a `Predicate`,
failing fast on bad args. This is what makes it serializable-from-day-one: the
spec is `[]struct{Name string; Args []string}` — trivially JSON, and the same
list is what a future TypeScript evaluator receives.

**What it can reference.** State property paths, move fields, proposer, phase,
chest constants (constructors resolve constant names to ints, as
`MaxNumComponentsConstructor` already does). **Where the line is:** predicates are
*leaves*. Composition (AND/OR/sequencing) is the engine's job via the ordered
list, not a predicate's — there is no user-authored boolean-expression AST to
interpret, no loops, no arithmetic beyond what a named constructor bakes in. If
you need real computation you drop to `LegalCustom` (§5). This keeps the
declarative surface analyzable (every predicate's `Refs` is statically known) and
keeps us out of the Turing tarpit.

---

## 2. Attachment & composition

Attachment is `With*` options writing specs into the config bag — identical to
`WithLegalPhases`. New sugar:

```go
// package moves
func WithPreconditions(preconditions ...preconditions.Spec) interfaces.CustomConfigurationOption
func WithPrecondition(name string, args ...string) interfaces.CustomConfigurationOption
```

`preconditions.Spec` is a tiny builder so authoring reads well:

```go
auto.Config(new(moveRevealCard),
    moves.WithPreconditions(
        preconditions.CardsLeftToReveal(),                  // p.CardsLeftToReveal >= 1
        preconditions.ComponentPresentAt("HiddenCards", "CardIndex"),
        preconditions.MayMoveToSlot("HiddenCards", "VisibleCards", "CardIndex"),
    ),
)
```

**Composition down the embedding chain.** This is the delicate part and the
engine cares about it most. Each level in the chain (`Default`,
`CurrentPlayer`, game move) contributes preconditions. Rather than each `Legal()`
calling `super.Legal()` imperatively (today's chain), the engine **collects the
full ordered list declaratively** by walking the config bags that
`auto.Config` merged. Order is: base-embedded first (phase, progression,
stack-constraints from `Default`; proposer==current from `CurrentPlayer`), then
move-specific, in declaration order. Because ordering is data, not call-stack
position, the engine can *re-sort within a stability-preserving grouping* by
`Cost` (§5) without changing semantics.

**Override/removal.** A subclass can:
- **Add** — the common case, just `WithPreconditions(...)`.
- **Remove** an inherited one by name: `WithoutPrecondition("legalInPhase")` sets
  a suppression entry the collector honors.
- **Replace** — remove + add.

Every built-in gets a stable name (`"legalInPhase"`, `"inProgression"`,
`"stackConstraints"`, `"proposerIsCurrentPlayer"`) so removal is addressable.
This is strictly more capable than today, where the three buried checks are
un-removable.

**auto.Config interaction.** Unchanged mechanically — options still write
namespaced keys. The engine reads `configPropPreconditions` (a `[]Spec`) at
manager-build, resolves once, caches the resolved `[]Predicate` on the move
*type*'s plan (not per-move-instance). `WithLegalPhases`/
`WithLegalMoveProgression` become thin shims that append the corresponding spec,
so old call-sites keep working during migration.

---

## 3. Layering

```
core boardgame ─────────────────────────────────────────────
  Predicate, PredicateConstructor, PreconditionInput,
  Refs, Cost, PreconditionFailure          (interfaces + value types)
  preconditionEngine                         (plan build, cache, eval)
  Move gains: Preconditions() []Predicate  +  LegalCustom (optional iface)
        │ (depends on)
        ▼
moves ──────────────────────────────────────────────────────
  WithPreconditions / WithPrecondition / WithoutPrecondition
  Default/CurrentPlayer/FixUp/StartPhase register their built-ins
        │
        ▼
preconditions (NEW — rhymes with constraints) ──────────────
  built-in library: CardsLeftToReveal, ComponentPresentAt,
  MayMoveToSlot, AllPlayersMatch, PropCompare, InPhase,
  InProgression, ProposerIsCurrentPlayer, ...
  + DefaultConstructors() for the registry
```

Dependency arrows point downward only. **Core gets minimal interfaces + the
engine** (the engine must live in core because it's called from `game.go`'s
apply/fixup loop and needs `Move`/`ImmutableState` internals). The *library* of
concrete predicates lives in `preconditions`, so core stays free of
game-semantics knowledge — exactly as `constraints` sits above core today.

The three buried checks land principled: `legalInPhase` → `preconditions.InPhase`,
`legalMoveInProgression` → `preconditions.InProgression`, `legalStackConstraints`
→ `preconditions.StackConstraints`. `Default.Legal()` stops being a hand-written
chain and becomes "run my resolved precondition plan" — one code path for all
moves.

Registry mirrors constraints exactly: `GameDelegate.ConfigurePredicateConstructors()
[]*PredicateConstructor`, with `base.GameDelegate` returning
`preconditions.DefaultConstructors()`.

---

## 4. Explainability

The error model is structured, not a string:

```go
type PreconditionFailure struct {
    Predicate string            // "cardsLeftToReveal"  (the spec name)
    Args      []string          // resolved args, for the client
    // MessageTemplate has {placeholders} filled from Bindings.
    MessageTemplate string      // "You have no cards left to reveal this turn"
    Bindings  map[string]string // {"remaining":"0"} — computed at eval time
    // Refs the failure actually touched — for logging / debugging fixups (#65).
    ReadPaths []string
}

func (f *PreconditionFailure) Error() string // renders template with bindings
```

A predicate returns this instead of `errors.New(...)`. `MessageTemplate` +
`Bindings` means the *server* ships the template and the raw bindings; the client
localizes/renders. Because the message is data, `moveRevealCard`'s three
distinct strings survive verbatim (§9) — each maps to a distinct predicate or a
distinct branch within one predicate carrying its own template.

**What the client receives** per move form:

```go
type moveForm struct {
    // ... existing ...
    Preconditions []preconditionResult // one per predicate in the plan
}
type preconditionResult struct {
    Predicate    string   `json:"predicate"`
    Args         []string `json:"args"`
    Passed       string   `json:"passed"`   // "yes" | "no" | "unknown"
    Message      string   `json:"message,omitempty"` // rendered, when "no"
    Bindings     map[string]string `json:"bindings,omitempty"`
    Evaluable    bool     `json:"evaluable"` // can the client re-evaluate locally? (§6)
}
```

**Fixup-loop logging (#65).** When the fixup loop rejects a candidate, the engine
already has the `*PreconditionFailure` (structured) — it logs
`move=X predicate=Y read=[paths] msg="..."` at debug level. Today the loop
discards the error string entirely; now every rejection is explainable, and the
"which predicate blocked the fixup" question is answerable without re-running.

---

## 5. Evaluation semantics & the escape hatch

**The plan.** Per move *type*, built once at manager-build:

```go
type PreconditionPlan struct {
    fieldIndependent []Predicate // Refs.MoveFields empty && !ReadsUnknown-on-fields
    fieldDependent   []Predicate
    custom           Predicate   // LegalCustom wrapper, or nil
    // union of all Refs.StatePaths, for coarse dirty-checks
    allStatePaths    map[string]bool
    readsUnknown     bool
}
```

Within each bucket, predicates are **stable-sorted by `Cost`** (Trivial→Expensive).
`custom` always runs last.

**Ordering & short-circuit.** Evaluation is: field-independent bucket (cheap→
expensive), then — only if we have a bound move — field-dependent bucket, then
custom. First failure short-circuits and returns its `PreconditionFailure`. This
is exactly #761's split: the phase/turn checks run *before* `DefaultsForState` /
field-binding; the `CardIndex-in-range` checks run after. For the fixup poll and
`LegalForAnyone`, only the field-independent bucket + any field-dependent
predicates whose fields have server-set defaults matter.

**Determinism.** Predicates are pure and `Cost`-ordered; the stable sort makes
evaluation order a deterministic function of the plan, so the *same* failure is
always reported for a given state (no flapping between two failing predicates).

**The escape hatch.** A move may implement:

```go
type CustomLegaler interface {
    // LegalCustom runs after all declarative preconditions pass. It is the
    // imperative residue (checkers capture graph, blackjack hand value).
    LegalCustom(state ImmutableState, proposer PlayerIndex) error
}
```

The engine wraps it in an opaque predicate: `Refs{ReadsUnknown:true}`,
`Cost:CostExpensive`, `Spec()→ok=false`. Consequences fall out automatically:
it runs last, it never caches (unknown refs ⇒ always dirty), and its
`preconditionResult.Passed` is `"unknown"` for the client when the server didn't
run it (or `"no"`/`"yes"` with `Evaluable:false` when it did). Coexistence is
clean: a checkers move is "declarative preconditions + a custom tail," and the
engine treats the tail as a black box it's honest about.

---

## 6. Sanitization-aware client story (designed-for)

Three-valued logic, computed server-side and shipped. For each predicate the
server knows two things the client can't: the predicate's `Refs.StatePaths`, and
the sanitization policy for each path in the *sanitized state the client will
receive*. The server computes `Evaluable` per predicate:

```
Evaluable(pred, viewingPlayer) =
    pred.Spec().ok                                  // has a serializable form
    && !pred.Refs().ReadsUnknown                    // not the escape hatch
    && every path in pred.Refs().StatePaths is Visible under the
       sanitization policy applied for viewingPlayer
```

If every path a predicate reads survives sanitization as `Visible`, the client
can re-run that predicate against fields the user is editing *without a round
trip* (#189/#213) — grays out the button live. If any path is `Hidden`/`Len`/
`Order`, `Evaluable=false` and the client shows the *server's* last verdict
(`Passed: yes|no|unknown`) as a static, non-live value. This is the honest
handling the brief demands: the client never guesses a hidden `0` vs hidden `7`;
it simply defers to the server for predicates it can't fully see.

The `Refs.StatePaths` metadata is the linchpin — it's *because* preconditions
declare their reads that the server can compute evaluability per-predicate rather
than all-or-nothing per-move. An opaque `Legal()` could only ever be
`Evaluable:false`.

---

## 7. Engine wins (the core of this lens)

Four concrete mechanisms, each keyed on the declared metadata.

### (a) Phase bucketing — turns O(moves) eval into O(1) index lookup
At manager-build, index every move by the phases its `InPhase` precondition
admits: `phaseIndex map[EnumKey][]Move`. The fixup loop and move-forms iterate
`phaseIndex[currentPhase]` (∪ ancestor phases for TreeEnums) instead of
`game.Moves()`. Moves whose phase can't match are never evaluated at all. For a
game with N moves across P phases, per-poll candidate set shrinks from N to ~N/P
with **zero predicate evaluations spent on filtering**. This alone attacks the
#640 fixup cost.

### (b) Field-independent memoization — evaluate once per (moveType, stateVersion)
The field-independent bucket depends only on state+proposer+phase — not on move
field values. So its result is cached keyed by `(moveType, stateVersion,
proposer)`:

```go
type fieldIndepCache struct {
    version int
    // moveType -> proposer -> result (nil = passed)
    results map[reflect.Type]map[PlayerIndex]*PreconditionFailure
}
```

Payoff across the three hot callers:
- **Fixup poll** (up to 256 recursions, but each is a *new* version, so cache is
  really within one poll's candidate set): the field-independent verdict for a
  move type is computed once even if the delegate lists it multiple times.
- **Move-forms** (server/api/main.go:1613 & :1621): today every move runs
  `Legal()` **twice** (player + admin). With the split, the *field-independent*
  portion is computed once per proposer and reused; the admin pass reuses
  everything except `ProposerIsCurrentPlayer`.
- **DefaultsForState + field re-binding**: field-independent verdict is stable
  while the client fiddles fields, so only the field-dependent bucket re-runs.

### (c) Dirty-tracking by referenced path — cache survives most transitions
Every `Move.Apply` mutates a knowable set of property paths. We capture the
**mutation write-set** cheaply: wrap the mutable state handed to `Apply` so that
each `SetXxxProp` / stack mutation records its namespaced path into a
`writeSet map[string]bool`. After Apply, the engine invalidates a cached
plan-result **only if `plan.allStatePaths ∩ writeSet ≠ ∅`** (plus always if the
plan `readsPhase` and the phase changed). Concretely: `moveRevealCard`'s Apply
writes `Players[current].CardsLeftToReveal` and the two card stacks; a *different*
move type whose preconditions read only `Game.Deck` keeps its cached verdict
across that transition. Instead of "every state change invalidates every move's
Legal," we get "a state change invalidates only the plans that read what changed."

Write-set capture is O(writes) and needs no author cooperation — it's
instrumentation on the existing mutable-reader path. Path granularity matches
`Refs.StatePaths` granularity because both use the same namespaced vocabulary.

### (d) Cost-ordered short-circuit — cheap rejections stay cheap
Because predicates carry `Cost`, the common rejection (wrong phase, not your
turn) is a `CostTrivial`/`CostCheap` check that fires first and short-circuits
before any `CostModerate` player-loop or `CostExpensive` graph search runs. The
blackjack "all players finished" loop (`CostModerate`) only runs once the phase
gate passed.

**Summary of complexity shift:**

| Operation | Today | This design |
|---|---|---|
| fixup candidate filtering | O(N moves) × Legal | O(1) phase index |
| move-forms per request | 2N × full Legal | N × (cached field-indep + field-dep) |
| verdict after a state change | invalidate all | invalidate only path-overlapping plans |
| custom/graph legality | every time | last, only if declarative gate passed |

What stays O(Legal): the imperative `LegalCustom` residue (checkers capture
graph) — by design uncacheable, but now gated behind cheap declarative checks so
it runs far less often.

---

## 8. Progression

Progression matching becomes **just another precondition**: `InProgression`.
Its `Refs` is `{StatePaths:["Game.moveHistory"], ReadsPhase:true}` — it reads the
move tape since the last phase transition. This folds it into the plan/cache/
explain machinery uniformly: it's `CostModerate` (walks the tape), so it runs
after the trivial phase/turn gates, and its verdict caches under dirty-tracking
(the tape only changes when a move applies, invalidating exactly the plans that
read `Game.moveHistory` — i.e. every progression-bearing move, correctly).

The tape memoization the current code TODOs about (default.go:475) is subsumed:
`historicalMovesSincePhaseTransition` becomes a per-version memoized read behind
the `Game.moveHistory` path, shared by every `InProgression` predicate in that
version's evaluation.

**#644 (state-dependent Repeat counts).** `Repeat(n)` today takes a static count.
Generalize the group combinators to accept a **count resolved from state** via a
predicate-adjacent value source: `RepeatFromProp("Game.RoundsThisTurn")`. Because
`InProgression` already declares `Refs.StatePaths`, adding the backing property
to its ref set is mechanical — the count-source path joins the read-set, so
dirty-tracking still knows when to invalidate. The group's `Satisfied(tape)`
signature gains access to `PreconditionInput` (it already gets state indirectly)
to resolve the dynamic count. No new caching story needed; it rides existing
metadata.

---

## 9. Migration (the acid test)

### 9.1 memory/moveRevealCard

**Before** (moves.go:39-62): imperative chain + three nil/range checks + MayMoveToSlot.

**After** — declarative preconditions replace the whole body; `Legal` is deleted.

```go
//boardgame:codegen
type moveRevealCard struct {
    moves.CurrentPlayer
    CardIndex int
}

// auto.Config:
auto.Config(new(moveRevealCard),
    moves.WithPreconditions(
        // CurrentPlayer contributes proposerIsCurrentPlayer automatically.
        preconditions.PropAtLeast("Players[current].CardsLeftToReveal", "1").
            WithMessage("You have no cards left to reveal this turn"),
        preconditions.RevealableCardAt("HiddenCards", "VisibleCards", "CardIndex"),
        preconditions.MayMoveToSlot("HiddenCards", "VisibleCards", "CardIndex"),
    ),
)
```

`RevealableCardAt` is a small built-in encoding memory's two-branch nil check,
preserving both strings exactly via `MessageTemplate` per branch:

```go
// package preconditions
func RevealableCardAt(hidden, visible, field string) Spec { /* name:"revealableCardAt" */ }

// resolved predicate's Eval:
func (p *revealableCardAt) Eval(in PreconditionInput) *PreconditionFailure {
    g := in.State.ImmutableGameState().Reader()
    idx := moveIntField(in.Move, p.field)             // "CardIndex"
    hidden, _ := g.ImmutableStackProp(p.hidden)
    if hidden.ImmutableComponentAt(idx) != nil {
        return nil
    }
    visible, _ := g.ImmutableStackProp(p.visible)
    if visible.ImmutableComponentAt(idx) == nil {
        return fail("noCardAtIndex", "there is no card at that index")
    }
    return fail("alreadyRevealed", "that card has already been revealed")
}
func (p *revealableCardAt) Refs() Refs {
    return Refs{StatePaths: []string{"Game." + p.hidden, "Game." + p.visible},
                MoveFields: []string{p.field}, Cost: CostCheap}
}
```

Error messages preserved verbatim; `MayMoveToSlot` reuses the existing
`ImmutableComponentInstance.MayMoveToSlot`. Every path is `Visible` under memory's
sanitization for the current player, so all three become `Evaluable:true` on the
client — reveal buttons gray out live.

### 9.2 blackjack/moveStartRoundCleanup

**Before** (moves.go:39-53): StartPhase chain + loop over active players.

**After** — one built-in captures "every active player matches predicate":

```go
//boardgame:codegen
type moveStartRoundCleanup struct {
    moves.StartPhase
}

auto.Config(new(moveStartRoundCleanup),
    moves.WithPreconditions(
        // StartPhase contributes its phase/progression preconditions automatically.
        preconditions.AllActivePlayers(
            preconditions.AnyOf(
                preconditions.PlayerBool("Eliminated"),
                preconditions.PlayerBool("Stood"),
            ),
        ).WithMessage("not all active players have finished their turn"),
    ),
)
```

`AllActivePlayers` skips inactive players (reusing `behaviors.PlayerIsInactive`)
and applies its inner predicate; `AnyOf` is the *one* allowed compositor,
implemented as a built-in constructor (not a user AST — it's a named predicate
with sub-specs, still fully serializable). `Refs`:
`{StatePaths:["Players[*].Eliminated","Players[*].Stood"], Cost:CostModerate}`.
Message preserved. This is `CostModerate`, so it runs only after StartPhase's
cheap gates pass — the player loop is skipped in the common wrong-phase case.

### 9.3 checkers/moveMoveToken (the residue case)

**Before** (moves.go:94-138): CurrentPlayer chain, swap-legality, nil check,
ownership, black-space, then capture-graph search.

**After** — the first four checks go declarative; the graph search stays
imperative behind `LegalCustom`:

```go
//boardgame:codegen
type moveMoveToken struct {
    moves.CurrentPlayer
    TokenIndexToMove enum.RangeVal `enum:"spaces"`
    SpaceIndex       enum.RangeVal `enum:"spaces"`
}

auto.Config(new(moveMoveToken),
    moves.WithPreconditions(
        // CurrentPlayer → proposerIsCurrentPlayer.
        preconditions.MaySwapByKey("Spaces", "TokenIndexToMove", "SpaceIndex"),
        preconditions.ComponentPresentAtKey("Spaces", "TokenIndexToMove").
            WithMessage("That space does not have a component in it"),
        preconditions.ComponentPropEqualsCurrentPlayer("Spaces", "TokenIndexToMove", "Color").
            WithMessage("that token isn't your token to move"),
        preconditions.SpacePredicate("SpaceIndex", "spaceIsBlack").
            WithMessage("you can only move to spaces that are black"),
    ),
)

// The genuinely-gnarly capture-graph search remains imperative and honest:
func (m *moveMoveToken) LegalCustom(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    g := state.ImmutableGameState().(*gameState)
    c := g.Spaces.ImmutableComponentAtKey(m.TokenIndexToMove.Value())
    t := c.Values().(*token)
    for _, space := range t.FreeNextSpaces(state, m.TokenIndexToMove.Value().Int()) {
        if m.SpaceIndex.Value().Int() == space {
            return nil
        }
    }
    for _, space := range t.LegalCaptureSpaces(state, m.TokenIndexToMove.Value().Int()) {
        if m.SpaceIndex.Value().Int() == space {
            return nil
        }
    }
    return errors.New("spaceIndex does not represent a legal space for that token to move to")
}
```

`SpacePredicate` names a registered board-geometry helper (`spaceIsBlack`),
keeping the check serializable. The four declarative preconditions are cheap and
run first; the expensive graph search runs only when they pass, and the client
sees the first four as `Evaluable:true` (grays the obviously-illegal moves) and
the fifth as `Passed:"unknown", Evaluable:false` — honest that the client can't
compute the capture graph. This is the coexistence story made concrete: ~80% of
illegal clicks are rejected client-side by the declarative gates; only the
graph-dependent residue needs the server.

---

## Risks & open questions

- **Write-set capture fidelity.** Dirty-tracking (§7c) assumes we can cheaply and
  completely record the paths an `Apply` mutates. Stacks mutate via component
  moves that touch two stacks; timers, dynamic component values, and
  behaviors-driven mutations must all funnel through instrumented setters or the
  cache goes stale-and-wrong. Needs an audit of every mutation path in core; a
  conservative fallback ("if unsure, invalidate everything this version") keeps
  it correct but erodes the win. This is the riskiest load-bearing claim.
- **Path vocabulary vs `Players[current]`.** Resolving `Players[current]` for
  `Refs` requires knowing the current player *before* eval, but `CurrentPlayer`
  can itself change. Simplest safe rule: `Players[current]` in a ref-set expands
  to `Players[*]` for invalidation purposes (coarser but sound).
- **`AnyOf`/`AllActivePlayers` re-open the AST door.** One layer of named
  compositors is enough for the surveyed moves, but the moment someone wants
  `AnyOf(AllOf(...))` nesting we're back toward the tarpit. Recommend: allow
  exactly one level of built-in compositor, everything deeper is `LegalCustom`.
- **Field-independent classification correctness.** A predicate that reads a move
  field only *sometimes* (data-dependent) must be classified field-dependent or
  the field-indep cache is wrong. `Refs()` must be a conservative over-approximation
  — verified by construction for built-ins, but the escape hatch for custom
  predicate authors needs a lint/test.
- **Cache lifetime & memory.** Per-version caches must be bounded to the game's
  live window (the same concern default.go:476 already notes). Keyed by
  `stateVersion`, evicted when the version is no longer the head — needs wiring
  into the game's version lifecycle.
- **Ordering-visible message changes.** Cost-reordering within a bucket can change
  *which* failure a user sees first vs today's hand-ordered chain. Mostly an
  improvement (cheapest/most-relevant first), but a few examples may report a
  different-but-equally-true message; migration tests should snapshot these.
