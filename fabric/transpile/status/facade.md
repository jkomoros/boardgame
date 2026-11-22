# Workstream 7: the facade

`server/static/src/engine/go/` now implements every function the freeze
stubbed, with Go's shapes and semantics, over the existing engine. One
function still throws, with a message starting `engine-gap:` (`stockApply`).
Every expected value in the tests comes from running Go
(`testdata/gen/facadevec`). Base: `d2cb55e4`.

## What was built

| File | What |
|---|---|
| `go/panic.ts` | Go's runtime panic texts and throwers: nil dereference, divide by zero, index out of range (`[i] with length n`, `[i]` when negative), slice bounds (all four of the runtime's forms, checked in its order), nil map write, and (new) `makeslice: len out of range` |
| `go/num.ts` | `idiv` (truncation, divide-by-zero panic), `imod`, `imul` (throws `go/num: integer overflow: ...` when the product is not a safe integer). No result is ever `-0` |
| `go/slice.ts` | Nil-aware slices: `appendTo` (a new array, so an alias keeps its old length as in Go; appending nothing to nil stays nil), `lenOf`, `at`/`setAt` (Go's index panic), `makeSlice` (Go's panic for a negative length), `copyInto` (overlap-safe), `subslice` (Go's bounds panics; `nil[0:0]` stays nil). Capacity is the length |
| `go/str.ts` | `join`, `repeat` (Go's negative-count panic), `split`, `toLower` (rune by rune, no context: `ΟΔΟΣ` gives `οδοσ`, `İ` gives `i`), `trimSpace` (Go's white space: U+0085 yes, U+FEFF no), `hasPrefix`, `hasSuffix`, `contains`, `strCompare` (-1/0/1 by code point), `utf8Len`, `itoa`, `atoi` (Go's scan order, syntax and range texts, the int64 bound as the value on a range error), `sprintf`, and (new) `quote` (`strconv.Quote`) and `quoteRune` |
| `go/api.ts` | Stacks and components in Go's shape (below), enums, range enums, grid graphs, behavior setters, `variantValue`, `legalError`, `stockApply` (gap) |

No file outside the brief's list was edited. `testdata/gen/regen.sh` gained
`facadevec`.

### The stack facade

A `Comp` is used the way Go uses a `ComponentInstance`: by identity (deck,
index, and the state it was read from), never by slot. Every operation finds
the component where it is now with the engine's `locateComponent` (Go's
`state.containingStack`) and then calls the engine's own slot-based method on
the stack it found, inside a `try` that turns a `StackError` into the
returned text. Any other exception propagates. So:

- `moveTo`, `secretMoveTo`, `moveToFirstSlot/LastSlot/NextSlot`,
  `slideToFirstSlot/LastSlot` return `string | null` with Go's text and
  change nothing on failure; a null component throws Go's nil-dereference
  text, as calling a method on a nil interface does.
- `mayMoveTo` reproduces `componentInstance.MayMoveTo`'s first checks in Go's
  order (different states; a merged destination is "destination stack is not
  attached: stack is not a physical stack"; the component's location), then
  delegates to the engine's `mayMoveToError`. `mayMoveToSlot` adds Go's slot
  texts.
- `moveAllTo`, `moveCountTo`, `shuffle`, `publicShuffle`, `swapComponents`,
  `swapComponentsByKey` (Go's is `SwapComponents(int(i), int(j))`),
  `sortComponents` (`less` over `Comp | null`), `setSize`, `expandSize`,
  `contractSize`, `sizeToFit`: the engine's methods, errors returned.
  `mayMoveAllTo`/`mayMoveCountTo` give Go's texts for a merged source ("... is
  not supported on MergedStacks") or destination ("destination is not a
  physical stack").
- `componentAt`, `componentAtKey` (Go's is `ComponentAt(int(k))`), `first`,
  `last` (Go's sized/growable/merged rules), `components` (a snapshot; null
  for an empty slot).
- `containingStack(state, c)`: `[StackRef | null, slot, err]`, the stack
  rebuilt as a facade handle from its owner path.
- `values(state, c)`: static values; for a generic component, the deck's
  generic values from `DeckSpec.generic` or `DeckDef.GenericValues`.
  `dynamic(state, c)`: the values of the state the component was read from; a
  generic component gets a fresh zero struct, as Go's `DynamicValues` does.
- A Go call that takes a `Stack` given a merged stack did not compile in Go;
  the facade throws `go/api: ...: ... is a merged stack, which Go cannot pass
  here`.

Operations that Go code cannot express on a merged stack (moving into one,
shuffling one) throw; reads of one work.

### Enums, ranges, graphs, behaviors

- `enumKey`, `setEnumKey` (Go's `That value is not valid for this enum`),
  `enumEquals` (same enum by name, same key), for owners `'game'`, a player
  index, `{ dynamicOf }`, `{ valuesOf }`. The enum is found from the
  definition's `PropDef.Enum` (or the deck's component props). A property
  that is absent (Go's nil `enum.Val`, e.g. in a generic component's zero
  dynamic values) throws the nil-dereference text.
- `rangeValue`, `rangeToValue`, `valueToRange`: implemented here from the
  stored strings and `EnumDef.Values` alone, exactly as `enum/range.go` does
  it (`keyForIndexes`, `indexesForKey`, `ValueFromString`). `ILLEGAL_ENUM_VALUE`
  is exported.
- `neighborsByKey(graph, key)`: implemented from `GraphRef.dimensions` and the
  Go direction names (`DirectionUp`, ... `DirectionDiagonal`), as
  `NewGridConnectedness` builds the graph.
- `setEliminated`, `setPlayerInactive`, `setSeatFilled` write what the
  behaviors write (`SetSeatFilled` sets `SeatClosed` too).
- `variantValue` is an own-property lookup, so `'constructor'` is `''`.
- `legalError(key, bindings)` is `failT(key, bindings)`, bindings omitted
  when absent as Go leaves them nil.
- Timers: the table maps `Active`, `Start`, `Cancel` to the engine's
  `TimerHandle` unchanged, so the facade has no timer function.

## Functions still throwing `engine-gap:`

| Function | What it needs, from which file |
|---|---|
| `stockApply(state, m)` | The loader: `rules/data.ts` `gameFromData` must resolve each `MoveData.Stock` through `rules/stockRegistry.ts` `stockMove(type, config)` and make that move's `StockParts.apply` reachable from the `MoveInstance` (or the compiled move) so the facade can call it and return its error as `string \| null` |

Not a gap, but a dependency to wire: **constraint-aware moves.** The facade
delegates every validity check after Go's first few to the engine
(`Stack.mayMoveToError`, `moveComponent*`, `transferError`). When the stack
stream adds `CheckConstraints`, the facade inherits it only if the check is
made inside those methods (as Go's `MayMoveTo` ends with
`dest.CheckConstraints`). If it lands elsewhere, `mayMoveTo` needs one line.
`componentAtKey`/`swapComponentsByKey` here are Go's own definitions and need
nothing from the parallel `stack.ts` work; if that work adds engine methods
of the same name they should agree (a key is a slot).

## Signatures changed from the freeze

| Signature | Change | Why |
|---|---|---|
| `rangeValue(...)` | returns `number[] \| null` (was `number[]`) | Go's `RangeValue` returns nil for a string that is not a range value |
| `valueToRange(...)` | returns `number[] \| null` | Go returns nil for a key the enum lacks |
| `neighborsByKey(...)` | returns `number[] \| null` | Go returns nil (never empty) for a key with no neighbors or not in the enum |
| `Owner` | `{ dynamicOf: Comp \| DynamicRef \| null } \| { valuesOf: Comp \| null }` | `c.DynamicValues().Mark` with `c` nil is a nil-dereference panic in Go; the translator passes a `Comp \| null` |
| `sliceBoundsText` | returns `''` for legal bounds | lets `subslice` share it |
| new exports | `ILLEGAL_ENUM_VALUE` (api), `MAKESLICE_LEN_TEXT` (panic), `INT_OVERFLOW_PREFIX` (num), `REPEAT_NEGATIVE_TEXT`, `quote`, `quoteRune` (str) | Go names the translator or a test needs |
| `Comp.slot` | documented as informational; `-1` for a component handed to `less` | Go's `ComponentInstance` has no slot; operations use identity |

## Tests

| Test | What it proves | Count |
|---|---|---|
| `go/lang.test.ts` | Against `testdata/facade.vectors.json`: 72 `idiv`/`imod` pairs over signs and zero (panic texts recovered from Go), 100 `imul` pairs (exact ones equal, others throw), 40 `atoi` inputs (texts, values, Go's scan order, the documented divergence for 6 non-safe int64s) and 8 `itoa`, 324 `strCompare` pairs (including non-BMP, where JavaScript's `<` disagrees), 18 `utf8Len`, 50 `sprintf` calls (verbs, flags, widths, slices, bad-verb/MISSING/EXTRA/NOVERB markers, floats), 72 `strings` calls, 22 runtime panics, 17 nil-aware slice programs serialized, 3 behavior setters, 4 range enums (every key and index tuple, in and out of range), 22 grid graphs (8x8 and 3x4, 11 direction sets, every key including -1 and n); plus a unit test of `variantValue`, `legalError`, `player`, `stockApply` | 12 tests |
| `go/stacks.test.ts` | Against `testdata/facade-stacks.vectors.json`: 13 scenarios, 934 facade calls on live vecgame states (36 kinds), replayed through the facade; after every call the result (error text, component, list, stack and slot) and the panic text match Go, and the whole flattened state matches, so every failure changed nothing. Includes a scripted prefix: a snapshot handle used after its component moved twice (Go moves it from where it is), a stale index after `ContractSize` (`that component's location is not public information` from `containingStack`, `moveTo`, `mayMoveTo`), nil handles in every method, merged stacks as source and destination. 137 calls panic in Go and 178 return an error | 14 tests |
| `go/purity.test.ts` | The facade's five modules use no platform API, BigInt, class, clock or randomness, import only engine modules, and nothing still says `not implemented`; any `engine-gap:` message names what it needs | 4 tests |

Mutation check, by hand: making a `Comp` slot-based (fail when the slot no
longer holds it) fails 8 of the 14 stack tests; breaking one direction filter
fails the graph test.

Runs (Node 24, Go 1.27.4 through `./scripts/go-local`):

- `./scripts/go-local build ./...`: passes. `go vet` of the generator: clean.
- `./scripts/go-local test ./boardgame-util/lib/conformance/...`: ok (both
  packages). `conformance export --check`: "The conformance corpus is current".
  No other Go package was touched (the generator lives under `testdata/`, which
  `go build ./...` skips).
- The generator is deterministic: two runs, identical SHA-1 of both files.
- `npm run type-check`: passes. `npx tsc --noEmit -p
  src/engine/tsconfig.sandbox.json`: passes.
- `npm run test:unit`: 1,199 of 1,199: the base's 1,168, plus 30 facade
  tests (12 + 14 + 4), plus 1 that node's runner counts for
  `go/testkit.test.ts`, a support file with no tests of its own.
- `npm run test:layers`: 11 of 11.
- `sandbox.test.ts` scans `go/` (it recurses) and passes.

## Judgment calls

1. **Identity, not slot.** See "The stack facade" and the next section.
2. **`atoi` beyond 2^53.** Go accepts any int64; the facade reports a value
   that is a valid int64 but not a safe integer as out of range, the same
   divergence `legal/goformat.ts` `goAtoi` documents. On a true range error
   the value is Go's bound (`math.MaxInt64`/`MinInt64`, which read as ±2^63).
   `str.ts` does not reuse `goAtoi` because it uses `BigInt`, which the engine
   rules forbid (`legal/` is outside the purity scans).
3. **`quote` is new** rather than `goformat.ts` `goQuote`, which passes
   non-printable non-ASCII through; Go escapes it (` `, `​`,
   ` `), and the vectors show it. Printability is JavaScript's
   `\p{L}\p{M}\p{N}\p{P}\p{S}` plus space, which is Go's `IsPrint` definition;
   the Unicode versions may differ at the margins.
4. **`sprintf` types.** A number is an `int` when it is an integer and a
   `float64` otherwise; Go's static type is not visible. So an integral
   `float64` prints as an int (Go: `1e+06`), and `null` prints `<nil>` (a nil
   interface or error), where a nil slice would print `[]` in Go. The vectors
   avoid both. Verbs other than `d s v q` that fmt knows (`%x`, `%f`, ...)
   throw rather than print something wrong; unknown letters print Go's
   `%!z(int=1)`. `%+v` drops the plus, `%06s` pads with zeros, as in Go.
5. **`neighborsByKey` order.** Go returns the keys in a map's iteration
   order, which changes run to run; the facade returns them ascending, and
   the vectors compare sorted lists. Go code whose result depends on the
   order has no single answer.
6. **`values` of a generic component with no generic values** throws its own
   text: Go's `c.Values().(*T)` panics with an interface-conversion message
   naming the game's type, which the facade cannot know.
7. **`dynamic` of a deck with no dynamic values** throws its own text, for the
   same reason (Go's type assertion on a nil `SubState`). The stack vectors
   ask only cards for dynamic values.
8. **The Go vectors drive the TypeScript through a stub `CompiledGame`**
   (`go/testkit.test.ts`): the vecgame definition, its enums and its decks'
   values, with no `GameSpec`. The facade reads nothing else from it.

## Where the proposal was wrong about the code

- **"A handle whose component has since moved fails as Go's does, with the
  engine's 'not public information' text."** It does not fail in Go. A
  `ComponentInstance` is a component and a state; `MoveTo` and the rest look
  the component up where it is now. Memory's capture loop works because of
  this. What fails is a component the state's index has lost
  (`sizedStack.ContractSize` does not refresh it) or a generic one. The
  scripted vectors show both.
- "`g.NeighborsByKey(k)` ... in Go's order": the order is a Go map's.
- `enum.IllegalValue` is `math.MaxInt64`, not a safe integer.
- `behaviors.Seat.SetSeatFilled` also sets `SeatClosed`; a property write of
  `SeatFilled` alone, as the table's "property writes" suggests, would differ.
- The table lists timers as a facade row, but every timer call is "the same"
  as the engine's, so there is nothing to write.
- `legal/goformat.ts`, which the brief suggests reusing, uses `BigInt`.

## Unverified

- `toLower` and `quote`'s printability beyond the vectors' characters depend on
  the JavaScript engine's Unicode tables.
- `values` of a generic component and `dynamic` of one: no Go vector (vecgame
  sets no generic values, and the stack vectors do not run on sanitized
  states).
- The facade on a sanitized (per-viewer) state: covered only by the engine's
  own refusal texts, not by a facade vector.
- Boards: there is no facade call for a board space (`Spaces.SpaceAt(i)` is
  not in the API table, and `gameStack` refuses a board property). A
  `StackRef` whose `at.Space` is set works, and the vectors use board spaces
  that way.
