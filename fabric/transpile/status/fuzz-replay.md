# Workstream 5: fuzz replay, coverage and mutations

`fabric/TRANSPILE-PROPOSAL.md`, "The differential fuzzer" (Format, The
divergence report, Coverage, Mutations). Base `cc6a81d6`.

## What was built

| Piece | Where |
|---|---|
| Fuzz replay and divergence report | `server/static/src/engine/testing/fuzz.ts` |
| The six hand ports against the fuzz corpus, plus `BOARDGAME_FUZZ_DIR` | `server/static/src/engine/games/fuzz.test.ts` |
| Mutations: the 20 historical ones as data, the body-mutant generator | `server/static/src/engine/testing/mutate.ts`, `testing/mutate.test.ts` |
| Mutation runner (Node, child process per mutant) | `server/static/scripts/fuzz-mutations.mjs` |
| Coverage: moves, predicates, visibility, Go blocks | `boardgame-util/lib/fuzz/coverage.go`, `coverage_test.go` |
| `boardgame-util fuzz coverage` | `boardgame-util/cmd_fuzz.go`, `cmd_fuzz_coverage_test.go` |
| Hand-port differences found | `fabric/transpile/HAND-PORT-BUGS.md` (none) |

## Exported interfaces

`testing/fuzz.ts`:

- `runFuzzDocument(game: CompiledGame, doc: FuzzDocument, options?: { spec?: () => GameSpec }): FuzzResult`.
  Calls `runDocument` unchanged for versions and views, then feeds the same
  outside inputs a second time and, after each settled version V, proposes
  every rejection with `atVersion === V` (`propose` with name, input,
  proposer, expectedVersion; `timer` as a forced tick; `seat` as a seat
  input for `input.TargetPlayerIndex`, else the next open seat). Each must
  return exactly Go's `error`, no versions, and leave the caller's state and
  history deep-equal to a copy taken before. A rejection at a version the
  replay never settled at counts as a mismatch.
- `FuzzResult { name, document: DocumentResult, rejections: {compared, mismatches, examples}, recordedRejections, passed, report?: DivergenceReport }`.
- `DivergenceReport { document, version, viewer?, field, path, want, got, move?, lastOutsideInput?, bodiesExecuted?, repro }`,
  the proposal's shape. `field` is one of `FIELDS`, `'rejection'` or
  `'replay'` (the replay could not continue). The first difference is the
  earliest version, then state, move, status, then each viewer in document
  order with game, move, legality, choices; a rejection after version V
  comes after every field of V. `want`/`got` are recomputed at the path
  (the header falls back to the runner's text). `repro` is
  `boardgame-util fuzz run --game G --seed N [--policy P] --max-proposals M --max-versions V --out DIR`,
  with the bounds read from `source.description` (V is the failing version,
  or `atVersion + 1` for a rejection, since the driver stops before acting
  at `MaxVersions`).
- `instrumentBodies(spec, count: BodyCounter): GameSpec`: a copy of the spec
  in which every function calls `count(path)` first, e.g.
  `moves[Place Token].apply`. With `options.spec`, a failing document is
  replayed once more on an instrumented copy and `bodiesExecuted` lists the
  bodies run by the step that produced the failing version plus that
  version's views (or by the refused proposal alone).
- `reproCommand`, `rejectionInput`, `valueAt`, `fuzzPassTable`, types `FuzzDocument`, `Rejection`, `BodyCounter`, `FuzzOptions`.

`testing/mutate.ts`: `HISTORICAL_MUTATIONS: SourceMutation[]` (`{id, description, failedOn, file, find, replace, note?}`),
`applyMutation`, `occurrences`, `bodyMutants(ts, fileName, source, functionName): BodyMutants`,
`LOOP_GUARD_LIMIT`, `LOOP_GUARD_COUNTER`.

`boardgame-util/lib/fuzz`: `LoadDocuments(dir, game)`, `Coverage(docs) (Report, error)`,
`Report{Game, Documents, Versions, Rejections, Moves []MoveCoverage, Predicates []PredicateCoverage, Visibility []VisibilityCoverage, Blocks []FuncCoverage}`
(JSON tags lower camel), `(Report).Gaps() Gaps`, `(Report).Text()`,
`ParseProfile`, `BlockCoverage(blocks, importPath, pkgDir)`, `ReadBlockCoverage`,
`BuildBlockProfile(root, workDir, goCmd, log)`, `GamePackage`, `ModulePath`.

CLI: `boardgame-util fuzz coverage [--game G] [--dir DIR] [--blocks [--work DIR] | --profile FILE] [--json] [--root DIR]`.

## Results (all run)

### The six hand ports against the fuzz corpus

All 48 checked-in documents pass on the first run; 97 rejections compared,
all equal (blackjack 9, debuganimations 22, memory 18, pig 25, tictactoe 5,
werewolf 18, the index's counts). Every version, every viewer of every
version and every rejection of each document is asserted compared. No hand
port and no engine file was changed; `HAND-PORT-BUGS.md` records none.
`EXPECTED_FAILURES` is empty.

A transient run, generated outside the repository with
`boardgame-util fuzz run --game G --seeds 100..111 --max-versions 150 --out DIR`
for each of the six games (default max proposals 300, refusal rate 1/8),
then replayed with `BOARDGAME_FUZZ_DIR=DIR node --test games/fuzz.test.ts`:
**72 of 72 documents pass**, 545 rejections compared, all equal.

| Game | Documents | Versions | Rejections | Pass |
|---|---|---|---|---|
| blackjack | 12 | 1,531 | 24 | 12 |
| debuganimations | 12 | 1,072 | 173 | 12 |
| memory | 12 | 1,638 | 128 | 12 |
| pig | 12 | 1,820 | 140 | 12 |
| tictactoe | 12 | 212 | 11 | 12 |
| werewolf | 12 | 548 | 69 | 12 |

(Seeds 100-111 take each game's derived setup as it comes, so werewolf's
unseated seeds and debuganimations' seated seeds, dead ends in Go, are
included; they still replay exactly.) The whole file then ran 125 tests in
41 s. The proposal's `--seeds 0..999` with default bounds was not run: a
first attempt with default bounds wrote 12 memory documents in about six
minutes and was stopped.

Added run time: `games/fuzz.test.ts` is 51 tests (one per document, plus the
corpus-shape check, the pass table and the divergence-report self-test);
alone it takes 7.3 s; inside `npm run test:unit` the 51 fuzz tests sum to
9.6 s of test time and the slowest single test is 1.0 s (the 10 s budget and
20 s timeout are untouched). `testing/mutate.test.ts` adds 4 tests, under
0.3 s.

The report was checked on recordings altered in memory
(`the divergence report names the first difference`): one state value of
`tictactoe/seed-0001` v5 gives `{version 5, field state, path
Game.Slots.Indexes[2], want 7, got 0, move Place Token by 0 (propose),
lastOutsideInput at 4, bodiesExecuted including moves[Place Token].*, repro
... --max-proposals 60 --max-versions 5 ...}`; one word of a rejection's text
gives `{field rejection, version 14, got {error: "it's not your turn"}}`.

### The 20 historical mutations against the fuzz corpus

`node scripts/fuzz-mutations.mjs historical` (from `server/static`): the
unmutated baseline passes 48 of 48, then each mutation runs in its own
process against all 48 documents. **All 20 are killed.**

| Mutation | File | Failed on (conformance corpus) | Fuzz documents failing | First field per failing document |
|---|---|---|---|---|
| Blackjack ace rule `>=` to `>` | games/blackjack.ts | view.game, view.legality | 5 of 48 | view.game 4, state 1 |
| The null-winners rule removed | testing/conformance.ts | status, view.game | 40 of 48 | status 40 |
| Memory's two fix-ups swapped | games/memory.ts | header, then the fix-up chain | 8 of 48 | header 8 |
| Hole card policy `len` to `order` | games/blackjack.ts | view.game | 8 of 48 | view.game 8 |
| Timer id drawn with 15 draws, not 16 | rules/state.ts | state, view.game | 11 of 48 | state 11 |
| Computed properties from the unsanitized state | rules/view.ts | view.game | 16 of 48 | view.game 16 |
| A move field's `sanitize` tag dropped | games/tictactoe.ts | view.move | 8 of 48 | view.move 8 |
| The hidden-name alias dropped | games/tictactoe.ts | header, view.move | 8 of 48 | header 8 |
| One extra draw before pig's starting player | games/pig.ts | replay stops | 7 of 48 | state 7 |
| An error string changed by one word | games/blackjack.ts | view.legality | 8 of 48 | view.legality 8 |
| Empty choice sets kept | rules/view.ts | view.choices | 16 of 48 | view.choices 16 |
| Round-robin's starting point off by one | rules/stock.ts | the deal order | 8 of 48 | state 8 |
| A seated player left active | rules/stock.ts | the gathering chain | 25 of 48 | state 25 |
| Repeatable moves not repeatable in a progression | rules/stock.ts | the deal chain | 16 of 48 | status 12, state 4 |
| `moves.AnyPlayer`'s target not set to the proposer | rules/stock.ts | replay stops | 8 of 48 | view.legality 8 |
| Role value shown regardless of visibility | rules/view.ts | view.game | 8 of 48 | view.game 8 |
| `same-role` dropped from a vote's visibility | games/werewolf.ts | view.move | 7 of 48 | view.move 7 |
| Tree enum default changed | games/debuganimations.ts | header, state | 8 of 48 | header 8 |
| Public shuffle made secret | games/debuganimations.ts | state, view.game | 6 of 48 | state 6 |
| Timer's move bound to the state being built | rules/state.ts | state | 8 of 48 | state 8 |

The run hashes every file under `src/engine/` and `conformance/` before and
after and fails if any changed (it did once, correctly, when I edited a file
during a run); the final run reported no change. `git status` after all runs
shows only the files listed in this report.

### The body-mutant generator

`node scripts/fuzz-mutations.mjs body --file F --function NAME --docs DIR`.
Each mutant is run until one document fails. Results against each game's
eight fuzz documents:

| Function | Mutants | Killed | Survivors |
|---|---|---|---|
| `games/pig.ts` `pig` | 31 | 24 | 7: `turnDone`'s DieCounted check (2); `state.g.TargetScore = DEFAULT_TARGET_SCORE` in finish (equivalent: the prop's `initial` already sets it, as Go's constructor does); the whole end condition (4: no pig run reaches 100 points) |
| `games/tictactoe.ts` `boardResult` | 32 | 25 | 7: both diagonal runs (deleted, or their indexes shifted): no corpus game is won on a diagonal |
| `games/tictactoe.ts` `placeToken` | 13 | 13 | none |
| `games/blackjack.ts` `handValue` | 31 | 28 | 3: `if (index === null) continue;` (2), `aces--` |
| `games/memory.ts` `revealCard` | 13 | 13 | none |
| `games/memory.ts` `hideCards` | 9 | 9 | none |
| `games/werewolf.ts` `resolveVotes` | 52 | 42 | 10: skipping inactive/eliminated voters (2), skipping non-wolves at night (2), `count === 0` skip (2), four literal shifts |
| `games/werewolf.ts` `castVote` | 34 | 30 | 4: the day-phase check of the night vote and the vote-target range check (each the `if` and its `return`) |

Survivors are corpus-adequacy findings, not hidden: each names a branch the
eight seeds never exercise (or, for `TargetScore`, an equivalent mutant).

### Coverage

`./scripts/go-local run ./boardgame-util fuzz coverage --blocks` over the
checked-in corpus. Block coverage under `./scripts/go-local` **works**, with
one condition the proposal did not know: the command's own main package must
be in `-coverpkg`. Built with `-coverpkg=.../examples/...` alone, the binary
ran and wrote **nothing** to `GOCOVERDIR` (empty directory, exit 0); with
`-coverpkg=github.com/jkomoros/boardgame/boardgame-util,github.com/jkomoros/boardgame/examples/...`
it writes `covmeta`/`covcounters` and `go tool covdata textfmt` converts
them. `--blocks` builds that binary, runs `fuzz corpus --check` under it
(the 48 runs regenerated in memory, so exactly the recorded runs), converts
the counters, and assigns blocks to functions with `go/parser`. Generated
files (`auto_*.go`) are counted but not listed. `--profile FILE` reads any
text profile instead.

Measures and their definitions:

1. Moves: per move type, versions applied (any trigger), accepted outside
   proposals (trigger `propose`), rejections naming it. Cross-checked against
   the driver's own counts in `index.json` (`TestCoverageMatchesIndex`).
2. Predicates: per precondition (move, position, name, args) of each declared
   plan, the pass/fail/unknown verdicts over every viewer of every version.
3. Go blocks: as above.
4. Visibility: a property is reported when some viewer other than the admin
   is, at some version, sent it without the `values` facet
   (`CurrentState.Visibility`). For it, the (version, viewer) pairs where a
   viewer other than the owner and the admin got it sanitized while the
   stored value was non-zero (null, false, 0, "", all-negative stack
   indexes, a timer without id, all-zero objects count as zero). The gap is
   a property never sanitized while non-zero: then hiding it was never
   tested on a value worth hiding. Also counted: non-owners shown its values
   while non-zero.

The report as printed, per game:

#### blackjack

8 documents, 527 versions, 9 rejections.

##### Moves

| Move | Fix-up | Applied | Proposed | Refused |
|---|---|---|---|---|
| Shuffle Discard Into Draw | yes | 0 | 0 | 0 |
| Seat Player | yes | 11 | 0 | 0 |
| Activate Inactive Player - Gathering | yes | 8 | 0 | 0 |
| Wait For Enough Players - Gathering | yes | 8 | 0 | 0 |
| Inactivate Empty Seat - Gathering | yes | 3 | 0 | 0 |
| Begin Game | yes | 8 | 0 | 0 |
| Start Round Cleanup | yes | 19 | 0 | 0 |
| Current Player Hit |  | 40 | 40 | 3 |
| Current Player Stand |  | 37 | 37 | 3 |
| Reveal Hidden Card | yes | 60 | 0 | 0 |
| Finish Turn | yes | 35 | 0 | 0 |
| Force Finish Turn |  | 0 | 0 | 3 |
| Activate Inactive Player | yes | 7 | 0 | 0 |
| Wait For Enough Players | yes | 25 | 0 | 0 |
| Inactivate Empty Seat | yes | 6 | 0 | 0 |
| Deal Initial Hidden Card | yes | 67 | 0 | 0 |
| Deal Initial Visible Card | yes | 67 | 0 | 0 |
| Start Phase Normal Play | yes | 25 | 0 | 0 |
| Accumulate Scores | yes | 19 | 0 | 0 |
| Collect Cards | yes | 19 | 0 | 0 |
| Reset Player For New Round | yes | 19 | 0 | 0 |
| Increment Rounds Completed | yes | 19 | 0 | 0 |
| Start Next Round | yes | 17 | 0 | 0 |

##### Predicates

| Move | # | Predicate | Pass | Fail | Unknown | Missing |
|---|---|---|---|---|---|---|
| Current Player Hit | 0 | inPhase(2) | 1072 | 1698 | 0 | unknown |
| Current Player Hit | 1 | proposerIsCurrentPlayer() | 1054 | 1716 | 0 | unknown |
| Current Player Hit | 2 | mayMoveFirstTo(game.DrawStack, player.VisibleHand) | 2770 | 0 | 0 | fail, unknown |
| Current Player Hit | 3 | playerBool(Eliminated, false) | 2554 | 216 | 0 | unknown |
| Current Player Hit | 4 | stackNotEmpty(game.DrawStack) | 2770 | 0 | 0 | fail, unknown |
| Current Player Hit | 5 | custom() | 2581 | 189 | 0 | unknown |
| Current Player Stand | 0 | inPhase(2) | 1072 | 1698 | 0 | unknown |
| Current Player Stand | 1 | proposerIsCurrentPlayer() | 1054 | 1716 | 0 | unknown |
| Current Player Stand | 2 | playerBool(Eliminated, false) | 2554 | 216 | 0 | unknown |
| Current Player Stand | 3 | playerBool(Stood, false) | 2423 | 347 | 0 | unknown |

##### Visibility

| Property | Sanitized for a non-owner while non-zero | Shown to a non-owner while non-zero | Facets non-owners got |
|---|---|---|---|
| Game.DiscardStack | 1317 | 0 | [count,nonempty] |
| Game.DrawStack | 2243 | 0 | [count,nonempty] |
| Players.HiddenHand | 3011 | 0 | [count,nonempty] |

##### Go blocks (hand-written functions with an unreached block)

137 of 151 blocks reached in hand-written files; 10 functions with an unreached block.

| Function | At | Reached | Unreached blocks |
|---|---|---|---|
| (*gameDelegate).Description | main.go:46 | 0/1 | 47.2-48.1 |
| (*gameDelegate).BeginSetUp | main.go:82 | 2/3 | 86.3-87.1 |
| (*gameDelegate).DistributeComponentToStarterStack | main.go:100 | 4/5 | 107.3-108.1 |
| (*moveAccumulateScores).Legal | moves.go:89 | 6/8 | 96.4-96.12 103.2-103.60 |
| (*moveCollectCards).Legal | moves.go:127 | 5/6 | 137.2-137.53 |
| (*moveCollectCards).Apply | moves.go:140 | 4/6 | 144.4-145.1 147.4-148.1 |
| (*moveResetPlayerForNewRound).Legal | moves.go:160 | 6/8 | 166.4-166.12 173.2-173.55 |
| (*moveIncrementRoundsCompleted).Legal | moves.go:192 | 7/9 | 201.4-202.1 207.3-208.1 |
| (*moveCurrentPlayerHit).Apply | moves.go:244 | 9/10 | 247.3-248.1 |
| (*moveRevealHiddenCard).LegalCustom | moves.go:320 | 4/5 | 327.3-328.1 |

##### Gaps

- Never applied: Shuffle Discard Into Draw; Force Finish Turn
- Never refused (non-fix-up): none
- Predicates missing a verdict: 10 of 10
- Hidden properties never sanitized for a non-owner while non-zero: none
- Hand-written functions with an unreached block: 10

#### debuganimations

8 documents, 168 versions, 22 rejections.

##### Moves

| Move | Fix-up | Applied | Proposed | Refused |
|---|---|---|---|---|
| Seat Player | yes | 0 | 0 | 0 |
| Move Card Between Short Stacks |  | 24 | 24 | 5 |
| Move Card Between Draw And Discard Stacks |  | 24 | 24 | 2 |
| Flip Card Between Hidden and Revealed |  | 12 | 12 | 1 |
| Move Fan Card |  | 13 | 13 | 2 |
| Visible Shuffle |  | 10 | 10 | 0 |
| Shuffle |  | 14 | 14 | 0 |
| Move Between Hidden |  | 12 | 12 | 0 |
| Move Token |  | 12 | 12 | 1 |
| Move Token Sanitized |  | 19 | 19 | 1 |
| Start Move All Components To Hidden |  | 7 | 7 | 5 |
| Start Move All Components To Visible |  | 1 | 1 | 4 |
| Shuffle Hidden |  | 12 | 12 | 1 |

##### Predicates

| Move | # | Predicate | Pass | Fail | Unknown | Missing |
|---|---|---|---|---|---|---|
| Visible Shuffle | 0 | stackCount(game.FanStack, >, 1) | 672 | 0 | 0 | fail, unknown |
| Shuffle | 0 | stackCount(game.FanStack, >, 1) | 672 | 0 | 0 | fail, unknown |
| Start Move All Components To Hidden | 0 | stackNotEmpty(game.AllVisibleStack) | 416 | 256 | 0 | unknown |
| Start Move All Components To Hidden | 1 | stackEmpty(game.AllHiddenStack) | 416 | 256 | 0 | unknown |
| Start Move All Components To Visible | 0 | stackNotEmpty(game.AllHiddenStack) | 256 | 416 | 0 | unknown |
| Start Move All Components To Visible | 1 | stackEmpty(game.AllVisibleStack) | 256 | 416 | 0 | unknown |
| Shuffle Hidden | 0 | stackNotEmpty(game.FanDiscard) | 504 | 0 | 0 | fail, unknown |

##### Visibility

| Property | Sanitized for a non-owner while non-zero | Shown to a non-owner while non-zero | Facets non-owners got |
|---|---|---|---|
| Game.AllHiddenStack | 192 | 0 | [count,occupancy,order,nonempty] |
| Game.DrawStack | 504 | 0 | [count,occupancy,order,nonempty] |
| Game.FanDiscard | 504 | 0 | [count,occupancy,order,nonempty] |
| Game.FirstShortStack | 387 | 0 | [count,occupancy,order,nonempty] |
| Game.HiddenCard | 402 | 0 | [count,occupancy,order,nonempty] |
| Game.HiddenStack | 504 | 0 | [nonempty] |
| Game.SanitizedTokensTo | 504 | 0 | [nonempty] |
| Game.SecondShortStack | 360 | 0 | [count,occupancy,order,nonempty] |

##### Go blocks (hand-written functions with an unreached block)

174 of 195 blocks reached in hand-written files; 15 functions with an unreached block.

| Function | At | Reached | Unreached blocks |
|---|---|---|---|
| (*gameDelegate).Description | main.go:44 | 0/1 | 45.2-46.1 |
| (*gameDelegate).DefaultNumPlayeres | main.go:48 | 0/1 | 49.2-50.1 |
| (*moveMoveCardBetweenShortStacks).Legal | moves.go:137 | 8/9 | 140.3-141.1 |
| (*moveMoveCardBetweenShortStacks).Apply | moves.go:163 | 6/7 | 176.3-177.1 |
| (*moveMoveCardBetweenDrawAndDiscardStacks).Legal | moves.go:202 | 8/9 | 205.3-206.1 |
| (*moveMoveCardBetweenDrawAndDiscardStacks).Apply | moves.go:228 | 6/7 | 241.3-242.1 |
| (*moveFlipHiddenCard).Legal | moves.go:257 | 5/8 | 260.3-261.1 266.3-267.1 270.3-271.1 |
| (*moveFlipHiddenCard).Apply | moves.go:276 | 6/7 | 289.3-290.1 |
| (*moveMoveCardBetweenFanStacks).Legal | moves.go:305 | 6/8 | 308.3-309.1 321.2-321.62 |
| (*moveMoveBetweenHidden).Legal | moves.go:395 | 6/8 | 398.3-399.1 411.2-411.53 |
| (*moveMoveToken).Legal | moves.go:436 | 6/8 | 439.3-440.1 452.2-452.54 |
| (*moveMoveTokenSanitized).Legal | moves.go:477 | 6/8 | 480.3-481.1 493.2-493.54 |
| (*moveStartMoveAllComponentsToHidden).Apply | moves.go:532 | 4/5 | 538.4-539.1 |
| (*moveStartMoveAllComponentsToVisible).Apply | moves.go:560 | 4/5 | 566.4-567.1 |
| (*moveShuffleHidden).Apply | moves.go:593 | 4/5 | 598.3-599.1 |

##### Gaps

- Never applied: Seat Player
- Never refused (non-fix-up): Visible Shuffle; Shuffle; Move Between Hidden
- Predicates missing a verdict: 7 of 7
- Hidden properties never sanitized for a non-owner while non-zero: none
- Hand-written functions with an unreached block: 15

#### memory

8 documents, 269 versions, 18 rejections.

##### Moves

| Move | Fix-up | Applied | Proposed | Refused |
|---|---|---|---|---|
| Seat Player | yes | 24 | 0 | 0 |
| Activate Inactive Player | yes | 24 | 0 | 0 |
| Reveal Card |  | 92 | 92 | 14 |
| Hide Cards |  | 36 | 16 | 4 |
| Finish Turn | yes | 39 | 0 | 0 |
| Capture Cards | yes | 3 | 0 | 0 |
| Start Hide Cards Timer | yes | 43 | 0 | 0 |

##### Predicates

| Move | # | Predicate | Pass | Fail | Unknown | Missing |
|---|---|---|---|---|---|---|
| Reveal Card | 0 | proposerIsCurrentPlayer() | 538 | 1144 | 0 | unknown |
| Reveal Card | 1 | propAtLeast(player.CardsLeftToReveal, 1) | 727 | 955 | 0 | unknown |
| Reveal Card | 2 | revealableCardAt(game.HiddenCards, game.VisibleCards, move.CardIndex) | 1682 | 0 | 0 | fail, unknown |
| Reveal Card | 3 | mayMoveToSlot(game.HiddenCards, game.VisibleCards, move.CardIndex, move.CardIndex) | 1682 | 0 | 0 | fail, unknown |
| Hide Cards | 0 | proposerIsCurrentPlayer() | 538 | 1144 | 0 | unknown |
| Hide Cards | 1 | propCompare(player.CardsLeftToReveal, <=, 0) | 955 | 727 | 0 | unknown |
| Hide Cards | 2 | stackNotEmpty(game.VisibleCards) | 916 | 766 | 0 | unknown |

##### Visibility

| Property | Sanitized for a non-owner while non-zero | Shown to a non-owner while non-zero | Facets non-owners got |
|---|---|---|---|
| Game.HiddenCards | 1413 | 0 | [count,occupancy,order,nonempty] |

##### Go blocks (hand-written functions with an unreached block)

130 of 243 blocks reached in hand-written files; 17 functions with an unreached block.

| Function | At | Reached | Unreached blocks |
|---|---|---|---|
| (*Agent).DisplayName | agent.go:45 | 0/1 | 46.2-47.1 |
| (*Agent).SetUpForGame | agent.go:50 | 0/7 | 51.2-54.1 56.2-57.1 58.2-58.15 59.3-60.1 62.2-62.16 63.3-65.1 67.2-67.13 |
| (*Agent).ProposeMove | agent.go:71 | 0/31 | 73.2-74.1 75.2-75.15 76.3-77.1 79.2-79.54 80.3-82.1 84.2-85.1 86.2-87.1 89.2-90.1 92.2-92.56 93.3-93.15 94.4-94.12 96.3-97.35 98.4-98.17 99.5-100.1 101.4-101.17 105.2-105.31 106.3-106.16 107.4-108.1 109.3-109.16 112.2-112.39 116.3-116.50 119.4-123.1 125.3-125.50 129.4-133.1 140.2-140.12 142.3-144.1 145.3-145.16 146.4-147.1 149.3-149.17 150.4-151.1 154.2-154.8 |
| (*agentState).PerhapsForgetCard | agent.go:158 | 0/5 | 159.2-159.26 160.3-161.1 162.2-162.40 163.3-165.1 166.2-166.14 |
| (*agentState).CullInvalidCards | agent.go:170 | 0/5 | 171.2-173.27 174.3-175.67 177.4-178.12 180.3-181.21 184.2-184.20 |
| (*agentState).CardSeen | agent.go:189 | 0/8 | 192.2-192.35 193.3-193.49 194.4-195.1 200.2-203.1 205.2-206.1 207.2-207.39 209.3-210.1 212.2-212.13 |
| (*agentState).FirstCardToFlip | agent.go:217 | 0/15 | 220.2-221.1 222.2-224.1 225.2-225.35 226.3-226.29 228.4-229.9 231.3-231.32 234.2-234.23 235.3-235.16 236.4-237.1 239.3-239.36 240.4-240.33 241.5-242.1 246.2-246.15 247.3-248.1 252.2-252.36 |
| (*agentState).PickRandomCard | agent.go:256 | 0/11 | 257.2-257.6 258.3-259.1 261.3-261.62 262.4-262.12 266.3-267.1 268.3-268.36 269.4-269.27 270.5-271.10 275.3-275.10 276.4-276.12 280.3-280.15 |
| (*agentState).SecondCardToFlip | agent.go:285 | 0/16 | 286.2-287.1 288.2-289.1 290.2-290.56 291.3-291.15 292.4-292.12 294.3-294.45 297.2-297.23 298.3-300.1 303.2-303.35 304.3-304.32 305.4-305.17 306.5-307.1 308.4-308.21 312.2-312.15 313.3-314.1 317.2-317.36 |
| (*gameDelegate).Description | main.go:44 | 0/1 | 45.2-46.1 |
| (*gameDelegate).DefaultNumPlayeres | main.go:48 | 0/1 | 49.2-50.1 |
| (*gameDelegate).BeginSetUp | main.go:141 | 9/12 | 154.3-154.21 158.3-159.1 161.3-162.1 |
| (*gameDelegate).FinishSetUp | main.go:176 | 23/28 | 211.4-212.1 226.5-227.1 229.5-230.1 233.5-234.1 236.5-237.1 |
| (*gameDelegate).GameEndConditionMet | main.go:284 | 3/4 | 291.2-291.13 |
| (*gameDelegate).PlayerScore | main.go:294 | 0/1 | 295.2-296.1 |
| (*moveRevealCard).Apply | moves.go:72 | 4/5 | 75.3-76.1 |
| (*moveHideCards).Apply | moves.go:186 | 6/7 | 195.5-196.1 |

##### Gaps

- Never applied: none
- Never refused (non-fix-up): none
- Predicates missing a verdict: 7 of 7
- Hidden properties never sanitized for a non-owner while non-zero: none
- Hand-written functions with an unreached block: 17

#### pig

8 documents, 413 versions, 25 rejections.

##### Moves

| Move | Fix-up | Applied | Proposed | Refused |
|---|---|---|---|---|
| Seat Player | yes | 14 | 0 | 0 |
| Activate Inactive Player | yes | 14 | 0 | 0 |
| Roll Dice |  | 91 | 91 | 10 |
| Done Turn |  | 83 | 83 | 15 |
| Count Die | yes | 99 | 0 | 0 |
| Finish Turn | yes | 104 | 0 | 0 |

##### Predicates

| Move | # | Predicate | Pass | Fail | Unknown | Missing |
|---|---|---|---|---|---|---|
| Roll Dice | 0 | proposerIsCurrentPlayer() | 826 | 2065 | 0 | unknown |
| Roll Dice | 1 | playerBool(DieCounted) | 2156 | 735 | 0 | unknown |
| Done Turn | 0 | proposerIsCurrentPlayer() | 826 | 2065 | 0 | unknown |
| Done Turn | 1 | playerBool(DieCounted) | 2156 | 735 | 0 | unknown |
| Done Turn | 2 | playerBool(Done, false) | 2309 | 582 | 0 | unknown |

##### Visibility

Every property is shown to every viewer.

##### Go blocks (hand-written functions with an unreached block)

65 of 69 blocks reached in hand-written files; 4 functions with an unreached block.

| Function | At | Reached | Unreached blocks |
|---|---|---|---|
| (*gameDelegate).Description | main.go:43 | 0/1 | 44.2-45.1 |
| (*gameDelegate).GameEndConditionMet | main.go:78 | 4/5 | 83.4-84.1 |
| (*gameDelegate).DynamicComponentValuesConstructor | main.go:118 | 2/3 | 124.2-124.12 |
| (*playerState).TurnDone | state.go:43 | 4/5 | 45.3-46.1 |

##### Gaps

- Never applied: none
- Never refused (non-fix-up): none
- Predicates missing a verdict: 5 of 5
- Hidden properties never sanitized for a non-owner while non-zero: none
- Hand-written functions with an unreached block: 4

#### tictactoe

8 documents, 134 versions, 5 rejections.

##### Moves

| Move | Fix-up | Applied | Proposed | Refused |
|---|---|---|---|---|
| Seat Player | yes | 8 | 0 | 0 |
| Activate Inactive Player | yes | 8 | 0 | 0 |
| Place Token |  | 59 | 59 | 5 |
| Finish Turn | yes | 51 | 0 | 0 |

##### Predicates

| Move | # | Predicate | Pass | Fail | Unknown | Missing |
|---|---|---|---|---|---|---|
| Place Token | 0 | proposerIsCurrentPlayer() | 266 | 0 | 0 | fail, unknown |
| Place Token | 1 | mayMoveFirstToSlot(players[move.TargetPlayerIndex].UnusedTokens, game.Slots, move.Slot) | 264 | 2 | 0 | unknown |

##### Visibility

Every property is shown to every viewer.

##### Go blocks (hand-written functions with an unreached block)

108 of 119 blocks reached in hand-written files; 9 functions with an unreached block.

| Function | At | Reached | Unreached blocks |
|---|---|---|---|
| (*Agent).DisplayName | agent.go:16 | 0/1 | 17.2-18.1 |
| (*Agent).SetUpForGame | agent.go:21 | 0/1 | 22.2-23.1 |
| (*Agent).ProposeMove | agent.go:26 | 0/1 | 27.2-28.1 |
| (*gameDelegate).DistributeComponentToStarterStack | main.go:28 | 5/6 | 39.2-39.59 |
| (*gameDelegate).Description | main.go:62 | 0/1 | 63.2-64.1 |
| checkGameFinished | main.go:228 | 21/24 | 240.3-241.1 281.3-282.1 287.3-288.1 |
| checkRunWon | main.go:305 | 8/9 | 308.3-309.1 |
| (*movePlaceToken).Apply | moves.go:57 | 5/6 | 60.3-61.1 |
| (*gameState).tokenValue | state.go:33 | 0/1 | 34.2-35.1 |

##### Gaps

- Never applied: none
- Never refused (non-fix-up): none
- Predicates missing a verdict: 2 of 2
- Hidden properties never sanitized for a non-owner while non-zero: none
- Hand-written functions with an unreached block: 9

#### werewolf

8 documents, 157 versions, 18 rejections.

##### Moves

| Move | Fix-up | Applied | Proposed | Refused |
|---|---|---|---|---|
| Seat Player | yes | 33 | 0 | 0 |
| Activate Inactive Player - Gathering | yes | 32 | 0 | 0 |
| Wait For Enough Players - Gathering | yes | 8 | 0 | 0 |
| Inactivate Empty Seat - Gathering | yes | 7 | 0 | 0 |
| Begin Game | yes | 8 | 0 | 0 |
| Cast Vote |  | 31 | 31 | 13 |
| Resolve Votes | yes | 8 | 0 | 0 |
| Resolve Day Votes on Timeout |  | 4 | 0 | 2 |
| Cast Night Vote |  | 8 | 8 | 3 |
| Resolve Night Votes | yes | 8 | 0 | 0 |
| Resolve Night Votes on Timeout |  | 2 | 0 | 0 |

##### Predicates

No move declares a legal plan.

##### Visibility

| Property | Sanitized for a non-owner while non-zero | Shown to a non-owner while non-zero | Facets non-owners got |
|---|---|---|---|
| Players.FellowWolves | 0 | 0 | [] |
| Players.NightVote | 3812 | 0 | [] |
| Players.Role | 3821 | 0 | [] |

##### Go blocks (hand-written functions with an unreached block)

186 of 195 blocks reached in hand-written files; 7 functions with an unreached block.

| Function | At | Reached | Unreached blocks |
|---|---|---|---|
| (*gameDelegate).Description | main.go:37 | 0/1 | 38.2-39.1 |
| (*moveBeginGame).Apply | moves.go:42 | 17/19 | 44.3-45.1 59.3-60.1 |
| startVoteTimer | moves.go:84 | 6/7 | 95.3-96.1 |
| populateFellowWolves | moves.go:101 | 7/8 | 113.5-114.1 |
| (*moveCastVote).Legal | moves.go:175 | 21/23 | 199.3-200.1 205.3-206.1 |
| (*moveResolveVotes).Legal | moves.go:277 | 14/15 | 286.3-287.1 |
| finishVoteResolution | moves.go:338 | 4/5 | 342.3-343.1 |

##### Gaps

- Never applied: none
- Never refused (non-fix-up): Resolve Night Votes on Timeout
- Predicates missing a verdict: 0 of 0
- Hidden properties never sanitized for a non-owner while non-zero: Players.FellowWolves
- Hand-written functions with an unreached block: 7


## Judgment calls

1. **`runDocument` is not changed**, so the rejections need a second feed of
   the outside inputs (`replay` in `fuzz.ts`, the same input construction as
   `runDocument`). The cost is one more replay without views per document.
2. **The report is located from `runDocument`'s first example per field** and
   the values recomputed from the second replay, instead of a second
   comparison engine. A header difference keeps the runner's text.
3. **The registry** is a function `game => (() => GameSpec) | undefined`;
   `games/fuzz.test.ts` consults a list of registries, hand ports last, so a
   generated game's registry can go in front. The mutation child process has
   its own copy of the six-game registry (a script cannot import a test).
4. **`BOARDGAME_FUZZ_DIR`** replays every `<game>/seed-*.json.gz` under it whose
   game is registered; other games' documents are skipped by design (the
   driver also writes checkers), not reported.
5. **Mutations run in child processes** with a `node:module` `registerHooks`
   load hook that substitutes the mutated text for one file; no file is
   written. The historical mutation "Timer's move bound to the committed
   state, not the one being built" names the correct binding (Go and the
   engine bind `MoveByName` to the committed state); the mutation binds it to
   the state being built.
6. **Loop guard, not a clock.** Body mutants get, in every loop of the target
   function, a counter that throws after 10^8 iterations per process, so a
   mutant that makes a loop run forever is killed by a count. A runaway loop
   elsewhere (outside the target function) is not guarded; the engine's own
   fix-up cap bounds the fix-up loop.
7. **Statement deletion skips declarations** (`const`, `let`, nested
   functions): deleting one only makes the next use throw, which the corpus
   kills trivially. The count skipped is reported.
8. **Integer literal shift is +1 only**, as the proposal says "by one".
9. **The `typescript` module is passed in**, not imported by `mutate.ts`, and
   its types come by `import('typescript')` type queries: a static import in
   a non-test source file counts as a client dependency in
   `src/vite-optimize-deps.test.ts`, which failed on my first full run until
   this was changed.
10. **Unknown verdicts**: no game's plan ever records `unknown` in these
    documents, so every predicate lists `unknown` as missing; the table says
    so rather than dropping the column.

## Where the proposal was wrong about the code

- "That this build works under `./scripts/go-local` was not tried": it works,
  but only with the main package instrumented too (above).
- The proposal's `bodiesExecuted` example names Go functions
  (`checkers.(*moveMoveToken).Apply`); with hand ports the names are spec
  paths (`moves[Place Token].apply`). A generated game's loader can wrap its
  bodies with the same `instrumentBodies` hook.
- "Read from the ledgers the documents already carry": the ledgers carry no
  `unknown` verdict for any of the six games.

## Unverified

- The transient acceptance run of the proposal (`--seeds 0..999`, default
  bounds) was not run; a smaller one was (72 documents, above).
- The `timer` and `seat` kinds of rejection replay: the corpus has only
  `propose` rejections, so those two paths ran in no test.
- `--blocks` is exercised by hand (the tables below), not by a Go test (it
  builds a binary and regenerates the corpus, about a minute).

## Checks run before reporting

- `./scripts/go-local build ./...`: passes.
- `./scripts/go-local test -count=1 ./boardgame-util/ ./boardgame-util/lib/fuzz/... ./boardgame-util/lib/conformance/...`:
  all ok (`lib/fuzz` 43.5 s, including the existing `TestCorpusIsCurrent`
  and the new `TestCoverageMatchesIndex`, `TestCoverageMeasures`,
  `TestIsZero`, `TestBlockCoverage`; `boardgame-util` including
  `TestFuzzCoverageReportsTheCorpus`, `TestFuzzCoverageJSON`).
- `conformance export --check`: "The conformance corpus is current". The
  fuzz corpus check is part of `TestCorpusIsCurrent` (passed) and of every
  `--blocks` run ("The fuzz corpus is current (48 documents)").
- `npm run type-check`: passes. `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json`: passes.
- `npm run test:unit`: 1,278 tests, 1,272 pass, 0 fail, 6 skipped (base
  1,223 with 6 skipped; +55: 51 in `games/fuzz.test.ts`, 4 in
  `testing/mutate.test.ts`). 26 s.
- `npm run test:layers`: 11 of 11.

## Edits outside my files

- `boardgame-util/cmd_fuzz_coverage_test.go` (new file, beside the driver
  stream's `cmd_fuzz_test.go`, which I did not edit).
- `server/static/src/engine/testing/mutate.test.ts` (new test of my harness).
