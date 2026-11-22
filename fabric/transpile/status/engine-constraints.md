# Workstream 8: stack constraints and enum-keyed access

Base `d2cb55e4`. What `fabric/TRANSPILE-PROPOSAL.md` workstream 8 asked for:
the TypeScript engine core enforces stack constraints exactly as Go does,
and has Go's by-key stack API, so checkers' state and moves have what they
need from the core.

## What was built

### `server/static/src/engine/constraints.ts` (new)

A port of `stack_constraint.go`, the constraint half of
`struct_inflater.go`'s `unpackStackStructTag`, and all four constructors of
`constraints.DefaultConstructors` (`max`, `unique`, `same`, `maxdistinct`),
including `constraints/prop_path.go` (`Prop`, `component.Prop`,
`dynamic.Prop`, case-insensitive prefix, values-then-dynamic fallback,
construction-time `validatePropPath`) and `helpers.go` `intEffectiveValue`
with `strconv.Atoi`'s error texts. A constraint is a data record, evaluated
by a `switch`; nothing at module scope but functions and constants.

Exported:

| Name | What |
|---|---|
| `StackConstraint`, `StackConstraintKind`, `PropPathPrefix` | The parsed constraint |
| `parseStackConstraints(propName, prop, def)` | `PropDef.Constraints` to constraints, or Go's boot error text (`X was a nil SizedStack and its struct tag was not valid: constraint same error: ...`) |
| `checkConstraints(dest, proposed)` | `Stack.CheckConstraints`: Go's error text or null; first failing constraint in tag order wins |
| `splitTagFields`, `parseConstraintExpr`, `constraintsFromStackTag` | The tag-parsing helpers (`constraintsFromStackTag` yields what `PropDef.Constraints` should hold for a whole tag) |
| `splitPropPath`, `goAtoi` | Helpers, exported for tests |

### `stack.ts`

- `Stack.constraints` (parsed, readonly), set by a new optional last
  parameter of `newStack`.
- `mayMoveToError` (Go's `componentInstance.MayMoveTo`) ends with
  `checkConstraints(dest, [component])`, after the slots-remaining check, as
  in Go. Everything that goes through it is therefore constrained, exactly
  Go's set: `moveComponent`, `secretMoveComponent`, the three
  first/last/next-slot moves, `mayMoveComponentTo`, and the transfers.
  Slides, swaps, shuffles, sorts and `insertStarterComponent` are not, as in
  Go.
- `transferError` (MoveAllTo/MoveCountTo and their May forms): with more than
  one component and a constrained destination it now runs Go's
  copy-and-simulate path (`validateComponentTransferPlan`), new
  `simulatedTransferError`: copy the state, find both stacks by owner path,
  perform the moves with `moveComonentImpl`'s checks, return the first error.
  Otherwise the existing live path.
- New: `componentAtKey(k)` on `StackView` (both `Stack` and
  `MergedStackView`), `swapComponentsByKey`, `maySwapComponentsByKey`, and
  `mayMoveComponentToSlot` (Go's `MayMoveToSlot`, which checkers'
  `movePlaceToken.LegalCustom` calls; `rules/binding.ts` has its own copy).

### `state.ts`

`index()` parses each stack or board property's `Constraints` once and
attaches them to the stack, or to every space of the board, as the struct
inflater does; a parse error throws a `StateError` with Go's text.
`register` takes an optional `constraints` argument. Copies and sanitized
copies re-index and so carry them.

## Vectors

`testdata/gen/stackvec` now also writes `testdata/constraints.vectors.json`
(from the new `constraints.go`, run after the existing generator with its own
seed, so `stacks.vectors.json` is byte for byte unchanged: regenerated and
compared). It holds:

- **Three definitions**, derived from live Go managers by `stackvec/def.go`
  (the shape `vecgame/defs.go` writes, plus `PropDef.Constraints` read from
  the real struct tags and `DeckDef.ComponentProps`): `cgame`, `tictactoe`,
  `checkers`. Checkers' carries `CapturedTokens: ["same(Color)"]`,
  tictactoe's `UnusedTokens: ["same(Value)"]`. `def.go` is a local copy, not
  a call into `vecgame/defs.go`, which workstream 1 is rewriting.
- **34 scenarios, 2,400 operations** on live states (26 cgame scenarios of
  80, 4 each for tictactoe and checkers of 40), 18 kinds of operation: the
  five moves, `MayMoveTo`, `MayMoveToSlot`, `MoveAllTo`, `MoveCountTo`,
  `MayMoveAllTo`, `MayMoveCountTo`, slides, `Shuffle`, `SwapComponents`,
  `SwapComponentsByKey`, `MaySwapComponentsByKey`, `ComponentAtKey` (and
  `ImmutableComponentAtKey` on a merged stack). 911 refused, 332 of them by a
  constraint. Each records Go's error or result, the stacks that changed,
  and the state after; the generator panics if a refused or May/lookup
  operation changes anything.
- **Coverage is enforced, not hoped for**: the generator keeps generating
  cgame scenarios (bounded at 200; it needed 127, keeping the 26 that added
  coverage) until, for each of the 11 component-moving operations and each
  of the 4 kinds, a constrained destination both accepted and refused; for
  each of the 4 transfers and each kind, a multi-component transfer was
  accepted and one was refused although each component passes alone (only
  the simulation finds those); plus swaps, slides and shuffles of
  constrained stacks, both by-key operations each way, `ComponentAtKey`
  found/nil/merged, and a stack with three constraints refusing. The test
  re-checks the same table from the file.
- **15 malformed tags** with Go's `NewStructInflater` error (wrong argument
  counts, non-int and out-of-range limits, a misspelled property, a
  `dynamic.`/`component.` prefix naming the wrong side, an unknown
  constraint, an unexpected field).

`cgame` (`stackvec/cgame/`, its own `auto_reader.go` from `boardgame-util
codegen`) has every kind on growable stacks, sized stacks, board spaces
(`maxdistinct(Color,1)` on a 3-space board), player stacks and a
dynamic-value stack, over enum, int, string, bool and int-slice values, a
deck lacking the property (always skipped), a spaced tag
(`maxdistinct(Suit, 2)`), and set-up that leaves a constraint violated (Go
does not check at set-up; the engine loads it and keeps refusing further
inserts with Go's "mixed values" text).

## Tests

`constraints.test.ts`, 8 tests: coverage table; the real constraint strings;
checkers' and tictactoe's constraints load into an `EngineState` on the
right stacks (and survive copy and sanitized copy); every one of the 2,400
operations matches Go (errors, results, changed stacks, secret-move counts,
no change on error, conservation, final state); the same replay with
`ComponentProps` removed (values read from the deck JSON); states load and
serialize back; Go's error for every malformed tag, both from the parser and
from `loadEngineState`; parser helpers.

Mutation check (run, then reverted): disabling `checkConstraints`, disabling
the simulation path, or removing the dynamic-value fallback each fails the
replay tests.

## Results (as run)

| Command | Result |
|---|---|
| `./scripts/go-local build ./...` | passes |
| `./scripts/go-local vet ./server/static/src/engine/testdata/gen/stackvec/...` | clean; `gofmt -l` clean |
| `./scripts/go-local test ./boardgame-util/lib/conformance/...` | ok, both packages |
| `./scripts/go-local run ./boardgame-util conformance export --check` | "The conformance corpus is current" |
| `stackvec` run twice | `constraints.vectors.json` identical; `stacks.vectors.json` unchanged from the base |
| `npm run type-check` | passes |
| `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | passes |
| `npm run test:unit` | 1,176 of 1,176 (1,168 before, plus 8); includes `games/conformance.test.ts`, `sandbox.test.ts`, both purity tests, the unchanged `stack.test.ts` |
| `npm run test:layers` | 11 of 11 |

No Go package outside `testdata/` was touched, so there are no other Go
tests to run.

## Core gaps checkers needs, and their status

From `examples/checkers/state.go`, `components.go`, `moves.go`, `main.go`.

| Need | Where checkers uses it | Status |
|---|---|---|
| `same(Color)` on `CapturedTokens` | `playerState` tag; `MoveToLastSlot` in `moveMoveToken.Apply` | **Done** (core) |
| `ComponentAtKey`, `ImmutableComponentAtKey` on a sized stack | `LegalCaptureSpaces`, `moveMoveToken.LegalCustom`/`Apply` | **Done** (core) |
| `SwapComponentsByKey`, `MaySwapComponentsByKey` | `moveMoveToken.Apply`; the `legal.MaySwapComponentsByKey` gate | **Done** (core); the legal predicate already existed (`legal/catalogStack.ts`) |
| `MayMoveToSlot` | `movePlaceToken.LegalCustom` | **Done** (core `mayMoveComponentToSlot`; `rules/binding.ts` had one for the evaluator) |
| `ContainingImmutableStack` | `token.Legal` | Existed: `EngineState.locateComponent` |
| Dynamic values (`Crowned`) read and written | `token.Dynamic`, `moveCrownToken.Apply` | Existed: `state.dynamicValues` |
| Component values (`Color`, an enum) | everywhere | Existed: `def.Decks.Tokens.Components` |
| `MoveTo`, `First`, `NumComponents`, `Len` | `movePlaceToken`, `GameScore` | Existed |
| A board keyed by an enum | — | **Not needed**: checkers' `Spaces` is a sized stack indexed by the range enum, not a `Board`. Go has enum boards (`board.go` `NewBoardForEnum`, `SpaceAtKey`, persisted `Enum`/`Keys`); no example game uses one. Not built |
| Range enum `spaces` (`RangeToValue`, `ValueToRange`, `RangeValue`, `enum.IllegalValue`, `RangeVal` move fields) | `components.go`, both moves | Other files: `lib/enumrange.ts` (workstream 9), move field types (`rules/`) |
| Grid graphs (`MustNewEnumGridConnectedness`, `NeighborsByKey`) | `AllNextSpaces` | Other files: `lib/graph.ts` (workstream 9) |
| `behaviors.PlayerColor`, `ComponentColor`, the color enum | state and components | Other files: `rules/stock.ts` (workstream 9) |
| `moves.FixUpMulti`, `moves.DefaultComponent`, `moves.CurrentPlayer`, `StartPhase` with `PhaseBehavior` | `ConfigureMoves` | Other files: `rules/stock.ts` (workstream 9); `startPhase` exists, `FixUpMulti` and `DefaultComponent` do not |
| `ComponentPresentAtKey`, `ComponentPropEqualsCurrentPlayer`, a game predicate (`checkers.spaceIsBlack`), `legal.Errorf` | `main.go` plans, `LegalCustom` | Exist in `legal/`; a game-registered predicate constructor is the loader's/translator's concern |
| `BOARD_SIZE` constant in the `sizedstack` tag | `gameState.Spaces` | A size, resolved by the exporter into `PropDef.Size`; not a core gap |

## Judgment calls

1. **Where the boot error surfaces.** Go fails when the manager boots. The
   engine has no boot; it fails when a state is built from the definition
   (`loadEngineState`, `copy`, anything that indexes), with the text Go's
   `NewStructInflater` gives for that property. The manager adds a further
   prefix (`Could not validate empty game state: `) that depends on the
   sub-state; it is not reproduced.
2. **`PropDef.Constraints` holds only constraint pieces**, trimmed, as the
   freeze says. A piece that is not `name(args)` is reported as Go's
   "unexpected field in struct tag". The size field is the exporter's to
   drop; `constraintsFromStackTag` shows the intended split.
3. **Constants.** A numeric argument may name a chest constant
   (`max(LIMIT)`). `GameDefinition` has no `Constants` field, so the parser
   reads an optional `Constants` from the definition object if one is ever
   added (typed read, no cast) and otherwise fails as Go does for an unknown
   name. No game in the repository uses a constant in a constraint.
4. **Values the engine cannot format.** Go compares properties by
   `fmt.Sprintf("%v", v)`. For a stack, board or timer property that prints
   a pointer; the engine refuses such a path at parse time with its own text
   (`... is a stack, board or timer, which the engine cannot compare`). No
   game does this.
5. **Reading component values.** Uses `DeckDef.ComponentProps` when present
   (Go's reader's property set), else the keys of the deck's JSON
   (`validatePropPath` on the first component, the value on each). The two
   agree on all vectors.
6. **`StackView.componentAtKey`** was added to the shared view interface, so
   merged views have it, as Go's `mergedStack.ImmutableComponentAtKey` does.
   Merged views still have no swap methods; Go's merged
   `MaySwapComponentsByKey` error stays the evaluator's/facade's to produce.
7. `constraints.ts` is not exported from `index.ts` (whose surface
   `purity.test.ts` pins); importers use the module directly.

## Unverified

- `strconv.Quote` is approximated by `JSON.stringify` in `goAtoi`'s errors;
  they differ only for non-ASCII or control characters in an argument.
- A constraint evaluated against a sanitized state's generic components
  reads the deck's `GenericValues` and no dynamic values; Go would read the
  generic values too and probably fault on dynamic ones. Not vectored (Go
  forbids modifying a sanitized state; only a `May...` call could get there).
- `goAtoi` beyond 2^53 is in range but inexact.

## Where the proposal was wrong about the code

- "boards keyed by an enum" are listed among the absent features that
  checkers needs; checkers has no board. Its enum-keyed access is a sized
  stack indexed by a range enum's keys.
- The proposal lists the constraint gap as "`same(...)`-style"; Go defines
  four kinds, and a transfer of several components into a constrained stack
  is checked by simulating it on a copy of the state, which matters
  whenever the components would each pass alone (every kind has such
  cases in the vectors).

## Edits outside my files

- `server/static/src/engine/purity.test.ts`: `'constraints.ts'` added to
  the expected list of core sources (one line).
- `server/static/src/engine/README.md`: a layout row for `constraints.ts`,
  two rows in the stack table (`MayMoveToSlot`, the by-key calls), and one
  paragraph on constraints.
- `server/static/src/engine/testdata/gen/regen.sh`: two comment lines naming
  the new output and `cgame`'s codegen.
- Not edited, for the owner of `definition.ts`: `PropDef.Constraints`'s doc
  comment still says "Not yet honored"; it now is.

## Files

New: `constraints.ts`, `constraints.test.ts`,
`testdata/constraints.vectors.json`, `testdata/gen/stackvec/constraints.go`,
`testdata/gen/stackvec/def.go`, `testdata/gen/stackvec/cgame/main.go`,
`testdata/gen/stackvec/cgame/auto_reader.go`. Changed: `stack.ts`,
`state.ts`, `testdata/gen/stackvec/main.go` (one call and a comment).
Regenerated: `testdata/stacks.vectors.json` (identical, so not in the diff).
`vecgame/` untouched.
