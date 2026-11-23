# Checkers fixed in Go, and its port regenerated

Base `ba30f2c4` (branch `claude/transpile-pipeline-e23924`); one local commit,
nothing pushed, no PR. The owner ruled that the checkers example is fixed in
Go and its TypeScript port regenerated through the pipeline. This stream had
the one exception to "existing Go is additive only": `examples/checkers/`.

## Part 1: the Go game

Every defect in the brief was confirmed first, by running the game before
any change (the base corpus, the base `checkersvec` vectors and the
2026-11 fuzz-driver status: 0 legal Move Tokens in 200 seeds), then fixed
with the smallest change. No move, property or declared-legality spec was
renamed or restructured; `ConfigurePredicateConstructors` is untouched (its
source hash is unchanged, so the agent-written block survived).

| # | Defect, as confirmed | Fix (file, function) |
|---|---|---|
| 1 | `spaceIsBlack` was `index % 2 == 0`: on an 8-wide board that is every even column, so tokens were set up in columns and no diagonal step landed on a "black" space | `components.go` `spaceIsBlack`: `(row + col) % 2 == 0`, row and column from the index, so (0,0) is black and colours alternate along rows and columns as its comment says |
| 2 | Nothing set a player's `Color`: both stayed the enum default, red, so the black tokens belonged to nobody and `componentPropEqualsCurrentPlayer` failed for every black token | `main.go` new `(*gameDelegate).BeginSetUp`: player 0 red, player 1 black |
| 3 | Red (placed at the top) used `graphUpward`, black `graphDownward`: each colour moved toward its own back edge | `components.go` `(*token).AllNextSpaces`: red moves down, black up; a crowned token adds the opposite graph, as before |
| 4 | `(*moveMoveToken).Apply` took the "middle" space as finish minus start; for most legal moves `RangeToValue` gave the illegal value and Apply returned "Invalid result from range to value" after Legal passed, otherwise it named an unrelated space | `moves.go` `(*moveMoveToken).Apply`: only a two-row move (a jump) has a middle, the midpoint `(start + finish) / 2`; a step captures nothing. Floor-midpoint for steps would have named a third square for a down-left/up-right step, so the jump test is needed |
| 5 | From the real starting position no move was legal for anyone | Follows from 1 to 3: red now has exactly its 7 opening steps, black none until red moves |

No other defect was found. Things that look odd but are the code's evident
intent, kept and left to the owner (below): captures are not mandatory, and
a multi-jump continuation is not restricted to the jumping token.

Who moves first: player 0 is the first current player and keeps red (the
enum's default). Standard draughts has black (dark) move first; giving
player 0 black would have changed player 0's colour, which the existing
`legal` package's corpus pins (`legal/testdata/conformance/propEquals.json`
and `propNotEquals.json` case 4/5 and `catalog_compare_test.go` expect
`checkersDefault`'s current player to be `Red`). Red first keeps those
tests, an existing Go package, unchanged.

### Go tests

- No existing test in `examples/checkers/` pinned the buggy behaviour:
  `legal_golden_test.go` finds its fixtures by searching the live position
  (first own token with a free next space, an opponent token, a non-black
  space) and passes unchanged; `main_test.go` unchanged.
- New `rules_test.go` (package `checkers`): the starting position
  (checkerboard; red on the black squares of rows 0-2, black on rows 5-7;
  players red and black, 12 tokens each; red's legal openings exactly the 7
  steps from row 2, black none), a step through the real game with Finish
  Turn passing the turn and black answering, a capture (the jumped token
  goes to the capturer's `CapturedTokens`, nothing else moves, the turn ends),
  a jump over one's own token is illegal, a multi-jump (turn stays open
  after the first jump, `TurnDone` errs, ends after the second), crowning
  (Crown Token's default, Legal and Apply; not legal before or after) and
  the crowned token's backward moves, a crowned token's backward step and
  backward jump versus a man's forward-only moves, the end condition at the
  twelfth capture and the score, and every legal Move Token from the start
  and from a crowded position applied on a copy with no Apply error.
- New `fuzz_test.go` (package `checkers_test`, `boardgame-util/lib/fuzz`'s
  random-legal policy at default bounds): seeds 0, 3 and 6 each reach `finished`,
  with no `apply-error` refusal, no legality panic, Move Token, Finish Turn
  and Crown Token each applied, and every Move Token proposed as legal
  applied.

`./scripts/go-local test ./examples/checkers/...`: ok (14 tests: TestManager, the three golden tests, 9 rules tests, TestFuzzGamesFinish; 73 s, nearly all the fuzz test).

## Part 2: everything derived from checkers

All by their commands; no generated file was edited by hand.

| Artifact | Command | Change |
|---|---|---|
| `conformance/definitions/checkers.json` | `gamedef export` | `SetUp.Begin` and the `BeginSetUp` function; new hashes for `spaceIsBlack`, `AllNextSpaces`, `(*moveMoveToken).Apply`; line numbers |
| `boardgame-util/lib/lint/testdata/declarative/checkers.json` | `lib/lint/testdata/gen/declarative` | 21 functions, 20 in the subset (was 20, 19) |
| `conformance/fuzz/checkers/*`, `index.json`, `README.md` | `fuzz corpus` | Bounds 80 proposals / 90 versions (were 40 / 60). 8 documents, 732 versions, 420 KB (was 160 KB); Move Token 255, Finish Turn 246, Crown Token 7 applied; 33 refusals; captures in all 8, multi-jumps in 4, crowning in 5; no game ends inside the bound. The other 48 documents are byte-identical |
| `lib/fuzz/corpus_test.go` | hand | The three checkers `expectedNeverApplied` entries removed: every checkers move is applied |
| `testdata/checkers.vectors.json.gz` | `testdata/gen/checkersvec` | Designed positions redone for the fixed rules (below); random positions now on black squares and with the real colour assignment most of the time. 48 positions, 152 legal Move Tokens, 1,631 applies (1,568 nil-error), 14 legal crowns, 7 legal Finish Turns |
| `rules/testdata/checkers/seed-000{0,1,6}.json.gz` | `fuzz run` | Seed 0 at default bounds (unseated, 481 versions, played to the end, 225 Move Tokens, 10 crownings, 33 refusals, 277 KB); seeds 1 (seated) and 6 at 40 proposals / 60 versions (62 versions each, 2 and 5 refusals, 36 KB each) |
| `games/checkers/{bodies,definition}.gen.ts` | `transpile ./examples/checkers` | New: `checkers_gameDelegate_BeginSetUp`. New hash headers: `spaceIsBlack`, `AllNextSpaces`, `(*moveMoveToken).Apply`. The agent block for `ConfigurePredicateConstructors` survived untouched (its Go is unchanged) |
| `fabric/transpile/REPORT.md`, `reports/checkers.json` | `port --all --dry-run` | checkers 25 units: 4 data, 20 mechanical, 1 agent, 0 engine gap; coverage 8 documents, 732 versions, 33 rejections, never-applied none, Go blocks 139 of 158 |

Designed `checkersvec` positions (all with player 0 red, player 1 black,
except one): the set-up; a step each way; a capture each way; a capture
then another (multi-jump start); mid-jump (turn open, another token could
also move); landing occupied; jump off the board; jump over one's own
token; a capture refused by `same(Color)` (a mixed `CapturedTokens`, the
only Legal-then-Apply-fails case left, and only from an impossible
position); crowned black and crowned red moving and capturing both ways;
crown candidates (each colour on its far row, crowned and not); a finished
turn; one capture from the end; the end; the colours swapped.

The hand-written stand-ins are gone. `rules/checkers.test.ts` now builds
checkers from `checkers.json` with the generated `bodies` and keeps the
loader-level assertions (every named FnRef has a body; a predicate with no
body is the GameSpecError naming it; every compiled plan equals Go's; the
predicate constructor's error text; the header and full replay of the three
documents with their refusals; the facade `{ move }` owner and
`componentInstance`). Its constructed-positions test was dropped: the
generated game's test replays the same positions against the same bodies,
in four slices. `testing/checkers.ts` lost `spaceIsBlackConstructor`;
`games/checkers/checkers.gen.test.ts` loses the stand-in wiring.

## Proving the port

All with zero differences (every version's state, move record and status;
every viewer's game, move, legality and choices; every recorded refusal
refused again with Go's text):

| What | Documents | Versions | Viewer-versions | Rejections | Games finished |
|---|---|---|---|---|---|
| Checked-in corpus `conformance/fuzz/checkers/` (`games/fuzz.test.ts`) | 8 | 732 | 2,928 | 33 | 0 |
| `rules/testdata/checkers/` (both `rules/checkers.test.ts` and `checkers.gen.test.ts`) | 3 | 605 | 2,420 | 40 | 1 |
| Constructed positions (`checkers.gen.test.ts`) | 48 positions | | | | |
| Transient: seeds 0-199, default bounds, outside the repository, `BOARDGAME_FUZZ_DIR` | 200 | 102,226 | 408,904 | 6,928 | 80 |

Constructed positions, as the replay counted them: 38,784 Move Token Legal
texts (152 legal), 5,247 full ledgers, 3,168 Crown Token Legal texts, 48
Finish Turn Legal texts, 1,631 applies (1,568 with the whole state after
compared; 89 captures), 48 end conditions and score lists.

The transient run in Go (10 parallel `fuzz run` processes of 20 seeds each,
merged; 17 minutes): 80 games finished, 118 stopped at 300 proposals, 2
ended `no-action` (seeds 122 and 170: the player to move had no legal
move); 88 seated. Applied: Move Token 47,757, Finish Turn 47,004 (so 753
multi-jump continuations), Crown Token 1,913. Refusals 6,928
(illegal-input 2,316, wrong-actor 2,273, stale-version 2,339); **no
`apply-error`: Apply never failed after Legal passed**; no legality panic.
A finished game took 107 to 262 Move Tokens (mean 211.6).

`port check 'checkers.(*gameDelegate).ConfigurePredicateConstructors' --file <the agent block as a standalone file>`:
PASS (unproven), 8 of 8 documents, 12 of 16 Go blocks, 15 of 24 mutants
killed (was 14 of 24 against the stuck corpus); the survivors are the
constructor-error and unknown-verdict branches no game reaches.

### Difference found, and its fix

1. **Replay harness: a recorded range-enum move field was re-proposed as
   its stored string.** The first documents in which a Move Token is ever
   applied failed at version 26: "the engine refused the recorded propose:
   move "Move Token" field "SpaceIndex": json: cannot unmarshal string into
   Go value of type int". A move record stores an enum field as its string
   (`"4_6"`), but a range-enum field's input codec is `integer`: Go's driver
   (and a client) proposes the key. The engine's `moveWithInput` was right
   (it reproduces Go's refusal of a string); `testing/conformance.ts`
   `runDocument` and `testing/fuzz.ts` `inputFor` passed the record's
   arguments through unchanged. Fixed in the harness: a new
   `proposedInput(game, name, args)` in `testing/conformance.ts`, used by
   both, takes the proposer-supplied fields and turns an `integer`-codec
   enum field's string back into its key. Regression test in
   `rules/checkers.test.ts`, against Go's `facade-moves.vectors.json`
   records (stored blob to key). No engine, loader, translator, facade or
   stock-registry difference was found: captures into the `same(Color)`
   stack, by-key swaps, crowning through the dynamic value, Finish Turn
   with continuation, and the end of the game with winners all matched Go
   on the first run once the harness proposed the right input.

## Coverage (`fuzz coverage --game checkers --blocks`)

| | Before | After |
|---|---|---|
| Documents, versions, rejections | 8, 224, 1 | 8, 732, 33 |
| Move Token / Finish Turn / Crown Token applied | 0 / 0 / 0 | 255 / 246 / 7 |
| Blocks reached, hand-written files | 109 of 180 | 166 of 187 |
| Blocks reached, the report's count (`reports/checkers.json`) | 82 of 151 | 139 of 158 |
| Functions with an unreached block | 16 | 12 |

Still unreached: error branches no play reaches (`BeginSetUp`'s `SetValue`
errors, Apply's swap and range errors, `Crown Token` on an empty space,
`(*token).Legal`'s position error), `Description`, `PlayerScore` (called by
the engine only when the game ends, which no corpus document reaches), the
predicate constructor's unknown-verdict branches, and the end condition's
true branch (likewise).

## Suites

| Command | Outcome |
|---|---|
| `./scripts/go-local build ./...` | passes |
| `./scripts/go-local test ./boardgame-util/... ./moves/... . ./legal/... ./server/api/...` (with `BOARDGAME_NODE` = Node 24) | all ok after one fix: `lib/lint` `TestDeclarativeReproducesBaseline` pinned checkers at 20 functions; see "Edits outside my files". 26 packages of `./boardgame-util/... ./moves/... .` ok, plus `legal`, `server/api`, `server/api/tablelease` |
| `./scripts/go-local test ./examples/checkers/...` | ok |
| `conformance export --check` | current |
| `fuzz corpus --check` | current, 56 documents; the other 48 byte-identical to the base |
| `gamedef export --check` | current |
| `transpile --check ./examples/...` | current |
| `port --all --dry-run --check` | current |
| `check-client` | "Client checks passed" |
| `npm run type-check` | passes |
| `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | passes |
| `npm run test:unit` | 1,426 of 1,426 (one rules test removed, one added) |
| `npm run test:layers` | 11 of 11 |
| `node scripts/fuzz-mutations.mjs historical` | baseline 56 of 56; 20 of 20 killed; no file changed |

## GO-BUGS rows this closes

`fabric/transpile/GO-BUGS.md` belongs to the pair-store stream and is not
on the base. The rows it should mark closed by this fix, if it lists them:
checkers `spaceIsBlack` parity (columns, not a checkerboard); both players
stored red (no player owns black); each colour's graph pointing at its own
back edge; `(*moveMoveToken).Apply` middle square as finish minus start
("Invalid result from range to value" after Legal); no legal move from the
starting position. The proposal's risk 1 (Apply failing after Legal) no
longer happens in play: 0 apply errors in 200 seeds.

## Left for the owner

1. Captures are not mandatory (a step is legal when a jump is available).
   The code never enforced it; kept.
2. Multi-jump continuation: after a capture with another capture available
   from the landing square, `FinishedTurn` stays false and the same player
   moves again, but any of their tokens may move, and a step ends the turn.
   Standard rules require the same token to keep jumping. Restricting it
   needs new state (which token is jumping); not added.
3. Who moves first: red (player 0), to keep the `legal` package's pinned
   fixture; standard draughts has black first.
4. A player with no legal move is stuck: the end condition is only "one
   player captured all 12". In the transient run 2 of 200 games
   ended `no-action`; standard rules make that player lose.
5. A man reaching the far row by a jump stops there (no backward
   continuation until crowned, and Crown Token runs as a fix-up after the
   move): standard.

## Edits outside my files

- `boardgame-util/lib/lint/declarative_test.go` (existing test in an
  existing package): `measuredDifferences["checkers"]` gains
  `"functions": 21, "inSubset": 20`, with a comment. The fix needs one new
  function (`BeginSetUp`; no existing hook can set a player's colour), and
  this table pins checkers' function count. Nothing else in the test
  changed.
- `server/static/src/engine/testing/conformance.ts`, `testing/fuzz.ts`: the
  `proposedInput` harness fix above.
- `server/static/src/engine/rules/data.test.ts`: the ComponentLegal test
  took the predicate from the deleted stand-in; it now takes every body
  from the generated `bodies` (it never evaluates the predicate).
- Regenerated by their own generators because they record live Go checkers
  states, now changed: `testdata/constraints.vectors.json` (`stackvec`:
  only the four checkers scenarios, 30-33), `legal/testdata/fixtures.json`,
  `plans.json`, `vectors.json` (`legal/testdata/gen`). `facadevec`,
  `libvec` and `defs` were rerun and wrote identical files. The TypeScript
  suites over all of them pass.
- `conformance/fuzz/index.json`, `README.md`: checkers entries only.
