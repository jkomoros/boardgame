# Fuzz corpus

Each `<game>/seed-NNNN.json.gz` here is a conformance document (the format in
`../schema.ts`, the same as the rest of `conformance/`) for one seeded run of
the fuzz driver, `boardgame-util/lib/fuzz`. The driver plays the game on the
Go engine, at each settled version choosing uniformly among the actions the
engine says are legal, and about one step in eight sends an input the engine
must refuse instead. The history is exported with `conformance.FromHistory`,
unchanged, so every version is recorded for every viewer exactly as in the
hand-picked corpus; the refused inputs are the document's `rejections`.

A document with any rejection has `schemaVersion: 2`; one without is
version 1. `source.kind` is `"scenario"` (the schema allows only that or
`"golden"`), and `source.description` starts with `fuzz:` and names the
game, seed, policy and bounds.

`index.json` lists every document (players, variant, seated or not,
versions, viewers, how the run ended, proposals, rejections) and, per game,
for each move type the times it was applied (fix-ups included), proposed
legally by the driver, and refused, plus rejections by kind and the
driver's reason for each refusal.

## What a run derives from its seed

All by SHA-256 of `"boardgame fuzz <use>:<game>:<seed>"`: the game id, the
secret salt, the driver's generator (`math/rand` seeded from the first eight
bytes), the player count within the game's bounds, a value for each variant
key, and whether players join through the seating rendezvous. The timer clock
is the manual clock fixed at `conformance.Clock` and never advances. Nothing
reads a wall clock, sleeps or retries; runs are bounded by counts only.

## Legal actions and refusals

At each settled version, for every seated player (every player in an
unseated run): each non-fix-up move's input fields that a proposer supplies
(`required`, `server-defaulted`) get a domain (both booleans; every player
index; every enum key; for an integer, the values of its declared choice
projection if it has one, every key if it is a range enum, otherwise the
default and 0 through the length of the longest stack; for a string, the
default only). The product is taken, capped at 4,096 candidates per move
(sampled with the generator above the cap). A candidate is legal for a
player when `Game.MoveWithInput` binds it, `Move.Legal` returns nil, and the
player is a valid (active) proposer. Firing the next timer is added when a
timer is active, and seating the first open seat in a seated run. A seated
run always seats a player first, so its history shows it is seated.

A refusal is one of: a candidate no player may propose (`illegal-input`), a
candidate proposed by a player for whom it is illegal while another may
propose it (`wrong-actor`), or a legal candidate at the previous version
(`stale-version`). A legal proposal the engine still refuses
(`apply-error`) and a timer whose move fails (`timer-error`) are recorded
too. Each records Go's error text, and the driver checks that the stored
state and version did not change.

## Seeds and bounds

Eight seeds per game: seeds 0 to 7, except where one setup is a dead end in
Go and the corpus keeps only the first eight seeds whose *derived setup* is
the other one (selected on the setup, never on the outcome):

| Game | Seeds | Max proposals | Max versions | Why |
|---|---|---|---|---|
| blackjack | 0-7 | 40 | 70 | Up to 9 viewers |
| debuganimations | first 8 unseated (0, 2, 3, 5, 6, 7, 10, 14) | 30 | 20 | Seated, its players are never activated, so nothing is legal |
| memory | 0-7 | 30 | 32 | Up to 8 viewers and large layouts |
| pig | 0-7 | 40 | 50 | |
| tictactoe | 0-7 | 60 | 60 | Every run finishes first |
| werewolf | first 8 seated (3, 7, 8, 16, 19, 20, 21, 26) | 60 | 80 | Unseated, every vote fails "your seat is not yet filled" |

The bounds keep the corpus near the size of the hand-picked one: about
2.7 MB gzipped against 1.75 MB for the rest of `conformance/`. By game:
blackjack 0.94 MB (existing 1.1 MB), memory 0.90 MB (0.49 MB), debuganimations
0.36 MB, pig 0.27 MB, werewolf 0.14 MB, tictactoe 0.07 MB; the last four have
only one to three hand-picked documents each, so eight seeds of any useful
length are larger.

Move types the checked-in corpus never applies, by design:

- blackjack `Shuffle Discard Into Draw`: needs the 52-card draw stack
  exhausted, several full rounds in (`conformance/blackjack/scenario-reshuffle`
  covers it).
- blackjack `Force Finish Turn`: legal only for the admin, and the driver
  never proposes as the admin (the exporter replays a non-seat admin move as
  a timer firing).
- debuganimations `Seat Player`: only unseated seeds are kept.

## Regenerating

```sh
./scripts/go-local run ./boardgame-util fuzz corpus          # rewrite
./scripts/go-local run ./boardgame-util fuzz corpus --check  # verify only
```

`./scripts/go-local test ./boardgame-util/lib/fuzz/...` regenerates all 56
documents and the index in memory (about 30 seconds) and fails if any
differs (decompressed bytes are compared), if a game records no refusal, or
if a move type outside the list above is never applied.

## Transient runs

```sh
./scripts/go-local run ./boardgame-util fuzz run --game checkers --seeds 0..999 --out /some/dir
```

writes `DIR/<game>/seed-NNNN.json.gz` and `DIR/index.json` (and with
`--results`, each run's refusal reasons as `seed-NNNN.result.json`). It
refuses to write inside `conformance/`. A run with a smaller
`--max-versions` is a prefix of the longer run, so the minimal document for
a failure at version V is the same seed with `--max-versions V`. Policies
are `random-legal` (default), `bot:tictactoe-random-empty` and `agent` (the
game's `boardgame.Agent`; memory's draws from the global `math/rand`, so an
agent run is not reproducible and is never checked in).
