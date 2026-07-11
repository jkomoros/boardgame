# Adversarial Critique: The Purely-Sugar Guarantee

**Spec:** `docs/superpowers/specs/2026-07-10-declarative-legality-design.md`
**Branch:** `declarative-legality-design` (design-only; no `legal/` package, no engine
code exists yet — every finding below is against the design text, ground-truthed against
the current engine).
**Angle:** The prime guarantee (spec §"The prime guarantee: purely sugar", lines 37–52):
*"A game ignoring [this system] behaves exactly as today... The engine treats such a move
as opaque and behaves exactly as it does now."* This critique hunts every place the design
silently breaks that.

**Verdict up front:** The guarantee holds for the *narrow* case the spec tests (a move with
a wholesale `Legal()` override and no embedding). It breaks in three ways the spec does not
address: (1) **opaque moves have no home in the phase index** and the spec never says where
they go; (2) the **embed-and-override pattern used by every real game today** produces
double-evaluation and a semantic re-ordering, and this is the *common* case, not the escape
hatch; (3) `ConfigurePredicateConstructors` is added to the `GameDelegate` **interface**,
which is a hard compile break for any delegate not embedding `base.GameDelegate`. Details
and fixes below.

---

## BLOCKING

### B1. Opaque moves have no bucket in the phase index — the spec never says where they go

**Spec claim** (§5 table, line 374):
> **Phase bucketing** — `phaseIndex map[phase][]moveType` built from each plan's `inPhase`
> spec (∪ TreeEnum ancestors) — Fixup loop and move-forms iterate `phaseIndex[currentPhase]`
> instead of all moves.

**Spec claim** (§4, `PreconditionPlan.opaque bool`, line 333) and (lines 42–43): an opaque
move "behaves exactly as it does now."

**The break.** Today, candidate iteration is *unconditional over every move*:

- `base/game_delegate.go:86` — `for _, move := range state.Game().Moves()` then
  `move.Legal(state, AdminPlayerIndex)`. There is **no phase pre-filter**; the phase check is
  *inside* `Legal()`. Every fixup move is polled every recursion.
- `defaultGameDelegate.ProposeFixUpMove` (`game_delegate_test.go:90`) — identical loop.
- `server/api/main.go:1597` `generateFormsWithLegality` — `for _, move := range game.Moves()`,
  again unconditional.

The optimization replaces "iterate all moves" with "iterate `phaseIndex[currentPhase]`." A
move's bucket is derived **from its `inPhase` spec** (line 374). But an opaque move — one that
overrides `Legal()` wholesale, or a hand-rolled move that implements no `PreconditionsProvider`
— **has no plan, therefore no `inPhase` spec, therefore no basis for bucketing.** If such a
move is silently omitted from `phaseIndex[p]`, the fixup loop and move-forms **stop
considering it entirely** — it can never fire, never appear in a form. That is not "behaves
exactly as today"; that is the move vanishing.

The spec's own `opaque bool` field proves the plan-builder *knows* a plan is opaque, yet §5
never states the corresponding rule: **an opaque move must be placed in every phase bucket**
(equivalently, kept in an always-considered fallback list that runs alongside the index). The
one line that would preserve the guarantee is absent.

Worse, a *migrated* move that legitimately declares no `inPhase` (memory has no phase enum —
`grep WithLegalPhases examples/memory/*.go` is empty; today `legalInPhase` returns nil at
`default.go:429` for zero-length `legalPhases`) is in the **same** situation as an opaque one:
zero-length legal-phases means "legal in all phases," which must map to "present in every
bucket," not "present in no bucket." The spec treats `inPhase` as always-present ("built from
each plan's `inPhase` spec") and never handles the empty/absent case.

**Evidence the bucketing assumes more than opaque moves can provide:**
- Determinism claim (line 340, "a given state always reports the same failure"): if opaque
  moves are appended to buckets in a non-stable order, or omitted, the "same failure" invariant
  breaks for the mixed opaque/declarative case.
- `noProgressionMoveTypesByGame` (`default.go:492–518`) already caches "which moves have no
  progression" per game and *keeps* them in the tape as always-legal. The phase index needs the
  analogous "opaque / all-phases moves are in every bucket" rule, and the spec does not port it.

**Fix.** State normatively in §5: *"A move whose plan is `opaque`, or whose `inPhase` spec is
absent/zero-length, is inserted into **every** phase bucket (and into a no-phase-enum game's
single implicit bucket). The phase index is an optimization over the current unconditional loop
and MUST NOT be able to hide a move the current loop would consider."* Add a plan test (the spec's
§9 "opaque fallback" test currently only asserts fallback *evaluation*, not fallback
*bucket membership* — extend it).

---

### B2. The embed-and-override pattern (every real game) double-evaluates and re-orders

**Spec claim** (lines 44–52): declaring *is* implementing; "An author never writes both a
declaration and the code enforcing it." And (§4, lines 293–294, 348–349): "`moves.Default.Legal()`
becomes exactly: evaluate plan; return first failure's `Verdict.Error()`."

**The break.** The spec's opt-out test is a move that *overrides `Legal()` wholesale and embeds
nothing that contributes* (line 52). But **no real game does that.** The universal pattern is
*embed `moves.CurrentPlayer`/`FixUp` **and** override `Legal()`, calling the embedded `Legal()`
super*:

- `examples/memory/moves.go:41` — `m.CurrentPlayer.Legal(state, proposer)` then imperative body.
- `examples/checkers/moves.go:95` — `m.CurrentPlayer.Legal(...)` then capture-graph search.
- `../games/murdermrmonroe/moves.go:69` — `m.CurrentPlayer.Legal(...)` then `CanDrawCard` check.
- `../games/valentine`, `metaltrader`, `pass`, `darwin` — same (`grep -c "CurrentPlayer.Legal\|FixUp.Legal"`
  across `../games` returns non-zero for 8 files).

Trace an **un-migrated** such move after the change. `moves.CurrentPlayer.Legal` (`current_player.go:37`)
still runs `c.Default.Legal(state, proposer)` at its top, then its own turn-check body
(`current_player.go:43–63`). After the change `Default.Legal` becomes "evaluate my plan"
(lines 293, 348). The plan for a `CurrentPlayer`-embedding move, per §2's `ContributedPreconditions`
chain (lines 251–262), is:

```
Default.ContributedPreconditions()  → [inPhase, inProgression, stackConstraints]
CurrentPlayer.ContributedPreconditions() → append(Default…, proposerIsCurrentPlayer)
```

So `Default.Legal` for a `CurrentPlayer` move now evaluates **`proposerIsCurrentPlayer`** (a turn
check) *because the plan for that move type includes it*. Then control returns to
`CurrentPlayer.Legal`, which runs the **imperative turn check again** (`current_player.go:43–63`).
**The proposer/turn check runs twice.** This is exactly the "author writes both the declaration
and the code enforcing it" that the guarantee (line 46) promises never happens — except here the
framework wrote the duplicate, silently, at the moment `Default.Legal`'s meaning changed.

Two concrete hazards:

1. **Ordering / message change.** Today `CurrentPlayer.Legal` runs `Default.Legal` (phase →
   progression → stack) *first*, then the turn check. If `Default.Legal` now evaluates the plan and
   the plan is **Cost-sorted** (line 339, "stable-sorted by Cost, Trivial → Expensive"),
   `proposerIsCurrentPlayer` is `CostTrivial` (line 128) and sorts *ahead of* `inPhase`
   (`CostCheap`) and `inProgression` (`CostModerate`, line 435). So an un-migrated move that
   today reports "Move is not legal in phase X" (phase checked before turn) may now report "it's
   not your turn" first — **a different first-failure for a move the spec calls opaque/unchanged.**
   The spec explicitly fences message change for *migrated* moves (lines 533–534, "Cost-reordering
   can legitimately change which failure is reported first") but the sugar guarantee says
   un-migrated behavior is *identical*. It is not, once `Default.Legal` internally reorders.

2. **Which body is authoritative?** `Default.Legal` under the new design must decide: is *this
   move type's* plan the plan for `moveMoveToken` (with author preconditions the un-migrated game
   never declared), or just `Default`'s own contributions? The spec's plan is built **per move
   type** (§4 line 326, "Per move type, built once"), keyed on the top-level struct. For an
   un-migrated `moveMoveToken` the per-type plan = ContributedPreconditions only (no
   `WithPreconditions`). But `moveMoveToken` *also* overrides `Legal()` wholesale (line 52's opt-out
   trigger) — so is it `opaque=true`? The spec's opt-out rule (line 52) says a wholesale `Legal()`
   override "opts out of the plan entirely." Yet the move's `Legal()` **calls
   `m.CurrentPlayer.Legal` which now evaluates the plan.** So the move is simultaneously "opaque
   (opted out)" at the engine's introspection layer **and** "evaluates the plan" through its own
   super-call. The design has no answer for a move that is opaque to the engine but whose
   hand-written body re-enters the plan evaluator. This is undefined behavior in the spec as
   written.

**Fix.** The spec must (a) define what `Default.Legal` evaluates *when reached via an imperative
super-call from an un-migrated override* — almost certainly it must evaluate **only the
ContributedPreconditions of the embedding chain, in today's fixed order (phase → progression →
stack → proposer), NOT Cost-sorted**, whenever the move type has no author `WithPreconditions`.
Cost-sorting may only kick in once a move is fully migrated (author declares preconditions and
deletes the override). (b) Make `CurrentPlayer.Legal`/`FixUp.Legal` *idempotent* super-calls: if
`Default.Legal` already evaluated `proposerIsCurrentPlayer` via the plan, `CurrentPlayer.Legal`
must not re-run it. The clean resolution is: **`CurrentPlayer` keeps its imperative `Legal` body
for the un-migrated compatibility path and stops delegating the proposer check into the plan** —
i.e. the ContributedPreconditions chain is *only* consulted when the engine introspects
`PreconditionsProvider`, and the imperative `X.Legal()` chain stays byte-for-byte what it is today.
Either way, §4 must add a "double-evaluation & ordering" subsection; right now it is silent and the
guarantee is violated for the majority of existing moves.

---

### B3. `ConfigurePredicateConstructors` on the `GameDelegate` interface is a compile break

**Spec claim** (line 167): *"Registry wiring mirrors constraints exactly:
`GameDelegate.ConfigurePredicateConstructors() []*legal.PredicateConstructor`, with
`base.GameDelegate` returning `legal.DefaultConstructors()`."*

**The break.** "Mirrors constraints exactly" is precisely the problem. `GameDelegate` is a Go
**interface** (`game_delegate.go`), and `ConfigureStackConstraintConstructors()` is a **member of
that interface** (`game_delegate.go:364`). Adding `ConfigurePredicateConstructors()` to the
interface the same way makes **every type that satisfies `GameDelegate` but does not embed
`base.GameDelegate` fail to compile.** The base-delegate default (line 168) only rescues types
that *embed* `base.GameDelegate`; it does nothing for a hand-rolled delegate.

Is that population empty? Not guaranteed — and the spec asserts additivity as if it were.
`base.GameDelegate` is the conventional embed, but the framework contract is the interface, and the
in-repo test suite already defines delegates directly against it (`game_delegate_test.go`,
`defaultGameDelegate`/`testGameDelegate` implement interface methods explicitly). Any external
game, or the framework's own test delegates, that satisfies `GameDelegate` structurally rather
than by embedding will break. This directly contradicts the sugar guarantee: a game that ignores
the entire legality system must still compile, and here it does not.

Note this is a *stronger* break than the `ConfigureStackConstraintConstructors` precedent, because
that method has existed since before the freeze; adding a **new** interface method is a source-
breaking change to a published interface, whereas the spec's own framing (lines 42–43, 60–61)
permits breaking the Go API only "in the additive-sugar sense... imperative Legal() keeps working
everywhere."

**Fix.** Do **not** add `ConfigurePredicateConstructors` to the `GameDelegate` interface. Make the
engine consume it via an **optional interface + type assertion** (the same shape as
`PreconditionsProvider`, line 316):

```go
type PredicateConstructorConfigurer interface {
    ConfigurePredicateConstructors() []*legal.PredicateConstructor
}
// at NewGameManager:
if c, ok := delegate.(PredicateConstructorConfigurer); ok { ... } else { use legal.DefaultConstructors() }
```

`base.GameDelegate` still implements it for the ergonomic default, but a hand-rolled delegate that
never heard of `legal` still compiles and gets the default catalog. The spec must call this out;
"mirrors constraints exactly" is the wrong precedent to copy for a net-new method.

---

## IMPORTANT

### I1. `CustomLegaler` vs wholesale `Legal()` override: two escape hatches, undefined interaction

**Spec** defines two opt-outs: `LegalCustom` (§4 line 355, "runs after all declarative
preconditions pass") and wholesale `Legal()` override (line 52, "opts out of the plan entirely...
falls back to today's behavior"). The spec never states what happens when a move **implements
both** `LegalCustom` **and** overrides `Legal()`. Given B2 shows the real-world pattern is "embed +
override `Legal()`," a half-migrated author will plausibly add a `LegalCustom` while an
override still exists (or an embedded type provides one). Which wins? If `Legal()` is overridden,
the engine's plan (and thus `LegalCustom`, which the spec says the engine wraps and runs last,
line 361–366) is **never reached** — so `LegalCustom` silently does nothing. That is a
foot-gun with no diagnostic. The teachability question in the brief ("Is the distinction
teachable?") answers itself: **no**, because the two hatches live at different layers (one is an
engine-introspected interface, one is a Go method override) and the spec gives no rule for their
coexistence.

**Fix.** Add a normative rule + boot-time validation: *if a move type both overrides `Legal()`
wholesale (detected: its `Legal` method is not `moves.Default.Legal`) and implements
`CustomLegaler`, `NewGameManager` fails fast with "move X overrides Legal() and also implements
LegalCustom; LegalCustom will never run — remove one."* Reflection can detect the override by
comparing method values, or by a marker. Document that `LegalCustom` is *only* meaningful for moves
that let `moves.Default.Legal` drive.

### I2. Client wire compat — `Preconditions` field is additive-safe, but `evaluable`/verdict enums are undocumented for existing consumers

**Spec** (§6, lines 400–409) adds a `Preconditions` array to move forms "alongside the preserved
`LegalForPlayer`/`LegalForAnyone`/`LegalForPlayerError`." Ground truth: the wire struct is
`moveForm` (`server/api/main.go:78`), TS mirror is `MoveForm` (`api.d.ts:73`). Both use
`json:",omitempty"` and an open TS interface — an added `Preconditions?:` field is tolerated by
existing consumers (unknown fields are ignored in JS; TS interface is structural). So the
*additive* claim holds. **But:**

- The spec preserves `LegalForPlayerError` (the pre-baked string, `main.go:1614`) *and* adds a
  structured `message` per predicate. For an **un-migrated** move, `LegalForPlayerError` today is
  the first-failure string from the imperative chain (`move.Legal(...).Error()`). If the server
  now derives `LegalForPlayerError` from *the plan's* first Cost-sorted failure instead of the
  chain's first failure, the string a legacy client displays changes — see B2(1). The spec must
  state that `LegalForPlayerError` for opaque/un-migrated moves is still sourced from
  `move.Legal()` verbatim, not synthesized from a plan. §6 does not say this.
- `verdict: "pass"|"fail"|"unknown"` and `evaluable: bool` are new enums with no `api.d.ts` entry;
  the spec should ship the type update in the same change so the two clients agree (the spec's §9
  ledger test asserts shape but not the TS type). Nit-adjacent, folded here.

**Fix.** §6: pin `LegalForPlayer`/`LegalForAnyone`/`LegalForPlayerError` semantics as
**byte-identical to today for every move (opaque or migrated) — sourced from `move.Legal()`**, and
declare `Preconditions` purely additive. Only migrated moves' *`Preconditions` array* may reflect
Cost-ordering; the three legacy fields must not, or un-migrated clients see behavior drift.

### I3. `DefaultsForState` timing vs the field-independent bucket — the server admin pass changes what gets evaluated when

**Spec** (§4, line 340): field-independent bucket runs "before `DefaultsForState`/field-binding,"
field-dependent after. Ground truth: today `DefaultsForState` is called inside
`NewMove`/`NewMoveType.NewMove` (`move.go:405`) — the move is *fully defaulted before* `Legal()`
is ever called, in both `ProposeFixUpMove` and `generateFormsWithLegality`. In today's server path
(`main.go:1613`, `1621`) the form move already has `DefaultsForState` applied, then `Legal` runs
against a bound move with real `move.*` fields.

The spec's split means the **field-independent bucket is now evaluated against `Move: nil`**
(§1, `Context.Move` "nil during field-independent evaluation," line 106) — i.e. *before* defaults
exist. For the sugar guarantee: a move whose *imperative* `Legal` reads a `move.*` field that
`DefaultsForState` populates (e.g. memory's `moveRevealCard.DefaultsForState` at
`examples/memory/moves.go:25` sets `CardIndex` to the first hidden card, and `Legal` reads
`m.CardIndex` at line 53) must NOT have that read hoisted into the field-independent bucket. For an
**opaque** move this is fine (whole `Legal` runs once, post-defaults, unchanged). But the spec's
`Reads()`-based split (line 116) is per-*predicate*; if a migrated predicate mis-declares a
`move.*` read as field-independent (the "conservative over-approximation" is a *lower* bound risk —
`Reads()` must be complete, line 378), it evaluates against `nil` Move and either panics or reads a
zero field. The spec flags `Reads()` conservativeness as a risk (line 558) but does **not** connect
it to the `Move: nil` evaluation window — a predicate that under-declares a `move.*` read is not
just a stale-cache bug, it's a **nil-deref in the field-independent pass.** For the admin/anyone
pass specifically (`main.go:1621`), nothing about *when* `DefaultsForState` runs changes for opaque
moves, so the guarantee holds there — but the spec should say so explicitly rather than leave it
inferred.

**Fix.** §4: state that the field-independent bucket evaluates with `Context.Move == nil` and that
**any predicate reading a `move.*` path is by definition field-dependent** (mechanically enforced,
not by-convention) — a `move.*` in `Reads()` forces the predicate into the field-dependent bucket
regardless of Cost. Add a boot-time check. Explicitly note opaque moves are unaffected (whole
`Legal` still runs once, post-`DefaultsForState`).

---

## NIT

### N1. "One code path for every declarative move" hides the progression-tape self-append

`legalMoveInProgression` (`default.go:574–577`) appends the *proposed* move to the tape before
matching ("Add ourselves to the end of the tape, since we're proposing adding ourselves"). The
`inProgression` predicate (§7, line 433) must replicate this self-append, and it can only do so
when it has the move's *name* — which is move-type metadata, available field-independently. Fine,
but the spec's §7 says `inProgression` has `Reads: [game.moveHistory]` and `CostModerate` and
implies it's field-independent (no `move.*`). The self-append of the current move-type name is a
subtle input the spec glosses; worth a sentence so the implementer doesn't forget the tape includes
the candidate.

### N2. Determinism claim vs opaque-move ordering

Line 340 promises "a given state always reports the same failure (no message flapping)" from stable
Cost-sort. Opaque moves and `LegalCustom` are `CostExpensive` and sort last — deterministic among
themselves only if the plan preserves a stable secondary order (declaration order). The spec says
"custom always last" (line 339) but for **multiple** opaque contributions (an embedded chain each
overriding `Legal`? rare but possible) the tiebreak is unspecified. Low severity; note declaration-
order tiebreak in §4.

### N3. Spec should assert the phase-index equivalence as a test, not just prose

§9 lists "phase-index correctness incl. TreeEnum ancestors" but not the **superset property** that
resolves B1: *for every state, `phaseIndex[currentPhase]` ⊇ the set of moves the current
unconditional loop would consider legal-candidates.* Make that the headline plan test; it is the
mechanical statement of the sugar guarantee for bucketing.

---

## Verdict

**The purely-sugar guarantee is not yet met.** The spec proves it for the one case it tests (a
lone move with a wholesale `Legal()` override and no contributing embed), but that case does not
exist in any real game. Every migrated-in-repo and `../games` move uses **embed + override-with-
super-call**, and for those the design silently (a) double-evaluates the proposer/phase check, (b)
re-orders first-failure reporting via Cost-sort inside `Default.Legal`, and (c) leaves "opaque to
the engine but re-enters the plan through its own super-call" undefined. Independently, opaque
moves have no defined phase-index bucket (they can vanish from the fixup loop and forms), and
`ConfigurePredicateConstructors` on the `GameDelegate` **interface** is a compile break for any
non-`base`-embedding delegate. Each is fixable and none is fatal to the architecture — but the
guarantee as written (lines 42–43, "behaves exactly as it does now") is false until they are.

### Top 3 changes

1. **Freeze the imperative chain; don't route un-migrated `Default.Legal` through the Cost-sorted
   plan.** The `X.Legal()` super-call path (`CurrentPlayer.Legal` → `Default.Legal`) must stay
   byte-for-byte today's phase→progression→stack→proposer order with no plan involvement until a
   move is *fully* migrated (author declares `WithPreconditions` and deletes the override). Add a
   §4 "double-evaluation & ordering" subsection. (Fixes B2, I2's ordering leak.)

2. **Normatively place opaque / no-`inPhase` moves in every phase bucket, and test the superset
   property.** State in §5 + §9 that the phase index can never hide a move the current
   unconditional loop would consider. (Fixes B1, N3.)

3. **Make `ConfigurePredicateConstructors` an optional interface consumed by type-assertion, not a
   `GameDelegate` interface member** — plus a fail-fast rule for a move implementing both
   `LegalCustom` and a wholesale `Legal()` override. (Fixes B3, I1.)
