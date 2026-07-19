# Transactional `MoveAllTo`

**Issue:** [#799](https://github.com/jkomoros/boardgame/issues/799)

**Branch:** `moveall-transaction-semantics`

## Outcome

`Stack.MoveAllTo` will have the same failure shape creators reasonably expect
from a single-component move:

```go
if err := source.MoveAllTo(destination); err != nil {
	// No framework-owned state was changed by MoveAllTo.
}
```

The transfer remains sequential for constraint purposes. A constraint sees the
destination before each proposed insertion, including components accepted by
earlier steps in the simulated transfer. On success, source order, destination
placement, component IDs, and component-location indexes are exactly those
produced by today's successful implementation.

The public API does not grow a transaction object, callback, or alternate
method. The safe operation is the obvious operation. `MayMoveAllTo` remains the
non-mutating predicate for `Legal`, and existing game code remains
source-compatible.

## The hole

`MoveAllTo` currently validates ownership, deck compatibility, and aggregate
capacity before moving components one at a time. Every individual move checks
the destination constraints. If the first component passes and a later one is
rejected, the earlier components remain moved even though `MoveAllTo` returns
an error.

The engine's normal proposal path masks this behavior: `Game.applyMove` runs
`Move.Apply` on a disposable whole-state copy and discards that copy when
`Apply` returns an error. The public stack API is nevertheless unsafe in three
ordinary cases:

- direct mutation during setup or tests;
- an `Apply` method which handles or accidentally ignores the returned error;
- library code which composes several mutations before deciding what error to
  return.

`MayMoveAllTo` already avoids the hole by copying the complete state, locating
the corresponding physical source and destination, and simulating the same
sequential moves on that copy. The missing piece is a commit boundary.

## Contract

### Framework-owned state

For a valid state and valid `StackConstraint` implementations:

1. If `MoveAllTo` returns an error, it makes no change to framework-owned
   state. This includes every stack, component-location indexes, dynamic
   component values, timers, callbacks, and the state's deterministic random
   generator.
2. If it succeeds, the only changes made by `MoveAllTo` are the transfer's
   source and destination membership, ordering, ID bookkeeping, and component
   location indexes.
3. Constrained validation observes a coherent whole-state simulation.
   Constraints on the second and later insertion see the effects of earlier
   simulated insertions, both through the destination and through the supplied
   state. An unconstrained transfer instead preflights every component without
   reconstructing unrelated state.
4. Within one `MayMoveAllTo` or `MoveAllTo` call, validation happens once per
   proposed insertion. Commit does not invoke constraints a second time. A
   normal proposal may still evaluate constraints once during `Legal` and once
   during `Apply`; constraints must never depend on total invocation count.
5. `MayMoveAllTo` and `MoveAllTo` use the same validation path, error
   categories, ordering, and slot selection. Their historical capacity strings
   retain different capitalization for compatibility.

“Framework-owned state” is deliberately broader than the two endpoint stacks.
A local snapshot-and-rollback cannot honestly satisfy this contract because a
constraint receives the whole state and may inspect relationships elsewhere.

### Constraint contract

The existing requirement that `StackConstraint` be pure and not panic becomes
more explicit. A valid constraint is a deterministic, copy-stable predicate:

- it may use immutable constructor/configuration constants and immutable
  component definitions;
- it otherwise bases its result only on persisted logical game, player,
  dynamic-component, and stack values reachable through the supplied
  destination, proposed components, and immutable state;
- it does not depend on `State.Version`, object identity, or live `Game`
  metadata (`Legal` sees the current version while `Apply` operates on the next
  version);
- it does not mutate supplied or captured state;
- it does not consume randomness, schedule callbacks, perform I/O, modify a
  captured variable, inspect pointer identity, consult clocks, or depend on
  invocation count;
- it does not retain supplied objects after returning; and
- it does not panic.

`MayMoveAllTo`, `Legal`, state replay, and the engine's copied-Apply boundary
already require deterministic validation. Transactional direct `MoveAllTo`
now exposes the same requirement. A custom constraint that previously captured
its live containing substate may therefore require migration; the canonical
pattern is to capture immutable configuration only and resolve runtime objects
from the supplied state.

Go cannot enforce this contract for a function closure. An `ImmutableState`
can be cast back to creator-owned concrete values, a closure can capture live
objects, and external effects cannot be rolled back. Behavior after a
constraint violates this contract is unsupported. The documentation must not
claim that `MoveAllTo` can transact a database write, global counter, or live
state deliberately mutated behind the immutable interfaces.

The simulation does provide useful containment: ordinary mutations made by
type-asserting the supplied copied state affect only the disposable copy. It
does not legitimize such constraints.

### Concurrency

Game states are single-writer objects and stack mutation is not concurrency
safe. Transactional means atomic with respect to returned errors, not safe for
concurrent goroutine mutation. Concurrent mutation remains unsupported and is
covered by the existing race test expectation.

## Selected design: preflight or simulate, then commit an infallible plan

The operation has two phases.

### 1. Prepare and validate

1. Resolve both values to attached physical stacks.
2. Validate distinct endpoints, exact state identity, mutation permission,
   matching decks, and aggregate capacity.
3. Return immediately for an empty source.
4. If the destination has no constraints, validate every source component
   against the live endpoints without mutation. Endpoint validation and
   aggregate capacity make every successive next slot structurally available.
5. If the destination has constraints, deep-copy the complete unsanitized
   state, locate copied endpoints through canonical stack-owner paths, and run
   the sequential checked transfer on the copy.
6. If any step fails, return that error without changing the live endpoints.

The constrained simulation is a private helper. It performs the current loop
of “first component to destination next slot” and therefore preserves the exact
order-dependent behavior of all four growable/sized source-destination
combinations. The unconstrained path still checks each component's state,
attachment, containing stack, and deck before commit; it does not merely trust
the aggregate count.

### 2. Commit

After simulation succeeds, repeat the same deterministic sequence on the live
endpoints using a private structural primitive:

```go
for remaining := plan.count; remaining > 0; remaining-- {
	component := source.removeComponentAt(source.firstComponentIndex())
	destination.insertComponentAt(destination.nextSlot(), component)
}
```

Commit performs no operation that returns an error and does not rerun
constraints. Endpoint identity, capacity, deck, and mutation permission were
validated immediately before the synchronous simulation; valid constraints
cannot mutate the live state; and concurrent mutation is unsupported. The
primitive remains private so creators cannot bypass normal validation.

This retains the established component-index maintenance in
`removeComponentAt` and `insertComponentAt`. It does not copy stack payloads
back from the simulated state, which would overwrite unrelated state identity,
ownership metadata, IDs, or references.

### Shared internal shape

Represent the prepared operation with a small private plan containing the
physical source, physical destination, and fixed transfer count. The count is
captured after endpoint and capacity validation, so commit does not use a
changing loop condition as its authority.

Use three narrowly named helpers rather than having public methods recursively
call one another:

- `validateMoveAllEndpoints`: structural endpoint validation;
- `prepareMoveAll`: endpoint/capacity validation and construction of the
  private plan;
- `validateMoveAllPlan`: constraint-free component preflight or whole-state
  constrained simulation;
- `commitMoveAllPlan`: exactly the planned number of non-failing structural
  moves on the original state.

`MayMoveAllTo` prepares and validates a plan, then stops. `MoveAllTo` prepares
and validates a plan, then commits it. This gives the two public methods one
source of truth without recursion, duplicate endpoint validation, or double
constraint evaluation inside one call.

The constraint-free path is part of the correctness design, not an unchecked
shortcut. It preserves the fundamental stack API for minimally attached
internal stacks, avoids rerunning unrelated creator construction hooks, and
makes common transfers scale with the transferred components rather than the
entire game. Unknown future physical stack implementations conservatively use
whole-state simulation until their constraint storage is understood.

## Rejected alternatives

### Roll back the two stacks

Restoring source and destination indexes does not restore another state field,
randomness, a pending callback, or an external effect touched by a bad
constraint. It would advertise stronger semantics than it implements and
would make component-index restoration another fragile path.

### Replace the live state with the successful copy

Direct callers retain stack and substate references. Replacing the state's
object graph would make those references stale, conflict with canonical stack
ownership, and move the engine's persistence boundary into a low-level stack
method.

### Check every component first against the current destination

Passing all components at once, or checking each against the unchanged
destination, changes sequential constraints. Maximum counts, uniqueness, and
cross-stack predicates must see earlier accepted components.

### Validate on the copy, then use normal moves on live state

This invokes constraints twice. A merely call-count-sensitive constraint could
disagree, and any invalid side effect would be doubled. More importantly, a
late live rejection recreates the partial-mutation hole.

### Document `MoveAllTo` as weak

The method reads as one operation, returns one error, and is used as a
convenience precisely to avoid manual sequencing. Making the common API
non-atomic places the burden on every creator and makes ignored errors
especially destructive.

### Add a public transaction API

The engine may eventually need general multi-mutation transactions, but
requiring a transaction wrapper for this common operation preserves the
footgun. A future general facility can subsume the private implementation
without changing the creator API.

## Errors and compatibility

- Preserve existing validation categories and the first failing constraint's
  error.
- Preserve the historical capacity strings, including their capitalization
  difference, until stack mutations expose typed errors. Existing errors are
  untyped, so string comparison is an unfortunate but plausible caller
  dependency.
- Do not recover constraint panics. The documented programming error should
  retain its stack trace, and recovery could not undo external effects anyway.
- Do not add backward-compatibility switches. The old partial mutation was an
  implementation defect, not useful API behavior.
- The canonical GAMES repository has no custom programmatic constraints and
  requires no migration. Other games with custom closures must audit for live
  state captures. Examples which ignore `MoveAllTo` errors should still be
  corrected so they demonstrate the intended style.

## Adversarial test matrix

### Atomic rejection

For growable and sized endpoints, construct a constraint that accepts one or
more insertions and then rejects. Assert after the error:

- exact source and destination raw order/holes are unchanged;
- every component still resolves to its original containing stack and slot;
- unrelated stacks and persisted scalar properties are unchanged;
- no pending framework callbacks or timers were added; and
- a subsequent valid single-component move succeeds, proving indexes remain
  coherent.

Cover growable-to-growable, growable-to-sized, sized-to-growable, and
sized-to-sized, including sparse sized sources and destinations.

### Successful equivalence

Run transactional `MoveAllTo` on one state and the old sequential checked
algorithm on an independent copy. Compare storage records and component
locations. Include destination constraints that depend on the accumulated
destination and state-level views of both endpoint stacks.

Assert constraints are invoked exactly once for each simulated insertion and
receive objects owned by the copied state, not the live state.

### Early failures

For nil/view endpoints, same stack, detached/stale stack, different states,
different decks, sanitized states, and insufficient capacity, assert the full
state storage record is unchanged. Existing ownership tests remain the primary
coverage for endpoint diagnostics.

### Constraint-contract boundaries

Add documentation-focused tests where useful, but do not encode unsupported
behavior as a guarantee. In particular:

- a constraint mutating the supplied concrete state should demonstrate that
  the mutation is confined to the discarded copy;
- an external call counter may verify no double evaluation, but its use is
  explicitly test instrumentation rather than an endorsed constraint pattern;
- a panic should propagate; and
- a constraint that captures and mutates the original state is invalid and is
  not promised rollback.

### Regression and performance

- Keep `MayMoveAllTo` non-mutating and constraint-error-equivalent.
- Run focused stack, ownership, constraint, and move tests under `-race`.
- Run the complete BOARDGAME suite.
- Run all canonical `../games` tests independently against the worktree module
  through a temporary task-local `go.work` if no game source changes are
  required; create a paired GAMES worktree before any source edit.
- Add a benchmark for a representative large state and bulk transfer. Record
  the copy cost, but prefer correctness until measured evidence supports an
  internal optimization.

Final implementation measurement on an Apple M1 (`-benchtime=100x`, three
runs) compares iterations containing two successful two-component transfers:

- unconstrained, normal state: 3.9–4.6 µs and 1.5 KB / 52 allocations;
- constrained, normal state: 138–165 µs and 47 KB / 678 allocations;
- unconstrained with an unrelated 10,000-slot stack: 4.2–4.4 µs and 1.5 KB /
  52 allocations; and
- constrained with that large unrelated stack: 342–483 µs and roughly 538 KB /
  678 allocations.

The common constraint-free path is independent of unrelated state size.
Constrained transfers pay the honest whole-state transaction cost; profile a
real game before introducing a more complex general transaction substrate.

## Adjacent APIs

This tranche inventories but does not broaden into every compound mutator:

- `SwapComponents`, shuffle, and public shuffle validate before structural
  mutation and expose no ordinary late error.
- `SortComponents` can record an internal swap error after earlier swaps, but
  its generated indexes and initial mutation check make that error unreachable
  for valid framework state. Comparator purity and panic behavior are a
  separate API-contract question.
- size contraction contains an internal “unexpected” late error guarded by a
  complete empty-slot count precondition.
- `moves.MoveAllComponents` intentionally represents multiple engine Moves and
  therefore multiple persistence/animation boundaries; it is not the same
  atomic operation as `Stack.MoveAllTo`.

If tests expose a reachable partial-error path in another compound mutator,
file it separately rather than quietly expanding this contract.

## Design self-critique

The first draft selected whole-state simulation but left three ambiguities that
would have weakened the implementation. This revision resolves them:

1. “Once” now means once within one API call. It does not imply that the engine
   skips the independent `Legal` and `Apply` checks.
2. Preparation returns an immutable private plan. Validation and commit cannot
   silently rediscover different endpoints or transfer a count chosen after
   validation.
3. The contract explicitly requires copy stability as well as absence of side
   effects. A pure predicate based on pointer identity or external mutable data
   could otherwise disagree between the copied validation graph and the live
   commit graph.
4. The permitted inputs exclude state version and live game metadata. Those
   values can differ between the independent `Legal` and `Apply` evaluations
   even when persisted game values have not changed.

Remaining risks are intentionally visible rather than papered over:

- constrained transfers copy the whole state and can be expensive in unusually
  large games or when repeated many times in one Apply;
- Go interfaces cannot prevent malicious mutation through type assertions or
  captured references; and
- an internal panic during the supposedly infallible commit would still be a
  framework invariant failure, not a recoverable transaction error.

## Implementation sequence

1. Add late-rejection tests that fail against the current implementation.
2. Split checked simulation from unchecked commit without changing public
   behavior.
3. Route `MayMoveAllTo` and `MoveAllTo` through shared simulation.
4. Strengthen API and constraint documentation.
5. Fix example error handling and audit all canonical game callers.
6. Run adversarial, race, benchmark, full BOARDGAME, and independent GAMES
   validation.
