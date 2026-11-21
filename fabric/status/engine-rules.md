# Status: engine rules layer (stream C, step 4)

Branch `claude/fabric-c-engine-rules-a18167`, 2026-11-21. Nothing pushed, no
PR, `master` untouched, no existing Go source changed except the corpus
definition (one new scenario).

The rules layer of the TypeScript engine is in
`server/static/src/engine/rules/` and the ported games in
`server/static/src/engine/games/`. A game defined there runs with no Go.
All six games in the conformance corpus are ported, and every document
passes at every version for every viewer. How to define a game and what a
host must do are in
[`server/static/src/engine/README.md`](../../server/static/src/engine/README.md).

## Result

| Check | Result |
|---|---|
| `npm run type-check` (Node 24.22.3) | passes |
| `npm run test:unit` (Node 24.22.3) | 1,164 tests, 1,164 pass (1,111 at baseline; 53 added) |
| Conformance corpus | 22 of 22 documents pass: 961 versions, 5,312 viewer-versions, every field |
| Go rejection vectors | 49 of 49 match, including both fix-up-cap cases |
| Legal corpus and Go predicate vectors, through the real state binding | 130 of 130 cases, 3,965 of 3,965 verdicts |
| State definitions against Go's reflection | memory, blackjack, tictactoe, pig match |
| `go build ./...`, `go test ./boardgame-util/lib/conformance/...` (go1.25.13) | pass |

The baseline on the fresh worktree was green before any change
(`npm ci`, type-check, 1,111 unit tests, the Go conformance tests), so the
merge of the first wave needed no fix.

The required games were memory, blackjack, tictactoe and pig. The two
stretch games, debuganimations and werewolf, were done after those were
green. Every document passed on its first full run except for one header
field (below, "What the first runs found"), which is why the runner was
then attacked with mutations (below, "Is the runner honest?").

## Commits

| Commit | Contents |
|---|---|
| `85877b19` | `rules/binding.ts`, `rules/enums.ts`: the legal evaluator's state-access interface over `EngineState` and the core stacks. `Stack.insertStarterComponent` added to the core |
| `19960e80` | The rules layer (`types`, `compile`, `state`, `engine`, `timers`, `time`, `view`, `progression`, `stock`, `sha256`), the four games, the conformance runner and its test |
| `6cbcbc3a` | `testdata/gen/rejectvec` and `rules/rejections.test.ts`; move names matched without regard to case |
| `4f280933` | `rules/engine.test.ts`, `rules/purity.test.ts`; `testdata/gen/defs` extended to pig and tictactoe |
| `612a0cfa` | New corpus scenario `blackjack/scenario-reshuffle` |
| `b90ad319` | debuganimations and werewolf; README |
| (this commit) | This report |

Files touched outside `server/static/src/engine/`:

- `boardgame-util/lib/conformance/corpus/corpus.go`, `scenarios.go`: one new
  scenario, as the task allowed.
- `conformance/blackjack/scenario-reshuffle.json.gz` (new) and
  `conformance/index.json` (one row added).
- This report.

Inside the engine directory, three first-wave files changed, additively:
`stack.ts` (`insertStarterComponent`), `sanitize.ts` (three group-membership
helpers exported), `definition.ts` (`EnumDef.Parents`, optional).

## Conformance pass table

Each cell is the number of comparisons, all equal. `header` is one
comparison of the static tables (move table, input and choice schemas and
their fingerprints, computed-property descriptors, variant configuration,
chest). `status` is settled, finished and winners per version. The four
`view.*` columns are compared once per viewer per version.

| Document | Versions | Viewers | header | state | move | status | view.game | view.move | view.legality | view.choices | Result |
|---|---|---|---|---|---|---|---|---|---|---|---|
| blackjack/golden-714104226cb9a317 | 30 | 4 | 1 | 30 | 30 | 30 | 120 | 120 | 120 | 120 | pass |
| blackjack/golden-82ef86840a83b9ed | 43 | 6 | 1 | 43 | 43 | 43 | 258 | 258 | 258 | 258 | pass |
| blackjack/golden-c93d3566c752df43 | 41 | 6 | 1 | 41 | 41 | 41 | 246 | 246 | 246 | 246 | pass |
| blackjack/golden-d2b8e7e47a412543 | 21 | 6 | 1 | 21 | 21 | 21 | 126 | 126 | 126 | 126 | pass |
| blackjack/golden-dac0bc2b57492b3a | 66 | 9 | 1 | 66 | 66 | 66 | 594 | 594 | 594 | 594 | pass |
| blackjack/scenario-bust-and-finish | 47 | 5 | 1 | 47 | 47 | 47 | 235 | 235 | 235 | 235 | pass |
| blackjack/scenario-reshuffle (new) | 171 | 9 | 1 | 171 | 171 | 171 | 1,539 | 1,539 | 1,539 | 1,539 | pass |
| memory/golden-6cb9690b3b55fe3d | 90 | 4 | 1 | 90 | 90 | 90 | 360 | 360 | 360 | 360 | pass |
| memory/golden-c369b3a9d5c0eda0 | 10 | 4 | 1 | 10 | 10 | 10 | 40 | 40 | 40 | 40 | pass |
| memory/golden-fec3529ffca41c00 | 6 | 4 | 1 | 6 | 6 | 6 | 24 | 24 | 24 | 24 | pass |
| memory/golden-ff0c8d5d3547687c | 30 | 4 | 1 | 30 | 30 | 30 | 120 | 120 | 120 | 120 | pass |
| memory/scenario-mismatch | 6 | 4 | 1 | 6 | 6 | 6 | 24 | 24 | 24 | 24 | pass |
| memory/scenario-full-game | 30 | 4 | 1 | 30 | 30 | 30 | 120 | 120 | 120 | 120 | pass |
| pig/golden-67672ddc7e8a71ad | 228 | 4 | 1 | 228 | 228 | 228 | 912 | 912 | 912 | 912 | pass |
| pig/golden-81137f3f9f5d40ba | 28 | 4 | 1 | 28 | 28 | 28 | 112 | 112 | 112 | 112 | pass |
| pig/scenario-short | 27 | 4 | 1 | 27 | 27 | 27 | 108 | 108 | 108 | 108 | pass |
| tictactoe/golden-0e7f8ddf5c1c7c65 | 14 | 4 | 1 | 14 | 14 | 14 | 56 | 56 | 56 | 56 | pass |
| tictactoe/golden-1b7c0a299f5b79c5 | 18 | 4 | 1 | 18 | 18 | 18 | 72 | 72 | 72 | 72 | pass |
| tictactoe/golden-568a53942f7d083d | 12 | 4 | 1 | 12 | 12 | 12 | 48 | 48 | 48 | 48 | pass |
| tictactoe/scenario-seated-win | 14 | 4 | 1 | 14 | 14 | 14 | 56 | 56 | 56 | 56 | pass |
| debuganimations/scenario-shuffles (stretch) | 16 | 4 | 1 | 16 | 16 | 16 | 64 | 64 | 64 | 64 | pass |
| werewolf/scenario-timed-vote (stretch) | 13 | 6 | 1 | 13 | 13 | 13 | 78 | 78 | 78 | 78 | pass |

`games/conformance.test.ts` emits this table as a test diagnostic on every
run.

### How the runner feeds a document

`testing/conformance.ts` creates the game from the recorded id, salt, player
count, variant and seating flag, then repeats only the outside input:

- a move with trigger `propose` is proposed by its recorded proposer with
  **only the fields a proposer supplies** (input disposition `required` or
  `server-defaulted`). Context-owned fields such as `TargetPlayerIndex` are
  left for the engine to derive, and the resulting move record must still
  equal the recording. `expectedVersion` is always passed;
- a move with trigger `seat` becomes a seat input for its recorded index;
- a move with trigger `timer` becomes a forced tick (the recording's clock
  never advances).

Set-up moves and fix-up moves are never fed. If the engine does not produce
one where the recording has one, the document fails.

### The one normalization the corpus does not declare

The runner honors each document's `normalization` list by construction:
omitted fields are never produced, and the engine's `now` is the document's
fixed clock. **One further rule is applied that no document declares: a
recorded `winners` or `Winners` that is `null` is treated as an empty list.**
It is generic (the same for every game) and is the only such rule.

The cause is in the exporter, not the engine.
`boardgame-util/lib/conformance/export.go:223-226` sets a version's
`Winners` only when that version is the head being captured, so every
unsettled version carries Go's nil slice, which marshals as `null`, both in
`versions[].winners` and in each view's `game.Winners`. `conformance/schema.ts`
declares both as arrays. Without the rule, all 19 documents of the four
required games fail on `status` and `view.game` at their unsettled versions
and on nothing else. Fixing the exporter (always call `winners()`) would let the rule be
deleted; I did not change it because the exporter is existing Go.

### What the first runs found

The first full run matched every state, move record and view in all 19
documents of the four required games. One thing differed:

- **The choice-projection fingerprint of a game with no projections.**
  `FingerprintMoveChoiceProjectionSchema` (`move_choice_projection.go`)
  hashes a copy made by appending to a nil slice, so an empty schema hashes
  `"projections":null`, not `[]`. Reproduced.

Two more differences surfaced from the rejection vectors, not the corpus:

- **Move names are matched without regard to case**
  (`game_manager.go:1468`). `place token` is accepted. Reproduced.
- **`encoding/json`'s wording for a non-integer number** ("cannot unmarshal
  number 1.5 into Go value of type int"). Reproduced.

### Is the runner honest?

A suite that passes at once deserves suspicion, so 20 single-line mutations
were made one at a time, each run against the corpus and reverted. All 20
made documents fail, each on the field it should:

| Mutation | Fails on |
|---|---|
| Blackjack ace rule `>=` to `>` | view.game, view.legality |
| The null-winners rule removed | status, view.game (all 19 documents then in the test) |
| Memory's two fix-ups swapped | header, then the fix-up chain |
| Hole card policy `len` to `order` | view.game |
| Timer id drawn with 15 draws, not 16 | state, view.game |
| Computed properties from the unsanitized state | view.game |
| A move field's `sanitize` tag dropped | view.move |
| The hidden-name alias dropped | header, view.move |
| One extra draw before pig's starting player | replay stops (wrong player's turn) |
| An error string changed by one word | view.legality |
| Empty choice sets kept | view.choices |
| Round-robin's starting point off by one | the deal order |
| A seated player left active | the gathering chain |
| Repeatable moves not repeatable in a progression | the deal chain |
| `moves.AnyPlayer`'s target not set to the proposer | replay stops |
| Role value shown regardless of visibility | view.game |
| `same-role` dropped from a vote's visibility | view.move |
| Tree enum default changed | header, state |
| Public shuffle made secret | state, view.game |
| Timer's move bound to the committed state, not the one being built | state |

## What each new test proves

| Test | Proves |
|---|---|
| `games/conformance.test.ts` (25) | The table above. Also that every version and every viewer of every version was compared, so a document cannot pass by comparing nothing |
| `rules/rejections.test.ts` (6) | For 47 refused proposals: Go's exact error text, no versions returned, and the caller's state and history bit-for-bit untouched. For the two cap cases: Go's 258 committed versions, the same move names, and the same final state |
| `rules/binding.test.ts` (2) | The legal corpus (130 cases) and all 3,965 Go predicate verdicts, with every fixture loaded as a real `EngineState`, so stack answers come from the core's `Stack` |
| `rules/engine.test.ts` (17) | A timer's deadline is the save time plus its duration; a tick is refused before the deadline, fires once, and is stale after firing, after a cancel, or with the wrong id; time left as a viewer sees it; a seat input for a named or the next open seat; seating as a creation input; a step is a function of its inputs and does not modify them; creation refuses what Go refuses, in Go's words; each game's state definition equals the one Go derives by reflection; instants; SHA-256 against the platform; progression matching |
| `rules/purity.test.ts` (3) | `rules/` and `games/` use no DOM, Node, crypto, timer, clock, `Math.random`, `BigInt` or asynchrony, and import only engine modules |

The progression-matching cases and the tick-validation cases are
hand-written: Go has no single answer to record for the latter (see
divergences 3 and 4), and the former is also exercised against real Go by
every blackjack document.

## How the rejection vectors were produced

`testdata/gen/rejectvec` creates each game with a fixed id and salt on the
real engine, applies accepted moves, makes one refused proposal, records
`err.Error()`, and itself asserts that the stored state, version and
finished flag did not change. Regenerate with

```sh
./scripts/go-local run ./server/static/src/engine/testdata/gen/rejectvec
```

(also part of `testdata/gen/regen.sh`). Coverage, 49 cases:

| Kind | Cases | Examples of Go's text |
|---|---|---|
| Wrong proposer | 11 | `it's not your turn`; `The proposer was the ObserverPlayerIndex, but observers may never make moves.`; `The proposer was not valid.`; `ForceFinishTurn can only be proposed by AdminPlayerIndex` |
| Stale version | 3 | `move proposal used stale game version 0; current version is 2` |
| Illegal arguments | 10 | `slot 4 is already occupied`; `99 is not an occupied slot in game stack "HiddenCards"`; `You have no cards left to reveal this turn` |
| Input the binder refuses | 8 | `move "Place Token" requires field "Slot"`; `... does not accept creator field "TargetPlayerIndex"`; `... json: cannot unmarshal string into Go value of type int`; `field "Slot" must not be null` |
| Unknown, engine-owned, foreign, out of phase, out of progression | 12 | `unknown move "No Such Move"`; `move "Finish Turn" is engine-owned and is not a creator move`; `That move is not configured for this game.`; `Move is not legal in phase Normal Play`; `Move name does not match: Begin Game is not Wait For Enough Players - Gathering` |
| After game over | 3 | `Game was already finished` |
| The fix-up cap | 2 | `we recursed deeply in fixup, which implies that ProposeFixUp has a move that is always legal` |

The cap cases use tictactoe wrapped with an always-legal fix-up, once legal
only after the first move (the cap is reached by a player's proposal) and
once legal everywhere (the cap is reached while the game is created).

## Where the TypeScript engine differs from Go

Numbered so they can be referred to. None of these is hidden by the
comparator; each is either outside what the corpus records or was asked for.

1. **`Finished`, `Winners`, `Version` and `ActiveTimers` are per version.**
   As the exporter's finding 1 recommended. `viewerOutput` computes them
   from the version being viewed; Go's `JSONForPlayer` (`game.go:280-292`)
   takes them from the head of the game.

2. **Legality at version 0 reads an empty history.** In Go,
   `Game.MoveRecords(upToVersion)` replaces any `upToVersion < 1` with the
   game's current version (`game.go:471-473`). So a legality check against
   the version-0 state, made after the game has moved on, matches move
   progressions against the whole history up to the head. That is what the
   server's first-load `/info` does (the exporter's finding 2). The engine
   judges a version by that version's history. No corpus document can show
   the difference: every move in a progression in these games is a fix-up,
   and fix-ups are not listed in legality.

3. **A tick is checked against the clock.** Go's scheduler decides when to
   fire and `fire` does not re-check the deadline (`timer.go:667`). The
   engine has no scheduler, so `propose` refuses a tick whose deadline is
   after `now` (`timer deadline has not passed`), unless `force` is set.
   Go's retry and back-off after a failed completion (`timer.go:699-744`)
   are the host's business and are not ported.

4. **A seat input that seats nobody is an error.** Go's `ForceFixUp` returns
   nil and leaves the seat request pending in the server. The engine returns
   `No fix-up move would seat the player`. Both texts are the engine's own.

5. **A legality function that throws is a refusal.** A Go `Legal` that
   panics outside a declared plan takes the process down. Here the thrown
   message becomes the error text.

6. **With several unacceptable input fields, the one named is the first in
   sorted order.** Go ranges over the input map (`move_input_bind.go:32`),
   so its choice is arbitrary.

7. **The committed state is a copy, not a re-read.** Go reads the state back
   from storage after each commit. The engine copies it, which also discards
   the component index, so the core's divergence 3 (`ContractSize`) does not
   carry from one move to the next, as in Go.

8. **Durations are whole milliseconds.** Go's are nanoseconds.

9. **Not ported**, because no corpus game needs them: stack constraints
   (tictactoe declares `same(Value)` on a stack nothing ever moves into;
   see "Unproven"), boards keyed by an enum, `RepeatFromProp`, the
   `BeforeLeavePhase`/`BeforeEnterPhase` hooks, move choice recording
   (`move_choice_recording.go`), agents, the team and color behaviors and
   their computed values, the game-administrator behavior, and the stock
   moves no corpus game uses (the collect, move-components, increment,
   graph and market moves, `CloseAllSeats`, `ActivateEmptySeat`,
   `ActivateFilledSeat`, the select moves, `AllPlayersSubmitted`). Of the
   tree-enum API only `Ancestors` is ported; the leaf check in
   `validateBeforeSave` (`state.go:1113`) is not.

## Things in the Go that look like bugs

The first four change what players experience. "Read, not run" marks a
conclusion from reading the source that nothing here demonstrates.

1. **Pig: the randomly chosen starting player never gets the first turn.**
   `examples/pig/main.go:69` picks a starting player, but nothing calls
   `ResetForTurn`, so every player starts with `DieCounted` false. The
   fix-up Count Die is therefore legal at once: it counts the die's
   constructor value of 1 as a roll, which is a bust, and Finish Turn passes
   play to the next player. All three pig documents open this way: version 1
   is Count Die for the chosen player, version 2 is Finish Turn, and the
   first proposal comes from the other player. Reproduced.

2. **`ForceFinishTurn` ignores the phase it is installed in.**
   `moves/force_finish_turn.go:102` overrides `Legal` without calling
   `Default.Legal`, so the phase and progression checks never run. Blackjack
   installs it with `AddForPhase(phaseNormalPlay, ...)`, yet the corpus
   records it as legal for the admin in Gathering, Initial Deal and Round
   Cleanup (for example `blackjack/scenario-bust-and-finish` versions 1, 3,
   21). An admin "skip turn" during the deal would advance the current
   player in the middle of it. Reproduced.

3. **Blackjack scores ten plus ace as 11** (`examples/blackjack/state.go:96`).
   The exporter's finding 11. Reproduced in `games/blackjack.ts`
   (`handValue`) and flagged there; a mutation of that one comparison fails
   two documents.

4. **`AnyPlayer.Legal` changes the move it is asked about, so
   `legalForAnyone` depends on who is asking.** `moves/any_player.go:53`
   writes the proposer into `TargetPlayerIndex` when the move still has its
   default. The server's form code calls `Legal` for the viewer and then for
   the admin on the same move object, so the admin check sees the viewer's
   index. In `werewolf/scenario-timed-vote` version 11, Cast Vote has
   `legalForAnyone: true` for player 1, `false` for player 0 (who has voted)
   and `false` for the admin and the observer, for one state. Reproduced,
   including the mutation.

5. **The fix-up cap leaves a game half-made.** `game.go:1194` returns
   `ErrTooManyFixUps` after the versions are committed: 258 of them, since
   the check is `recurseCount > 256` on a count that starts at 0. When this
   happens during creation, `NewGame` returns an error while a 258-version
   game sits in storage with no main loop. Reproduced as far as the engine
   can: `createGame` returns the versions and the error together.

6. **A default-score game whose scores are all negative has no winner.**
   `base/game_delegate.go:563` starts the maximum at 0, so if every eligible
   player's score is below zero nobody matches it and the game ends with no
   winners. Read, not run. Reproduced as written.

7. **`DealCountComponents.Legal` can index the player list with -1.**
   `moves/deal_components.go:249` uses `nextPlayerIndex` unchecked. In a
   game whose delegate reports no current player, the starter is the
   observer index, `Previous` and `Next` leave it alone, and the index
   panics inside `ProposeFixUpMove`. Read, not run. The engine returns the
   panic's text as a refusal.

8. **The "moves with no progression" cache is keyed by game name, in a
   package variable** (`moves/default.go:19`, `:570`). Two managers whose
   delegates share a name but differ in moves share one answer, whichever
   was computed first. Wrapped delegates in tests and tools all have to be
   named after their package (the generators here name theirs `main`), so
   this is easy to hit. Read, not run.

9. **`MoveRecords` slices past the end for a state under construction.**
   `game.go:496` returns `cachedHistoricalMoves[:upToVersion]`. A caller
   that passes the version of the state being built (one more than the
   game's) gets a slice extended into spare capacity or a panic.
   `countMovesApplied` and `RoundRobin.lastMoveName` take a state's version;
   today both are reached only from `Legal`, with a committed state. Read,
   not run.

10. **`MoveWithInput` names moves two ways in its errors.**
    `move_input_bind.go` quotes the move's own name in "does not accept
    creator field" and the name as the caller typed it in "requires field"
    and in field errors. With case-insensitive lookup the two can differ.
    Reproduced.

11. **A timer that fired keeps its deadline; one that was canceled does
    not.** `Timer.Cancel` on a timer that is not active clears only the id
    (`timer.go:160`). Memory's Hide Cards cancels the timer whether the
    player or the timer proposed it, so the stored record ends as
    `{Status: "fired", Deadline: <kept>}` in one case and
    `{Status: "canceled", Deadline: <zero>}` in the other. Harmless.
    Reproduced.

12. **Blackjack's turn order runs backwards by one each round.** The player
    whose stand ends a round is still the current player when the next
    round is dealt, so they act first again. In
    `blackjack/scenario-reshuffle` the proposers are 0123456, 6012345,
    5601234, 4560123. It may be intended. Reproduced.

## Unproven

- **Stack constraints.** Not ported. Tictactoe's `UnusedTokens` is
  constrained in Go (`same(Value)`); the port is correct only because
  nothing moves into that stack after set-up, and set-up bypasses
  constraints in Go too. A game that moves into a constrained stack cannot
  be ported yet.
- **Timers against a moving clock, in Go's terms.** The corpus clock is
  fixed. The engine's tick validation is tested, but against its own
  contract, not a Go recording. More than one active timer, and ordering
  between timers, appear in no document.
- **Legality with non-default arguments.** Each legality row is evaluated
  with the default-bound move, as in the corpus. Other arguments are covered
  by projected choices (memory, werewolf), by the rejection vectors, and by
  the proposals themselves.
- **Seating at "the next open seat".** Every corpus seat step names its
  index. The unnamed form is covered by a hand-written test only.
- **Variants and player counts outside the corpus.** Memory's `cardset`
  values other than `all`, and its `large` layout only through one golden.
  Blackjack with jokers cannot occur.
- **Werewolf and debuganimations have one short document each** (13 and 16
  versions). Werewolf's night phase, an elimination, a tie, two werewolves
  and the end of the game are not in the corpus, so those parts of
  `games/werewolf.ts` are a reading of the Go, not a proven match.
  Debuganimations' three state-conditional moves are each taken in one
  direction only.
- **Rejections in games other than the four.** The vectors use memory,
  blackjack, tictactoe and pig.
- **Refusals whose text Go does not fix.** A panic's message, and which of
  several bad input fields is named (divergences 5 and 6).
- **Checkers**, the largest declarative-legality example, is not in the
  corpus and is not ported. It needs boards.
- **Performance.** Nothing here is tuned. The slowest document
  (`blackjack/scenario-reshuffle`, 171 versions and 9 viewers) takes about
  one second, most of it producing all nine viewers' output for every
  version; a real host produces one version's views per proposal.

## For whoever picks this up

- **The exporter's `winners: null`** (above) is a one-line fix in
  `export.go` followed by regenerating the corpus. Then delete
  `winnersOrEmpty` in `testing/conformance.ts`.
- **Regenerating the corpus with go1.25.13 rewrites every `.gz`** even when
  nothing changed, because its compressor's output differs from the Go that
  wrote the checked-in files. The decompressed bytes are identical, and the
  staleness test compares those. I restored the untouched files after
  regenerating so the commit adds only the new document.
- **The new scenario slows the Go corpus test** from about 31 s to about
  90 s (`go test ./boardgame-util/lib/conformance/corpus`), since each
  document is exported twice there. It is the only document with a shuffle
  after set-up in a real game. If the time matters more than that coverage,
  remove `blackjackReshuffle` from `corpus.go` and delete the document.
- **`GameSpec` is the surface a Common Fabric pattern would import.** The
  engine is one function per invocation, `propose`, over three stored
  values (`GameRecord`, the move history, the state). `viewerOutput` is the
  per-seat release. The record's `secretSalt` and the stored state must stay
  sealed.
- **A game's `checkGameFinished` is recomputed from the stored state** at
  the start of each call rather than persisted. It must therefore be a
  function of the state alone, which Go's signature already implies.
