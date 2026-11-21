# Status: engine core (stream C, step 1)

Branch `claude/fabric-c-engine-core-2cb79c`, 2026-11-21. Nothing pushed.

The engine core is in `server/static/src/engine/`. It reproduces the Go
core's generator, ids, stacks, conservation and sanitization, and matches Go
on every known answer below, with one corner where Go has no single answer
(divergence 1). Its API is described in
[`server/static/src/engine/README.md`](../../server/static/src/engine/README.md).

Not built here, by design: moves, legality, phases, the fix-up loop, timers,
computed properties, stack constraints.

## Result

| Check | Result |
|---|---|
| `npm run type-check` | passes |
| `npm run test:unit` | 869 tests pass, 0 fail (787 before this work; 82 are the engine's) |
| `testdata/sanitize` fixtures | all 11 cases of Go's `TestSanitization` pass, read from the original patch files |
| Golden histories | 9 histories, 337 versions (memory 4 / 136, blackjack 5 / 201): ids, conservation, set-up shuffles, timer ids and every viewer's view match Go at every version |
| Go build | `go vet` clean on the generators; `go build ./...` and the repository-hygiene tests unaffected; no existing Go file changed |

### Which commands ran under which Node

`npm run test:unit` needs Node 23 or newer (it runs `.ts` files natively; CI
uses 24.x). This session did not install any toolchain.

| When | Runtime | Command (in `server/static`) | Outcome |
|---|---|---|---|
| Baseline | Node 20.20.0 | `npm ci` | ok |
| Baseline | Node 20.20.0 | `npm run type-check` | passes |
| Baseline | Node 20.20.0 | `npm run test:unit` | **could not run as written**: `Could not find 'src/**/*.test.ts'` (Node 20 neither expands the glob nor runs `.ts`). nvm then held only 18.11, 18.17, 20.20 and 21.7, none of which can |
| During development | Node 20.20.0 | `npm run type-check`, `npx tsc --noEmit` | passes at each commit |
| During development | Deno 2.5.6 (already on the machine) | `deno test -A --no-check src/engine/` | the engine's tests only, as a stand-in runner. Not the project's runner: no test budget, and none of the other 787 tests |
| Final | Node 24.22.3 | `npm run type-check` | passes |
| Final | Node 24.22.3 | `npm run test:unit` | 869 pass, 0 fail |
| Final | Node 24.22.3 | `node --test --test-timeout=20000 --import=./scripts/test-budget.mjs 'src/engine/**/*.test.ts'` | 82 pass, 0 fail |

Node 24.22.3 was found already present under `~/.nvm/versions/node` partway
through the work; this session did not install it and does not know who did.
The pre-existing count of 787 is the Node 24 total minus the engine's 82; the
baseline suite was never run on its own under a Node that can run it.

Go commands ran through `./scripts/go-local` with go1.27.4.

## Commits

| Commit | Contents |
|---|---|
| `eec42a96` | `rand`: `math/rand` v1 and FNV-1 64 ports; `u64` helpers; generators `cooked`, `randvec` |
| `d066619c` | `sha1`, component ids; the Go vector game and its harness; generator `idsvec` |
| `e75e2d11` | `definition`, `state`, `stack`: stacks, merged views, conservation; generators `defs`, `stackvec` |
| `7555a520` | `sanitize`: policies, groups, facets, `viewFor`; the `testdata/sanitize` harness; generator `sanitizevec` |
| `91af1b81` | Golden replay; generator `goldenvec` |
| `77584c46` | `gosort` (`sort.Sort`); Go's `null` `Indexes`; `index.ts`; purity guard; `regen.sh`; generator `sortvec` |
| (this commit) | README, this document, sanitized-state loading test |

Files touched outside `server/static/src/engine/`: this document only. No
existing file was edited; the engine's tests are picked up by the existing
`src/**/*.test.ts` glob and `tsconfig.test.json`.

## What each test proves

All counts are tests in `npm run test:unit`.

**`rand.test.ts` (15).** Against `rand.vectors.json`, for 60 seeds including
0, ±1, `MaxInt32`, `MaxInt32+1`, `MinInt64`, `MaxInt64`:

- the raw `Uint64` stream, 1,300 draws deep for eleven seeds (past two full turns
  of the 607-word register) and 20 deep for the rest;
- `Int63`, `Uint32`, `Int31`; `Int31n` for 14 moduli including powers of two
  and both ends of the range; `Int63n` for 12 moduli up to 2^52; `Intn` on both
  sides of 2^31;
- `Float64` and `Float32`, compared exactly;
- `Perm` and `Shuffle` for sizes 0 to 1,000;
- one generator used for interleaved `Perm`/`Intn`/`Float64`/`Int63` calls;
- FNV-1 64 for 79 strings (ASCII, multi-byte, astral), and that seeding the
  generator from the hash gives Go's `Perm`.

**`gosort.test.ts` (2).** Against `sort.vectors.json`, 286 inputs up to 3,000
elements with many equal keys: the exact permutation `sort.Sort` produces.
Asserts that most of these differ from a stable sort, so the vectors really
pin the unstable behavior.

**`ids.test.ts` (5).** SHA-1: the FIPS examples including one million `a`,
and `crypto/sha1` for every length 0 to 200. Component ids: 210 ids read from
live Go states after shuffles and secret moves, with secret-move counts up to
9, over 7 id/salt pairs including ones where the unseparated concatenation is
ambiguous (`"12"+"3"` versus `"1"+"23"`).

**`stack.test.ts` (12).** Against `stacks.vectors.json`: 20 scenarios of 50
operations each on live Go states, 19 kinds of operation (moves, secret
moves, first/last/next slot, slides, `MoveAllTo`, `MoveCountTo`, `Shuffle`,
`PublicShuffle`, swaps, three sorts, four resizes), about a third of them
refused. After every operation: every stack's JSON (physical and merged),
`SecretMoveCount`, Go's exact error text, no change on error, and
conservation. Also: a recorded state loads and serializes back identically;
several shuffles on one state continue one generator sequence while a copy
restarts it; Go's wording for each conservation violation.

**`sanitize.test.ts` (14).** Against `sanitize.vectors.json`: 176 cases over
3 states (2, 3 and 4 players), every kind of viewer, each of the five
policies on every game, player and dynamic property, struct-tag policies, and
random mixtures over the groups `all`, `other`, `self`, `same-color`,
`different-color` and the enum groups. 127 cases have one Go answer and
`viewFor` equals it exactly. 45 have several (divergence 1). 4 fail in Go and
fail in the engine. Also: Go's own parse of each struct tag; least
restrictive wins; the facet table; a Go view loads as a sanitized state and
serializes back unchanged.

**`sanitize-fixtures.test.ts` (14).** Reproduces
`sanitization_test.go`'s `TestSanitization`: a TypeScript port of
`internal/patchtree` composes `testdata/base.json` with the chain of
`modification.patch` files, the input loads as a state of Go's test game
(transcribed by hand, since it lives in `_test.go` files) with the test's
fixed id and salt, the case's policies are installed, and the sanitized JSON
is compared with the composed expected blob. All 11 cases, covering every
directory under `testdata/sanitize` and the three dynamic-value fixtures.
Because loading discards recorded ids, these fixtures also check id
computation against values recorded years ago.

**`golden.test.ts` (16).** For each golden history under
`examples/memory/testdata/golden` and `examples/blackjack/testdata/golden`,
at every version:

- the engine's own expansion of the golden file (full-state and jsondiffpatch
  delta encodings, including text patches) equals Go's, by digest;
- every recorded id of every physical stack recomputes from the recorded salt
  and counts (tens of thousands of slots), and a loaded state serializes back to the recorded
  JSON, which also checks every merged view;
- conservation holds;
- `viewFor` for admin, observer and each player equals Go's
  `Game.JSONForPlayer` view: by digest at every version (1,888 views), and as
  full JSON at sampled versions;
- every shuffle in the corpus is reproduced. All of them are in set-up.
  Blackjack: the recorded draw stack is exactly `Perm(52)` of the version-0
  generator. Memory: a replay of `FinishSetUp` (shuffle the unused cards,
  deal pairs by card set using `MoveToNextSlot` and `SlideToLastSlot`,
  shuffle the hidden row) reproduces the whole recorded version-0 state,
  including both shuffles drawn from one generator. A guard fails if a
  history ever records a shuffle after version 0;
- every recorded timer id in memory (17 of them) is the first 16 `Intn(16)`
  draws of that version's generator.

**`purity.test.ts` (4).** The engine's sources import only each other and use
no `BigInt`, `crypto`, timer, clock, `Math.random`, browser or Node global,
host API, asynchrony or dynamic code; the public surface is as documented.

## How each vector file was produced

Everything under `server/static/src/engine/testdata/` except `gen/` is output
of a Go program under `gen/`. Regenerate all of it with

```sh
server/static/src/engine/testdata/gen/regen.sh
```

| File | Generator | What Go code produced the values |
|---|---|---|
| `../rand-cooked.ts` | `gen/cooked` | The `rngCooked` table, read from `$GOROOT/src/math/rand/rng.go` (it is unexported) |
| `rand.vectors.json` | `gen/randvec` | `math/rand` and `hash/fnv` |
| `sort.vectors.json` | `gen/sortvec` | `sort.Sort` |
| `defs/*.def.json` | `gen/defs` | `vecgame.DefinitionFor(manager)`: property types and stack shapes from `manager.ExampleState()`, policies from `StructInflater.PropertySanitizationPolicy`, enums and decks from the chest |
| `ids.vectors.json` | `gen/idsvec` | `crypto/sha1`; `ComponentInstance.ID()` on live states |
| `stacks.vectors.json` | `gen/stackvec` | The public stack and component methods on a mutable copy of a loaded state; `state.StorageRecord()` after each |
| `sanitize.vectors.json` | `gen/sanitizevec` | `Game.JSONForPlayer`, 60 times per case, keeping each distinct outcome |
| `golden/*.golden.json` | `gen/goldenvec` | `record.New(file).State(v)` for states; `Game.JSONForPlayer(viewer, game.State(v))` for views |

`gen/vecgame` is a game written for this purpose: growable, capped and sized
stacks, both kinds of merged stack, a board, dynamic component values that
hold other components, enums, slices, a timer, and a group enum. Its
`auto_reader.go` is from `boardgame-util codegen`. Its delegate can override
the policy of a property per call, which is how one game exercises every
policy; the resolution over groups after that override is still Go's.

The generators reach states the way the server does: a blob is saved to an
in-memory store at a chosen version and loaded through `manager.Game(id)`.
A loaded state is copied with `Copy(false)` and asserted to
`boardgame.State` to get a mutable one with the real game id, salt and
version, so ids and `state.Rand()` are the real ones.

Run twice, `regen.sh` leaves every file identical except possibly
`sanitize.vectors.json` (see divergence 1: the set of outcomes it records for
an order-dependent case depends on which map orders Go happened to take in
60 runs).

## Divergences and ambiguities found in the Go code

Numbered so they can be referred to. "Reproduced" means the engine does what
Go does and a vector proves it.

**1. Sanitization of dynamic component values depends on Go map order, and
can panic.** `sanitizeStateObj` ranges over `Props()`, a map
(`sanitization.go:383`). For each visible stack it marks that stack's
components as having visible dynamic values (`sanitization.go:398-421`). A
visible *merged* stack is read through its parts, which may or may not have
been sanitized yet in this iteration order. So for a visible merged stack
over non-visible stacks of a deck that has dynamic values, one input has
several outcomes:

- merged stack examined before its parts: the dynamic values of every
  component the view shows are revealed, although the parts are hidden;
- examined after, parts `hidden`: nothing is revealed;
- examined after, parts `order`/`len`/`nonempty`: the view yields the generic
  component, index -2 is recorded as visible, and
  `transativelyMarkDynamicComponentsAsVisible` panics with
  `index out of range [-2]` (`sanitization.go:467`).

Observed in 45 of 176 vector cases. No game in this repository is affected:
memory and blackjack have merged stacks but no dynamic values.

**The engine cannot match this**, because there is no single behavior. By
default it never reveals dynamic values through a merged stack (the
"examined last" answer without the panic), which is the conservative choice
for a system whose point is hiding state.
`mergedStacksRevealDynamicValues: true` gives the "examined first" answer.
Tests: with the option, the engine's view is one of Go's recorded outcomes in
all 45 cases; without it, the view equals Go's outside `Components` in all
45, and is itself one of Go's outcomes wherever Go has a non-revealing one.

**2. A non-visible policy on a merged stack panics.** `applyPolicy` asserts
its input to `Stack` (`sanitization.go:579`), which `*mergedStack` does not
implement: `interface conversion: *boardgame.mergedStack is not
boardgame.Stack`. The engine throws `SanitizeError` instead. In practice a
merged stack is always left visible and hidden through its parts.

**3. `sizedStack.ContractSize` leaves the component index stale.** It removes
empty slots, shifting later components left, without calling
`updateIndexForAllComponents` (`stack.go:2561`; compare
`growableStack.removeComponentAt`, `stack.go:2007`). If the state's lazily
built component index already exists, a shifted component then cannot be
moved in the same state: `containingStack` finds nothing at the old slot and
returns "that component's location is not public information". A fresh copy
of the state (the next move) is fine. **Reproduced**: the engine ports the
component index literally (`EngineState.locateComponent`), so it refuses the
same operations with the same text. Found by the stack vectors.

**4. Resizing one part of an overlap makes state serialization panic.**
`mergedStack.IDs` indexes each part by the first part's length
(`stack.go:1274`). After `ExpandSize`/`ContractSize` on one of two overlapped
sized stacks, marshalling the state panics with `index out of range`. Not
reproduced: the engine serializes (treating the missing slot as empty), and
`conservationError()` reports the view as invalid, as Go's commit-time
validation would. The stack vectors avoid resizing those stacks, since Go
has no answer to record.

**5. An empty concatenation serializes `"IDs": null`.** `mergedStack.IDs`
appends to a nil slice (`stack.go:1283`), so a concatenated view with no
slots marshals `null`, where every other stack marshals `[]`. **Reproduced**
(blackjack's `Hand` is such a view at the start of each round).

**6. Sorting an empty stack leaves `"Indexes": null`.** `sortComponentsImpl`
copies with `append([]int(nil), stack.indexes...)` (`stack.go:2122`,
`:2134`), which stays nil for an empty stack; the stored state then carries
`null` until a component is inserted. **Reproduced**; this is why
`StackJSON.Indexes` is `number[] | null`.

**7. `SortComponents` is not stable but is deterministic.** It uses
`sort.Sort`. For elements that compare equal, the result is fixed by Go's
pdqsort, including its xorshift "pattern breaking". **Reproduced** by porting
`sort.Sort` (`gosort.ts`). This means a ported game's `less` must return the
same answers as its Go original for every pair, not merely induce the same
ordering of distinct keys.

**8. A merged stack's `IDsLastSeen` drops ids last seen at version 0.**
`if val > result[key]` with a zero default (`stack.go:1332`) never admits a
value of 0. **Reproduced** (visible in every golden history's version 0).

**9. Under `len`, a sized stack's ids and indexes stop lining up.** Ids are
sorted, which moves every `""` (empty slot) to the front, while `Indexes`
keeps the real occupancy (`sanitization.go:827`, `:766`). So `len` on a sized
stack still reveals which slots are occupied. **Reproduced**; noted because
"only the length is visible" is not quite what `len` does there.

**10. Two components that hold each other would hang sanitization.**
`transativelyMarkDynamicComponentsAsVisible` appends to its work list without
checking for repeats (`sanitization.go:485`). Conservation permits card A in
B's dynamic stack and B in A's. The engine visits each component once and
terminates. Not exercised by a vector, since Go does not return.

**11. A hidden `PlayerIndex` becomes 0,** a real player, not the observer
(`sanitization.go:548`). Reproduced; noted because 0 is indistinguishable
from "player 0".

**12. A player state is judged by its own player's groups.** For enum-valued
group names, the membership passed to the policy for `Players[i]` is player
i's (`sanitization.go:268`), not the viewer's. So `Red:visible` on a player
property means "visible to everyone if this player is Red". On the game
state the same tag means "visible if the viewer is Red". Reproduced; both
are covered by vectors.

**13. Timer ids come from `state.Rand()`.** `timer.Start` draws 16 values
(`timer.go:148`). A move that starts a timer and also shuffles must do so in
Go's order, or everything after diverges. The engine has no timers yet; the
golden test confirms the draw for every recorded id.

**14. `randPermForStack` (`sanitization.go:780`) is dead code** and is not
ported.

### Limits of the port

- `GoRand.int63n` and `intn` accept n up to 2^52 and throw beyond it, to stay
  exact without `BigInt`. Go accepts any positive `int64`. Nothing in the
  core, `moves/`, `components/` or `examples/` passes a value near that
  (callers are `Perm` in `stack.go`, `Intn` in `components/dice`,
  `enum.RandomValue`, `game.go` `randomString`, `examples/pig`, and `Perm` in
  `examples/werewolf`).
- `Rand.Int`, `NormFloat64`, `ExpFloat64` and `Read` are not ported; nothing
  uses them.
- Stack constraints are not ported: `CheckConstraints`
  (`component.go:542`) and the copy-and-simulate path of
  `validateComponentTransferPlan` (`stack.go:1610`). The engine behaves as Go
  does for stacks with no constraints.
- Sanitizing an already sanitized state (the facet intersection in
  `state_visibility.go:44`) is not ported; `sanitizedForPlayer` refuses.
- `viewFor` omits `Computed`. The golden views were compared with `Computed`
  removed on the Go side.
- When several properties have an invalid policy, which one Go's error names
  depends on map order. The engine reports the first in sorted order.
- Where Go's component-centric API cannot express a call at all (moving the
  component at a slot that has none), the engine throws a `StackError` whose
  text is its own.

## Open questions

1. **Divergence 1 needs a decision.** Is the engine's default (never reveal
   dynamic values through a merged stack) the intended semantics? If so Go
   should be fixed to match and the option removed. The exporter planned for
   `boardgame-util` will hit the same nondeterminism for any game with both
   features.
2. **Should Go's `ContractSize` be fixed (divergence 3)?** The engine
   reproduces the defect today. If Go is fixed, the literal index port can be
   replaced by a scan.
3. **`Computed`.** Views cannot match Go in full until the engine has computed
   properties; that belongs with game rules. The exporter's comparison needs
   to either include them then or keep ignoring them.
4. **Size of the vectors.** About 7.6 MB of JSON is checked in under
   `testdata/` (`sanitize` 3.5 MB, `stacks` 1.3 MB, `golden` 1.3 MB, `rand`
   0.9 MB, `sort` 0.4 MB). They can be thinned further, or generated in CI
   instead of committed. Whether `boardgame-util build static` copies
   `src/engine/testdata` into build output was not checked.
5. **Stack constraints.** Not in this task's list. If a ported game uses
   them they need porting with the move layer.
6. **Go 1.27's `encoding/json` is v2 underneath.** The vectors reflect it.
   Structural comparison makes key order irrelevant, but a `null`-versus-`[]`
   difference (divergences 5 and 6) is exactly the kind of thing a different
   Go version could change.

## For whoever builds the move layer on this

- Apply a move as `next = current.copy(current.version + 1)`, mutate `next`,
  then `next.conservationError()`. That is `applyMove` in `game.go`.
- Draw only through `next.rand()`, in Go's order. Set-up runs at version 0.
- `state.game` and `state.players[i]` are the JSON; enums are their strings.
- `testdata/defs/memory.def.json` and `blackjack.def.json` are ready-made
  definitions, and `golden.test.ts` shows memory's `FinishSetUp` replayed on
  the stack API.
- `testing/golden-record.ts` reads a golden history into per-version states
  and moves.
