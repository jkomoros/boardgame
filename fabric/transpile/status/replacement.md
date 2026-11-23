# Replacing the hand ports with the generated games

Base `d66bbed7` (branch `claude/transpile-pipeline-e23924`). One local
commit; nothing pushed. The owner ruled that a generated game is the one in
use wherever it is fully semantically equivalent to its hand port. This
stream established that for all six hand-ported games, then made the
generated games the ones in use and removed the six hand ports
(`server/static/src/engine/games/{memory,blackjack,tictactoe,pig,
debuganimations,werewolf}.ts`).

**The comparison below is a one-time record.** It compared files that no
longer exist; no test can repeat it. What stays tested is each generated
game against Go (its corpus, the rejection vectors, the engine tests), which
is what the hand ports were held to.

## Step 1: equivalence, game by game

All six are equivalent. None was kept.

| Game | Structural differences (compiled / spec) | Unexplained | Conformance + fuzz corpus (generated) | Rejection vectors | Engine tests | Fresh transient fuzz, hand port and generated |
|---|---|---|---|---|---|---|
| memory | 21 / 76 | 0 | 14 documents, 441 versions, 18 rejections | 9 of 9 | pass | 24 of 24 and 24 of 24 |
| blackjack | 32 / 161 | 0 | 15, 946, 9 | 9 of 9 | pass | 32 of 32 and 32 of 32 |
| tictactoe | 17 / 41 | 0 | 12, 192, 5 | 27 of 27, and both fix-up-cap cases (on tictactoe) | pass | 1,000 of 1,000 and 1,000 of 1,000 |
| pig | 19 / 65 | 0 | 11, 696, 25 | 2 of 2 | pass | 100 of 100 and 100 of 100 |
| debuganimations | 32 / 114 | 0 | 9, 184, 22 | none exist | not covered (none before) | 32 of 32 and 32 of 32 |
| werewolf | 21 / 80 | 0 | 9, 170, 18 | none exist | not covered (none before) | 128 of 128 and 128 of 128 |

### The structural comparison

A one-time script (in the session's scratch directory, not checked in;
it imported both games while both existed) compiled each hand port and
each generated game with `compileGame` and walked both as JSON trees:
the whole `GameSpec` (functions as a presence marker, every progression
evaluated on currentCount 0 to 16 at the length its group evaluates it, as
`rules/data.test.ts` does), the display name and `variantConfig`, the
compiled `GameDefinition` (state definition, chest: decks, components,
enums, generic values), the enum registry, the component schema, every
compiled move (index, fields, field names, schema, field and name policies,
declared flag, legal plan with its custom check's presence, input fields),
the canonical move names, the phase enum, the merged legal templates, the
no-progression set, the move input schema and its fingerprint, the choice
schema and its fingerprint, and the legal registry's keys. Every differing
path was listed and classified. No path is unclassified.

**Equal in every game** (no differing path): the display name, the variant
config, every enum, every deck's components in order, every property's
type, deck, size, cap and merged stacks, the move list and its order, each
move's name, help text, fix-up flag, base, fields and their types, legal
phases, progression, `seatsPlayer` and `gatheringStart`, each move's
compiled plan (every spec, argument and message, and whether a custom check
exists), each move's field and name policies, the input fields, the move
input schema and the choice schema and both fingerprints, the templates,
the phase enum, the no-progression set and the registry.

The differences, by class. The left half is in what `compileGame` derives;
the right half is only in how the `GameSpec` object spells the same thing.

| Class | memory | blackjack | tictactoe | pig | debuganimations | werewolf | Why it is not a behavior difference |
|---|---|---|---|---|---|---|---|
| Compiled: a sanitize tag written as its policy map (`'order'` against `{ all: 'order' }`, absent against `{ all: 'visible' }`) | 13 | 20 | 9 | 14 | 24 | 11 | `policyMapFor` gives the same map for both; the compiled field and name policies are equal |
| Compiled: a move field's `codec` and `input` explicit | 6 | 10 | 6 | 4 | 6 | 10 | The hand port left them to the default; the input schema, its fingerprint and the input fields are equal |
| Compiled: a deck's generic values present | 1 | 1 | 0 | 0 | 1 | 0 | **Closer to Go.** Go's `SetGenericValues`; the hand ports omitted them. No recorded view differs (both pass every document); `rules/engine.test.ts` now compares them with Go's exported definition |
| Compiled: a component schema for the deck | 1 | 0 | 1 | 1 | 1 | 0 | **Closer to Go.** The exported `ComponentProps`; the hand ports declared none where every value is a string or a slice |
| Compiled: `Overlap: false` written against omitted | 0 | 1 | 0 | 0 | 0 | 0 | The same default |
| Compiled: a stack constraint Go declares | 0 | 0 | 1 | 0 | 0 | 0 | **Closer to Go.** tictactoe's `UnusedTokens` is `stack:"tokens,same(Value)"` (`examples/tictactoe/state.go:55`); the hand port omitted the constraint. Every token a player's stack ever receives is that player's, so no recorded move reaches it |
| Spec: the sanitize tags (as above) | 13 | 20 | 9 | 14 | 24 | 11 | As above |
| Spec: codec and input (as above) | 6 | 10 | 6 | 4 | 6 | 10 | As above |
| Spec: legality as the loader writes it (a `plan` with the base check, `planEnabled`, stock parts in `stock`) against the hand port's `preconditions`, `suppress` and base | 22 | 40 | 5 | 18 | 28 | 12 | The compiled plan of every move is equal, spec for spec; the stock `apply`/`legal` the loader takes from `rules/stockRegistry.ts` is what the hand port inlined |
| Spec: a `false` flag explicit (`allowMultipleInProgression`, `planEnabled`) | 9 | 36 | 6 | 8 | 21 | 20 | The same default |
| Spec: `base: 'default'` explicit | 5 | 20 | 3 | 3 | 13 | 9 | The same default |
| Spec: display text explicit (variant value display names, descriptions) | 7 | 11 | 0 | 1 | 0 | 3 | The compiled `variantConfig` and display name are equal |
| Spec: an empty `fields` record explicit | 4 | 18 | 1 | 4 | 10 | 6 | The same |
| Spec: an empty table explicit (`enums`, `componentEnums`, `computed`, `constants`, `variants`, `legalTemplates`, `traits`) | 5 | 2 | 5 | 6 | 6 | 6 | The same |
| Spec: initial values as data (`initialValues`) against an `initial` function | 2 | 2 | 3 | 5 | 2 | 3 | Both give the same version-0 state in every recorded document |
| Spec: deck generic values (as above; tokens decks `null`) | 1 | 1 | 1 | 1 | 2 | 0 | As above |
| Spec: deck component props (as above) | 1 | 0 | 1 | 1 | 1 | 0 | As above |
| Spec: an end hook Go declares that the hand port left to the engine default | 1 | 0 | 0 | 0 | 1 | 0 | memory `end.playerScore` is Go's `PlayerScore` (`GameScore()`), which the engine's default (the `gameScore` trait) is; debuganimations `end.check` is Go's `CheckGameFinished`, `return false, nil`, which an absent `end` is |
| Spec: `Overlap` (as above) | 0 | 1 | 0 | 0 | 0 | 0 | As above |
| Spec: the tictactoe constraint (as above) | 0 | 0 | 1 | 0 | 0 | 0 | As above |
| **Total differing paths** | **97** | **193** | **58** | **84** | **146** | **101** | |

So every difference is spelling, an explicit default, or the generated game
carrying something of Go's the hand port left out (generic values, a
component schema, a stack constraint, two end hooks). None is a difference
in behavior a recording shows, and the corpora below are where behavior is
held.

### The rejection vectors and the engine tests

`rules/rejections.test.ts` (49 vectors from Go: 47 refused proposals on
memory, blackjack, tictactoe and pig, and the two fix-up-cap cases on a
wrapped tictactoe) and `rules/engine.test.ts` (timers against a moving
clock, ticks, seat inputs, purity of a step, creation refusals, a finished
game, and each game's state definition against Go's `DefinitionFor`) were
switched to the generated games while the hand ports still existed, and
run: every case passes, as every case did on the hand ports in the base
run (1,510 of 1,510). Identical results: Go's error text, an untouched
state, the same versions and final state at the cap.

One engine test needed a change and is stronger for it: the state
definition test compared each deck with Go's `DefinitionFor`, which carries
no generic values; the generated memory and blackjack carry them (as Go
does). The deck is now compared without them against `DefinitionFor`, and
the generic values against Go's exporter (`conformance/definitions/`).

### Fresh transient fuzz runs, replayed against both

Seeds not used before (the earlier runs were 1000.. and 100..), bounded as
those were, written outside the repository:

```
./scripts/go-local run ./boardgame-util fuzz run --game tictactoe       --seeds 2000..2999 --max-versions 150 --out DIR   #  54 s
./scripts/go-local run ./boardgame-util fuzz run --game werewolf        --seeds 2000..2127 --max-versions 150 --out DIR   # 102 s
./scripts/go-local run ./boardgame-util fuzz run --game pig             --seeds 2000..2099 --max-versions 150 --out DIR   # 129 s
./scripts/go-local run ./boardgame-util fuzz run --game memory          --seeds 2000..2023 --max-versions 150 --out DIR   # 150 s
./scripts/go-local run ./boardgame-util fuzz run --game blackjack       --seeds 2000..2031 --max-versions 150 --out DIR   # 129 s
./scripts/go-local run ./boardgame-util fuzz run --game debuganimations --seeds 2000..2031 --max-versions 150 --out DIR   # 149 s
```

Each document was replayed by `testing/fuzz.ts runFuzzDocument` (every
version's state, move record and status, every viewer's game object,
sanitized move, legality and choices, every rejection proposed again) once
against the hand port and once against the generated game, by a one-time
script:

| Game | Documents | Versions | Viewer-versions | Rejections | Hand port passed | Generated passed |
|---|---|---|---|---|---|---|
| tictactoe | 1,000 | 17,346 | 69,384 | 1,172 | 1,000 | 1,000 |
| werewolf | 128 | 6,973 | 54,395 | 877 | 128 | 128 |
| pig | 100 | 15,147 | 90,118 | 978 | 100 | 100 |
| memory | 24 | 3,436 | 20,544 | 305 | 24 | 24 |
| blackjack | 32 | 3,346 | 22,208 | 74 | 32 | 32 |
| debuganimations | 32 | 2,316 | 9,264 | 358 | 32 | 32 |
| **Total** | **1,316** | **48,564** | **265,913** | **3,764** | **1,316** | **1,316** |

### Also, before removal

- The two Fabric patterns' table bytes (`fabric/patterns/tools/measure-bytes.mjs
  memory` and `blackjack`: what the table stores and each viewer is handed,
  for the deal and every move of a fixed game) are byte for byte the same
  output with the hand port and with the generated game.
- The pattern Node tests (below) passed with the generated games while the
  hand ports were still present.

## Step 2: the generated games in use

### The export name

`boardgame-util/lib/transpile/files.go` now emits, in each
`games/<game>/index.ts`, `export function <game>(): GameSpec` (`memory()`)
in place of `<game>Generated()`. Nothing needed the old name once its users
(below) were switched, so it is gone. `indexFile` refuses a game named like
one of the file's own imports (`bodies`, `definition`, `gameFromData`,
`GameSpec`); `TestIndexExport` holds both. All seven games were regenerated
with `./scripts/go-local run ./boardgame-util transpile ./examples/...`;
only the seven `index.ts` files changed. The reports did not change
(`port --all --dry-run --check` current).

### Consumers switched

| File | Now |
|---|---|
| `fabric/patterns/memory/game.ts`, `fabric/patterns/blackjack/game.ts` | `games/<game>/index.ts`; comments say the game is generated |
| `fabric/patterns/memory/README.md`, `fabric/patterns/blackjack/README.md` | The file table, and a paragraph: the game is generated, the switch gave the table a new policy identity, tables on the old one are stranded |
| `games/conformance.test.ts` | The six generated games |
| `games/fuzz.test.ts` | One registry of the seven generated games (the hand-port registry and the layering are gone) |
| `games/<game>/<game>.gen.test.ts` (7) | Import `<game>`; memory's no longer describes a hand-port stand-in it stopped using |
| `games/generated.testkit.test.ts` | Comments |
| `rules/rejections.test.ts`, `rules/engine.test.ts` | The generated games (engine: the generic-values change above) |
| `rules/data.test.ts` | Reworked; see "Tests removed or reduced" |
| `rules/stockRegistry.test.ts` | Reduced; see below |
| `rules/purity.test.ts` | Scans each `games/<game>/` directory; expects the 21 generated files; allows a generated file's imports (`./*.gen.ts`, `../../rules/`, `../../go/`, `../../legal/`) |
| `lib/lib.test.ts` | The generated pig's die and Roll Dice, the generated blackjack's deck and enums |
| `testing/mutate.ts`, `testing/mutate.test.ts` | See "The 20 mutations" |
| `scripts/fuzz-mutations.mjs` | Loads every game from `games/<game>/index.ts` |
| `scripts/port-check.mjs` | Reads the `<game>` export |
| `boardgame-util/lib/port/check_test.go` | A comment naming the removed file |
| `rules/stock-inventory.json` | `verifiedBy` notes: the generated games, not the hand ports or the stand-in rebuild; DealCountComponents' note |
| `server/static/src/engine/README.md` | "Defining a game" has both ways; the layout and test tables |
| `fabric/transpile/HAND-PORT-BUGS.md` | A closing note |

Searched and not needing a change: `fabric/guest`,
`fabric/patterns-declarative`, `server/static/packages` and the webapp
import no engine game; `conformance/fuzz/README.md` says "hand-picked"
(the corpus), not hand port. Status and proposal documents that describe
the hand ports historically were left as written.

`git grep` finds no import of a removed file. What still names one:
comments (`check_test.go`, the mutation notes), and
`server/static/scripts/pairs-extract.mjs`, the pair store's one-time
extraction, which reads the hand ports and now fails if run; that is the
intended signal (`status/pairs.md`, "For the hand ports' removal"), and
`games/pairs.test.ts`'s currency test now skips saying so.

### Tests removed or reduced

`npm run test:unit`: **1,510 of 1,510 before; 1,513 after, 1,512 pass, 1
skipped, 0 fail.** By name (`node --test` output, before against after):
22 names went and 25 came.

Removed, with no successor, because their subject is gone:

| Test | Why |
|---|---|
| `registry parts reproduce {memory,blackjack,tictactoe,pig,debuganimations,werewolf} on its corpus` (6) | Rebuilt each hand port with its stock moves swapped for the registry's parts and replayed its corpus. A generated game is built from the registry's parts already (`gameFromData`), so the same proof is each `<game>.gen.test.ts`, with Go's exported configuration rather than the test's hand-read table |
| `every hand port had stock moves replaced` | Counted those swaps. Replaced by `every generated game runs stock moves from the registry` (blackjack at least 13, every game at least 1, checkers included) |

Reduced or re-expressed (the old name went, a new one came):

| Old | New | What changed |
|---|---|---|
| `{game} exported definition: gameFromData with the hand port's functions passes the corpus` (6) | `{game} exported definition: moves through move: compile as the generated game and pass the corpus` (6) | The old test loaded the exported definition with the hand port's functions as bodies (an adapter by role, with two test-written bodies for functions the hand ports left to engine defaults) and compared it with the hand port, in two passes. The bodies pass with the generated bodies IS the generated game (its own suite replays it), so what is left is the `move:` hook: every move whose generated MoveSpec carries no stock Apply goes through it, compiles as the generated game, and passes the conformance corpus; a no-stand-in check stays. A move whose translated `apply` calls its stock Apply cannot go through `move:`, which carries behavior and not stock parts (`rules/data.ts`); those moves keep their bodies and are listed. tictactoe has none that can (all four moves are stock), so for it this is the bodies pass |
| (new) | `the move: hook served moves of every exported game but tictactoe` | So the reduction above cannot go vacuous |
| `{pig,blackjack} fixture: gameFromData with the hand port's functions passes the corpus` | `... with the generated game's functions passes the corpus` | The hand-written fixtures now take the generated bodies; a fixture move whose hand-written stock configuration the registry cannot build, or builds without the Apply Go's configuration yields, gets the registry's parts from Go's exported configuration through `stock:` (blackjack: 7 moves; pig: none). Compared with the generated game instead of the hand port; both passes (bodies, `move:`) kept |
| The eight contract tests in `rules/data.test.ts` | Same names | Bodies from the generated pig and blackjack instead of the hand ports |
| `blackjack's inlined deck and enums are playingcards'`, `pig's inlined die and roll are dice's` | `blackjack's deck and enums are playingcards'`, `pig's die and roll are dice's` | The same assertions on the generated games |
| `the fuzz corpus holds eight documents for each hand port, as its index lists` | `... for each game, ...` | The seven registered games |
| `the rules layer and the four games are present` | `the rules layer and the seven generated games are present` | The 21 generated files |
| `the 20 historical mutations each apply to exactly one place` | `... to exactly the places they name` | One mutation occurs twice by design; a swap keeps length |
| `applyMutation refuses a text that does not occur exactly once` | `... as often as it must` | Plus the new count |
| (new) | `applyMutation swaps two blocks` | The swap the move-order mutation needs |
| `bodyMutants: a real hand-port function` | `bodyMutants: a real generated function` | `blackjack_handValue` |
| `the store is what pairs-extract.mjs cuts from the hand ports today` | (same, now **skipped**) | Skips by design: "the hand ports are gone (0 of 6 present)" |

`rules/data.test.ts` lost its adapter (`handPortBodies`, `TEST_BODIES`);
the generated bodies need neither.

### The 20 mutations

`node scripts/fuzz-mutations.mjs historical` (in `server/static`), against
the 56-document fuzz corpus: baseline 56 of 56; **20 of 20 killed**; no
file under `src/engine/` or `conformance/` changed.

The ten made to hand ports, re-expressed (each keeps its id, description
and expected failure; its `note` gives the hand-port text it replaces):

| Mutation | Now in | Text | Fuzz documents failing (first field) |
|---|---|---|---|
| blackjack-ace | `games/blackjack/bodies.gen.ts` | `if (currentValue >= 11) {` to `>` (the generated `handValue` has `TargetScore - 10` folded) | 5 of 56 (view.game 4, state 1) |
| memory-fixup-order | `games/memory/definition.gen.ts` (data) | The Capture Cards and Start Hide Cards Timer move literals swapped in `Moves` (a new `swapWith`/`blockEnd` form) | 8 (header) |
| hole-card-policy | `games/blackjack/definition.gen.ts` (data) | `HiddenHand`'s `Sanitize: { other: 'len' }` to `'order'` | 8 (view.game) |
| move-field-sanitize | `games/tictactoe/definition.gen.ts` (data) | Place Token's `Slot` field loses `Sanitize: 'self:visible'` | 8 (view.move) |
| hidden-name-alias | `games/tictactoe/definition.gen.ts` (data) | Place Token's `HiddenAnimationKey: 'Hidden Action'` removed (the move's own key, not the copy in its stock configuration) | 8 (header) |
| pig-extra-draw | `games/pig/bodies.gen.ts` | One `state.rand().intn(state.numPlayers())` before the starting player's | 7 (state) |
| error-string | `games/blackjack/bodies.gen.ts` | `'Current player is already at target scores'` loses its last letter | 8 (view.legality) |
| vote-same-role | `games/werewolf/definition.gen.ts` (data) | `VoteTarget`'s `Sanitize` loses `same-role:visible`, at **both** votes (`occurs: 2`): the hand port's one vote factory served Cast Vote and Cast Night Vote | 7 (view.move) |
| tree-enum-default | `games/debuganimations/definition.gen.ts` (data) | The phase enum's `DefaultValue: 1` to `0` | 8 (header) |
| public-shuffle-secret | `games/debuganimations/bodies.gen.ts` | `publicShuffle(...)` to `shuffle(...)` | 6 (state) |

**Two engine mutations moved**, because the first run after the switch
showed them surviving (18 of 20): each mutated a line that only a hand port
ran.

| Mutation | Was | Now | Why |
|---|---|---|---|
| round-robin-start | `rules/stock.ts`, `dealCountComponents`'s own round robin (`apply(state)`) | `rules/stock.ts`, `roundRobinParts` (`apply(state, move)`), the round robin of the registry's `dealComponents`, which the generated blackjack's deals run | 0 of 56 failed on the old line; 8 of 56 (state) on the new |
| progression-repeatable | `rules/stock.ts`, the installer `addOrderedForPhase` (`allowMultiple: item.allowMultipleInProgression === true`) | `rules/data.ts` `progressionFromData` (`allowMultiple: move.AllowMultipleInProgression`), where every generated game's progressions are built | 0 of 56 on the old line; 24 of 56 (status 12, state 12) on the new |

The other eight engine mutations are unchanged and killed.

## Step 3: the Fabric patterns

**The policy identity of both patterns changed.** `fabric/patterns/memory`
and `fabric/patterns/blackjack` now import the generated game; the engine
modules a table imports are part of its policy, so each table has a new
identity, and a table created on the old one (in flight) is stranded. Any
future change to a generated game's text does the same
(`fabric/TRANSPILE-PROPOSAL.md`, risk 6). The owner accepted this.

### Suites run

| Suite | Command (repository root unless said) | Result |
|---|---|---|
| Pattern Node tests | `node --test fabric/patterns/memory/table.test.ts fabric/patterns/memory/parity.test.ts fabric/patterns/blackjack/parity.test.ts` | 39 of 39 (the count in `fabric/status/viewer-bound-port.md`); run with the hand ports present and again after their removal |
| Table bytes | `node fabric/patterns/tools/measure-bytes.mjs memory`, `... blackjack` | Byte-identical output, hand port against generated |
| The sandbox rules | `server/static/src/engine/sandbox.test.ts`, `purity.test.ts`, `rules/purity.test.ts` (in `npm run test:unit`) | Pass with the generated games scanned |
| The pattern compiler's option | `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` (in `server/static`) | Passes |

### Not run

No labs checkout was looked for beyond what the status files name: they
name only a labs worktree that was removed
(`.claude/worktrees/boardgame-viewer-port` in the labs repository,
`viewer-bound-port.md`) and placeholder paths (`/path/to/labs`); neither
`$LABS` nor `CF_LABS_ROOT` was set. Nothing was installed or fetched.

| Suite | Command | Why not |
|---|---|---|
| Memory parity in the sandbox | `EXPERIMENTAL_VIEWER_BOUND_RELEASE=true CF_LABS_ROOT=/path/to/labs /path/to/labs/bin/cf test fabric/patterns/memory/parity.test.tsx --root "$PWD"` | Needs labs |
| Blackjack parity in the sandbox | the same with `fabric/patterns/blackjack/parity.test.tsx` | Needs labs |
| Memory, multi-user (202) | the same with `fabric/patterns/memory/multi-user.test.tsx`, with `EXPERIMENTAL_SERVER_EXECUTION` false and true | Needs labs |
| Blackjack, multi-user (92) | the same with `fabric/patterns/blackjack/multi-user.test.tsx` | Needs labs |
| The display boundary | `cp fabric/patterns/blackjack/attacks/display-fit.test.ts /path/to/labs/runtime/html/test/boardgame-blackjack-display-fit.test.ts`, then `(cd /path/to/labs/runtime/html && BOARDGAME_ROOT="$OLDPWD" deno test -A test/boardgame-blackjack-display-fit.test.ts)` | Needs labs |
| Guest | `npm run type-check`, `npm test` in `fabric/guest` | `fabric/guest/node_modules` is absent here (not installed); the guest imports no engine game, so the switch does not reach it |
| Browser games | `fabric/patterns/{memory,blackjack}/tools/play-*.mjs` | Needs labs servers and a deployed piece |

So whether the generated games load in a real pattern sandbox is
**unverified**; the labs-free half (`sandbox.test.ts`, the purity tests,
the sandbox type check) passes.

### Could the generated modules' size or shape matter to a pattern?

From the repository's own record of the pattern sandbox
(`fabric/patterns/memory/FINDINGS.md`, "The engine in the sandbox"; labs
`SES_SANDBOXING_SPEC.md` section 4.2 as quoted there): a pattern module's
top level may hold a function, freezable data, imports and types. Each
generated file's top level is imports and `export function`s:
`definition()` builds its (large) literal when called, as the hand ports
built their specs; `bodies(ref)` is a `switch`. `sandbox.test.ts` checks
every one of those forms and passes on the 21 files. No rule recorded there
constrains identifier names, so the `$` in names such as
`memory_gameDelegate_ConfigureComputedProperties$1` (and the translator's
`$s2`, `$i1` locals) is ordinary JavaScript; that labs' verifier accepts
them is not shown here, since no labs run was possible.

Size: the patterns are deployed as source (`cf ... --root <repository
root>`, compiled by labs), not bundled by a script in this repository, so
the measure is the static import closure of each pattern's `main.tsx` and
`table.tsx`:

| Pattern | Before (hand port) | After (generated) |
|---|---|---|
| memory | 47 modules, 591,394 bytes | 55 modules, 729,972 bytes |
| blackjack | 47 modules, 594,449 bytes | 55 modules, 765,944 bytes |

The eight more modules are the generated game's three files, `rules/data.ts`
and the `go/` facade. The only size limit the repository records is the
4 MiB server-execution upload (FINDINGS, "The guest bundle", which applies
to the iframe guest's assets); whether the pattern program falls under a
limit, and how labs' type check and compile time change with a 61 KB
definition literal, is not recorded and was not measured.

## Before reporting

Go with `BOARDGAME_NODE` set to Node 24.22.3; Node 24.22.3 for the rest.

| Command | Result |
|---|---|
| `./scripts/go-local build ./...` | passes |
| `./scripts/go-local vet ./boardgame-util/lib/transpile/` | clean |
| `conformance export --check` | "The conformance corpus is current" |
| `fuzz corpus --check` | "The fuzz corpus is current (56 documents)" |
| `gamedef export --check` | "The game definitions are current" |
| `transpile --check ./examples/...` | "The generated games are current" |
| `port --all --dry-run --check` | "The port reports are current" (memory and checkers: 1 refused, 1 agent-written each) |
| `./scripts/go-local test ./boardgame-util/... ./moves/... . ./legal/... ./examples/checkers/...` | 29 packages ok, none failing (`lib/transpile/pairs` 128.8 s, with the substitution proof; `examples/checkers` 79.0 s) |
| `./scripts/go-local test -run 'TestSubstitutionProof\|TestBrokenRuleFailsProof\|TestRuleAgainstPairs' ./boardgame-util/lib/transpile/pairs/` | ok, 113.9 s: `TestSubstitutionProof` 86 of 86 substituted pairs passing (memory 15, blackjack 21, tictactoe 7, pig 9, debuganimations 24, werewolf 10; 11 helpers through callers; 0 failing); `TestBrokenRuleFailsProof` all four bundling pairs fail as they must; `TestRuleAgainstPairs` "not yet a rule" for the one carried refusal |
| `npm run type-check` | passes |
| `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | passes |
| `npm run test:unit` | 1,513 tests: 1,512 pass, 1 skipped (the pair store currency test, by design), 0 fail (1,510 of 1,510 at the base) |
| `npm run test:layers` | 11 of 11 |
| `node scripts/fuzz-mutations.mjs historical` | baseline 56 of 56; 20 of 20 killed; no source file changed |

## What is no longer verified now that the hand ports are gone

- **Independence.** Each game was held to Go by two ports written two ways
  (by hand and by machine), and they agreed on every recording. Now one
  port is held to Go; a recording-level difference would still be caught,
  but nothing else compares the generated game with an independent reading
  of the Go.
- **The structural comparison** above cannot be rerun.
- **Code only the hand ports ran.** These engine exports are now used by no
  shipped game and by no test: `rules/stock.ts` `dealCountComponents`
  (and its own round robin), `defaultRoundSetup`, `eliminationProps`,
  `scoreProps`, `roleProps`; `legal/` `firstOccupied`, `occupiedOrdinal`,
  `mayMoveToSameSlot`, `playerBoolIs` (spec constructors; the predicates
  they name are still evaluated, from the generated games' plans). These
  are used only by `rules/stock.test.ts`: `currentPlayerProps`,
  `inactivePlayerProps`, `seatProps`, `phaseProps`, `roundRobinProps`,
  `combine`. The installers and hand-written-spec API in general (the
  "Written by hand" half of the engine README) are now exercised only by
  the games the tests write by hand (`rules/stock.test.ts`'s, replayed
  against Go's `testdata/stock/` recordings, and the small specs of other
  `rules/` tests); no shipped game is written by hand. None of this was deleted (not in this stream's
  files).
- **The `move:` hook with a whole hand-written move.** It was exercised with
  every move of six hand ports; now with the generated moves that carry no
  stock Apply (3 to 12 per game, none in tictactoe).
- **The pair store's currency** against the hand ports
  (`games/pairs.test.ts` skips); the store itself, its proof that each game
  built from it passes its corpus, and the substitution proof still run.
- The labs suites listed above, on the generated games.

## Judgment calls

1. **The old export name was dropped**, not kept beside the new one:
   everything that used it (`port-check.mjs`, `fuzz-mutations.mjs`, the
   generated tests, `fuzz.test.ts`) was switched.
2. **`games/conformance.test.ts` and `games/fuzz.test.ts` were kept and
   switched**, though each `<game>.gen.test.ts` replays the same documents
   on the same game: they keep the whole-corpus pass tables, the corpus
   index check, checkers' fuzz replay (its own test replays three other
   documents) and the divergence-report test. The unit suite's run time did
   not grow (26.6 s after, 27.1 s before): the hand-port replays went.
3. **The two moved engine mutations** are a change to what those two
   mutations mean in file terms; in behavior they mean the same (the deal's
   round-robin start, repeatable moves in a progression), now on the code
   the shipped games run.
4. **The README keeps the hand-written API's memory excerpts**, labeled as
   the hand port's: that half documents writing a `GameSpec` by hand, which
   the generated game does not do. Memory's state and its Hide Cards move
   are shown from the generated files in the new "From Go" part.
5. **`engine.test.ts` compares generic values with Go's exporter.**

## Edits outside my files

None. Every file changed is one the brief names, a consumer it asks for
(`scripts/port-check.mjs`, `rules/stock-inventory.json`), or this file.
