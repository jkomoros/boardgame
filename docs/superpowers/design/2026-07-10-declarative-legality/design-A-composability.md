# Design A — Composability: One Precondition Algebra

## Executive summary

The framework already speaks a small language three times: the **groups** algebra
(`Serial`/`Parallel`/`Repeat` combinators over a tape), the **constraints** registry
(name+args → `StackConstraint`, serializable, chest-aware), and the **With\*** config
idiom (namespaced key writes read at runtime). Legal() today is a fourth, ad-hoc
dialect written in raw Go. This design unifies all of it into **one precondition
algebra**: a `Cond` — a serializable predicate over `(state, move, proposer)` that
returns a three-valued result (`Pass`/`Fail`/`Unknown`) plus a structured, templated
failure. `Cond`s are built from **five atom families** (Phase, Turn, Compare, Stack,
Progression) and **three combinators** (`All`, `Any`, `Not`) — the *same* algebraic
shape as groups, so the framework reads as one language, not three. Every atom is
registered by name+args exactly like a `StackConstraint`, so serialization, the future
TS evaluator, and struct-tag authoring all fall out of one registry. Preconditions
attach via a new `WithLegal(...Cond)` option that **accumulates** down the embedding
chain (Default's phase/turn/stack/progression checks become *pre-registered* `Cond`s on
the base move, not buried Go), and a subclass may `WithoutLegal(name)` to prune an
inherited atom. The imperative escape hatch survives as a single `CustomLegal` atom
wrapping a Go closure that always reports `Unknown` to the client. The engine indexes
`Cond`s by referenced phase and property-path so the fixup loop and move-forms
computation skip the expensive `Cond`s whose inputs did not change. Three orthogonal
concepts — **atom, combinator, registry** — cover all nine acid tests.

Layering: core `boardgame` gains only the `Cond` interface + `Eval` result type
(~40 lines). The `moves` package gains the atom constructors and the `WithLegal`
plumbing. Progression stops being special: it becomes the `InProgression` atom, and
#644's state-dependent counts become an argument that reads a property path.

---

## 1. Representation: what is a precondition?

A precondition is a **`Cond`** — a named, serializable predicate. Core interface
(new file `cond.go` in package `boardgame`, ~40 lines total):

```go
package boardgame

// Trit is three-valued logic. Unknown means "a referenced input was sanitized
// away, so this proposer cannot decide locally" — never a failure, always a
// deferral to the server.
type Trit uint8

const (
	Unknown Trit = iota // must be zero value: an un-evaluated Cond is Unknown
	Pass
	Fail
)

// Eval is the structured outcome of evaluating one Cond.
type Eval struct {
	Result   Trit
	Cond     string            // atom/combinator name, e.g. "player.turn"
	Args     []string          // the serialized args that were in play
	Bindings map[string]string // resolved values for message placeholders
	Sub      []Eval            // children (for All/Any/Not), for explainability
}

// Cond is a serializable precondition. Eval is the runtime check; Marshal is
// the name+args serialization (mirrors StackConstraint's registry contract).
type Cond interface {
	Eval(state ImmutableState, move ImmutableMove, proposer PlayerIndex) Eval
	Serialize() SerializedCond // {Name, Args, Sub}
}

type SerializedCond struct {
	Name string           `json:"name"`
	Args []string         `json:"args,omitempty"`
	Sub  []SerializedCond `json:"sub,omitempty"` // combinator children
}
```

**What a `Cond` may reference** — deliberately bounded to keep it out of the Turing
tarpit. Every reference is a **string path** resolved by the shared resolver
(reused from `constraints/prop_path.go`), never arbitrary Go:

| Reference | Path syntax | Example |
|---|---|---|
| Game property | `game.Prop` | `game.DrawStack` |
| Current-player property | `player.Prop` | `player.CardsLeftToReveal` |
| Named-player property | `players[i].Prop` | (rare; `player.` covers 8 of ~8 cases) |
| Move field | `move.Field` | `move.CardIndex` |
| Proposer / turn | atom-level | `player.turn` |
| Phase | atom-level | `phase.legal(Playing)` |
| Chest constant | literal arg resolved via chest | `spaceIsBlack` → **not expressible** |
| Computed property | `game.ComputedProp` (same resolver) | `game.HandValue` |

**The line before the tarpit.** Atoms are *relations over resolved paths and
literals*: equality, ordering, membership, count, presence. There is **no** user
arithmetic, no loops the author writes, no lambdas in the serialized form. Anything
needing real computation (checkers' capture graph, blackjack's ace-aware hand value)
lands behind the single `CustomLegal` atom (§5) or is *precomputed into a state
property* the atom then compares against. This is the deliberate boundary: **if you
can't say it as a relation over a path, you push the computation into state (a
computed property) or into the escape hatch.** That keeps the declarative surface
small, serializable, and TS-portable, and it makes "declarative vs imperative" a
one-line decision for the author.

**Both name+args AND AST — because they're the same thing.** A combinator is just
an atom whose args are other `Cond`s (`Sub`). So `SerializedCond` is simultaneously
the name+args record (leaf) and the AST node (branch). One format, no duplication.

---

## 2. Attachment & composition

**Attachment: one option, accumulating.** A new `moves.WithLegal(conds ...Cond)`
appends to a namespaced slice in the config bag (rhymes exactly with existing
`With*`). Auto-config already threads options; nothing new to learn:

```go
auto.Config(new(moveRevealCard),
	WithMoveName("Reveal Card"),
	WithLegalPhases(phasePlaying),           // sugar → still a Cond under the hood
	WithLegal(
		Compare("player.CardsLeftToReveal", GTE, Lit(1)).
			Else("You have no cards left to reveal this turn"),
		StackHasComponentAt("game.HiddenCards", "move.CardIndex").
			Else("there is no card at that index"),
		MayMoveToSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
	),
)
```

**Composition down the embedding chain — accumulation, not override.** The base
`moves.Default` no longer *hard-codes* phase/progression/stack checks in Go. Instead
its constructor pre-registers them as named `Cond`s in the bag:

```go
// moves.Default seeds these named Conds (author never writes them):
//   "phase.legal"     ← from WithLegalPhases
//   "move.progression" ← from WithLegalMoveProgression
//   "stack.mayMove"   ← from WithSource/DestinationProperty
// moves.CurrentPlayer additionally seeds "player.turn".
```

`Default.Legal()` becomes a thin, uniform evaluator over the **union** of the bag's
`Cond` slice — collected by walking the embedding chain from base to leaf. Because
each layer *appends* named atoms, a game move that embeds `CurrentPlayer` and calls
`WithLegal(...)` automatically has `[phase.legal, player.turn, ...its own...]` in
evaluation order. **No `if err := m.CurrentPlayer.Legal(...)` boilerplate ever
again** — the chain composes declaratively.

**Subclass removal/override.** `WithoutLegal(name)` prunes an inherited atom by name
(e.g. a variant that removes the turn restriction: `WithoutLegal("player.turn")`).
Override = prune + re-add. This is only possible *because* atoms are named — the
name is the same identity used for serialization and for engine indexing. One naming
scheme, three payoffs.

**Interaction with auto.Config.** `WithLegal`/`WithoutLegal` are ordinary
`CustomConfigurationOption`s. The auto-configurer, when it materializes a move,
resolves the accumulated `Cond` slice once and caches it on the move type (they're
immutable per move type). No per-call allocation.

---

## 3. Layering (dependency arrows)

```
                 ┌─────────────────────────────────────────┐
   core          │ boardgame: Cond, Eval, Trit,            │  (~40 lines, no deps
   (minimal      │ SerializedCond, condRegistry            │   on moves/constraints)
    interfaces)  │ + reuses existing prop-path resolver    │
                 └───────────────▲─────────────────────────┘
                                 │ implements Cond
       ┌─────────────────────────┴───────────────────────────┐
       │ moves/legal: atom constructors (Phase, Turn,        │
       │ Compare, Stack, Progression, Custom), combinators   │
       │ (All/Any/Not), WithLegal/WithoutLegal, the uniform  │
       │ Default.Legal() evaluator                           │
       └───▲───────────────────────────▲─────────────────────┘
           │ registers atoms by name    │ shares path resolver
           │                            │
   ┌───────┴────────┐          ┌────────┴──────────┐
   │ constraints    │          │ examples/../games │  (author-facing)
   │ (StackConstraint│          │ WithLegal(...)    │
   │  registry —     │          └───────────────────┘
   │  the pattern we │
   │  rhyme with)    │
   └─────────────────┘
```

Core gets **only** the interface + result types + a name→constructor registry
(structurally identical to `constraintConstructors` in `game_manager.go`). Per #761's
instinct, no atom *logic* lives in core. The three buried checks migrate out of
`Default.Legal()`'s Go body and into named atoms in `moves/legal` — principled,
because they now live beside every other atom, are serializable, and are indexable.

---

## 4. Explainability: the error model

Every atom carries an optional `.Else("template")`. The template uses `{path}` and
`{binding}` placeholders filled from `Eval.Bindings`:

```go
Compare("player.CardsLeftToReveal", GTE, Lit(1)).
	Else("You have no cards left to reveal this turn")

MayMoveToSlot(...).Else("card at {move.CardIndex} can't move there: {reason}")
```

The evaluator produces a **tree of `Eval`s** (via `Sub`), so a failure is fully
structured: which atom failed, with which args and resolved bindings, nested under
which combinator. `Default.Legal()` adapts the top failing `Eval` into the existing
`errors.Friendly` (preserving the current server contract):

```go
func evalToError(e Eval) error {
	f := errors.NewFriendly(renderTemplate(e))       // player-facing message
	return f.WithError(e.Cond + " " + strings.Join(e.Args, ",")) // dev detail
}
```

**Fixup-loop logging (#65).** When `ProposeFixUpMove` finds no legal fixup, it now
has the `Eval` trees for every rejected candidate. It logs, per candidate, the
*first failing atom name + bindings* — turning "no fixup was legal, good luck" into
"`captureCards` failed at `stack.count(game.VisibleCards)==2` (was 1)". This is the
#65 win, and it's free: the structured failure already exists.

**What the client receives** (per move form, unchanged envelope, richer payload):

```json
{
  "name": "Reveal Card",
  "legal": "fail",
  "failed": {
    "cond": "player.compare",
    "args": ["player.CardsLeftToReveal", "gte", "1"],
    "bindings": {"player.CardsLeftToReveal": "0"},
    "message": "You have no cards left to reveal this turn"
  }
}
```

---

## 5. Evaluation semantics & the escape hatch

**Ordering** is the declared slice order, which is *already* cheap→expensive by
construction: base atoms (phase, turn — field-independent, per #761) come first
because the chain seeds them first; the author's field-reading atoms come last. The
evaluator formalizes #761's split with a per-atom `FieldDependent() bool`:
field-independent atoms evaluate in the pre-`DefaultsForState` pass, field-dependent
ones after. **Short-circuit:** `All` stops at first `Fail`; `Any` stops at first
`Pass`. **Determinism:** atoms are pure over their resolved inputs — no clocks, no
RNG, so identical `(state, move)` always yield identical `Eval`.

**Three-valued short-circuit** (this is the subtle part): under `All`, `Fail`
dominates `Unknown` dominates `Pass`. So a move with one locally-unknowable atom but
one locally-failing atom is still confidently `Fail` on the client — the client only
reports `Unknown` when the *aggregate* can't be decided. This maximizes what the
client can gray out without a round trip.

**The escape hatch — one atom.** `CustomLegal` wraps a Go closure:

```go
func CustomLegal(name string, fn func(state ImmutableState, move ImmutableMove,
	proposer PlayerIndex) error) Cond
```

It evaluates `fn` server-side to `Pass`/`Fail`, and **always serializes as
`Unknown`** to the client (the closure isn't portable). So the checkers capture
graph stays exactly as imperative Go as today — it's just *wrapped* rather than
special-cased, and it composes in the same slice as declarative atoms. The client
sees `legal: "unknown"` for that move and shows it as "may be legal — will confirm on
submit," honestly. Crucially, if a *declarative* atom in the same `All` fails first
(e.g. wrong token color), the client still grays out confidently before ever hitting
the `Unknown`.

---

## 6. Sanitization-aware client story (designed-for)

Each atom knows its referenced paths (§7's index gives us this for free). Before
serializing move forms, the server computes, **per proposer**, whether each
referenced path is `Visible` under that proposer's sanitization policy. An atom whose
inputs are all visible is marked `local: true`; otherwise `local: false`. The client
evaluates the `local: true` atoms itself (using the TS twin of the registry — same
name+args, same resolver semantics) and treats `local: false` atoms as `Unknown`,
composing via the same three-valued `All`/`Any`. Because `Fail` dominates `Unknown`,
the client still grays out most illegal moves locally; it only defers when a *hidden*
input is genuinely decisive.

The server tells the client *which* atoms are local by annotating the serialized
form — no separate protocol:

```json
{"name": "player.compare", "args": ["player.CardsLeftToReveal","gte","1"],
 "local": true}
```

This is the whole reason for name+args-not-closures: the client can re-run the atom
because it's data, not code. (TS evaluator is explicitly a follow-up; this design
*ships the format that makes it a small follow-up*.)

---

## 7. Engine wins

Each atom declares its **referenced phase(s)** and **property paths** (it must, to
resolve them — so it's zero extra author work). At move-registration the engine
builds two indexes per move type:

- `phaseIndex`: phases in which the move's `phase.legal` atom can pass.
- `pathIndex`: the set of property paths any atom reads.

**Fixup loop / move-forms become dirty-tracked.** Today: after every move, every
candidate's full `Legal()` reruns (up to 256×; move-forms run every non-fixup move's
`Legal()` twice per /info). With the index:

1. If the current phase isn't in a move's `phaseIndex`, skip it entirely — **O(1)**,
   no `Cond` evaluation. This alone eliminates most candidates each fixup pass.
2. Track which property paths a state mutation touched (the mutable state layer
   already knows what was written). A move whose `pathIndex` is disjoint from the
   dirty set **keeps its cached `Eval`** — its legality provably didn't change. Only
   moves reading a mutated path re-evaluate.
3. `CustomLegal` atoms have an unknown `pathIndex` (opaque closure), so they
   conservatively always re-evaluate — the escape hatch pays its own cost, nothing
   else does.

So: phase/turn/stack/compare/progression checks become **O(cheap, cached, indexed)**;
only genuinely custom logic stays **O(Legal)**, and only when its (unknown) inputs
might have moved. That directly retires the #640 perf concern for the declarative
majority (~5+8+6+4+3 of the ~28 surveyed checks).

---

## 8. Progression: just another atom

Progression stops being a separate code path. The move-tape matcher becomes the
`InProgression` atom:

```go
func InProgression(group MoveProgressionGroup) Cond
```

Its `Eval` runs the existing `matchTape` (verbatim logic from `default.go:615`), so
zero behavior change — it's a **wrapper, not a rewrite**. `WithLegalMoveProgression`
becomes sugar that appends `InProgression(group)` to the legal slice. The groups
algebra (`Serial`/`Parallel`/`Repeat`) is *untouched*; it's now nested one level
under the precondition algebra, which is the correct relationship — "the tape matches
this pattern" is one precondition among several.

**#644 (state-dependent Repeat counts).** `Repeat(n, ...)` gains a sibling
`RepeatFromState(path, ...)` whose count resolves a property path at eval time via
the same resolver every atom uses. Because progression is now inside the `Cond`
world, it inherits the resolver — the state-dependent count is *the same mechanism*
as `Compare("player.X", ...)`. One resolver, every count and comparison. This is the
composability payoff: a feature requested for progressions falls out of the algebra
for free.

---

## 9. Migration: the three acid tests, verbatim

### 9a. memory/moveRevealCard

```go
// moves.go — the entire Legal() method is DELETED. Config gains:
auto.Config(new(moveRevealCard),
	WithMoveName("Reveal Card"),
	WithLegalPhases(phaseNormalPlay),
	WithLegal(
		Compare("player.CardsLeftToReveal", GTE, Lit(1)).
			Else("You have no cards left to reveal this turn"),
		StackHasComponentAt("game.HiddenCards", "move.CardIndex").
			ElseWhen("game.VisibleCards", HasComponentAt,
				"that card has already been revealed").
			Else("there is no card at that index"),
		MayMoveToSlot("game.HiddenCards", "game.VisibleCards", "move.CardIndex"),
	),
)
```

`player.turn` is inherited from `CurrentPlayer` (no boilerplate). The three original
error strings are preserved exactly. The `nil`-vs-already-revealed branch maps to
`ElseWhen`: if the source slot is empty *but* the visible slot is full, use the
"already revealed" message; else "no card at that index." `MayMoveToSlot` reuses the
component's existing `MayMoveToSlot` under the hood, so its native error passes
through. **Net: ~24 lines of Go → 0 lines of Go, 6 lines of declarative config, all
messages intact, all client-evaluable except the sanitization-dependent slot check.**

### 9b. blackjack/moveStartRoundCleanup

```go
auto.Config(new(moveStartRoundCleanup),
	WithMoveName("Start Round Cleanup"),
	WithPhaseToStart(phaseCleanup, phaseEnum),
	WithLegal(
		// "for every active player, Eliminated OR Stood"
		ForEachActivePlayer(
			Any(
				Compare("player.Eliminated", EQ, Lit(true)),
				Compare("player.Stood", EQ, Lit(true)),
			),
		).Else("not all active players have finished their turn"),
	),
)
```

This needs one new **quantifier atom**, `ForEachActivePlayer(inner Cond)`, which
evaluates `inner` against each non-inactive player state (reusing
`behaviors.PlayerIsInactive`) and is `Pass` iff all pass. It's a bounded, total
quantifier — not a loop the author writes, so it stays out of the tarpit. It earns
its place: the survey shows "all active players satisfy P" is a recurring shape
(blackjack here, likely others). The `StartPhase` chain is inherited; the error
string is preserved verbatim. **Net: ~14 lines of Go → 0, plus one reusable atom.**

### 9c. checkers/moveMoveToken

```go
auto.Config(new(moveMoveToken),
	WithMoveName("Move Token"),
	WithLegal(
		Compare("player.Color", EQ,
			StackValueAtKey("game.Spaces", "move.TokenIndexToMove", "Color")).
			Else("that token isn't your token to move"),
		SpaceIsBlack("move.SpaceIndex").
			Else("you can only move to spaces that are black"),
		// The capture graph stays imperative — wrapped, not special-cased:
		CustomLegal("checkers.legalDestination", legalDestination),
	),
)

// legalDestination is the SURVIVING imperative residue, lifted verbatim from
// the old Legal() body (FreeNextSpaces + LegalCaptureSpaces graph walk):
func legalDestination(state boardgame.ImmutableState, move boardgame.ImmutableMove,
	proposer boardgame.PlayerIndex) error {
	m := move.(*moveMoveToken)
	g := state.ImmutableGameState().(*gameState)
	if err := g.Spaces.MaySwapComponentsByKey(
		m.TokenIndexToMove.Value(), m.SpaceIndex.Value()); err != nil {
		return err
	}
	c := g.Spaces.ImmutableComponentAtKey(m.TokenIndexToMove.Value())
	if c == nil {
		return errors.New("That space does not have a component in it")
	}
	t := c.Values().(*token)
	for _, sp := range t.FreeNextSpaces(state, m.TokenIndexToMove.Value().Int()) {
		if m.SpaceIndex.Value().Int() == sp {
			return nil
		}
	}
	for _, sp := range t.LegalCaptureSpaces(state, m.TokenIndexToMove.Value().Int()) {
		if m.SpaceIndex.Value().Int() == sp {
			return nil
		}
	}
	return errors.New("that is not a legal space to move to")
}
```

`spaceIsBlack` is board geometry, so it's a small **domain atom** the game registers
(`SpaceIsBlack`), not core — exactly as constraints let games register custom
`StackConstraintConstructor`s. The two cheap checks (ownership, black-space) become
declarative and *client-evaluable*, so the client grays out most illegal clicks with
zero round trips; only the genuinely gnarly capture graph stays imperative and
reports `Unknown`. **Net: the two easy checks migrate; the hard one is honestly
labeled — the design's central claim, proven on the hardest case.**

---

## Risks & open questions

1. **`ForEachActivePlayer` and quantifiers risk tarpit creep.** One bounded, total
   quantifier is fine; a general `ForEach(collection, predicate)` invites nesting and
   arbitrary iteration. I'd ship *only* the specific quantifiers the survey demands
   (active-player-forall) and force everything else through `CustomLegal` until a
   second real use case appears. Watch this boundary.
2. **`ElseWhen` (the memory nil-vs-revealed branch) is a mini conditional.** It's the
   thin end of a wedge toward if/else in the DSL. Alternative: express it as
   `Any(sourceHasCard, Not(visibleHasCard).Else("already revealed"))`. Needs a
   usability pass — the branch logic is genuinely awkward and may reveal that
   "structured message selection" wants a first-class (but still bounded) form.
3. **Dirty-path tracking granularity.** The win in §7 assumes the mutable state layer
   can report touched property *paths* cheaply. If it can only report touched
   *stacks/sub-states* coarsely, the cache invalidation is coarser (still a big win
   over today's unconditional re-eval, but less surgical). Needs verification against
   the actual mutation-tracking surface.
4. **Move-field reference safety pre-`DefaultsForState`.** Field-dependent atoms must
   not be evaluated before fields are bound; the `FieldDependent()` split handles it,
   but a mis-declared atom would read a zero field. Consider deriving
   `FieldDependent` automatically from whether any arg path starts with `move.`
   (mechanical, unspoofable) rather than trusting a hand-set flag.
5. **TS twin drift.** Ship-server-first means the Go registry and the future TS
   registry can diverge. Mitigation: the atoms are so small (relations over paths)
   that a shared conformance-test corpus (`{state, move, cond} → expected Eval`) can
   pin both. Not built here, but the format is designed to make it cheap.
6. **`Unknown` UX contract.** "May be legal, confirm on submit" is a new client
   state between legal and illegal. It must not read as an error. This is a client
   design question deferred with the TS evaluator, but the three-valued core is what
   makes it representable at all.
