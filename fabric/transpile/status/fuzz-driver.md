# Workstream 4: the fuzz driver

`fabric/TRANSPILE-PROPOSAL.md`, "The differential fuzzer" (Driver,
Enumerating legal moves in Go, Bounds, Format), plus the Go-only checkers
experiment. Base `d2cb55e4`.

## What was built

| Piece | Where |
|---|---|
| Driver, legal-move enumeration, policies, corpus writer and checker | `boardgame-util/lib/fuzz/` (`config.go`, `driver.go`, `enumerate.go`, `policy.go`, `document.go`, `corpus.go`, `games.go`) |
| Tests | `boardgame-util/lib/fuzz/driver_test.go`, `corpus_test.go`; `boardgame-util/cmd_fuzz_test.go` |
| Subcommands `fuzz run` and `fuzz corpus [--check]` | `boardgame-util/cmd_fuzz.go`, two lines in `cmd_base.go` |
| Checked-in corpus: 48 documents and `index.json` | `conformance/fuzz/` |
| Corpus README (derivation, enumeration, seeds, bounds, sizes) | `conformance/fuzz/README.md` |

`lib/conformance` is not changed: `Document.Rejections` and `SchemaVersion`
are exported fields, so `lib/fuzz` calls `conformance.FromHistory`
unchanged and then sets `Rejections` and, only when there are any,
`SchemaVersion = conformance.SchemaVersionWithRejections`.

## Exported interfaces (`boardgame-util/lib/fuzz`)

- `Config{Game, Seed int64, Policy, MaxProposals, MaxVersions, RefusalRate float64}`.
  Zero values take the defaults `DefaultMaxProposals` (300),
  `DefaultMaxVersions` (600), `DefaultRefusalRate` (1/8); a negative
  `RefusalRate` means none. `MaxCandidatesPerMove` = 4,096.
- `Derive(manager, seed, hasSeatMove) (*Derived, error)`; `Derived{ID, SecretSalt, NumPlayers, Variant, Seated}`.
  Each is SHA-256 of `"boardgame fuzz <use>:<game>:<seed>"` with uses
  `id`, `salt`, `generator`, `players`, `variant <key>`, `seated`
  (scenario.Run's scheme with the fuzzer's own prefixes).
- `Run(cfg)`, `RunWithDelegate(newDelegate, cfg) (*Result, error)`;
  `Result{Config, Derived, History, Refusals []Refusal, End, Proposals, Moves map[string]*MoveStats, MoveOrder, Uncovered, LegalPanics}`;
  `Refusal{conformance.Rejection; Reason}`; `MoveStats{Applied, ProposedLegally, Refused}`.
  Reasons `ReasonIllegalInput`, `ReasonWrongActor`, `ReasonStaleVersion`,
  `ReasonApplyError`, `ReasonTimerError`. Ends `EndFinished`, `EndNoAction`,
  `EndMaxProposals`, `EndMaxVersions`.
- `Document(cfg) (*conformance.Document, *Result, error)`; `DocumentFor(newDelegate, result)`; `Name(game, seed)` = `"<game>/seed-NNNN"`.
- Policies: `PolicyRandomLegal`, `PolicyBotPrefix` + a name from `Bots()`
  (one: `tictactoe-random-empty`, `examples/tictactoe/bot.RandomEmpty`),
  `PolicyAgent`.
- `Games()`, `Delegate(game)`: the six ported games plus checkers.
- Corpus: `CorpusDir`, `CorpusIndexFile`, `CorpusSeeds`, `CorpusGames []CorpusGame{Game, MaxProposals, MaxVersions, Seating}`,
  `CorpusConfigs()`, `CorpusFile(cfg)`, `RunAll(configs, sink)`,
  `Summarize(outputs) *Index`, `(*Index).Text()`, `GenerateCorpus`,
  `WriteCorpus(root)`, `CheckCorpus(root, configs, all)`,
  `WriteRuns(configs, outDir, withResults)`. Index types `Index`,
  `IndexDocument`, `GameSummary`, `MoveSummary`.

CLI: `boardgame-util fuzz run --game G (--seed N | --seeds A..B) [--policy P] [--max-proposals N] [--max-versions N] [--refusal-rate P] [--results] --out DIR`
(refuses `--out` at or under the repository's `conformance/`, symlinks
resolved) and `boardgame-util fuzz corpus [--check]`, which print the
per-game summary (per move type: applied, proposed legally, refused;
rejections by kind; refusals by reason; never-applied moves).

## Test results (all run)

- `./scripts/go-local build ./...`: passes.
- `./scripts/go-local test ./boardgame-util/lib/fuzz/...`: ok, 4 tests
  (41.5 s): `TestCorpusIsCurrent` (all 48 documents and the index,
  26.7 s, plus the coverage assertions below), `TestEveryGameRuns` (7
  subtests), `TestRunIsDeterministic` (memory, tictactoe, werewolf: two runs,
  identical bytes), `TestTruncatedRunIsPrefix` (memory, pig, blackjack:
  `MaxVersions/3` gives the same versions byte for byte and exactly the
  rejections recorded before its last version).
- `./scripts/go-local test -count=1 ./boardgame-util/`: ok, including the 3
  new tests (refusal of `conformance/` paths, a `fuzz run` to a temp
  directory, the command failing inside the corpus).
- `./scripts/go-local test ./boardgame-util/lib/conformance/...`: ok, both
  packages, unchanged.
- `./scripts/go-local run ./boardgame-util conformance export --check`:
  "The conformance corpus is current".
- `./scripts/go-local run ./boardgame-util fuzz corpus --check`: "The fuzz
  corpus is current (48 documents)".
- Node: skipped. No TypeScript changed (`conformance/schema.ts` already had
  `rejections` from the freeze). The TypeScript conformance test reads
  `conformance/index.json` only, so `conformance/fuzz/` is not picked up by
  it.

## Acceptance, as measured (`conformance/fuzz/index.json`)

| Game | Seeds | Finished | Rejections (all `propose`) by reason | Never applied |
|---|---|---|---|---|
| blackjack | 0-7 | 2 | illegal-input 3, stale-version 5, wrong-actor 1 | `Shuffle Discard Into Draw`, `Force Finish Turn` |
| debuganimations | 0, 2, 3, 5, 6, 7, 10, 14 | 0 | illegal-input 9, stale-version 13 | `Seat Player` |
| memory | 0-7 | 0 | illegal-input 8, stale-version 7, wrong-actor 3 | none |
| pig | 0-7 | 0 | stale-version 15, wrong-actor 10 | none |
| tictactoe | 0-7 | 8 | illegal-input 3, stale-version 1, wrong-actor 1 | none |
| werewolf | 3, 7, 8, 16, 19, 20, 21, 26 | 8 | illegal-input 12, stale-version 2, wrong-actor 4 | none |

The three never-applied moves, with reasons (also asserted by name in
`TestCorpusIsCurrent`): blackjack's reshuffle needs the 52-card draw stack
exhausted (covered by `conformance/blackjack/scenario-reshuffle`);
`Force Finish Turn` is admin-only (`ForceFinishTurn can only be proposed by
AdminPlayerIndex`) and the driver never proposes as the admin; debuganimations
seeds are unseated only. No `apply-error`, `timer-error` or seat rejection
occurred in any game. No field was default-only (no game has a supplied
string field). The corpus is 2.7 MB gzipped (the rest of `conformance/` is
1.75 MB); sizes per game and the bounds are in the README.

## Judgment calls

1. **The admin is not an actor.** The proposal enumerates for every seated
   player "and the admin". `conformance.FromHistory` replays a history move
   with a negative proposer as a seat (if it is a seat move) or a timer
   firing; an admin proposal would be replayed as `ForceNextTimer` and fail.
   So admin proposals cannot be in a history the exporter accepts unchanged.
   Consequence: `Force Finish Turn` is never applied.
2. **Legal means accepted.** A candidate counts as legal for a player only
   if `Legal` returns nil *and* the player is a valid proposer
   (`PlayerIndex.Valid`: within bounds and `PlayerMayBeActive`). The engine
   checks the latter before `Legal`; without it every seated-but-inactive
   debuganimations player's moves came back from `Apply` as "The proposer
   was not valid." Such candidates are refusal material instead.
3. **A fresh move per actor.** `moves.AnyPlayer.Legal` writes the proposer
   into `TargetPlayerIndex` on its first call, so one bound move cannot be
   checked for several players (werewolf's votes all looked illegal before
   this). Each (candidate, player) binds its own move.
4. **Integer-coded enums take every enum key** (checkers' range-enum
   `TokenIndexToMove`/`SpaceIndex`), not "0 to one past the longest stack":
   the binder refuses a value outside the enum anyway. Checkers' product is
   exactly 64 x 64 = 4,096, so no sampling.
5. **Projected domains** are the union over seated players of every
   candidate value of the projection (available or not), in first-seen
   order, because one candidate list is shared by all players.
6. **Seating**: one seat action, for the first open seat, offered only when
   the game's seat move is `Legal` for it. A seated run seats first, before
   any bound applies: the exporter infers seating from a seat move in the
   history, so a seated history must have one. `Document` checks
   `doc.Game.Seating` against the derived value.
7. **Refusal draw**: one `Float64` per step, always, so the generator's
   sequence does not depend on which refusals are available; if none is,
   the step proposes a legal action. Kinds are chosen uniformly among
   those available. If the engine accepts an input the driver expected it
   to refuse, the run fails (never happened).
8. **Runs are sequential.** Parallel runs crashed in `NewGameManager`
   (a torn string read): example delegates memoize their name in a package
   variable (`memoizedDelegateName` in `examples/blackjack/main.go`), so the
   engine plus games are not safe for concurrent managers.
9. **Corpus seed selection** by derived setup for two games (README table):
   werewolf unseated is a dead end ("your seat is not yet filled"), and
   debuganimations seated is a dead end (it configures `Seat Player` but no
   `Activate Inactive Player`, so seated players are inactive forever).
10. **`source.kind` is `"scenario"`**, with `description` starting `fuzz:`,
    because `conformance/schema.ts` types it `'scenario' | 'golden'` and this
    stream changes no TypeScript. A later stream may add `'fuzz'`.
11. **Bounds** check `Version >= MaxVersions` before each step, so a run ends
    at the first settled version at or past the bound (a fix-up chain is
    never cut). A refusal's `name` is `""` for a timer; its `expectedVersion`
    is the version named (current, or current minus one for stale).
12. **`--refusal-rate 0`** on the CLI means none (the struct's zero means the
    default).

## Where the proposal was wrong about the code

- "For each actor (every seated player, and the admin)": see call 1.
- "Keep the candidates for which `MoveWithInput` succeeds and `Legal`
  returns nil": `applyMove` also rejects an invalid (inactive) proposer, a
  timer guard and an input outside a declared choice domain before or
  beside `Legal`; and `Legal` can mutate the move (calls 2, 3).
- Risk 1 (checkers): the `Apply` bug is real in the source but unreachable;
  checkers cannot make a single move (below).
- The fix-up cap case was never reached by any run.

## Checkers in Go alone

Command (transient, outside the repository):
`boardgame-util fuzz run --game checkers --seeds 0..199 --results --out <scratch>/checkers200`
(default bounds: 300 proposals, 600 versions, refusal rate 1/8; 48 s).

- **Can a full game be played? No. Not one move.** All 200 runs end
  `no-action` at the first playing position: 112 unseated runs at version 25
  (24 `Place Token` fix-ups and `Start Phase Playing`), 88 seated runs at
  version 29 (plus two seats, each followed by `Activate Inactive Player`).
  `Move Token` was applied 0 times; `Crown Token` and `Finish Turn` 0 times;
  no run finished.
- **Why** (from enumerating all 4,096 `Move Token` inputs for both players at
  that position, seeds 0 and 1, seated and unseated, identical each time):
  player 0 has 0 legal inputs: "That space does not have a component in it"
  2,520, "that token isn't your token to move" 756, "you can only move to
  spaces that are black" 384, "spaceIndex does not represent a legal space
  for that token to move to" 372, "i and j were the same" 64. Player 1 gets
  "it's not your turn" for all 4,096 and never gets a turn. The position is
  red tokens in columns 0, 2, 4, 6 of rows 0-2 and black tokens in columns
  0, 2, 4, 6 of rows 5-7: `spaceIsBlack` is `index % 2 == 0`, which on an
  eight-wide board is every even column, as risk 1 said. Red uses
  `graphUpward` and black `graphDownward`, i.e. each colour moves toward
  its own back edge, where every straight neighbour is occupied and every
  diagonal neighbour is an odd column (that "up" is toward row 0 is
  inferred: otherwise red's row-2 tokens would have a legal straight move).
  Both players' `Color` is `Red` in the stored state, so the black tokens
  belong to no player (where colours should be assigned was not
  investigated).
- **`Apply` errors after `Legal` passed: 0, out of 0 legal proposals.** The
  middle-space bug in `(*moveMoveToken).Apply` (finish minus start passed to
  `RangeToValue`) is never reached, because no `Move Token` is ever legal.
- **Captures: none**, so whether a capture removes the wrong token cannot be
  observed. `CapturedTokens` stays empty in every run.
- **Refusals recorded: 38, all `illegal-input`, all `propose`:** "it's not
  your turn" 15 (e.g. seed 8, version 29, player 1, `{TokenIndexToMove: 61,
  SpaceIndex: 60}`); "that token isn't your token to move" 9 (seed 6, version
  25, player 0, `{48, 54}`); "That space does not have a component in it" 8
  (seed 9, version 25, player 0, `{11, 62}`); "spaceIndex does not represent
  a legal space for that token to move to" 3 (seed 64, version 27, player 0,
  `{16, 12}`); "you can only move to spaces that are black" 3 (seed 105,
  version 25, player 0, `{6, 27}`).
- So a faithful port of checkers is a port of a game that deadlocks at its
  first turn. To exercise `moveMoveToken.Apply` at all, directed scenarios
  would need a constructed position; none was built here, and no Go was
  changed.

## Unverified

- The TypeScript replay of `rejections` (workstream 5) has not run against
  these documents.
- The `apply-error`, `timer-error` and seat-failure paths of the driver were
  never exercised: no game produced one.
- `agent` policy: one transient run (memory seeds 0-3, max 100 versions)
  completed and exported; it is not reproducible by design.
- `bot:tictactoe-random-empty`: one transient run (seeds 0-19) completed.

## Edits outside my files

- `boardgame-util/cmd_base.go`: two lines registering `fuzz`.
