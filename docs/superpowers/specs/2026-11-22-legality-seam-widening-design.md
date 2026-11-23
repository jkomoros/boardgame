# Declarative Legality — Seam Widening — Design

**Date:** 2026-11-22
**Base:** `2026-07-10-declarative-legality-design.md` (v1, normative for
everything not restated here) and `2026-07-12-legality-completeness-design.md`.
**Scope:** limit (1) of the remaining `legal/doc.go` limits: only an allowlist
of eight framework base types could opt in to declarative legality.

## Provenance

The first draft was critiqued adversarially against the live code before
implementation (verdict: sound with amendments). The critique is checked in at
`docs/superpowers/design/2026-11-22-legality-seam-widening/critique.md`. This
spec is the amended design. Each finding's disposition is in the
"Critique dispositions" section.

## Related history

- **#761** (closed, "Allow declarative legal methods") asked for moves that opt
  in to being declarative by "nesting the sub-types' legal statements within
  their own". That is the contribution chain. This work finishes it for
  every framework base type, not just eight.
- **#189 / #213** (client-side rejection of obviously illegal moves) are
  advanced. Every base type now reports named, ordered atoms in the ledger,
  even where the atom stays server-only. The design also names which atoms
  are worth making client-evaluable next (see "Follow-ups").
- **#65** (log FixUp rejections) is advanced. A fix-up rejection on an
  opted-in framework move now names the framework atom that failed
  (`finishTurn`, `applyUntil`, ...), not just a string.
- **#64** (closed, protection against ProposeFixUp never ceasing) is the
  existing safety net. This design does not lean on it, because framework
  atoms cannot be deleted (below).
- **#624** (Deal/Move/CollectComponents should reject when nothing is left to
  move) is untouched in behavior. The checks that already exist
  (`dealCountComponents`, `moveCountComponents`) become named atoms.
- **#473 / #640** (parallel legal / progression performance) are unaffected.
  Framework atoms are never memoized, which matches the frozen chain's cost.

## Problem

A game move opts in with `moves.WithLegalPreconditions(...)` (or by
implementing `LegalCustom`). At `NewGameManager` the plan is assembled as

    base contributions (ContributedPreconditions, base-first)
    + authored specs
    − WithoutLegalPrecondition suppressions
    + LegalCustom residue (last)

`moves.Default.Legal()` evaluates that plan *instead of* its frozen imperative
chain. That is only honest if the move's whole legality is in the plan. A
framework type with its own `Legal()` override (`FinishTurn`, `RoundRobin`,
`AnyPlayer`, ...) runs extra imperative checks after its super-call. Those
checks are invisible to the plan and to the ledger.

So the seam was an allowlist (`legalSupportedMovesBaseTypes`). Embedding
anything else while declaring was a boot error. Game authors had to learn the
allowlist ("which bases are on the seam?"). It was restated in three error
messages and two docs, and needed its own doc-drift test. It also blocked the
most common real bases: `FinishTurn` (every example game),
`DealCountComponents` (blackjack, metaltrader), `RoundRobin` (valentine,
pass), `AnyPlayer` (werewolf, darwin), `MoveOnGraph`/`AdvanceToken`
(murdermrmonroe), the seat-management moves, and the `ApplyUntil` family.

## Survey: every framework type with real legality logic

Legend:
- **C**: expressible today as contributed catalog specs.
- **G**: expressible with a small, principled catalog addition or a
  re-explanation of the type.
- **O**: inherently opaque. It calls game-supplied Go, or computes something
  rule 3 forbids in serialized form.

"Form?" says whether the type is a player-facing move (rendered as a form,
where client evaluability pays off) or a fix-up.

| Type | Own check beyond its super-call | Class | Form? |
|---|---|---|---|
| `Default`, `CurrentPlayer`, `DrawToPlayer`, `MoveComponentToSlot` | already contribute | C | yes |
| `FixUp`, `FixUpMulti`, `StartPhase`, `RecordCurrentPlayerChoice` | none | C | — |
| `FinishTurn` | current player valid; game hook `TurnDone()` nil | O | fix-up |
| `ForceFinishTurn` | proposer is Admin; current player valid; **super-calls nothing** | G | admin |
| `ApplyUntil` (+`ApplyUntilCount`, `ApplyCountTimes`, which only override `ConditionMet`) | game hook `ConditionMet()` **non-nil** (negated) | O | fix-up |
| `RoundRobin` (+`RoundRobinNumRounds`) | not started ⇒ last move ≠ me; started ⇒ ApplyUntil's | O (conditional) | fix-up |
| `DealCountComponents` (+3 subclasses) | next round-robin player's stack accepts 1 | O | fix-up |
| `MoveCountComponents` (+3 subclasses) | ApplyUntil **without** Default's stackConstraints; remaining transfer feasible | O | fix-up |
| `AnyPlayer` | target defaults to proposer (**mutates the move**); in range; proposer≡target; seat filled | G, blocked by the mutation | yes |
| `AdminPlayer` | if the player has GameAdministrator, must be admin | G | yes |
| `SelectColor` / `SelectRole` / `SelectTeam` | value non-nil and valid; optionally unique among seated players | O (quantifier with move-field compare) | yes |
| `AllPlayersSubmitted` | every active player `HasSubmitted()` | G (message names the index) | fix-up |
| `ResetAllPlayerSubmissions` | some player has submitted (∃) | G | fix-up |
| `SeatPlayer`, `CloseEmptySeat`, `InactivateEmptySeat`, `ActivateEmptySeat`, `ActivateFilledSeat`, `ActivateInactivePlayer` | storage-injected data; delegate calls | O | fix-up |
| `CloseAllSeats`, `WaitForEnoughPlayers` | seat counts vs `TargetCount`; `delegate.ReadyToStart` | O | mixed |
| `AdvanceToken`, `HopAlongPath`, `MoveOnGraph` | graph walking; game hooks | O (rule 3) | mixed |
| `DefaultComponent` | the component's own `LegalComponent.Legal` | O | fix-up |
| `ShuffleDiscardIntoDraw`, `ReplenishMarket` | behavior-discovered stacks; branchy | O | fix-up |

Conclusions:

1. **Most framework types are O, and that is their point.** They are reusable
   because they call back into game Go (`TurnDone`, `ConditionMet`,
   `TargetCount`, `ReadyToStart`). Cataloguing them would either break rule 3
   or force authors to restate their hooks as state properties. That changes
   the authoring contract of every game using them.
2. **The negated/conditional types are not special once we stop decomposing
   them.** `ApplyUntil`'s "legal *until*" and `RoundRobin`'s "if started ...
   else ..." stay inside one opaque atom each. We add no `not`/`if`
   compositor. The anti-tarpit rules are untouched.
3. **Opacity is cheap where it occurs.** Nearly every O type is a fix-up and
   is never rendered as a form, so client evaluability buys nothing there.
   The player-facing G family (`AnyPlayer`, `AdminPlayer`) is where catalog
   re-expression pays off. Its blocker is the `TargetPlayerIndex` mutation
   inside `Legal`, not missing catalog vocabulary (see "Follow-ups").
4. **The allowlist's real criterion was never "declares no `Legal()`".** It was
   "**its whole legality is in its contribution**". `DrawToPlayer` and
   `MoveComponentToSlot` already met it by contributing a defining atom. The
   fix is to make every type meet it.

## Design: a base type's legality is its contribution

### The rule (what an author needs to know)

> Every framework move type contributes its own legality as named
> preconditions. Embedding any of them and adding `WithLegalPreconditions` or
> `LegalCustom` composes exactly as embedding `moves.Default` always has.
> A framework type's own check can be *moved* (suppress it, then re-author it
> where you want it), never *removed*.

There is no allowlist and no "unsupported base type" error.

### Concept mapping: no new author concepts

Every piece below is an existing concept used in a new place:

| Need | Existing concept reused |
|---|---|
| A base type's imperative check inside a plan | `LegalCustom`: the escape hatch, now also contributable by a base type, with a name and a position |
| Its ordering and suppression | the contribution chain plus `WithoutLegalPrecondition` |
| "Move, don't remove" | the existing defining-atom suppress-then-re-author recipe (`requireDefiningPreconditionReauthored`) |
| Its message | the frozen error text, rendered exactly as a plain `LegalCustom` error is today |

### Mechanism

**(a) Contributed residue atoms (core).** Core gains one constructor:

```go
// NewLegalResiduePredicate wraps imperative Go as a named, opaque plan atom.
func NewLegalResiduePredicate(name string, check func(LegalContext) error) *LegalPredicate
```

It is the `LegalCustom` wrapper with a name. Its properties:

- It is opaque: unserializable and not client-evaluable. Its verdict stays
  server-side.
- Its error translation is shared with the `LegalCustom` wrapper: a
  `*LegalError` keeps its structured verdict, and a plain error becomes a
  one-off template that renders as its own text. First-failure messages are
  therefore byte-identical to the frozen chain.
- Plan assembly files every opaque predicate (including any compositor
  containing one) as field-dependent: **never memoized**. A residue calls
  game hooks whose reads cannot be declared, and some read the move's own
  fields (`AnyPlayer`). This replaces the draft's "no Reads ⇒ unknown" rule.
  Opacity is the honest signal, it composes through `any`, and it leaves the
  F3 undeclared-read probe untouched for game predicates.

**(b) Framework residue atoms (moves).** Each framework type whose `Legal()`
adds checks after its super-call extracts that tail, verbatim, into a private
per-type method (`legalFinishTurn`, `legalApplyUntil`, ...). The names are
distinct so promotion never shadows. A registry constructor named after the
type wraps the method with `NewLegalResiduePredicate`. It calls the method on
`ctx.Move`, the concrete move, through a per-type interface that promotes
through embedding. The constructors are registered beside `inProgression` in
`moves.FrameworkConstructors()`, so the `legal` catalog does not grow.

Each atom has an exported `PreconditionName` constant (`PreconditionFinishTurn`
= `"finishTurn"`, ...) and an exported spec constructor
`moves.FrameworkPrecondition(name)` for re-authoring.

**(c) The contribution mirrors the super-call chain.** Each type overrides
`ContributedPreconditions()` as *parent's contribution + own atom*, in the
order its `Legal()` runs them. Where the frozen chain does not simply append,
the contribution says so explicitly:

- `MoveCountComponents` drops `stackConstraints` (its chain calls
  `legalWithoutStackConstraints`), then appends `applyUntil`, then
  `moveCountComponents`.
- `RoundRobin` appends a single `roundRobin` atom that owns the whole
  conditional. It does not append `applyUntil` unconditionally, because
  RoundRobin's legality does not run it unconditionally.
- `ForceFinishTurn` super-calls nothing, so it contributes only
  `forceFinishTurn`. It drops the inherited `inPhase`, `inProgression` and
  `finishTurn`, and documents that `WithLegalPhases` on it is ignored (as it
  always was).

**(d) The frozen `Legal()` defers to the plan.** Each override takes one
shape:

```go
func (f *FinishTurn) Legal(state boardgame.ImmutableState, proposer boardgame.PlayerIndex) error {
    if err := f.Default.Legal(state, proposer); err != nil {
        return err
    }
    if legalPlanOwnsChain(state, f.Name(), PreconditionFinishTurn) {
        return nil
    }
    return f.legalFinishTurn(state, proposer)
}
```

For a move that did not opt in, `legalPlanOwnsChain` is false and the frozen
code runs exactly as before. `CurrentPlayer.Legal` already had this shape.

`ForceFinishTurn.Legal` never super-calls, so it calls the shared plan entry
(`Default.legalPlanEntry`, extracted from `Default.Legal`) before its frozen
body.

`RoundRobin`'s frozen residue called `ApplyUntil.Legal`. That would re-enter
the plan from inside the plan and recurse forever. It now calls
`legalApplyUntil` directly. This is the one edit to a frozen path for
non-opted moves. It drops a redundant second evaluation of `Default`'s chain
on the same `(state, proposer)` that passed a few lines earlier. That chain
is pure apart from idempotent memos, so the result cannot be observed. The
differential tests cover it.

**(e) Probe proves the call path matches the type (core).** During the boot
probe, each level's guard records the atom it owns
(`LegalProbeReachResidue`). After the probe, every framework residue atom in
the plan must have been reached. Otherwise boot fails, naming the move and
the skipped level. This catches a game override that super-calls a
*non-immediate* ancestor (skipping, say, `FinishTurn.Legal`). Without the
check, opting in would silently add the skipped level's check.

**(f) Move, don't remove (core).** Two boot errors, both generic over residue
atoms:

- Suppressing a residue atom without re-authoring it in
  `WithLegalPreconditions`. This unifies with the defining-atom rule: the
  suppression can only reorder.
- Authoring a residue atom the move does not contribute (and suppress). A
  `finishTurn` atom on a move that does not embed `FinishTurn` would always
  be Unknown.

So the security gates (`forceFinishTurn`'s admin check, `seatPlayer`'s,
`anyPlayer`'s proposer≡target and auto-correct) and the loop terminators
(`applyUntil`) can never be deleted by configuration.

**(g) Ledger mode does not run residues past a failure (core).** In
full-ledger evaluation, a residue atom after the first non-Pass reports
Unknown "not evaluated: an earlier precondition failed" instead of running.
Residues assume their predecessors passed; for example, `adminPlayer` indexes
the target that `anyPlayer` validated. The hot path already short-circuits.
The `LegalCustom` tail keeps its current behavior; changing it is out of
scope (see open questions).

**(h) Enforcement.**

- *Structural (source-parse) test.* Every `moves` type that declares `Legal()`
  (except `Default`, the plan entry) must:
  - declare `ContributedPreconditions()`;
  - have a `Legal()` body that is either the three-statement template above
    (super-call; `legalPlanOwnsChain(...)` returning nil; `return
    x.legal<Type>(...)`) or begins with `legalPlanEntry`.

  The atom named in the guard must also be registered.
- *Differential golden tests (moves).* A purpose-built fixture installs every
  framework base type, opted in with zero authored specs, on one manager. For
  each state, proposer (Observer, Admin, each player) and move-field variant
  (sentinel, each player, out of range), the test compares:
  - **frozen**: `Legal()` evaluated under core's `LegalWithPlansSuspended`
    hook, which makes the plan invisible so the frozen chain runs;
  - **plan**: the real opted-in `Legal()`.

  Verdicts, first-failure text and post-call move fields must match. This
  exercises the real opted-in path: the short-circuits, the RoundRobin
  change, and ForceFinishTurn's entry. States include two-failure cases:
  - wrong phase plus a failing residue;
  - ForceFinishTurn in a wrong phase (must still pass);
  - MoveCount with a failing first-component stack check (must still pass);
  - RoundRobin unstarted, last-move-self, and started with condition met.
- *Game migrations* use the existing `legal_golden_test.go` convention: a
  hand-copied snapshot of the pre-migration `Legal()` compared with the
  migrated move.

### What is removed

`legalSupportedMovesBaseTypes`, its prose helper,
`LegalSupportedMovesBaseTypeNames`, `legalUnsupportedMovesBaseType`, and the
unsupported-base boot error. `LegalCustom` on a move that embeds no `moves`
type at all stays an error, because nothing would evaluate the plan. The
seam-allowlist doc-drift test is deleted (it has nothing left to drift);
the structural test replaces it.

## Implementation notes (what building it found)

- **Where things live.**
  - Core: `legal_residue.go` holds `NewLegalResidueConstructor`, opacity
    bucketing, probe reach, the move-don't-remove validation and
    `LegalWithPlansSuspended`.
  - moves: `moves/framework_checks.go` holds the 27 `Precondition*`
    constants, `FrameworkPrecondition`, `legalPlanOwnsChain` and the
    registry. Each base type's own check is a verbatim `legal<Type>` method
    beside its `Legal()`.
  - The golden harness is a new test-helper package, `legal/legaltest`
    (`OptInAll`, `Compare`/`CompareAt`, `Walk`). It depends only on core, so
    the moves tests, the example games and GAMES can all use it.
- **The differential harness must compare every version, not just settled
  states.** Fix-ups are judged at intermediate versions that a settled game
  never shows. `Walk` therefore compares at `game.State(v)` for every
  version. Mutation checks against the harness:
  - Swapping the order of DealCountComponents' contribution goes undetected
    by a walk of well-formed games: no reachable state has two failing atoms.
    The "two failures" fixture (moves restricted to a phase the game is not
    in, whose own check also fails) catches it.
  - Contributing `applyUntil` unconditionally under RoundRobin is caught at
    boot by the reach probe (e).
- **The harness compares only pure contributions.** It skips moves with
  authored specs, suppressions, `LegalCustom`, or a plan-only base. A
  game's own legality exists only in its plan, so "frozen" is not a
  baseline for it; it is fenced by the game's snapshot golden test.
  `DrawToPlayer`/`MoveComponentToSlot` were always plan-only and now say so
  (`LegalPlanOnly()`), so the harness skips them too.
- **The boot probe contains panics.** During the probe every framework level
  returns right after its super-call, so a game override's own code after
  that super-call runs on a move no framework check normalized. werewolf's
  `moveCastVote` indexed `players[-1]` because `AnyPlayer` never defaulted
  the target. Reach is recorded before any such panic, so the probe now
  recovers.
- **A latent desync was fixed.** `MoveCountComponents` previously inherited
  Default's `stackConstraints` contribution, even though its `Legal()` has
  always skipped that check (`legalWithoutStackConstraints`). Its
  contribution now mirrors the call path.
- **Interplay with enforced Reads** (landed on master during this work).
  Opaque predicates are exempt from read enforcement, so residue atoms still
  receive `ctx.Move`. A residue nested in `any` would be withheld the move,
  so nesting one is a boot error.
- **Codegen bug found and fixed.** `boardgame-util codegen` treated a
  delegate-shaped struct in a `_test.go` file as the package delegate. It
  emitted `ConfigureEnums` for it into the non-test `auto_enum.go`, which
  then could not compile. It now ignores test files (`findDelegateName`).
- **Stale generated reader in GAMES.** murdermrmonroe's `auto_reader.go`
  for `MoveMoveToRoom` (on `MoveOnGraph` → `CurrentPlayer`) predated
  codegen's promotion of `TargetPlayerIndex`. The move could not opt in, and
  its `TargetPlayerIndex` was never serialized. Regenerated.

## Migrations

- **boardgame examples:**
  - werewolf `moveCastVote` (`AnyPlayer`): the eliminated-voter gate is
    declared. The per-phase "already voted" read, the phase/role
    disjunction of conjunctions and the target checks stay `LegalCustom`.
    Fenced by `legal_golden_test.go`: 14,896 cases over a game played into
    the night.
  - Every example game gains `TestDeclarativeEquivalenceEveryMoveOptedIn`.
- **GAMES:**
  - darwin `moveSubmitFoodCard` (`AnyPlayer`) is fully declarative
    (`PropEquals` phase, `StackEmpty`, `ComponentPresentAt`), fenced by a
    snapshot golden.
  - Every game gains the opt-everything-in differential walk. This covers
    `MoveOnGraph`, `AdvanceToken`, `RoundRobin`/`RoundRobinNumRounds` and
    the Deal family on real games.
  - No other GAMES move on a newly composable base has a `Legal()` override
    to migrate.

## Alternatives considered

1. **Keep the allowlist and add types one by one with hand-written catalog
   specs.** Rejected. Most types are O, so each addition would be either
   dishonest or a change to the authoring contract.
2. **Partial contribution plus an imperative tail that keeps running after
   the plan.** Rejected. This is the desync the old docs warned about: the
   ledger says Pass while `Legal()` fails, and the tail cannot be reordered.
3. **One generic "whatever the base's Legal says" atom per move.** Rejected.
   It is unnamed, cannot be moved per check, and recurses into the plan entry.
4. **`not`/`if` compositors for ApplyUntil/RoundRobin.** Rejected. They
   violate the anti-tarpit rules only to wrap opaque leaves.
5. **"No declared Reads ⇒ unknown read-set" (the first draft).** Rejected per
   critique 2. It does not compose through `any`, weakens the F3 probe, and
   reads as vacuously evaluable. Opacity is the existing honest concept.
6. **Uniform, deletable suppression (the first draft).** Rejected per
   critique 1, because it deletes security gates and loop terminators.
7. **Always-plan: every framework type forces `LegalPlanEnabled()`, and only
   `Default` declares `Legal()`** (DrawToPlayer already does this). This is
   the most elegant end state: one path, no guards, and a trivial structural
   test. It is deferred, not rejected, because it would change behavior for
   un-migrated games:
   - a ledger wire change;
   - panics become Unknown;
   - boot validations fire on games that never opted in;
   - memo-mutex overhead in the fix-up loop.

   This design is the scaffolding for it. Once every type's contribution is
   proven equivalent, flipping the default is a local change.
8. **Re-explain AnyPlayer now (propose-time target defaulting, then catalog
   atoms).** Deferred to a follow-up. Moving the mutation out of `Legal`
   changes AnyPlayer's behavior and has its own blast radius (server form
   defaults, bots).

## Critique dispositions

1. Adopted as (f).
2. Adopted as (a). The atoms are described as contributed `LegalCustom`, not
   rule 2.
3. Adopted as (e).
4. Adopted as (g) for residue atoms. The custom tail is left unchanged and
   flagged.
5. Survey amended with the "Form?" column. AnyPlayer and AllPlayersSubmitted
   catalog re-expression are follow-ups.
6. Adopted. The differential harness replaces the shadow plan, and the AST
   test checks a strict template.
7. Corrections applied (FinishTurn's super-call, ForceFinishTurn's drops,
   RoundRobin's frozen-path edit disclosed).
8. Named as the endpoint (alternative 7).

## Follow-ups

- AnyPlayer: move target defaulting out of `Legal`, then contribute
  `ProposerIsPlayerFromMove` / `PlayerSeatIsFilled(PlayerFromMove(...))`
  atoms. This makes werewolf/darwin forms client-evaluable.
- AllPlayersSubmitted / ResetAllPlayerSubmissions as catalog atoms, once
  message parity is acceptable.
- Always-plan (alternative 7).

## Open questions for the user

- Should the `LegalCustom` tail also report "not evaluated" in ledger mode
  after an earlier failure, matching (g)?
- Is "move, don't remove" the right default for *every* framework atom, or
  should some (`finishTurn`) be deletable, now that `ForceFinishTurn` is the
  sanctioned way to skip `TurnDone`?
