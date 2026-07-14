# Design C — Delight: Declarative Move Legality

**Lens: Developer Experience & Explainability.** The author should write a
precondition that reads like a rulebook sentence; the player should read a
rejection that reads like a helpful friend; a newcomer should learn the whole
system from one doc page.

## Executive summary

A **precondition** is a serializable, named `Predicate` value — the direct
sibling of the `StackConstraint` that landed recently, rhyming with it beat for
beat (name + string args, a constructor registry, struct-tag sugar). A move
carries an ordered **`[]Predicate`** attached three ways that all funnel to the
same slice: `With(...)` options in `auto.Config`, struct tags for the common
five, and a fluent `Require(...)` for one-offs. Each predicate answers one
question against a small, explicit **`Context`** (state, move fields, proposer,
phase, chest) and returns not a bare `error` but a **`Verdict`**: `Pass`,
`Fail{template, bindings}`, or `Unknown{reason}` (the last is what makes the
client story honest under sanitization). Player-facing text lives in **message
templates with named bindings** — `"You have {left} cards left to reveal"` —
authored once, rendered server-side today and client-side tomorrow from the
same serialized form. The engine indexes predicates by **referenced property
paths and phase**, so the fixup loop and move-forms recomputation skip any move
whose inputs didn't change: the common case drops from O(Legal) to O(1 phase
compare). `Legal()` survives as a **first-class escape hatch** — the checkers
capture graph stays imperative and simply reports `Unknown` to the client. The
existing three buried checks (phase / progression / stack) become three
built-in predicates, so nothing is special-cased anymore. Progression stays a
distinct matcher but is *wrapped* as one predicate so it composes uniformly.

Core `boardgame` gets two tiny interfaces and nothing else. Everything with
opinions lives in a new **`legal`** package that depends on `boardgame` and is
depended on by `moves`.

---

## 1. Representation

A precondition is a **`Predicate`**: a named, serializable check that reads a
`Context` and returns a `Verdict`. It is deliberately the same shape as
`constraints.StackConstraint`, so an author who has met one already knows the
other.

```go
// package legal (new)

// Verdict is the three-valued result of evaluating a Predicate.
type Verdict struct {
    // Outcome is Pass, Fail, or Unknown.
    Outcome Outcome
    // Message is set on Fail (and optionally Unknown): a template key plus the
    // bindings to render it. Nil on Pass.
    Message *Message
    // Reason, on Unknown, says WHY we couldn't decide (e.g. "reads hidden
    // property HiddenCards"). Drives the client story in Q6.
    Reason string
}

type Outcome int
const ( Pass Outcome = iota; Fail; Unknown )

// A Message is a template key + named bindings. Never a pre-baked string, so
// it can be re-rendered in any locale, on server or client. See Q4.
type Message struct {
    Template string                 // "reveal.no_cards_left"
    Bindings map[string]any         // {"left": 0}
}

// Predicate is one legality question. Evaluate is pure: same Context in, same
// Verdict out — no state mutation, no I/O.
type Predicate interface {
    // Name is the registered name, e.g. "player_prop_at_least". Powers
    // serialization and the constructor registry.
    Name() string
    // Args are the string arguments, e.g. ["CardsLeftToReveal", "1"]. Together
    // Name()+Args() round-trip through the registry.
    Args() []string
    // Reads returns the property paths this predicate touches, so the engine
    // can index it (Q7) and the sanitizer can decide evaluability (Q6). Empty
    // slice = depends only on move fields / proposer (always client-evaluable).
    Reads() []PropPath
    // Evaluate answers the question. Cheap; called a lot.
    Evaluate(ctx Context) Verdict
}
```

The `Context` is the *entire* vocabulary a predicate may reference — the line
before the Turing tarpit is drawn here, by construction:

```go
type Context struct {
    State    boardgame.ImmutableState  // property paths, stacks, phase via delegate
    Move     boardgame.Move            // move fields (CardIndex, SpaceIndex…)
    Proposer boardgame.PlayerIndex
    Chest    *boardgame.ComponentChest // enum values, constants
}

// A PropPath names a readable location, reusing constraints' path grammar so
// authors learn one syntax: "game.HiddenCards", "player.CardsLeftToReveal",
// "player:current.Color". The "current" selector resolves against CurrentPlayer.
type PropPath string
```

**Where the line is.** A `Predicate` is *not* arbitrary Go — you cannot write a
loop or call out. It is a leaf question drawn from a fixed catalog
(`prop_compare`, `stack_size`, `component_may_move`, `phase_is`,
`proposer_is_current`, `all_players`, `progression`…). Combination is
structural, not procedural: `All(...)`, `Any(...)`, `Not(...)` are themselves
predicates. Anything that can't be expressed drops to the imperative escape
hatch (Q5) and honestly reports `Unknown` to the client. This is the same
bargain the constraints package already struck, and it's why it stayed
delightful instead of sprawling.

**Two serialized forms, both first-class** (mirroring constraints exactly):

```jsonc
// name + args (the wire + struct-tag form)
{"name": "player_prop_at_least", "args": ["CardsLeftToReveal", "1"]}
// composite AST (nesting is just predicates-with-predicate-args)
{"name": "all", "children": [
  {"name": "proposer_is_current", "args": []},
  {"name": "player_prop_at_least", "args": ["CardsLeftToReveal", "1"]}
]}
```

---

## 2. Attachment & composition

Three authoring surfaces, one destination — a move's ordered `[]Predicate`.
All three are sugar over `WithPreconditions`, so there is exactly one thing to
learn and two shortcuts.

**(a) Struct tags** for the common five (phase, current-player, stack size,
prop compare, may-move). This is the "reads like the field it guards" surface:

```go
type moveRevealCard struct {
    moves.CurrentPlayer
    CardIndex int `legal:"index_into(game.HiddenCards)"`
}
```

**(b) `With(...)` options** in `auto.Config`, identical in feel to the existing
`WithLegalPhases` / `WithSourceProperty` an author already uses:

```go
auto.MustConfig(new(moveRevealCard),
    moves.WithPlayerPropAtLeast("CardsLeftToReveal", 1),
    moves.WithComponentMayMove("game.HiddenCards", "@CardIndex", "game.VisibleCards", "@CardIndex"),
)
```

`WithPlayerPropAtLeast` is a thin constructor returning a
`CustomConfigurationOption` that *appends* to `configPropPreconditions` (same
append idiom `WithLegalPhases` already uses at with.go:97-99):

```go
func WithPlayerPropAtLeast(prop string, min int) CustomConfigurationOption {
    return WithPreconditions(legal.PlayerPropAtLeast(prop, min))
}
func WithPreconditions(preds ...legal.Predicate) CustomConfigurationOption {
    return func(c boardgame.PropertyCollection) {
        prev, _ := c[configPropPreconditions].([]legal.Predicate)
        c[configPropPreconditions] = append(prev, preds...)
    }
}
```

**(c) Fluent `Require`** for the case an author reaches for once and never
abstracts — reads like a sentence, no registration needed for locally-built
composites:

```go
moves.WithPreconditions(
    legal.All(
        legal.PhaseIs("Play"),
        legal.Not(legal.StackEmpty("game.DrawStack")),
    ),
)
```

**Composition down the embedding chain.** Preconditions **accumulate** — a game
move's slice is `Default`'s ∪ `CurrentPlayer`'s ∪ its own, in embedding order,
so `CurrentPlayer` contributing `proposer_is_current` "just works" without the
`if err := m.CurrentPlayer.Legal(...)` boilerplate every game move writes today.
Accumulation happens once at config time, not per-eval.

**Removing an inherited precondition** is explicit and rare:
`moves.WithoutPrecondition(name)` drops matching entries from the accumulated
slice (matched by `Name()`+`Args()`). This is the escape valve for "I embed
CurrentPlayer but this one move is legal for anyone."

**Interaction with `auto.Config`.** None that's new: preconditions ride in the
same `PropertyCollection` bag as every other `With*` option today. `auto.Config`
doesn't need to know they exist — it already forwards the whole bag into
`CustomConfiguration()`.

---

## 3. Layering

```
core boardgame  ──  two interfaces, zero opinions:
                    • Move optionally implements Preconditioned{ Preconditions() []Predicate }
                    • Predicate/Verdict/Context/PropPath type defs? NO — see below.

     ▲ (imports)                    ▲ (imports)
     │                              │
  package legal  ─────────────  the whole system:
     Predicate, Verdict, Context, Message, the catalog
     (PhaseIs, PlayerPropAtLeast, StackSize, ComponentMayMove,
      AllPlayers, Progression-wrapper, All/Any/Not), the registry,
      the renderer, the evaluator, the indexer.

     ▲ (imports)
     │
  package moves  ──  the With* sugar + struct-tag parsing; the built-in
                     moves whose buried checks now delegate to legal predicates.
```

Following #761's instinct, **core `boardgame` gets the *minimum*.** It defines a
single interface so the engine can find preconditions without importing
`legal`:

```go
// package boardgame
type Preconditioned interface {
    // Preconditions returns the move's accumulated, ordered predicates. The
    // engine treats each as an opaque Evaluator; the concrete Predicate type
    // lives in package legal.
    Preconditions() []Evaluator
}
// Evaluator is the core-visible face of a legal.Predicate: just enough for the
// engine to evaluate and index without knowing what a predicate IS.
type Evaluator interface {
    Evaluate(state ImmutableState, move Move, proposer PlayerIndex) LegalResult
    Reads() []string // property paths, for dirty-tracking
}
type LegalResult struct {
    Legal   bool
    Unknown bool
    Error   error // rendered message; nil when Legal
}
```

`legal.Predicate` satisfies `boardgame.Evaluator`. Core stays ignorant of
templates, bindings, and the catalog. **The three existing buried checks**
(`legalInPhase`, `legalMoveInProgression`, `legalStackConstraints`) move out of
`moves.Default.Legal()` and become three registered predicates in `legal`
(`phase_is`, `progression`, `stack_source_dest`), auto-attached by the same
`With*` options that configure them today. `Default.Legal()` shrinks to: "if I
implement `Preconditioned`, the engine already ran them; return nil." Nothing is
special-cased anymore — phase and progression are ordinary catalog entries.

---

## 4. Explainability

The error model is a **`Verdict`**, never a bare string. On `Fail` it carries a
`Message{Template, Bindings}`. Templates are authored per-game (and per-builtin)
in a flat registry the delegate exposes, keyed by dotted names:

```go
// A game registers templates once. Builtins ship defaults; games override.
var memoryTemplates = legal.Templates{
    "reveal.no_cards_left":   "You have no cards left to reveal this turn",
    "reveal.no_card_here":    "There is no card at position {index}",
    "reveal.already_revealed":"The card at position {index} has already been revealed",
}
```

A predicate names its template and supplies bindings; it never formats prose:

```go
func (p *playerPropAtLeast) Evaluate(ctx legal.Context) legal.Verdict {
    have := ctx.PlayerInt("current", p.prop) // p.prop == "CardsLeftToReveal"
    if have >= p.min { return legal.Pass() }
    return legal.FailT("reveal.no_cards_left", legal.B{"left": have, "min": p.min})
}
```

**Rendering** is one function, `legal.Render(msg, templates)` →
`"You have no cards left to reveal this turn"`. Called server-side today;
shipped to the client as `(template, bindings)` tomorrow so the *same* failure
renders locally without a round trip and in the player's locale.

**What the client receives** per move (superset of today's `MoveForm`):

```jsonc
{
  "Name": "Reveal Card",
  "Legal": false,                       // Pass?  (was LegalForPlayer)
  "Failure": {                          // first failing precondition
    "Template": "reveal.no_cards_left",
    "Bindings": {"left": 0, "min": 1},
    "Rendered": "You have no cards left to reveal this turn"
  },
  "Preconditions": [                     // full ledger, for gray-out + tooltips
    {"name": "proposer_is_current", "verdict": "pass",    "clientEvaluable": true},
    {"name": "player_prop_at_least","verdict": "fail",    "clientEvaluable": true,
       "template": "reveal.no_cards_left", "bindings": {"left": 0}},
    {"name": "component_may_move",  "verdict": "unknown", "clientEvaluable": false,
       "reason": "reads hidden property game.HiddenCards"}
  ]
}
```

**Fixup-loop logging (#65).** When `ProposeFixUpMove` polls a fixup and it
fails, the engine logs the structured `Verdict` — *which* precondition failed
with *which* bindings — behind the delegate's existing debug hook:

```
[fixup] StartHideCardsTimer rejected: precondition stack_size(game.VisibleCards)==2
        failed: got 1, want 2
```

Nine years of "why didn't my fixup fire?" (#65) becomes greppable, because the
failure is data, not a string returned and discarded.

---

## 5. Evaluation semantics & the escape hatch

**Ordering** is per #761's two classes, sorted at config time so it's free at
runtime:

1. **Field-independent** predicates first (`Reads()` touches only state/phase/
   proposer — `phase_is`, `proposer_is_current`, `all_players_finished`). These
   are checkable *before* `DefaultsForState` and field-binding, and dominate the
   move-forms fast path (a whole move can be culled before any field logic).
2. **Field-dependent** predicates next (`Reads()` mentions `@Field` — `index_into`,
   `component_may_move`). Checked after fields are bound.
3. **Imperative residue** (`Legal()`) last, only if the move still implements it.

Within each class, **cheap→expensive** by a static `cost` the catalog declares
(a `prop_compare` is cost 1; `component_may_move` walks a stack, cost 10).
**Short-circuit** on first `Fail`. `Unknown` does **not** short-circuit — we
keep evaluating so the client gets the *fullest* ledger possible; the move's
overall verdict is `Fail` if any `Fail`, else `Unknown` if any `Unknown`, else
`Pass`. **Determinism** is guaranteed because `Evaluate` is pure and ordering is
fixed at config time.

**The escape hatch.** `Legal(state, proposer) error` stays a legal method on any
move. The checkers capture-graph logic (genuinely gnarly) simply keeps its
`Legal()`. The engine runs preconditions first, then — only if all passed —
calls `Legal()`. To the client, a move that still has an imperative `Legal()`
tail reports one synthetic precondition `{"name": "custom", "verdict":
"unknown", "reason": "server-only rule"}`, so the button shows enabled-but-
tentative (not grayed) and a click still round-trips. Authors opt a custom check
*out* of "unknown" by declaring it can't be client-evaluated *and* that a
server confirm is required — the honest default.

---

## 6. Sanitization-aware client story (designed-for)

The three-valued `Verdict` exists precisely for this. A predicate declares its
`Reads()` paths; the server cross-references them against the **sanitization
policy applied to the state it just sent the client**:

```go
func clientEvaluable(pred Evaluator, policy SanitizationPolicy) bool {
    for _, path := range pred.Reads() {
        if policy.IsHidden(path) || policy.IsLen(path) /* zeroed shape */ {
            return false
        }
    }
    return true
}
```

- If **all** a predicate's reads are Visible in the sanitized state → the client
  can evaluate it locally (tomorrow, in TS) and gray the button *without a
  round trip*. Marked `clientEvaluable: true`.
- If **any** read is Hidden/Len → the predicate is `clientEvaluable: false`; the
  server sends its *server-computed* verdict as advisory, and the client treats
  it as `Unknown` for local re-evaluation (it can't recompute `CardsLeftToReveal`
  logic on a card it can't see, but it can *display* the server's verdict).

So `memory`'s `component_may_move` over `HiddenCards` is honestly `Unknown` on
the client (it reads hidden cards); `player_prop_at_least("CardsLeftToReveal")`
is `clientEvaluable` iff that counter is Visible under policy. The client never
guesses "0 means hidden 7" — the server told it which reads it may trust. This
is the whole reason the representation is data with declared reads, not a
closure.

---

## 7. Engine wins

Today the fixup loop and `generateFormsWithLegality`
(server/api/main.go:1590-1629) call `Legal()` on every move after every state
change — and each `Legal()` re-runs phase + progression + stack + custom. With
declarative predicates we index at manager-build time and dirty-track at
runtime.

- **Phase index.** Every move's `phase_is` predicates are known statically.
  Build `map[EnumKey][]Move`. When computing move-forms, the fixup loop, or
  candidates, we only consider moves legal in the current phase — the ~5 phase
  checks across the survey become **one map lookup**, not N phase evaluations.
- **Read-path dirty tracking.** Each predicate declares `Reads()`. After a move
  applies, the engine diffs which property paths changed (it already produces
  state deltas for storage). A move's cached verdict is **reused** unless a
  changed path intersects its predicates' reads. In the fixup loop's 256-deep
  recursion, most candidate fixups read a handful of stacks; if those didn't
  change, they're skipped at **O(1)**.
- **What becomes O(cheap):** phase gate (map lookup), field-independent
  predicates over unchanged state (cache hit), the whole move-forms recompute
  when only one player's private stack changed (other moves' reads didn't
  intersect).
- **What stays O(Legal):** any move with an imperative `Legal()` tail
  (checkers capture graph) — but even those get the phase gate for free, so
  they're only evaluated when actually in-phase.

Concretely: memory's per-request move-forms goes from "run 6 moves' full
`Legal()` twice" to "1 phase lookup + re-evaluate only the predicates whose
reads changed since the cached version." The common idle-refresh is a cache hit.

---

## 8. Progression

Move-tape progression becomes **one predicate**, `progression`, that *wraps* the
existing `MoveProgressionGroup` matcher (moves/groups.go) — it does **not**
dissolve into leaf predicates, because the tape match is inherently stateful
across the move history and doesn't fit the pure-`Context` leaf shape. Wrapping
it lets it compose uniformly (it sits in the same `[]Predicate` slice, gets the
same `Verdict`/template treatment, shows up in the client ledger) while keeping
its specialized matcher internally:

```go
func Progression(group moves.MoveProgressionGroup) legal.Predicate // wraps groups.go
// Reads() returns the move-history path so the indexer knows it's history-
// dependent (re-evaluate whenever a move is applied, never a plain state read).
```

**#644 (state-dependent Repeat counts).** A `Repeat(n)` in a progression can
today only take a constant `n`. We let `n` be a **`legal.IntExpr`** — the same
`prop_path` grammar predicates use — so `Repeat(legal.Prop("game.NumRounds"))`
reads the count from state at match time. The IntExpr is evaluated against the
`Context` at progression-match time; because it's the same path grammar, an
author who learned `player.CardsLeftToReveal` for a `prop_compare` already knows
how to write a state-dependent repeat count. No new concept, just letting an
existing hole take an existing expression type.

---

## 9. Migration — the three acid tests

### (1) memory/moveRevealCard

```go
type moveRevealCard struct {
    moves.CurrentPlayer
    CardIndex int
}

// Preconditions replace the whole Legal() body. Ordered field-independent →
// field-dependent by the engine automatically.
func configRevealCard(auto *moves.AutoConfigurer) boardgame.MoveConfig {
    return auto.MustConfig(new(moveRevealCard),
        // proposer==current comes free from embedded CurrentPlayer.
        moves.WithPlayerPropAtLeast("CardsLeftToReveal", 1),        // field-independent
        moves.WithComponentPresent("game.HiddenCards", "@CardIndex", // field-dependent
            "reveal.no_card_here", "reveal.already_revealed_if(game.VisibleCards)"),
        moves.WithComponentMayMove("game.HiddenCards", "@CardIndex",
            "game.VisibleCards", "@CardIndex"),
    )
}

var memoryTemplates = legal.Templates{
    "reveal.no_cards_left":    "You have no cards left to reveal this turn",
    "reveal.no_card_here":     "There is no card at that index",
    "reveal.already_revealed": "That card has already been revealed",
}
```

`DefaultsForState` and `Apply` are unchanged. **No `Legal()` method at all.**
Every original error string is preserved verbatim as a template — including the
tricky "no card here vs. already revealed" branch, which becomes the
`WithComponentPresent` predicate that checks the source slot and, when empty,
consults the destination slot to pick which of the two templates to emit. The
`c.MayMoveToSlot` pre-check is the `component_may_move` predicate — and it's
correctly flagged `Unknown` on the client because it reads `HiddenCards`.

### (2) blackjack/moveStartRoundCleanup

```go
type moveStartRoundCleanup struct {
    moves.StartPhase
}

func configStartRoundCleanup(auto *moves.AutoConfigurer) boardgame.MoveConfig {
    return auto.MustConfig(new(moveStartRoundCleanup),
        moves.WithStartPhase(phaseRoundCleanup),
        // "every active player satisfies (Eliminated OR Stood)"
        moves.WithEveryActivePlayer(
            legal.Any(
                legal.PlayerPropTrue("Eliminated"),
                legal.PlayerPropTrue("Stood"),
            ),
            "cleanup.players_unfinished",
        ),
    )
}

var blackjackTemplates = legal.Templates{
    "cleanup.players_unfinished": "not all active players have finished their turn",
}
```

`WithEveryActivePlayer` is the `all_players` catalog entry with the active-player
filter built in (it reuses `behaviors.PlayerIsInactive` under the hood). The loop
vanishes; the rule reads as the sentence it is. **No `Legal()` method.** Because
`Eliminated`/`Stood` are typically Visible, this precondition is
`clientEvaluable` — blackjack can gray the phase-advance affordance locally.

### (3) checkers/moveMoveToken

```go
type moveMoveToken struct {
    moves.CurrentPlayer
    TokenIndexToMove enum.RangeVal `enum:"spaces"`
    SpaceIndex       enum.RangeVal `enum:"spaces"`
}

func configMoveToken(auto *moves.AutoConfigurer) boardgame.MoveConfig {
    return auto.MustConfig(new(moveMoveToken),
        // proposer==current free from CurrentPlayer.
        moves.WithComponentPresent("game.Spaces@key", "@TokenIndexToMove",
            "checkers.empty_space"),
        moves.WithProposerOwnsComponent("game.Spaces@key", "@TokenIndexToMove",
            "Color", "checkers.not_your_token"),
        moves.WithMoveFieldSatisfies("@SpaceIndex", "space_is_black",
            "checkers.not_black"),
        // capture-graph logic stays imperative — see Legal() below.
    )
}

var checkersTemplates = legal.Templates{
    "checkers.empty_space":   "That space does not have a component in it",
    "checkers.not_your_token":"that token isn't your token to move",
    "checkers.not_black":     "you can only move to spaces that are black",
    "checkers.illegal_dest":  "spaceIndex does not represent a legal space for that token to move to",
}

// The gnarly part stays imperative — and honestly reports Unknown to the client.
func (m *moveMoveToken) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    // Preconditions above already ran and passed (ownership, black, presence,
    // MaySwap). Only the capture/free-space graph walk remains.
    g := state.ImmutableGameState().(*gameState)
    t := g.Spaces.ImmutableComponentAtKey(m.TokenIndexToMove.Value()).Values().(*token)
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
    return legal.Errorf("checkers.illegal_dest", nil) // still a template!
}
```

Three of the four checks become declarative predicates with their exact
original strings; `space_is_black` is registered once as a named boolean
`MoveFieldSatisfies` predicate (it's pure over a move field, so it's fully
client-evaluable — checkers can gray non-black targets with zero round trips).
The `MaySwapComponentsByKey` pre-check folds into `WithComponentPresent`. Only
the genuine graph walk stays in `Legal()` — and even it returns a *template*
error, so its rejection is still structured, localizable, and greppable. To the
client this move is "custom / Unknown," so the destination-legality tooltip says
"the server will confirm" for the capture logic while black/ownership gray
instantly.

---

## Risks & open questions

- **`@Field` binding syntax vs. type safety.** `"@CardIndex"` is a stringly-typed
  reference resolved at eval time; a typo isn't caught until runtime. Mitigation:
  validate all `@Field` and `PropPath` references at `NewGameManager` time
  against the move's reader + chest (constraints' `validatePropPath` already
  proves this pattern works). Still, struct tags are less safe than method calls.
- **Template registry ergonomics.** Flat string keys can collide across games in
  a shared server. Namespacing by move/game prefix is a convention, not enforced.
  Worth deciding whether templates live per-game-delegate or in a global chest.
- **`all_players` / `every_active_player` `Reads()`** touches every player state,
  so its dirty-tracking is coarse (any player change re-evaluates). Acceptable,
  but it's the predicate most likely to blunt the engine-win in player-heavy
  games.
- **Progression as a wrapped opaque** means the client can't locally evaluate
  tape position without the move history + the group tree shipped down. Deferred
  to the TS follow-up; today it's server-computed and marked `clientEvaluable:
  false`. Honest, but it's the biggest gap in the "no round trip" story.
- **Does the imperative `Legal()` tail double-evaluate work** the predicates
  already did (e.g. re-reading the token)? Minor; the graph walk needs the token
  anyway. But it means an author must remember preconditions ran — documented,
  not enforced.
- **`Unknown` UX is a product decision, not just an API one.** "Enabled but
  tentative" vs. "grayed with spinner" for `Unknown` moves needs a client design
  pass; this doc specifies the data, not the pixels.
