# Transactional Counted Component Transfers

**Issues:** [#624](https://github.com/jkomoros/boardgame/issues/624), follow-up to
[#799](https://github.com/jkomoros/boardgame/issues/799)

**Branch:** `transactional-count-moves`

## Outcome

Game creators get the missing operation between moving one component and moving
an entire stack:

```go
if err := source.MayMoveCountTo(destination, 3); err != nil {
	return err
}

if err := source.MoveCountTo(destination, 3); err != nil {
	// No framework-owned state was changed by MoveCountTo.
	return err
}
```

`MoveCountTo` moves exactly `count` components, in source order, into successive
next slots in the destination. It is atomic with respect to returned errors and
represents one notional game move, so clients animate the transferred components
together. `MayMoveCountTo` performs the identical validation without mutation.

The existing `moves.MoveCountComponents` family keeps its distinct purpose: it
moves one component per engine move, preserving separate persistence, move
records, and animation boundaries. Its `Legal` method will preflight every
remaining predictable transfer so a repeated fix-up cannot begin a sequence
that is already known to run out of components, capacity, or constraint-valid
insertions.

## Why this is a separate primitive

The current API has two endpoints:

- `ComponentInstance.MoveTo*` moves one selected component.
- `Stack.MoveAllTo` atomically moves every component.

Creators who need exactly N components either write a loop around `First()` and
`MoveToNextSlot`, which can partially mutate before a later failure, or abuse a
repeated move type when they actually want one logical move. Both choices hide
intent and invite ignored-error bugs.

The counted primitive makes the common expression direct while retaining the
important distinction between one compound move and multiple engine moves.
There is no public transaction object, callback, or plan type to learn.

## Public contract

Add the following methods:

```go
type ImmutableStack interface {
	MayMoveCountTo(dest ImmutableStack, count int) error
}

type Stack interface {
	MoveCountTo(dest Stack, count int) error
}
```

The argument order follows `MoveAllTo(destination)` and keeps the destination
adjacent to the verb; the optional quantity comes last, like other Go APIs that
extend an operation with a count.

For both methods:

1. `count < 0` is a loud programmer error returned as an error.
2. `count == 0` is a successful no-op only after validating that both endpoints
   are real, attached, distinct, mutable physical stacks in the same state and
   deck. A zero must not make `nil` or stale endpoints silently valid.
3. `count > source.NumComponents()` fails before constraint evaluation.
4. Components are selected from first non-empty to last non-empty source slot.
   Sparse sized stacks therefore have unsurprising deterministic behavior.
5. Destination placement uses each successive `nextSlot`, exactly like
   `MoveAllTo`.
6. Insufficient aggregate destination capacity fails before constraint
   evaluation.
7. Constraints observe an ordered simulation: the second insertion sees the
   first simulated insertion, and so on.
8. A returned error leaves framework-owned live state unchanged. On success,
   constraints are not evaluated again during commit.
9. Merged stacks remain read-only views and are rejected as either endpoint.
10. The deterministic, pure, copy-stable `StackConstraint` contract documented
    by transactional `MoveAllTo` applies unchanged.

`MoveAllTo` will share the counted implementation with
`count == source.NumComponents()`. Existing `MoveAllTo` behavior and historical
error text remain compatible; the new counted methods receive clear lowercase
diagnostics without introducing a parallel typed-error hierarchy in this
tranche.

## Declarative legality

Add the catalog predicate:

```go
legal.MayMoveCountTo(
	"game.DrawStack",
	"player.Hand",
	"move.Count",
)
```

The first two paths resolve immutable stacks and the third resolves an integer.
Resolution, type, negative-count, ownership, capacity, and constraint errors are
reported through the normal declarative legality diagnostics. The predicate is
server-evaluated because custom stack constraints are arbitrary Go predicates.

Do not add a constant-count variant yet. Built-in moves cover the common fixed
count through `moves.WithTargetCount`; custom input moves naturally carry their
count as a typed field. Another constructor would enlarge the catalog without a
demonstrated authoring case.

## Internal design

Generalize the private `moveAllPlan` into a counted component-transfer plan:

```go
type componentTransferPlan struct {
	source      Stack
	destination Stack
	count       int
}
```

Preparation validates endpoints, count, available source components, and
aggregate destination capacity. Validation retains the existing split:

- unconstrained destinations preflight the selected components directly;
- constrained destinations copy the whole state, locate corresponding physical
  endpoints, and perform exactly `count` checked transfers on the copy.

Commit performs exactly `count` private structural removals and insertions and
cannot return an error. Public methods all use the same prepare/validate path:

- `MayMoveCountTo`: prepare, validate, stop;
- `MoveCountTo`: prepare, validate, commit;
- `MayMoveAllTo`: capture current component count, prepare, validate, stop;
- `MoveAllTo`: capture current component count, prepare, validate, commit.

Keeping the plan private prevents callers from retaining stale endpoints or
trying to commit a plan after unrelated mutations.

## Built-in move integration

### `MoveCountComponents` and subclasses

`Legal` computes the remaining number of applications from the concrete move's
`Count` and `TargetCount`. The existing `ApplyUntilCount` contract says each
application moves the count exactly one step toward its target, so the absolute
difference is the number of remaining component transfers. Negative counts or
targets become explicit legality errors.

It then calls `source.MayMoveCountTo(destination, remaining)`. This catches, at
the first proposal:

- an undersized source;
- a destination that cannot hold the complete remainder; and
- an order-dependent constraint that rejects a later component.

`Apply` continues to move exactly one component, but uses
`source.MoveCountTo(destination, 1)` so the mutation and preflight paths share
one implementation. The engine still commits one version per application.

### Deal/collect round robins

`DealCountComponents` and `CollectCountComponents` can have custom player
conditions, player selection, stack selection, and `RoundRobinAction` behavior.
Pretending that their entire future multi-version sequence is one statically
knowable transaction would be a false guarantee and could execute creator code
during `Legal`.

Their `Legal` methods will therefore validate the next scheduled transfer with
`MayMoveCountTo(destination, 1)`, and their default actions will use
`MoveCountTo(destination, 1)`. This fully addresses #624 for every individual
engine move without changing round-robin semantics. A future explicit
round-robin-plan API would require its own design and evidence.

## Compatibility and foot-gun controls

- No existing method is removed or weakened.
- No automatic conversion from multiple move records to one move record occurs.
- Exact-count failure is explicit; the API never means “up to count.”
- Zero is a no-op, while negative counts and invalid endpoints remain loud.
- `May*` and mutating methods have identical validation order and constraint
  results within one state.
- Examples and tutorial material will contrast `MoveCountTo` with
  `moves.MoveCountComponents` so authors choose intentionally.
- The declarative predicate uses a typed integer move field; malformed or
  non-integer paths fail configuration/legality loudly.

## Adversarial test matrix

### Fundamental API

Cover growable and sized sources/destinations, including sparse sized sources:

- negative, zero, one, exact-all, and too-large counts;
- nil, same, detached/stale, cross-state, cross-deck, immutable, and merged
  endpoints;
- insufficient growable max size and sized empty slots;
- successful order and slot placement;
- component-location index consistency after success and failure;
- `MayMoveCountTo` non-mutation and error parity with `MoveCountTo`;
- late order-dependent constraint rejection with exact whole-state equality;
- constraint invocation exactly once per proposed insertion within one call;
- successful equivalence to sequential checked moves on an independent copy;
- `MoveAllTo` regression parity through the shared implementation.

### Built-in moves

- `MoveCountComponents` rejects an insufficient source before its first move.
- It rejects insufficient destination capacity and late constraint failures
  before its first move.
- Its successful path still creates distinct move records/versions.
- Until-count-reached, until-count-left, and move-all subclasses derive the
  correct remaining count.
- Deal and collect reject an empty source and an invalid destination before
  their action without changing round-robin bookkeeping.
- Custom stack selectors retain their existing behavior.

### Declarative legality

- Bindings, conformance corpus, documentation snippets, and client metadata.
- Correct count, zero, negative, missing field, wrong type, capacity, and
  constraint diagnostics.
- Predicate evaluation never mutates the live state.

## Validation

1. Focused stack, constraint, move, legal-catalog, and example tests.
2. Focused race tests for fundamental stack and move packages.
3. `go vet` for changed Go packages.
4. Complete BOARDGAME test suite through `./scripts/go-local`.
5. Complete GAMES suite through the paired workspace.
6. Publish/pin Boardgame and rerun GAMES with `GOWORK=off` before landing if
   companion source or module metadata changes.

## Implementation sequence

1. Add failing fundamental API tests, then generalize the existing private
   `MoveAllTo` plan without changing its behavior.
2. Add `MayMoveCountTo` and `MoveCountTo` and complete the adversarial matrix.
3. Add the declarative predicate and its generated/client-facing metadata.
4. Integrate the built-in move family while preserving separate move records.
5. Update tutorial, package documentation, examples, and issue references.
6. Run focused, race, full framework, paired-game, and standalone validation.
7. Rebase on the latest targets, regenerate if needed, retest, and land in
   dependency order.

## Deliberate non-goals

- A general public transaction builder.
- Atomicity across multiple engine move records or database commits.
- Running arbitrary custom `RoundRobinAction` implementations speculatively.
- A best-effort “move up to N” operation.
- Typed mutation errors for every historical stack method.
