# engine

A TypeScript reproduction of the Go core's game-state semantics, held to the
Go implementation by known answers that real Go produced.

This directory is the engine **core**: the generator, ids, stacks, and
sanitization. Moves, legality, phases and the fix-up loop are not here yet;
`legal/` belongs to a separate module. The plan this implements is
`fabric/PROPOSAL.md`, "The engine: reproducing the core in TypeScript". What
has been verified, how, and every place Go's behavior was surprising is in
`fabric/status/engine-core.md`.

## Constraints

The engine runs inside a sandboxed JavaScript handler. Every file directly in
this directory (not `testing/`, not `*.test.ts`) is pure TypeScript:

- no DOM, no Node APIs, no `crypto`, no timers, no clock, no `Math.random`;
- no `BigInt`, no `TextEncoder`, no asynchrony;
- no dependencies: engine modules import only each other.

`purity.test.ts` enforces this by reading the sources.

State is plain JSON in exactly the shape the Go server serializes, so a
recorded Go state loads directly, and what the engine serializes is what Go
would have.

## Layout

| File | What it is | Go source it ports |
|---|---|---|
| `u64.ts` | 64-bit integers as uint32 pairs; UTF-8 encoding | — |
| `fnv.ts` | 64-bit FNV-1 | `hash/fnv` `New64` |
| `rand.ts`, `rand-cooked.ts` | `math/rand` v1: the generator and `Int63`, `Uint32`, `Int31`, `Int31n`, `Int63n`, `Intn`, `Float64`, `Float32`, `Perm`, `Shuffle`; and the seed the core derives for a state | `math/rand`; `state.go` `(*state).Rand` |
| `sha1.ts` | SHA-1 | `crypto/sha1` |
| `gosort.ts` | `sort.Sort` (pdqsort), including its order among equal elements | `sort` |
| `ids.ts` | Component ids; the empty-slot and generic-component sentinels | `component.go` `ID`, `secretMoveCount` |
| `definition.ts` | `GameDefinition`: decks, enums, property schema and sanitization policies; struct-tag parsing | `struct_inflater.go`, `sanitization.go` `policyFromString` |
| `stack.ts` | `Stack` (growable and sized) and `MergedStackView` | `stack.go` |
| `state.ts` | `EngineState`: the state JSON plus identity, version, generator, stack handles, conservation | `state.go`, `stack_ownership.go`, `game_manager.go` `stateFromRecord` |
| `sanitize.ts` | Policies, groups, visibility facets, `sanitizedForPlayer`, `viewFor` | `sanitization.go`, `state_visibility.go`, `base/game_delegate.go` |
| `index.ts` | The public surface | |
| `testing/` | Test support that uses Node: a port of `internal/patchtree`, and a reader for golden history files | |
| `testdata/` | Known answers from Go, and the Go programs that produce them (`testdata/gen/`) | |

## What a game definition provides

A `GameDefinition` (`definition.ts`) is JSON. It is what Go discovers by
reflection over struct tags:

```ts
const def: GameDefinition = {
  Name: 'memory',
  Decks: { cards: { Size: 120, Components: [/* static values, by deck index */] } },
  Enums: { phase: { Values: { 0: 'Gathering', 1: 'Playing' }, DefaultValue: 0 } },
  GroupEnum: 'group',            // optional: the enum whose values are player groups
  Game: { Props: {
    CurrentPlayer: { Type: 'playerIndex' },
    HiddenCards:  { Type: 'sizedStack', Deck: 'cards', Size: 40, Sanitize: 'order' },
    VisibleCards: { Type: 'sizedStack', Deck: 'cards', Size: 40 },
    Cards:        { Type: 'mergedStack', Deck: 'cards', Overlap: true, Stacks: ['VisibleCards', 'HiddenCards'] },
  } },
  Player: { Props: {
    Hand: { Type: 'stack', Deck: 'cards', Sanitize: 'other:len' },
  } },
  DynamicComponentValues: {},    // deck name -> schema of per-component dynamic values
};
```

`Sanitize` takes the same string a Go `sanitize:"..."` struct tag takes, or
the parsed map (`{ other: 'len' }`). Property types are `int`, `bool`,
`string`, `playerIndex`, `enum`, the four scalar slices, `enumSlice`, `timer`,
`stack`, `sizedStack`, `mergedStack` and `board`.

You do not have to write one by hand for a game that exists in Go:
`testdata/gen/defs` derives the definition from the game's live `GameManager`
(`vecgame.DefinitionFor`), and `testdata/defs/` has the results for memory and
blackjack.

## What a host does

```ts
import { EngineState, viewFor, OBSERVER_PLAYER_INDEX } from './index.ts';

// The game's identity is the only secret input: ids and every random draw
// are derived from it.
const identity = { gameId, secretSalt };

// Load a state the Go server stored (or one the engine wrote earlier).
const current = EngineState.load(def, identity, version, stateJSON);

// To apply a move: copy at the next version, mutate, check, store.
const next = current.copy(version + 1);
const draw = next.gameStack('DrawStack');
draw.moveComponentToNextSlot(0, next.playerStack(0, 'Hand'));
next.gameStack('DiscardStack').shuffle();        // draws from next.rand()
if (next.conservationError() !== null) throw new Error('...');
const stored = next.toJSON();                    // what Go would persist

// What each viewer may see.
const forPlayer0 = viewFor(next, 0);
const forObservers = viewFor(next, OBSERVER_PLAYER_INDEX);
```

### State

- `EngineState.load(def, identity, version, json, options?)` copies the JSON,
  recomputes every id (recorded ids are ignored, as in Go), and throws if
  component conservation fails. `{ sanitized: true }` loads a per-viewer view
  instead: ids are kept as given and nothing can be modified.
- `state.copy(version?)` is Go's `state.copy(false)`. Passing a version is
  what `applyMove` does before applying a move.
- `state.rand()` is `state.Rand()`: seeded from game id + salt + version and
  memoized, so all draws while applying one move continue one sequence.
  **Anything that draws must draw in the order Go does.** Note that starting a
  timer draws 16 values for its id.
- `state.game`, `state.players[i]`, `state.dynamicValues(deck, index)` are the
  sub-states' JSON objects. Scalars, slices and enums (stored as their string
  values) are read and written there directly.
- `state.toJSON()` is the storage record: a copy, merged views recomputed,
  `SecretMoveCount` present only when non-empty.
- `state.conservationError()` returns Go's error text, or `null`.

### Stacks

`state.gameStack(prop)`, `state.playerStack(player, prop)`,
`state.dynamicStack(deck, index, prop)`, `state.gameBoardSpace(prop, space)`
(or `state.stack(ref)`) return a `Stack`: a handle over the JSON, not a copy.
`state.view(ref)` returns a read-only `StackView` for any stack property,
including merged ones.

Go's moves are methods on the component (`stack.ComponentAt(i).MoveTo(...)`).
Here they are methods on the stack that take the component's slot:

| Go | Engine |
|---|---|
| `c.MoveTo(dest, slot)` | `stack.moveComponent(i, dest, slot)` |
| `c.SecretMoveTo(dest, slot)` | `stack.secretMoveComponent(i, dest, slot)` |
| `c.MoveToFirstSlot/LastSlot/NextSlot(dest)` | `stack.moveComponentToFirstSlot/LastSlot/NextSlot(i, dest)` |
| `c.SlideToFirstSlot/LastSlot()` | `stack.slideToFirstSlot/LastSlot(i)` |
| `c.MayMoveTo(dest)` | `stack.mayMoveComponentTo(i, dest)` (error text or `null`) |
| `s.MoveAllTo(dest)`, `s.MoveCountTo(dest, n)` | `stack.moveAllTo(dest)`, `stack.moveCountTo(dest, n)` |
| `s.MayMoveAllTo`, `s.MayMoveCountTo` | `stack.mayMoveAllTo`, `stack.mayMoveCountTo` |
| `s.Shuffle()`, `s.PublicShuffle()` | `stack.shuffle()`, `stack.publicShuffle()` |
| `s.SwapComponents(i, j)` | `stack.swapComponents(i, j)` |
| `s.SortComponents(less)` | `stack.sortComponents(less)` (`less` gets deck indexes) |
| `s.ExpandSize/ContractSize/SetSize/SizeToFit` | `stack.expandSize/contractSize/setSize/sizeToFit` |

A failed operation throws `StackError` with Go's message and changes nothing.

A component is a `(deck name, deck index)` pair. `stack.componentAt(i)`
returns the deck index, `GENERIC_COMPONENT` (-2) in a sanitized stack, or
`null` where Go returns nil. Static values are `def.Decks[deck].Components[i]`.

### Views

- `viewFor(state, viewer, options?)` is the JSON a viewer receives as the
  current state (Go's `Game.JSONForPlayer(...).CurrentState`), with
  `Visibility`. `viewer` is a player index, `OBSERVER_PLAYER_INDEX`,
  `ANY_PLAYER_INDEX` (treated as the observer) or `ADMIN_PLAYER_INDEX`.
  **The admin view is the unsanitized state, `SecretMoveCount` included; never
  release it.**
- `sanitizedForPlayer(state, viewer)` returns the sanitized `EngineState`.
- `Computed` is not produced: computed properties are game rules. Pass
  `options.computed` to supply them.
- `options.mergedStacksRevealDynamicValues` selects between the two answers Go
  can give in one corner; see the status document.

## Tests

`npm run test:unit` (Node 23 or newer, as for the rest of the suite) runs
everything. Each `*.test.ts` here says at its top what it proves.

The expected values are never hand-derived. They come from Go:

```sh
server/static/src/engine/testdata/gen/regen.sh
```

regenerates every vector file and `rand-cooked.ts`. The generators live under
`testdata/gen/`; a directory named `testdata` is skipped by `go build ./...`
and `go test ./...`, so they are not part of the repository's Go build.
`testdata/gen/vecgame` is a game definition written only to give Go one of
every kind of property to serialize.
