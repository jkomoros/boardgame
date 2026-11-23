# Transpile pipeline: the per-game report

Written by `boardgame-util port --all --dry-run` from `fabric/transpile/reports/<game>.json`; do not edit by hand. `port --all --dry-run --check` fails when this file or a report is stale. Fields are those of fabric/TRANSPILE-PROPOSAL.md, "The per-game report". A field that cannot be produced yet is `null` in the JSON, with its reason under `nulls`, and "null" here.

## Headline: data, mechanical, agent, engine gap

| Game | Units | Data | Mechanical | Agent | Engine gap |
|---|---|---|---|---|---|
| memory | 20 | 3 (15.0%) | 16 (80.0%) | 1 (5.0%) | 0 (0.0%) |
| blackjack | 39 | 16 (41.0%) | 23 (59.0%) | 0 (0.0%) | 0 (0.0%) |
| pig | 13 | 3 (23.1%) | 10 (76.9%) | 0 (0.0%) | 0 (0.0%) |
| tictactoe | 14 | 3 (21.4%) | 11 (78.6%) | 0 (0.0%) | 0 (0.0%) |
| werewolf | 20 | 4 (20.0%) | 16 (80.0%) | 0 (0.0%) | 0 (0.0%) |
| debuganimations | 25 | 1 (4.0%) | 24 (96.0%) | 0 (0.0%) | 0 (0.0%) |
| checkers | 25 | 4 (16.0%) | 20 (80.0%) | 1 (4.0%) | 0 (0.0%) |
| **Total** | **156** | **34 (21.8%)** | **120 (76.9%)** | **2 (1.3%)** | **0 (0.0%)** |

How each bucket is counted. The units of a game are its move configs that
are data only plus its functions that reach engine output (the manifest of
boardgame-util/lib/gamedef, the declarativeness lint's "functions reaching
engine output"). A move config with game code is not a unit of its own: its
configuration is exported as data like any other, and its code is counted as
the functions it holds, so counting the config too would count that code
twice. Every unit is in exactly one bucket:

- **Data**: a data-only move config (its type is a stock move, or a game type
  that embeds one and declares no method) whose stock type the engine has
  (MoveConfigStatic.StockPorted).
- **Mechanical**: a function the translator translates now (a dry run of
  lib/transpile, verdict subset) that has no checked-in agent-written block.
- **Agent**: a function with a checked-in agent-written block whose Go source
  hash is current, or a function the translator refuses for a reason other
  than an engine gap, which is bound for an agent (functions.boundForAgent).
- **Engine gap**: a function refused only for engine gaps (gamedef bucket
  engineGap: every refusal is engine-gap, or callee-refused naming such a
  function), and a data-only move config whose stock type the engine lacks.

The headline is a dry run: it does not need a generated game to be checked
in. functions.mechanical below counts only what is checked in.

## Per game

| Game | movesData | movesWithCode | functions | data | mechanical | translatable | agent | agentNowMechanical | engineGap | boundForAgent | notPorted | stale | definition | TS replay |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| memory | 3 | 4 | 17 | 14 | 16 | 16 | 1 | 0 | 0 | 0 | 11 | 0 | current | 14 of 14 pass (checked-in) |
| blackjack | 16 | 7 | 23 | 13 | 23 | 23 | 0 | 0 | 0 | 0 | 1 | 0 | current | 15 of 15 pass (checked-in) |
| pig | 3 | 3 | 10 | 12 | 10 | 10 | 0 | 0 | 0 | 0 | 1 | 0 | current | 11 of 11 pass (checked-in) |
| tictactoe | 3 | 1 | 11 | 14 | 11 | 11 | 0 | 0 | 0 | 0 | 5 | 0 | current | 12 of 12 pass (checked-in) |
| werewolf | 4 | 7 | 16 | 11 | 16 | 16 | 0 | 0 | 0 | 0 | 1 | 0 | current | 9 of 9 pass (checked-in) |
| debuganimations | 1 | 12 | 24 | 23 | 24 | 24 | 0 | 0 | 0 | 0 | 1 | 0 | current | 9 of 9 pass (checked-in) |
| checkers | 4 | 3 | 21 | 14 | 20 | 20 | 1 | 0 | 0 | 0 | 0 | 0 | current | 8 of 8 pass (checked-in) |

The one number to watch per game is agent − agentNowMechanical; it should fall.

## Coverage

| Game | Fuzz documents | Versions | Rejections | Moves never applied | Predicates missing a verdict | Hidden props never sanitized while non-zero | Go blocks reached | Surviving mutants |
|---|---|---|---|---|---|---|---|---|
| memory | 8 | 269 | 18 | none | 7 of 7 | none | 79 of 91 | null |
| blackjack | 8 | 527 | 9 | Shuffle Discard Into Draw; Force Finish Turn | 10 of 10 | none | 93 of 106 | null |
| pig | 8 | 413 | 25 | none | 5 of 5 | none | 36 of 38 | null |
| tictactoe | 8 | 134 | 5 | none | 2 of 2 | none | 67 of 73 | null |
| werewolf | 8 | 157 | 18 | none | 0 of 0 | Players.FellowWolves | 143 of 151 | null |
| debuganimations | 8 | 168 | 22 | Seat Player | 7 of 7 | none | 135 of 154 | null |
| checkers | 8 | 732 | 33 | none | 6 of 7 | none | 139 of 158 | null |

## Named functions

### memory

- agent-written: `memory.(*moveStartHideCardsTimer).Apply`
- null `coverage.mutants`: no body-mutant run is recorded for this game's checked-in bodies (port check runs the mutants of one function; a run over every function is not part of port)

### blackjack

- null `coverage.mutants`: no body-mutant run is recorded for this game's checked-in bodies (port check runs the mutants of one function; a run over every function is not part of port)

### pig

- null `coverage.mutants`: no body-mutant run is recorded for this game's checked-in bodies (port check runs the mutants of one function; a run over every function is not part of port)

### tictactoe

- null `coverage.mutants`: no body-mutant run is recorded for this game's checked-in bodies (port check runs the mutants of one function; a run over every function is not part of port)

### werewolf

- null `coverage.mutants`: no body-mutant run is recorded for this game's checked-in bodies (port check runs the mutants of one function; a run over every function is not part of port)

### debuganimations

- null `coverage.mutants`: no body-mutant run is recorded for this game's checked-in bodies (port check runs the mutants of one function; a run over every function is not part of port)

### checkers

- agent-written: `checkers.(*gameDelegate).ConfigurePredicateConstructors`
- null `coverage.mutants`: no body-mutant run is recorded for this game's checked-in bodies (port check runs the mutants of one function; a run over every function is not part of port)

