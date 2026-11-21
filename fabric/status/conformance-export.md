# Conformance export: status

Branch `claude/fabric-c-conformance-export-e76c46`. Additive Go tooling that
exports, for a game run on the real engine, every state version as every
viewer sees it. This is the "new exporter" row of the oracle table in
`fabric/PROPOSAL.md`.

## Commits

| Commit | What |
|---|---|
| `5d07767d` Add a conformance exporter for the game engine | `boardgame-util/lib/conformance` (replay, capture, format, normalization), the `conformance export` subcommand, `server/api/conformance_export.go` and its test. The subcommand imports the corpus package added by the next commit, so this commit does not build alone |
| `f1215547` Check in the conformance corpus and its staleness test | `boardgame-util/lib/conformance/corpus` (corpus list, five new scenarios, `TestCorpusIsCurrent`, `TestExportIsDeterministic`), `conformance/` (21 documents, `index.json`, `schema.ts`, `README.md`) |
| (the commit adding this file) | This status note |

Nothing was pushed; no PR; `master` is untouched.

Test results: `./scripts/go-local build ./...` passes. `./scripts/go-local
test ./...` has 43 packages passing and 4 failing, the same 4 that fail on
the base commit (finding 12 below). The two new packages pass.

## What exists

| Piece | Where |
|---|---|
| Exporter package (replay, capture, format, normalization list) | `boardgame-util/lib/conformance/` |
| Corpus definition, scripted scenarios, staleness test | `boardgame-util/lib/conformance/corpus/` |
| Subcommand `boardgame-util conformance export [--check] [--plain --out DIR] [--root DIR]` | `boardgame-util/cmd_conformance.go` |
| Checked-in corpus, `index.json`, `README.md` | `conformance/` |
| Document format as TypeScript types (no imports) | `conformance/schema.ts` |

Regenerate: `./scripts/go-local run ./boardgame-util conformance export`.
Verify without writing: add `--check`.

### Existing files touched

- `boardgame-util/cmd_base.go`: two lines, to register the subcommand.
- `server/api/conformance_export.go` is a **new file in an existing package**,
  which goes slightly beyond "new packages only". It adds one exported
  function, `MoveFormsForViewerAtState`, that no server path calls. The reason:
  the legality wire shape (`moveForm`, the preconditions ledger, the private
  move filter) is unexported in `server/api`. The alternative was a second
  copy of about 150 lines of "what the server reports" inside the exporter,
  which would be free to drift. `server/api/review_snapshot.go` is the
  precedent. `server/api/conformance_export_test.go` pins it to
  `generateFormsWithLegality` for the head state.

No existing behavior changed.

## How a document is produced

Both inputs reduce to a history. A `scenario.Spec` is run by `scenario.Run`,
and its `History` is replayed the same way a golden file is. Replay recreates
the game from the history's id, salt, player count and variant, then repeats
only the outside input: a player move is proposed, a seat is injected through
the seating rendezvous, a timer is forced. Fix-ups are never proposed; the
engine has to produce them. Every replayed state and move is compared with
the history, and the export fails on any difference.

After each driving step the game is settled, and every version the step
passed through is captured: the stored state, the move record, and for each
viewer `Game.JSONForPlayer`, `Game.MoveJSONForPlayer`, the move forms with
legality, and (settled versions, seated players only) projected move choices.

Three checks back the claim that this is what a client receives:

- `TestExportAgreesWithScenarioReplay`: `scenario.Run` captures each viewer's
  API snapshot live, at the moment a version is the head. For the memory and
  werewolf scenarios the export's game object, legality and projected choices
  equal those snapshots at every frame and viewer.
- Inside the exporter, the `ActiveTimers` rebuilt from the stored state is
  compared with the scheduler's own payload at every settled version.
- `TestMoveFormsForViewerAtStateMatchesServerAtHead` (in `server/api`).

## Corpus inventory

21 documents, 1.15 MB gzipped (about 55 MB of JSON). Viewers are always admin,
observer and every player, so "viewers" is players + 2. "Versions" counts
version 0.

| File (under `conformance/`) | Game | Versions | Viewers | Exercises |
|---|---|---|---|---|
| `blackjack/golden-714104226cb9a317` | blackjack | 30 | 4 | Two players, one round, the next deal |
| `blackjack/golden-82ef86840a83b9ed` | blackjack | 43 | 6 | Two of four seats filled through seating; empty seats inactivated |
| `blackjack/golden-c93d3566c752df43` | blackjack | 41 | 6 | Four players, one round, the next deal |
| `blackjack/golden-d2b8e7e47a412543` | blackjack | 21 | 6 | Four players, part of a round |
| `blackjack/golden-dac0bc2b57492b3a` | blackjack | 66 | 9 | Seven players, one round, the next deal |
| `blackjack/scenario-bust-and-finish` (new) | blackjack | 47 | 5 | A bust; hidden card dealt and revealed, with the `len`-policy hand as seen by self, other, observer, admin; two rounds; game over, winner player 2 |
| `debuganimations/scenario-shuffles` (new) | debuganimations | 16 | 4 | Mid-game secret shuffles (`SecretMoveCount` rises, ids re-scramble), a public shuffle, a hidden-name move with no alias (other viewers get a null move), `order` and `nonempty` stacks, a move argument tagged `all:visible` |
| `memory/golden-6cb9690b3b55fe3d` | memory | 90 | 4 | Full game to game over (winner player 0): 7 timer-fired hides, 3 player-proposed hides, 10 captures |
| `memory/golden-c369b3a9d5c0eda0` | memory | 10 | 4 | A player-proposed hide, a capture |
| `memory/golden-fec3529ffca41c00` | memory | 6 | 4 | One non-match, player-proposed hide |
| `memory/golden-ff0c8d5d3547687c` | memory | 30 | 4 | Large layout, timer-fired and player-proposed hides |
| `memory/scenario-mismatch` (existing scenario) | memory | 6 | 4 | Non-match, timer started by a fix-up, timer fires |
| `memory/scenario-full-game` (new) | memory | 30 | 4 | Small game to game over with a winner: non-match hidden by the timer, non-match hidden by the player (timer status `canceled`), all five pairs captured |
| `pig/golden-67672ddc7e8a71ad` | pig | 228 | 4 | Full game to game over (winner player 0); 83 dice rolls from `State.Rand` |
| `pig/golden-81137f3f9f5d40ba` | pig | 28 | 4 | Partial game |
| `pig/scenario-short` (new) | pig | 27 | 4 | A few turns, including a banked turn |
| `tictactoe/golden-0e7f8ddf5c1c7c65` | tictactoe | 14 | 4 | Win by player 0; hidden-name move with alias |
| `tictactoe/golden-1b7c0a299f5b79c5` | tictactoe | 18 | 4 | Draw: finished with no winners |
| `tictactoe/golden-568a53942f7d083d` | tictactoe | 12 | 4 | Win by player 1 |
| `tictactoe/scenario-seated-win` (new) | tictactoe | 14 | 4 | Both players seated, then a win. `Place Token` is visible by name only to its proposer: others get `AnimationKey: "Hidden Action"` and no arguments |
| `werewolf/scenario-timed-vote` (existing scenario) | werewolf | 13 | 6 | Four seats, hidden roles, a vote whose target only the voter and same-role players see, deadline timer, projected move choices for the actor |

The task asked for goldens of memory and blackjack. Pig and tictactoe goldens
are included too because they are small and add a draw, a second winner and a
long run of dice rolls.

How the required cases are met:

- **Blackjack full round with a bust and the hidden-card reveal**:
  `blackjack/scenario-bust-and-finish`. Player 0 hits at version 12 and busts
  (`Eliminated: true`, turn ends by fix-up). `HiddenHand` under `len` shows
  the real index to its owner and the admin and `-2` to everyone else;
  `Visibility` and the computed `HandValue` differ by viewer accordingly.
- **Memory capture, and non-match with the timer firing**: `memory/scenario-full-game`
  and four goldens.
- **Game over with winners**: blackjack scenario, memory scenario and golden,
  pig golden, tictactoe (two winners and a draw).
- **Secret move or shuffle mid-game**: `debuganimations/scenario-shuffles`.
  No example calls a secret component move; shuffles are the only source of
  `SecretMoveCount` changes in the examples. Blackjack's own reshuffle
  (`Shuffle Discard Into Draw`) never fires in the corpus, because exhausting
  a 52-card draw stack takes several long rounds.
- **A move a viewer may not see the name of**: tictactoe (`Place Token`, with
  alias) and debuganimations (`Shuffle Hidden`, no alias). Hidden arguments:
  werewolf `Cast Vote`.

## Normalization

The machine-readable list is the `normalization` section of every document
and `conformance.Normalization()`. In the same order:

| Field | Action | Why |
|---|---|---|
| Move `Timestamp` | omitted | Wall clock (`game.go:1094`) |
| Game `Created`, `Modified` | omitted | Wall clock |
| `state.Timers[].Deadline` | fixed | Manual clock fixed at `engine.clock` (2020-01-01T00:00:00Z) that never advances; deadline = clock + duration. A golden file's own wall-clock deadlines are not compared |
| Timer ids | kept | Drawn from `State.Rand`, so deterministic |
| `state.Timers[].Move.Timestamp` | kept | Always Go's zero time |
| View `game.Diagram` | omitted | Go-rendered debug string |
| View `game.ActiveTimers` | derived | Comes from the manager's scheduler, not a state version. Rebuilt from the version's stored timer records; checked against the scheduler at every settled version |
| View `game.Version` | rewritten | Wire value is the head version at request time; set to the version viewed |
| View `game.Finished`, `Winners` | rewritten | Wire values are the game's current ones; set to the values as of the version |
| `legality[].defaults` | rewritten | Wire `Fields` array is in Go map order and repeats static types; reduced to `{field: defaultValue}` |
| Form `HelpText`, `IsGatheringStart` | omitted | Static; in `moves[]` |
| Legality where `settled` is false | derived | The server reports legality only at the settled head; elsewhere it is evaluated against that version's state with moves bound to that state |
| `projectedMoveChoices` | kept | Only where the server sends it |
| All embedded JSON | rewritten | Keys sorted, whitespace removed |
| `game.id`, `game.secretSalt` | kept | From the golden, or SHA-256 of the scenario seed |
| `/info` session fields (`Players`, `CompanionInfo`, ...) | omitted | Server session data, not engine output |

Not a normalization but worth knowing: files are gzipped with no timestamp or
name in the header, and the staleness test compares decompressed bytes.

## Engine findings

Things that were surprising or inconsistent while exporting. None was changed.

1. **`JSONForPlayer` mixes a historical state with head-of-game values**
   (`game.go:280-292`). `Finished`, `Winners`, `Version` and `ActiveTimers`
   come from the game object, not from the `state` argument. The server calls
   it with historical states for animation bundles
   (`server/api/main.go:1284`), so a client replaying version 3 of a finished
   game is told `Finished: true` and given the final `Version`. The exporter
   rewrites these four fields to the viewed version; a TypeScript engine
   should define them per version and not copy this.
2. **First-load `/info` evaluates legality on state 0 with moves bound to the
   head** (`server/api/main.go:1639-1651` picks `game.State(0)`; line 1761
   passes it to `generateFormsWithLegality`, which gets its moves from
   `Game.Moves()`, bound to `CurrentState` at `game.go:858`). Defaults come
   from one version and the verdict from another. The export binds moves to
   the state being evaluated, so for version 0 of a game with a setup chain it
   will not match that particular live response.
3. **Form `Fields` order is not deterministic on the wire**
   (`server/api/main.go:2425` ranges over a map). Nothing can compare it
   byte-for-byte.
4. **Computed properties are computed from the sanitized state**
   (`state.go:1273`, lazily, per state copy), so they differ by viewer.
   In blackjack, another player's `HandValue` counts their hidden card as 0
   until it is revealed. A port must compute after sanitizing, not before.
5. **The admin view carries `SecretMoveCount`** (`state.go:1224`: emitted
   whenever the state is not sanitized, and admin is not). The salt is never
   sent, so ids still cannot be recomputed from it alone.
6. **`len`-policy stacks keep their component ids** (sorted), replacing only
   the indexes with `-2`. An observer of a one-card `len` hand is sent that
   card's id.
7. **Whether seating will ever happen is an input to the game, held outside
   state** (`moves/seat_player.go:15`, read from
   `Storage().FetchInjectedDataForGame`). Blackjack with seating off starts
   at once with all seats active; with seating on it waits and inactivates
   empty seats. It is recorded as `game.seating`; nothing in the state or the
   move history says so directly.
8. **Version 0 is not a settled version for games with a setup chain.**
   Blackjack and pig run fix-ups during creation, so their first settled
   version is between 2 and 19, not 0.
9. **A timer's stored intent move has `Version` set to the version that
   started the timer and `Initiator` 0** (`timer.go:142` serializes an
   unapplied move). It is not the version the move will get when it fires.
10. **`MoveJSONForPlayer` inflates the stored move against the current state**
    (`move_sanitization.go:123`, `record.inflate(g)` binds defaults from
    `CurrentState`) while taking sanitization groups from the pre-move state.
    The stored blob overrides the defaults, so the corpus shows no effect,
    but a move property absent from an old blob would pick up a head default.
11. **Blackjack scores a ten plus an ace as 11, not 21**
    (`examples/blackjack/state.go:96`: `currentValue >= targetScore - 10`
    stops the ace from being promoted exactly when promotion would make 21).
    Visible in `blackjack/scenario-bust-and-finish` version 8, player 1. A
    port has to reproduce it to match; it looks like a bug in the example.
12. **Baseline**: `boardgame-util/lib/build/api`, `lib/build/static`,
    `lib/gamepkg` and `lib/lint` fail before and after this change, because
    the pinned
    `golang.org/x/tools` cannot read Go 1.27 export data ("export data
    version 5 is greater than maximum supported version 4").

## What a TypeScript engine still cannot check from this corpus

- **Rejections.** Every recorded proposal succeeded. Stale-version errors,
  wrong proposer, illegal arguments, a proposal after game over, and the
  fix-up cap (`ErrTooManyFixUps`) are not exercised.
- **Legality beyond default-bound moves.** Each form is evaluated with the
  engine's default arguments, as the move tray is. Other arguments are
  covered only where a game declares choice projections (memory, werewolf),
  and then only for the actor at settled versions.
- **Input binding.** Documents record the arguments a move ended up with,
  not the raw input or the binder's rejections of unknown or context-owned
  fields.
- **Legality on unsettled versions as the server sends it.** Those rows are
  state-pure derivations; the server never reports them.
- **Timers against a moving clock.** The clock is fixed, so `TimeLeft` is
  always the full duration, no timer is ever overdue, there is never more
  than one active timer, and nothing checks ordering between timers.
- **Features no exported example uses**: dynamic component values
  (`Components` is always `{}`), a secret component move, boards, agents,
  blackjack's reshuffle, seating at "next available seat" (every seat step
  names its index), and any variant or player count not in the inventory.
  Checkers, the largest declarative-legality example, is not exported.
- **Storage and recovery**: reload, proposal-frontier reconciliation,
  concurrent proposals.
- **Wire details normalized away**: `Diagram`, form `Fields` order, and the
  head-of-game `Version`/`Finished`/`Winners` on historical bundles.
