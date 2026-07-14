# Acid-Test Critique: Declarative Move Legality (Designs A / B / C)

Adversarial pass. Each design walked line-by-line through the three acid-test
migrations against the framework as it actually exists, then checked Q5 (escape
hatch) and Q8 (progression / #644). Findings first, praise second — praise costs
findings.

## Ground truth established before critiquing (verified in-repo)

These facts decide several claims below, so they're stated once here:

1. **Memory sanitization is NOT "hidden cards."** `examples/memory/state.go:16-17`:
   `HiddenCards` is `sizedstack:"cards,40" sanitize:"order"` and `VisibleCards`
   has *no* sanitize tag (fully visible). `sanitize:"order"` hides card *values
   and order*, but **slot occupancy — which index holds a component vs nil — is
   visible.** `CardsLeftToReveal` (playerState) has no sanitize tag → visible.
   Consequence: memory's three checks (`CardsLeftToReveal >= 1`, the
   nil-vs-already-revealed slot disambiguation, and even the `MayMoveToSlot`
   slot-occupancy pre-check) are ALL client-evaluable. Nothing decisive is
   hidden. This directly contradicts the client-story claims in every design.

2. **The prop-path resolver the designs claim to "reuse" resolves ONE component
   instance, not state paths.** `constraints/prop_path.go:19` —
   `resolvePropValue(c ImmutableComponentInstance, propPath)` with only
   `component.`/`dynamic.` prefixes. There is no `game.X`, `player.X`, `move.X`,
   or `players[current].X` resolution anywhere. Every design leans on this
   resolver as if it already spoke those paths. It does not — that resolver is a
   *net-new build*, and the hardest, least-specified part of all three designs.

3. **No write-set / dirty-path tracking exists in core.** `grep` for
   `writeSet|dirtyPath|touchedPath` across the tree: zero hits. Every design's
   §7 "engine win" #3 depends on instrumenting the mutable state layer to record
   touched paths. This is un-built and non-trivial.

4. `errors.Friendly` (`errors/main.go:27`) carries `friendlyMsg` + `secureMsg` +
   `msg` + `fields` — genuinely supports the structured→Friendly adaptation all
   three propose. This claim holds.

5. Acid-test source verified: memory `Legal()` (`moves.go:39-62`), blackjack
   `moveStartRoundCleanup` (`moves.go:39-53`), checkers `moveMoveToken`
   (`moves.go:93-138`). Error strings quoted by all three designs match verbatim.

---

## DESIGN A — Composability (One Precondition Algebra)

### Line-by-line through the acid tests

**9a memory.** The `Cond` slice is plausible and `Compare("player.CardsLeftToReveal",
GTE, Lit(1))` maps cleanly. But the disambiguation is hand-waved. A's answer is
`.ElseWhen("game.VisibleCards", HasComponentAt, "that card has already been
revealed")` (lines 370-372). This `ElseWhen` operator appears **exactly once in
the entire design**, is never defined in §1's atom families, has no `Serialize()`
form, and A's own Risk #2 (line 476) confesses "it's the thin end of a wedge
toward if/else in the DSL... needs a usability pass." So the single gnarliest
line of the single most-cited acid test is powered by an operator the author
admits is unprincipled. The disambiguation *survives as text* but not *as
mechanism*. Given ground-truth fact #1 (both slots visible), the honest move was
to express it as `Any(sourceHasCard, Not(visibleHasCard))` — which A mentions in
the risk section but does NOT use in the migration. The migration shows the worse
version.

**9b blackjack.** Introduces `ForEachActivePlayer` (line 396) — a new bounded
quantifier atom. This is legitimate and matches `behaviors.PlayerIsInactive`
(verified to exist). Error string preserved. Clean.

**9c checkers.** The `CustomLegal` wrapper is correct and the imperative residue
is lifted verbatim — but note A's residue body (lines 434-453) reintroduces the
`MaySwapComponentsByKey` and nil-check that the design claims elsewhere become
declarative atoms. So checkers actually keeps MORE imperative code than the
prose implies: `SpaceIsBlack` and the color compare are pulled out, but the swap
and presence checks are re-embedded in the closure rather than being the
`ComponentPresentAtKey`-style atoms B uses. Minor inconsistency, but it means A's
"two easy checks migrate" undersells how much stays imperative here.

**`StackValueAtKey("game.Spaces", "move.TokenIndexToMove", "Color")`** (line 419)
— this atom must resolve a stack property *by move-field-indexed key* and pull a
named component value. That is well beyond the verified resolver (fact #2) and is
never specified. Hand-wave.

### Q5 (escape hatch): answered. `CustomLegal` as a single atom that always
serializes `Unknown` is the cleanest of the three treatments — it composes in the
same slice, so a declarative atom failing *before* it still grays out the client.
Genuinely good.

### Q8 (progression / #644): answered well. `InProgression` wraps `matchTape`
verbatim (fact: `matchTape` at `default.go:615` is real). `RepeatFromState(path,...)`
for #644 is the strongest #644 answer of the three because it explicitly reuses
"the same resolver every atom uses" — *if* that resolver existed. It doesn't yet
(fact #2), so #644 is answered in principle but rests on the same unbuilt
resolver as everything else. Not deferred, but not free either.

### (1) Best idea worth stealing
**The unification thesis: combinator = atom whose args are other Conds, so
`SerializedCond` is simultaneously the name+args leaf record AND the AST node
(§1, lines 100-104).** One serialization format covers leaves and composites with
zero duplication. Whoever wins should adopt this — B and C both carry an awkward
separate "AnyOf is a special compositor" seam that this dissolves.

### (2) Worst flaw (quoted)
The `ElseWhen` conditional, §9a lines 370-372:
> `StackHasComponentAt("game.HiddenCards", "move.CardIndex").ElseWhen("game.VisibleCards", HasComponentAt, "that card has already been revealed").Else("there is no card at that index")`

An undefined, un-serialized, author-confessed-unprincipled if/else operator doing
the load-bearing work of the headline acid test. The two-stack disambiguation
"survives" only by inventing DSL surface the design elsewhere disowns.

### Scores
Elegance 8 · Explainability 7 · Engine-leverage 7 · Migration-credibility 5

---

## DESIGN B — Engine & Performance lens

### Line-by-line through the acid tests

**9.1 memory.** B is the ONLY design that gives the disambiguation a real,
inspectable implementation. `RevealableCardAt` (lines 483-503) shows actual Go:
checks `hidden.ImmutableComponentAt(idx) != nil` → pass; else checks
`visible.ImmutableComponentAt(idx) == nil` → "there is no card at that index";
else → "that card has already been revealed". This is the memory `Legal()` body's
exact branch structure (verified against `moves.go:53-59`), both strings verbatim,
`Refs()` declared. **This is the one migration across all three designs where the
two-stack disambiguation genuinely survives as mechanism, not prose.** It does so
by NOT trying to express it in a combinator DSL — it's a named built-in with a
hand-written `Eval`. That's the honest engineering answer.

**9.2 blackjack.** `AllActivePlayers(AnyOf(PlayerBool("Eliminated"),
PlayerBool("Stood")))` (lines 526-531). Clean, string preserved, `Refs`
`Players[*]` declared. `AnyOf` as "the *one* allowed compositor" is a defensible
line-drawing.

**9.3 checkers.** Best-structured of the three: `MaySwapByKey`,
`ComponentPresentAtKey`, `ComponentPropEqualsCurrentPlayer`, `SpacePredicate`
each map to a distinct original check with its exact string, and `LegalCustom`
(the interface, lines 302-308) takes ONLY the graph walk. The residue (lines
574-589) is minimal and correct. The `Passed:"unknown", Evaluable:false` client
treatment for the graph walk is exactly right.

**The load-bearing lie B is honest about:** `Refs.StatePaths` uses paths like
`"Players[current].CardsLeftToReveal"` and `"Game.HiddenCards"` (lines 84-88).
These do not resolve today (fact #2). B at least *names* the vocabulary precisely
and flags in Risk #2 (line 612) that `Players[current]` resolution is hard and
falls back to `Players[*]`. It's the most honest of the three about the resolver
gap — but it's still the foundation the whole plan/cache/dirty-track edifice sits
on.

### Q5 (escape hatch): answered, and best-integrated. `CustomLegaler` interface
with `Refs{ReadsUnknown:true}` auto-deriving "runs last, never caches, ships
unknown" (lines 311-318) is the most mechanically complete escape hatch — the
consequences fall out of the metadata rather than being special-cased. Superior
to A's and C's.

### Q8 (progression / #644): answered. `InProgression` with
`Refs{StatePaths:["Game.moveHistory"], ReadsPhase:true}` folds it into the cache
correctly, and B alone notes it subsumes the existing `default.go:475` tape-memo
TODO (verified — that TODO is real, lines 475-478). For #644, `RepeatFromProp`
plus "the group's `Satisfied(tape)` gains access to `PreconditionInput`" (line
447) — this is the one #644 answer that acknowledges the *plumbing* problem
(`Satisfied` currently only sees the tape, not state), instead of assuming a
resolver already reaches it. Most credible #644.

### (1) Best idea worth stealing
**Per-move-type `PreconditionPlan` splitting field-independent from
field-dependent buckets, evaluated in that order (§5, lines 274-296), and the
field-independent bucket memoized per `(moveType, stateVersion, proposer)`
(§7b).** This operationalizes #761's split into an actual cache key and is the
single most concrete performance mechanism proposed by anyone. Steal this
regardless of winner.

### (2) Worst flaw (quoted)
§7c dirty-tracking, lines 388-401 — the entire cache-survives-transitions win
rests on:
> "wrap the mutable state handed to `Apply` so that each `SetXxxProp` / stack
> mutation records its namespaced path into a `writeSet`"

This instrumentation **does not exist** (fact #3), and B's own Risk #1 (line
605-611) calls it "the riskiest load-bearing claim" requiring "an audit of every
mutation path in core" for timers, dynamic values, and behaviors-driven
mutations. Half of B's headline engine wins (the "invalidate only path-overlapping
plans" complexity-table row) are contingent on un-built, admittedly-risky
instrumentation whose conservative fallback ("invalidate everything") erases the
win entirely. B is honest about it, but the whole performance thesis is a promissory
note.

### Scores
Elegance 6 · Explainability 8 · Engine-leverage 9 · Migration-credibility 8

---

## DESIGN C — Delight (DX & Explainability)

### Line-by-line through the acid tests

**(1) memory.** Here is where C hand-waves the headline case worst. The
disambiguation is delegated to a *template name with a function-call embedded in
the string*:
`WithComponentPresent("game.HiddenCards", "@CardIndex", "reveal.no_card_here",
"reveal.already_revealed_if(game.VisibleCards)")` (lines 465-466). The string
`"reveal.already_revealed_if(game.VisibleCards)"` is a template key with an
inline conditional-on-a-stack-path baked into it. This is **less principled than
A's `ElseWhen`** — it's a mini-DSL smuggled into a string literal, with no parser,
no serialization story, and no explanation of how `_if(game.VisibleCards)`
evaluates. The prose (lines 483-485) says the predicate "consults the destination
slot to pick which of the two templates to emit," but the *shown code* encodes
that as an un-parsed string argument. The two-stack disambiguation survives as an
aspiration, not a mechanism. Weakest migration of the headline case.

**(2) blackjack.** `WithEveryActivePlayer(Any(PlayerPropTrue("Eliminated"),
PlayerPropTrue("Stood")), "cleanup.players_unfinished")` (lines 498-505). Fine,
string preserved via template. Comparable to A/B.

**(3) checkers.** Solid. Three checks declarative, `Legal()` kept for the graph
walk, and C's distinctive touch — `Legal()` *still returns a template error*
(`legal.Errorf("checkers.illegal_dest", nil)`, line 564) so even the imperative
residue's rejection is structured/localizable/greppable. That's a genuinely nice
detail the other two miss.

**`@Field` stringly-typed references** (`"@CardIndex"`, `"@TokenIndexToMove"`) —
C's Risk #1 (lines 583-587) honestly flags these aren't caught until runtime and
proposes `NewGameManager`-time validation "against the move's reader + chest
(constraints' `validatePropPath` already proves this pattern works)." But
`validatePropPath` (verified, `prop_path.go:86`) validates against *component
deck props*, not move fields or game/player state — so the cited precedent does
NOT cover the paths C needs. Same resolver-gap as A/B, with a mis-cited precedent.

### Q5 (escape hatch): answered — arguably the *most* faithful to the brief's
"Legal() remains ground-truth." C keeps `Legal()` as a real method the engine
calls after preconditions pass (lines 344-352), and synthesizes a `{"name":
"custom", "verdict": "unknown"}` precondition so the client shows "enabled-but-
tentative." One subtlety C gets right that A/B blur: `Unknown` does NOT
short-circuit (line 340), so the client gets the *fullest* ledger. Good.

### Q8 (progression / #644): answered but thinnest. `Progression` wraps the
matcher (correct, matches `groups.go`), and `Repeat(legal.Prop("game.NumRounds"))`
via an `IntExpr` (lines 438-445). But C never addresses the *plumbing* problem B
caught — that `MoveProgressionGroup.Satisfied` doesn't currently receive state to
resolve `IntExpr` against. C asserts "The IntExpr is evaluated against the
`Context` at progression-match time" without showing how `Context` reaches inside
`Satisfied`. #644 is answered at the syntax layer, deferred at the plumbing layer.

### (1) Best idea worth stealing
**Message templates as `{Template, Bindings}` pairs shipped to the client, with
even the imperative `Legal()` residue returning a template
(`legal.Errorf("checkers.illegal_dest", nil)`, §9.3 line 564).** Structured,
localizable, greppable failures *everywhere* — including the escape hatch — is
the cleanest explainability story and directly retires #65 without exception.
The full-ledger client payload (§4, lines 290-306) with per-precondition
`clientEvaluable` flags is the best-specified client contract of the three.

### (2) Worst flaw (quoted)
§9.1 line 466, the disambiguation encoded as a string-embedded conditional:
> `moves.WithComponentPresent("game.HiddenCards", "@CardIndex",
> "reveal.no_card_here", "reveal.already_revealed_if(game.VisibleCards)")`

A conditional-on-a-stack-path (`_if(game.VisibleCards)`) smuggled into a template-
key string literal, with no parser, no grammar, no serialization. The headline
acid test's hardest branch is "solved" by hiding a second DSL inside a string.
Worse than A's confessed-ugly `ElseWhen` because C doesn't even flag it as a risk.

### Scores
Elegance 7 · Explainability 9 · Engine-leverage 6 · Migration-credibility 5

---

## Cross-cutting verdicts

- **Two-stack disambiguation survives cleanly in exactly ONE design: B**, via a
  named built-in (`RevealableCardAt`) with a hand-written `Eval` mirroring the
  original branch structure. A and C both try to express it in DSL surface
  (`ElseWhen` / `_if()`-in-a-string) and both hand-wave. The lesson: this branch
  wants a purpose-built predicate, not a combinator — B saw that.
- **All three under-sell one gift and over-sell another.** They all miss that
  memory's stacks are `sanitize:"order"` (slot occupancy visible), so the memory
  client story is *better* than any of them claim — the disambiguation is fully
  client-evaluable. And they all over-claim the prop-path resolver as "reuse"
  when `constraints/prop_path.go` resolves a single component instance, not state
  paths — a large net-new build masked as free in every design.
- **The engine wins are a shared promissory note.** Dirty-path tracking (fact #3)
  is un-built in all three. Only B prices it honestly as its riskiest claim; A
  and C wave it through.

---

## Scores summary

- **Design A — Composability:**   Elegance 8 · Explainability 7 · Engine-leverage 7 · Migration-credibility 5
- **Design B — Engine:**          Elegance 6 · Explainability 8 · Engine-leverage 9 · Migration-credibility 8
- **Design C — Delight:**         Elegance 7 · Explainability 9 · Engine-leverage 6 · Migration-credibility 5

## Overall ranking

**B > C > A**: B is the only design whose migrations actually survive line-by-line
scrutiny (real disambiguation code, honest resolver/dirty-track risks, best
escape hatch and #644 plumbing), C wins explainability but hides a second DSL in a
string on the headline case, and A's elegant unification thesis is undercut by an
unprincipled `ElseWhen` doing the load-bearing work of its own showcase migration.
