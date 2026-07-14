# Spec Critique — Contact With the Real Codebase

**Target:** `docs/superpowers/specs/2026-07-10-declarative-legality-design.md`
**Angle:** Verify the spec's structural claims against the actual code, file:line.
**Reviewer stance:** Adversarial. The design is elegant on paper; several of its
load-bearing claims do not survive contact with `moves/` and `examples/`.

Verdict up front: the *representation* (§1) and the *prime-sugar guarantee* (§0)
are sound and honestly costed. The *composition seam* (§2) and the *migration
survey* (§8) are materially wrong — not in flavor but in the counts and the
shape of the embedding graph they assume. Two of the spec's own anti-tarpit
rules (`no first-class not`, `depth-1 compositor`) collide head-on with the
negated / conditional Legal() bodies that already ship in `moves/`.

---

## BLOCKING

### B1. The composition seam names 4 contributors; there are ~24, and the chain isn't linear super-calls

**Spec (§2, §3):**
> "ContributedPreconditions chain on Default/CurrentPlayer/FixUp/StartPhase"
> "Embeddable types chain explicitly — the same pattern as today's Legal()
> super-calls, but written ONCE in the framework"

**Code evidence — the four named types are the wrong four:**

- `FixUp` (moves/fixup.go:13) and `StartPhase` (moves/start_phase.go:27) **do
  not override `Legal()` at all.** They inherit `Default.Legal()` verbatim.
  Listing them as legality contributors is misleading; they contribute *nothing*
  beyond Default. `StartPhase.Legal()` in the acid tests (§8 blackjack:
  "StartPhase contributes its phase/progression preconditions") is really just
  `Default.Legal()` — correct outcome, wrong attribution.
- Meanwhile **~24 move types in `moves/` override `Legal()` and add real,
  independent legality checks** the spec's chain would silently drop. A
  representative slice, each verified:
  - `AnyPlayer.Legal` (moves/any_player.go:36) — target-corrects to proposer,
    `Equivalent(proposer)`, seat-filled check. `AdminPlayer`, `SelectColor/Role/
    Team` all embed this, not `Default`.
  - `FinishTurn.Legal` (moves/finish_turn.go:46) — `PlayerTurnFinisher.TurnDone()`.
  - `SeatPlayer.Legal` (moves/seat_player.go:184), `CloseEmptySeat` (:312),
    `InactivateEmptySeat` (:401), `ActivateEmptySeat` (:498), `CloseAllSeats`
    (:584), `WaitForEnoughPlayers` (:732) — seat-state, admin, ReadyToStart.
  - `MoveOnGraph.Legal` (moves/move_on_graph.go:44) — shortest-path + movement
    budget.
  - `DefaultComponent.Legal` (moves/default_component.go:116) — delegates to the
    *component's* `Legal(state, legalType)` (this is the checkers `moveCrownToken`
    path — legality lives in `checkers/components.go:79`, not in a move at all).
  - `AllPlayersSubmitted` (moves/all_players_submitted.go:46), `AdvanceToken`
    (moves/advance_token.go:37), `ReplenishMarket` (moves/replenish_market.go:77),
    `ShuffleDiscardIntoDraw` (:34), `HopAlongPath` (:32),
    `ResetAllPlayerSubmissions` (:38), `ActivateInactivePlayer`
    (moves/inactive_player.go:46), `Deal*`/`Move*` component families
    (deal_components.go:243, move_components.go:142).

**Fix:** The spec must either (a) implement `ContributedPreconditions()` on the
full set of Legal-overriding intermediates (a much larger surface than "written
ONCE"), or (b) reframe the guarantee: only Default + CurrentPlayer are migrated
to the plan in v1; **every other built-in stays opaque** (falls to today's
`Legal()` via the §0 escape). Option (b) is honest and preserves the sugar
guarantee, but it guts the "phase bucketing / caching for all moves" engine-wins
table (§5), because seat moves, gathering moves, and component moves — the
bulk of a real game's fixup traffic — remain O(Legal). Pick one and re-cost §5.

---

### B2. `ForceFinishTurn` calls NO super — the "always super-call" chain model breaks a security contract

**Spec (§2):** the chain is "the same pattern as today's Legal() super-calls."

**Code (moves/force_finish_turn.go:74-83), verbatim comment above the method:**
> "deliberately do NOT call FinishTurn.Legal — that's exactly the check
> (TurnDone) we want to bypass. We don't need to call Default.Legal either
> because the admin-only gate is stricter"

```go
func (f *ForceFinishTurn) Legal(...) error {
    if proposer != boardgame.AdminPlayerIndex { return errors.New(...) }
    // ... no Default.Legal, no phase check, no progression check ...
}
```

A `ContributedPreconditions()` model that assumes each type appends to its
embedded parent's contributions will **re-introduce** the phase/progression
checks that `ForceFinishTurn` intentionally strips — inverting its behavior.
This move is not a corner case; it's a deliberate admin bypass whose whole point
is to *skip* the inherited plan.

**Fix:** The plan model needs an explicit "replaces-chain" opt-out at the
built-in level, not just the wholesale `Legal()` override at the game-author
level. `ForceFinishTurn` must be able to declare an empty (or admin-only) plan
that does not inherit `Default`'s specs. The spec's `WithoutPrecondition(name)`
(§2) is per-name suppression and does not express "inherit nothing."

---

### B3. `ApplyUntil` / `RoundRobin` are *negated* predicates — collides with the anti-`not` rule

**Spec (§1, anti-tarpit rule 1):**
> "no first-class `not` (a Kleene-`Not` is a TS-conformance liability … `not`
> doubles the surface where they can disagree)."

**Code (moves/apply_until.go:48-66):** `ApplyUntil.Legal()` returns `nil`
(legal) *when `ConditionMet` is NOT met*, and an error *when it IS met*:
```go
if err := conditionMet.ConditionMet(state); err != nil {
    return nil            // condition NOT met ⇒ legal
}
return errors.New("the condition was met, so the move is no longer legal")
```
`RoundRobin.Legal()` (moves/round_robin.go:292-324) is worse: it **branches on
runtime state** (`roundRobinHasStarted(state)`) to decide *whether* to even
consult `ApplyUntil.Legal`, and adds a "last move wasn't me" anti-infinite-loop
guard. Its contribution is *conditional*, not a static precondition list.

These bodies are exactly the "branchy logic" the spec says becomes a
purpose-built named predicate (rule 2) — fine — but they are also *inherently
negated / conditional*, and the whole `ApplyUntil`/`RoundRobin`/`FixUpMulti`
family is the backbone of every deal/collect/round-robin move. Modeling them as
plan predicates means either (a) a first-class `not`/conditional (violating
rule 1), or (b) a purpose-built predicate per family that internally negates
(fine, but then `ConditionMet` is an opaque imperative hook — i.e. these moves
are `LegalCustom`, not declarative). The spec claims the latter category is
"2 genuinely custom" (§8); this family alone is larger than that.

**Fix:** Explicitly classify the `ApplyUntil`/`RoundRobin`/component-move family
as opaque (escape-hatch) in v1 and say so in §8's migration scope. Do not
pretend `ConditionMet` reduces to a serializable relation.

---

### B4. §8 migration survey undercounts "genuinely custom" by 2×–5×

**Spec (§8):** "~5 phase, ~8 current-player, ~6 stack-size/presence, ~4 property
comparisons, ~3 MayMoveTo — all covered; **2 genuinely custom** stay in
`LegalCustom`."

**Code — full survey of the six example games' move `Legal()` bodies.** Hard-
custom (arithmetic / graph search / cross-component-value comparison), each
irreducible to a relation-over-a-path:

1. `memory/moveStartHideCardsTimer` (examples/memory/moves.go:88) — compares two
   revealed cards' `Values().(*cardValue).Type` across slots.
2. `memory/moveCaptureCards` (examples/memory/moves.go:143) — same cross-
   component type comparison, inverted.
3. `blackjack/moveCurrentPlayerHit` (examples/blackjack/moves.go:194) —
   `currentPlayer.HandValue() >= targetScore`: a sum-with-ace-logic aggregation
   over a hand.
4. `checkers/moveMoveToken` (examples/checkers/moves.go:94) —
   `FreeNextSpaces`/`LegalCaptureSpaces` board-graph neighbor traversal +
   capture computation. (This *is* the acid test's `LegalCustom` residue, so
   it's counted — but it's #4, not one of two.)

Plus a **player-quantifier fold** category the catalog does not cover — ∀/∃ over
players with per-player predicates, several phase-conditional:

5. `blackjack/moveStartRoundCleanup` (moves.go:39) — ∀ active: `Eliminated||Stood`
   (this is an acid test in §8; the spec's `AllActivePlayers`+`Any` handles it,
   *if* `AllActivePlayers` is a real quantifier predicate — but that predicate is
   not in the catalog nor a stated primitive).
6. `blackjack/moveAccumulateScores` (moves.go:63) — ∃ active with non-empty hand.
7. `blackjack/moveCollectCards` (moves.go:101) — ∃ player with cards.
8. `blackjack/moveResetPlayerForNewRound` (moves.go:130) — ∃ active: Elim||Stood.
9. `blackjack/moveIncrementRoundsCompleted` (moves.go:162) — ∀: hands empty.
10. `werewolf/moveResolveVotes` (examples/werewolf/moves.go:143) — ∀ eligible
    (active ∧ ¬eliminated ∧ (werewolf-if-night)): voted. Phase-conditional
    eligibility filter — genuinely custom.
11. `werewolf/moveCastVote` (examples/werewolf/moves.go:78) — cross-references
    voter vs. chosen target indices, phase-conditional role check. Borderline.

Charitable count (only arithmetic/graph/cross-value): **4**. Realistic count
(quantifier folds need an escape hatch or new predicates): **10-11**. Either way
the "2" is wrong. Note also that `AllActivePlayers` and `Any` (the blackjack
acid test) are the *only* place a player-quantifier appears in the whole spec,
yet ~6 blackjack fixups need exactly that shape — so the catalog is under-spec'd
even for the games the spec claims to have surveyed.

**Fix:** Revise §8's "2 genuinely custom" to the real number, add a
"player-quantifier predicate" to the catalog (`AllActivePlayers`/`AnyPlayer` fold
with a per-player sub-predicate) as a first-class primitive, and note the
cross-component-value and hand-value cases as escape-hatch residue.

---

## IMPORTANT

### I1. `spaceIsBlack` has no named-function registry — the acid-test call can't resolve

**Spec (§8 checkers):** `legal.SpacePredicate("move.SpaceIndex", "spaceIsBlack")`.

**Code:** `spaceIsBlack` is an **unexported free function** in
`examples/checkers/components.go:52` (`func spaceIsBlack(spaceIndex int) bool`).
Nothing in the spec's registry (§1: `ConfigurePredicateConstructors()`) provides
a way to register a *game-specific named scalar function* and reference it by
string `"spaceIsBlack"` from a `Spec`. The registry resolves *predicate
constructors* (name → Predicate), not arbitrary `func(int) bool` hooks. As
written, `SpacePredicate(..., "spaceIsBlack")` has no resolution path.

**Fix:** Either (a) `spaceIsBlack` becomes a registered predicate constructor in
checkers' `ConfigurePredicateConstructors()` (then it's just a normal purpose-
built predicate, and `SpacePredicate` is unnecessary sugar), or (b) drop
`SpacePredicate` and show checkers registering a `blackSpaceOnly` predicate.
The spec should show the registration, since "reference a game function by
string" is otherwise a new mechanism the §1 registry does not describe.

### I2. `ProposerIsCurrentPlayer` is field-*dependent* and two-sided, contradicting §4 ordering

**Spec (§2):** `CurrentPlayer.ContributedPreconditions()` appends
`legal.ProposerIsCurrentPlayer()` (no args). **Spec (§4):** field-independent
bucket runs "before … field-binding" (DefaultsForState).

**Code (moves/current_player.go:37-64):** the check compares
`c.TargetPlayerIndex` — a **move field** set by `DefaultsForState`
(current_player.go:68) — against *both* `state.CurrentPlayerIndex()` *and*
`proposer`:
```go
if !targetPlayerIndex.Equivalent(currentPlayer) { return "it's not your turn" }
if !targetPlayerIndex.Equivalent(proposer)      { return "it's not your turn" }
```
So (a) it reads `move.TargetPlayerIndex` ⇒ `Reads()` includes a `move.*` path ⇒
it is **field-dependent**, and must run *after* DefaultsForState, not in the
field-independent bucket the name implies; and (b) it's a two-way equivalence,
not the one-liner "proposer is current player" the name suggests. In the fixup
loop (`base.GameDelegate.ProposeFixUpMove`, base/game_delegate.go:98) `Legal` is
called on the shared `state.Game().Moves()` instance **without DefaultsForState**
— so `TargetPlayerIndex` is zero-valued there. Today that's fine because fixups
aren't CurrentPlayer moves; but the plan model must not assume `move.*` fields
are populated at Legal time in every call site.

**Fix:** Mark `proposerIsCurrentPlayer` field-dependent (its `Reads()` includes
`move.TargetPlayerIndex`), and confirm the plan-builder's field-independent /
field-dependent split (§4) puts it after binding. Rename or document that it
enforces target==current ∧ target==proposer (the `Equivalent` wildcard handles
`AdminPlayerIndex`/`AnyPlayerIndex`, which is what makes `LegalForAnyone`'s admin
pass work in server move-forms — see I5; that behavior must be preserved
exactly).

### I3. §2 seam omits `WithUnique`/`WithAllowDuplicates`/`WithRequireAdmin` config-driven legality

**Spec (§2):** "WithLegalPhases / WithLegalMoveProgression / WithSourceProperty
keep working — they are now thin shims that produce these specs."

**Code (moves/with.go:238-265):** three more Legal-affecting config options
exist and the spec never mentions them: `WithUnique` (rejects already-claimed
selections — select_role.go:58 loop), `WithAllowDuplicates`, and
`WithRequireAdmin` (gathering.go:32 `checkRequireAdmin` → `PlayerIsAdmin`). These
feed `SelectRole/Team/Color` and `CloseAllSeats` legality. If the seam only
shims the three named options, uniqueness and admin checks silently fall out of
the plan (or force those moves opaque).

**Fix:** Enumerate *all* config-driven legality shims in §2, or explicitly list
`Select*`/seat moves as opaque in v1. The uniqueness check is itself a player-
quantifier fold (I1/B4), reinforcing the need for that primitive.

### I4. Behavior-preservation of `Default.Legal()` — the three checks translate, but ordering is already fixed

**Spec (§4):** "Buckets are stable-sorted by Cost … Cost-reordering can
legitimately change *which* failure is reported first" (§8).

**Code (moves/default.go:339-388):** today's order is fixed and hand-written:
`legalInPhase` → `legalMoveInProgression` → `legalStackConstraints`. By Cost that
maps to Cheap → Moderate → Cheap. A naive `Cost`-sort would move
`legalStackConstraints` (CostCheap, single stack read) *ahead* of
`legalMoveInProgression` (CostModerate, tape walk) — **changing which error a
game currently reports** when both a stack constraint and a progression are
violated. The spec acknowledges this in the abstract but the golden-equivalence
tests (§9) must pin the *existing* three-check order, not just the migrated
examples. Also note `legalStackConstraints` silently returns `nil` on missing
stacks / empty source (default.go:373-385) — the translated `MayMoveToSlot`
predicate must reproduce the "absent ⇒ pass" degradation, not "absent ⇒ fail."

**Fix:** Assign the three built-in specs Cost values that preserve today's order
(phase < progression < stack), and add golden tests over *existing* games that
trigger two failures simultaneously.

### I5. Server move-forms: the double-pass admin trick relies on `Equivalent`, memo must respect proposer

**Spec (§5):** "Field-independent memo, keyed `(moveType, stateVersion,
proposer)`."

**Code (server/api/main.go:1608-1621):** `generateFormsWithLegality` calls
`move.Legal(state, playerIndex)` (player pass) and `move.Legal(state,
AdminPlayerIndex)` (structural pass). The admin pass passes proposer checks via
`PlayerIndex.Equivalent` (state.go:626: `AdminPlayerIndex` ⇒ always true). The
memo key includes `proposer`, so player-pass and admin-pass are *different* keys
— good, no collision. But note the "field-independent half computed once" claim
(§5) is only true for predicates whose `Reads()` has no proposer dependency; the
`proposerIsCurrentPlayer` predicate (I2) *is* proposer-dependent, so it cannot be
shared across the two passes. The memo saves the *phase/progression/stack* half,
not the proposer half — which is the accurate but smaller win. §5's phrasing
("computes the stable half once") is right only if "stable" excludes proposer-
dependent predicates; make that explicit.

**Fix:** State that proposer-dependent predicates (`proposerIsCurrentPlayer`,
admin checks) are *not* memoized across the player/admin double pass; only the
field-independent ∧ proposer-independent subset is.

### I6. Phase index: moves with NO declared phases — bucket membership unspecified

**Spec (§5):** "phaseIndex map[phase][]moveType built from each plan's `inPhase`
spec (∪ TreeEnum ancestors) … Fixup loop and move-forms iterate
`phaseIndex[currentPhase]` instead of all moves."

**Code (moves/default.go:429-433):** a zero-length `legalPhases` means "legal in
ALL phases":
```go
if len(legalPhases) == 0 { return nil } // legal everywhere
```
Many moves declare no phases (memory's moves, tictactoe, seat moves, most
fixups). The spec's `phaseIndex[currentPhase]` lookup would **exclude** these
"legal everywhere" moves unless they are inserted into *every* bucket — which the
spec never says. If the index only contains moves with an explicit `inPhase`
spec, the fixup loop and move-forms would stop considering all phase-agnostic
moves — a correctness regression (e.g. `SeatPlayer`, `WaitForEnoughPlayers`
run outside any phase).

**Fix:** Specify a "phase-agnostic" bucket that is unioned into every
`phaseIndex[phase]` lookup (or a wildcard entry iterated alongside the current
phase). This is essential and currently unhandled.

---

## NITS

### N1. `ImmutableComponentAt` / `ImmutableStackProp` APIs check out
The acid-test builder Eval (`revealableCardAt`, §8) uses
`ImmutableComponentAt(idx)` (stack.go:47, returns
`ImmutableComponentInstance`, nil for empty slot — exactly the branch the memory
code relies on, examples/memory/moves.go:53-59) and `ImmutableStackProp`
(property_reader.go:39). Both exist and behave as the code assumes. The memory
acid test is faithful — except it should use `MayMoveToSlot` (what the real code
calls at moves.go:61), not bare `MayMoveTo`; the spec's `MayMoveToSlot(...)`
builder name is correct, good.

### N2. `behaviors.PlayerIsInactive` exists as claimed
`behaviors/inactive_player.go:43` — `func PlayerIsInactive(playerState
boardgame.ImmutableSubState) bool`. The blackjack acid test's `AllActivePlayers`
reusing it is feasible.

### N3. Path resolver honesty is correct
§1's claim that `constraints/prop_path.go` is single-component-instance-only and
the state path resolver is a *net-new build* is accurate:
`constraints/prop_path.go:18` `resolvePropValue(c ImmutableComponentInstance, …)`
resolves `component.`/`dynamic.` prefixes against one component, with no
`game.`/`player.`/`players[*].`/`move.` grammar. This is the spec's most honest
section — keep it.

### N4. `errors.Errorf` / `legal.Errorf` is new
`errors/main.go` has `New`/`NewFriendly`/`NewSecure`/`Extend`, no `Errorf`.
`legal.Errorf` (§4, §6) is net-new in the new package — fine, but the spec
implies it "adapts to the existing errors.Friendly" (§1); confirm the adapter
maps `Message{Template,Bindings}` → `errors.Friendly` fields (errors/main.go:112
`Fields()`), which is plausible but unspecified.

### N5. Latent bug spotted (not spec's fault, worth flagging)
`examples/pig/moves.go:35` — `moveRollDice.Legal` does `return nil` where it
means `return err` on base-Legal failure. Migrating pig to declarations would
*fix* this silently (the plan would enforce the base checks), which is a
behavior change the golden-equivalence test (§9) would flag as a "regression"
when it's really a fix. Note it in the migration.

---

## VERDICT

**Revise before implementation.** The representation, layering, and honest
path-resolver costing are strong and can survive. But the composition seam (§2)
and migration survey (§8) rest on an incorrect model of the `moves/` embedding
graph: FixUp/StartPhase contribute nothing, ~24 other types contribute real
legality logic, three core families (`ApplyUntil`/`RoundRobin`/component-moves)
are inherently negated/conditional and collide with the anti-`not` rule, one move
(`ForceFinishTurn`) deliberately inherits nothing, and the "2 genuinely custom"
figure is off by 2×–5×. The engine-wins table (§5) is priced against "all moves
join the plan," which the real embedding graph won't support in v1.

### Top 3 changes

1. **Reframe §2/§0 to "Default + CurrentPlayer migrate in v1; everything else
   opaque,"** and re-cost §5's engine wins against that reality (seat/gathering/
   component/round-robin moves stay O(Legal)). Add an explicit "inherit-nothing"
   opt-out for `ForceFinishTurn`-style bypasses — `WithoutPrecondition(name)` is
   insufficient (B1, B2).

2. **Fix the §8 survey: real "genuinely custom" is 4 (hard) to ~11 (incl.
   quantifier folds), not 2.** Add a first-class **player-quantifier predicate**
   (`AllActivePlayers`/`AnyPlayer` with a per-player sub-predicate) to the
   catalog — ~6 blackjack fixups, `WithUnique`, and the blackjack acid test all
   need it. Classify `ApplyUntil`/`RoundRobin`/`ConditionMet` and the cross-
   component-value / hand-value cases as escape-hatch residue (B3, B4, I3).

3. **Specify the phase-agnostic bucket and the field-dependent classification of
   proposer checks.** `len(legalPhases)==0` moves must be unioned into every
   `phaseIndex[phase]` lookup or they vanish from candidate iteration (I6);
   `proposerIsCurrentPlayer` reads `move.TargetPlayerIndex` and is proposer-
   dependent, so it's field-dependent and un-memoizable across the double pass —
   fix its bucket and preserve today's phase<progression<stack error ordering via
   Cost assignment (I2, I4, I5).
