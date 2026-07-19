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
3. Validation observes a coherent whole-state simulation. Constraints on the
   second and later insertion see the effects of earlier simulated insertions,
   both through the destination and through the supplied state.
4. Within one `MayMoveAllTo` or `MoveAllTo` call, validation happens once per
   proposed insertion. Commit does not invoke constraints a second time. A
   normal proposal may still evaluate constraints once during `Legal` and once
   during `Apply`; constraints must never depend on total invocation count.
5. `MayMoveAllTo` and `MoveAllTo` use the same validation path, error semantics,
   ordering, and slot selection.

“Framework-owned state” is deliberately broader than the two endpoint stacks.
A local snapshot-and-rollback cannot honestly satisfy this contract because a
constraint receives the whole state and may inspect relationships elsewhere.

### Constraint contract

The existing requirement that `StackConstraint` be pure and not panic becomes
more explicit. A valid constraint is a deterministic, copy-stable predicate:

- it bases its result only on logical values reachable through the supplied
  destination, proposed components, and immutable state;
- it does not mutate supplied or captured state;
- it does not consume randomness, schedule callbacks, perform I/O, modify a
  captured variable, inspect pointer identity, consult clocks, or depend on
  invocation count;
- it does not retain supplied objects after returning; and
- it does not panic.

This is not a new practical restriction: `MayMoveAllTo`, `Legal`, state replay,
and the engine's copied-Apply boundary already require deterministic validation.
It is a clarification of the semantic contract that those APIs depend on.

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

## Selected design: validate a copy, commit an infallible plan

The operation has two phases.

### 1. Prepare and validate

1. Resolve both values to attached physical stacks.
2. Validate distinct endpoints, exact state identity, mutation permission,
   matching decks, and aggregate capacity.
3. Return immediately for an empty source.
4. Deep-copy the complete unsanitized state.
5. Locate copied endpoints through the canonical stack-owner paths.
6. Run the existing sequential checked transfer on the copied endpoints.
7. If any step fails, return that error and discard the copy.

The checked simulation is a private helper. It performs the current loop of
“first component to destination next slot” and therefore preserves the exact
order-dependent behavior of all four growable/sized source-destination
combinations.

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
- `validateMoveAllPlan`: whole-state copy, copied-stack lookup, and sequential
  checked transfer on the disposable state;
- `commitMoveAllPlan`: exactly the planned number of non-failing structural
  moves on the original state.

`MayMoveAllTo` prepares and validates a plan, then stops. `MoveAllTo` prepares
and validates a plan, then commits it. This gives the two public methods one
source of truth without recursion, duplicate endpoint validation, or double
constraint evaluation inside one call.

No fast path will initially bypass copying when a destination has no
constraints. The optimization would couple transaction correctness to private
constraint storage and make future per-component validation easy to omit. Add
it only if a benchmark shows state copying is material in realistic Apply
workloads.

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
- Normalize the duplicate capacity messages to lower-case
  `not enough space in the target stack`; callers must not depend on error text.
- Do not recover constraint panics. The documented programming error should
  retain its stack trace, and recovery could not undo external effects anyway.
- Do not add backward-compatibility switches. The old partial mutation was an
  implementation defect, not useful API behavior.
- Existing games require no migration. Examples which ignore `MoveAllTo`
  errors should still be corrected so they demonstrate the intended style.

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

- Keep `MayMoveAllTo` non-mutating and error-equivalent.
- Run focused stack, ownership, constraint, and move tests under `-race`.
- Run the complete BOARDGAME suite.
- Run all canonical `../games` tests independently against the worktree module
  through a temporary task-local `go.work` if no game source changes are
  required; create a paired GAMES worktree before any source edit.
- Add a benchmark for a representative large state and bulk transfer. Record
  the copy cost, but prefer correctness until measured evidence supports an
  internal optimization.

Initial implementation measurement on an Apple M1 (`-benchtime=100x`, three
runs) is 0.23–0.43 ms and roughly 47 KB / 678 allocations per benchmark
iteration. Each iteration performs two successful two-component transactional
transfers, including two whole-state validations. This is visible overhead but
not evidence for a more fragile fast path; profile a real large game before
special-casing constraint-free destinations.

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

Remaining risks are intentionally visible rather than papered over:

- whole-state copying may be expensive for unusually large games, so the
  implementation includes measurement before considering a constraint-free
  fast path;
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
