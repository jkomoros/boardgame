# Status: transpile pipeline

Branch `claude/transpile-pipeline-e23924`, local only. Nothing pushed, no
pull request, `master` and `fabric-factoring` untouched. The plan is
[`../TRANSPILE-PROPOSAL.md`](../TRANSPILE-PROPOSAL.md). A coordinator session
assigned each workstream to an Opus sub-agent in its own worktree, read each
diff, re-ran its tests and merged it. Each stream's own report is under
[`../transpile/status/`](../transpile/status/). A number below is from the
coordinator's own run unless the line says "agent's report".

One stream is not merged: the pair store and rule proof (workstream 11) was
still running when this was written. See "Not done".

## Result

All seven example games are generated from their Go by the pipeline, checked
in beside the six hand ports, and match Go on every recording there is.

| Game | Units | Data | Mechanical | Agent | Engine gap |
|---|---|---|---|---|---|
| memory | 20 | 3 | 16 | 1 | 0 |
| blackjack | 39 | 16 | 23 | 0 | 0 |
| pig | 13 | 3 | 10 | 0 | 0 |
| tictactoe | 14 | 3 | 11 | 0 | 0 |
| werewolf | 20 | 4 | 16 | 0 | 0 |
| debuganimations | 25 | 1 | 24 | 0 | 0 |
| checkers | 24 | 4 | 19 | 1 | 0 |
| **Total** | **155** | **34 (21.9%)** | **119 (76.8%)** | **2 (1.3%)** | **0** |

A unit is a data-only move config or a function that reaches engine output.
The table and how each bucket is counted are in
[`../transpile/REPORT.md`](../transpile/REPORT.md), written by
`boardgame-util port --all --dry-run`; the per-game JSON is in
`../transpile/reports/`.

The engine-gap column is zero because the gaps were closed during this work,
not because there were none: checkers started with 8 functions blocked on
engine features. Fifteen stock move types and seven behaviors are still
unported (`server/static/src/engine/rules/stock-inventory.json`); no example
game uses them.

## What was built

| Part | Where | Acceptance, as run by the coordinator |
|---|---|---|
| Go descriptions | `moves/describe.go`, `legal_plan_export.go`, `moves/resolved_config.go` | `moves` and root tests pass |
| Exporter (run-time and static halves) | `boardgame-util/lib/gamedef`, `gamedef export [--check]`, `conformance/definitions/*.json` | Seven definitions written; `--check` current; `testdata/defs` byte-identical after regeneration |
| Subset checker and lint | `lib/transpile/subset`, `lint --declarative [--json] [--max-refused N]` | Measured table below; goldens per game |
| Loader | `engine/rules/data.ts` (`gameFromData`, the body calling convention) | Each of the six hand ports rebuilt from its exported definition with its own functions as bodies passes its whole conformance corpus, with no stock stand-in |
| Facade | `engine/go/` | Vector tests from Go: language answers and 934 stack calls on live states (agent's report for the counts) |
| Engine gaps | `constraints.ts`, `stack.ts`, `rules/stock.ts`, `stockRegistry.ts`, `engine/lib/`, game predicates and component `Legal` in `rules/` | Constraint vectors (2,400 operations), stock games replayed, checkers vectors |
| Fuzz driver | `lib/fuzz`, `fuzz run`, `fuzz corpus [--check]`, `conformance/fuzz/` | 56 documents current under `--check`; deterministic; a truncated run is a prefix |
| Fuzz replay, coverage, mutations | `engine/testing/fuzz.ts`, `games/fuzz.test.ts`, `fuzz coverage [--blocks]`, `scripts/fuzz-mutations.mjs` | All 20 historical mutations fail the fuzz corpus |
| Translator | `lib/transpile`, `transpile [--check] [--dry-run] [--stats]` | `--check` current for seven games; accept and refuse sets equal the lint's |
| Orchestrator and agent harness | `lib/port`, `port`, `port check`, `port accept`, `port brief`, `--adopt` | `port --all --dry-run --check` current; `TestCheckMemory` passes for real |
| Generated games | `engine/games/<game>/{definition.gen.ts,bodies.gen.ts,index.ts}` | See "Generated games against Go" |

## Suites on the head commit

Run by the coordinator on the head commit (Go 1.27.4, Node v24.22.3).

| Suite | Result |
|---|---|
| `./scripts/go-local build ./...` | passes |
| `./scripts/go-local test ./boardgame-util/... ./moves/... .` | 26 packages ok, none failing (the four Node-backed `lib/port` tests need `BOARDGAME_NODE` set to Node 24, or they skip) |
| `conformance export --check` | current (the 22 documents are byte for byte as they were) |
| `fuzz corpus --check` | current, 56 documents |
| `gamedef export --check` | current |
| `transpile --check ./examples/...` | current |
| `port --all --dry-run --check` | current |
| `npm run type-check` | passes |
| `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | passes |
| `node scripts/fuzz-mutations.mjs historical` | baseline 56 of 56; 20 of 20 mutations killed; no source file changed |
| `npm run test:unit` | 1,426 tests, 1,426 pass, 0 skipped (1,168 at the start) |
| `npm run test:layers` | 11 of 11 |

Not run by anyone in this work: the webapp build, the e2e and renderer
suites, `go test ./...` over the whole repository, the Fabric pattern tests.
Whether a generated game loads in a real Common Fabric pattern was not
tried; `sandbox.test.ts` and the sandbox type check are the only evidence.

## Generated games against Go

Checked-in recordings (in the unit suite, which the coordinator ran). The
per-game counts are from the agents' reports.

| Game | Documents | Versions | Rejections | Differences |
|---|---|---|---|---|
| pig | 3 conformance + 8 fuzz | 696 | 25 | 0 |
| blackjack | 7 + 8 | 946 | 9 | 0 |
| debuganimations | 1 + 8 | 184 | 22 | 0 |
| tictactoe | 4 + 8 | 192 | 5 | 0 |
| werewolf | 1 + 8 | 170 | 18 | 0 |
| memory | 6 + 8 | 441 | 18 | 0 |
| checkers | 8 fuzz, 3 more fuzz documents, 44 constructed positions | see below | | 0 |

Transient runs, hand port and generated game side by side (agent's report;
not re-run by the coordinator): tictactoe 1,000 seeds, werewolf 128, pig
100, memory 24, blackjack 32, debuganimations 32, each at
`--max-versions 150`, and checkers 200 seeds at default bounds. No
difference in any. The proposal's 1,000 seeds at default bounds for every
game was not run: twelve memory documents took about six minutes in Go.

No hand-port bug was found by running anything. One was found by reading:
memory's capture discards the error from `MoveToNextSlot` in Go, and the
hand port would throw. The stack is uncapped, so no game reaches it
(`../transpile/HAND-PORT-BUGS.md`).

## The agent step

Two functions fall outside the translator, the two the proposal predicted.
Each was given to a coding agent that could read only its brief and its own
output file and run only `port check`.

| Function | Outcome |
|---|---|
| `memory.(*moveStartHideCardsTimer).Apply` (`package-var-read`) | PASS on the first check run: 14 of 14 documents, 4 of 4 Go blocks, 2 of 3 mutants (the survivor deletes `return null`, which changes nothing). Accepted with an `@generated agent:` header |
| `checkers.(*gameDelegate).ConfigurePredicateConstructors` (`predicate-constructor`) | Static and type checks passed; `port check` had no recordings to replay. The coordinator added checkers to the checked-in fuzz corpus; the check is then `PASS (unproven)`, 14 of 24 mutants killed. In the generated checkers test it matches the Go-recorded constructed positions, 4,702 ledgers included |

What the two agents said the briefs lacked: the value of the package
variable that caused memory's refusal (derived from the move table); an
import rule that reads as contradictory; for a predicate, no mapping of
framework constants, no API-table rows for package `legal`, no rule for a
nil slice or a failed type assertion; a reference to a "Panics" section
that does not exist. None of this was fixed in the brief writer.

## Checkers

Asked: does `Apply` fail after `Legal` passes? It does, but no real game
gets that far.

- **No move is legal from the starting position.** `Move Token`,
  `Crown Token` and `Finish Turn` are applied zero times in 200 seeds
  (agent's report), in 20 seeds the coordinator ran, and in the eight
  checked-in seeds. An agent tried all 4,096 inputs for the player to move:
  none is legal. Its causes, not re-derived by the coordinator:
  `spaceIsBlack` is `index % 2 == 0`, which marks columns; each color's
  graph points at its own back edge; both players' `Color` is red in the
  stored state.
- **From constructed positions** (44, built on a real Go state by
  `testdata/gen/checkersvec`): 109 `Move Token` proposals pass `Legal`. Of
  those `Apply` succeeds for 22, returns "Invalid result from range to
  value" for 86 (the finish-minus-start "middle" arithmetic), and one
  capture is refused by `same(Color)`. Agent's report.
- The generated checkers reproduces all of it with no difference. No Go was
  changed. A faithful port is a port of a game that stops after set-up.

## The measured baseline

`boardgame-util lint --declarative ./examples/...`, against the proposal's
hand count in parentheses where they differ.

| Game | Configs | Data | Code | Specs | Functions | Subset | Outside | Engine gap | Not ported |
|---|---|---|---|---|---|---|---|---|---|
| memory | 7 | 3 | 4 | 9 | 17 (16) | 16 (15) | 1 | 0 | 11 |
| blackjack | 23 | 16 | 7 | 6 | 23 (22) | 23 (22) | 0 | 0 | 1 |
| pig | 6 | 3 | 3 | 4 | 10 | 10 | 0 | 0 | 1 |
| tictactoe | 4 | 3 | 1 | 0 | 11 | 11 | 0 | 0 | 5 (6) |
| werewolf | 11 | 4 | 7 | 0 | 16 | 16 | 0 | 0 | 1 |
| debuganimations | 13 | 1 | 12 | 7 | 24 | 24 | 0 | 0 | 1 |
| checkers | 7 | 4 | 3 | 5 | 20 | 19 | 1 | 0 (about 11) | 0 |

A computed property is two bodies (the closure and the method it calls),
which is the extra function in memory and blackjack.

## Decisions made by the coordinator

1. Existing conformance documents stay `schemaVersion` 1. A document is
   version 2 only when it carries `rejections`.
2. `LegalPlanSpecs` returns `(specs, declared, hasCustom, ok)`.
3. `RepeatFromProp` and any foreign group are described as `unknown`.
4. A counter is fitted at the length its group passes it, on the grid 0 to
   64; `any` at length 1 is described as `all`.
5. A stock type configured directly is its own stock ancestor.
6. Fuzz seeds for werewolf are the first eight that derive a seated game and
   for debuganimations the first eight unseated, chosen by the derived
   set-up, never by the outcome. The driver proposes no admin move.
7. The loader takes a stock `Apply` only when the registry supplies one; a
   base type's placeholder is not one.
8. The engine-gap table follows `stock-inventory.json`. `RoundRobin` stays a
   gap: it is ported as parts but cannot be built from configuration.
9. One more read-only Go accessor than the proposal listed:
   `moves.ResolvedSlotField`, because Go recomputes that value and never
   stores it.
10. `GameData` gained `ComponentLegal` (deck to the component type's `Legal`).
11. A move's own fields are an enum owner in the facade under two names,
    `fieldsOf` (what the translator emits) and `move`; two agents each added
    one and both were kept.
12. Refused functions are never written into generated output by the
    translator or by hand; only `port accept` adds an agent block.
13. Checkers joined the checked-in fuzz corpus (8 seeds, 160 KB), so that
    `port check` has something to replay. The corpus is 56 documents.
14. The owner's defaults were followed: additive read-only Go accessors
    only; generated games beside the hand ports, none deleted; checkers
    fuzzed in Go alone first and no Go bug fixed; generated TypeScript
    checked in with per-function hash headers.

## Where the proposal was wrong about the code

- "SchemaVersion becomes 2" would have rewritten the checked-in corpus.
- Checkers has no enum-keyed board; `Spaces` is a sized stack indexed by the
  range enum's keys.
- Go has four stack-constraint kinds, and a multi-component transfer into a
  constrained stack is validated on a copy of the state.
- A `ComponentInstance` is the component's identity, so a snapshot handle
  whose component has moved still works. Only a component the index lost
  fails as "not public information".
- `NeighborsByKey` has no Go order (map iteration); the facade returns
  ascending.
- The agent's OUTPUT rule "imports only from `../../go` and `../../lib`"
  cannot hold: signatures need types from `rules/` and `legal/`, and callees
  come from `./bodies.gen.ts`.
- A header's `first:` commit cannot name the commit that adds the block; it
  names HEAD at generation.
- Risk 1 predicted `Apply` failing after `Legal` in play. It cannot happen
  in play; see "Checkers".
- Counts: 52 stock move types, not about 50; 121 functions, not 119;
  `testdata/defs` holds four games and vecgame, not five games; memory's and
  debuganimations' misspelled `DefaultNumPlayeres` returns the base default
  anyway.
- A conformance ledger exists only for a move offered in a view, so fix-up
  plans cannot be compared with it.
- Example games are not safe for concurrent managers (a delegate memoizes
  its name in a package variable), so fuzz runs are sequential.
- Go block coverage under `./scripts/go-local` works only when the command's
  own main package is in `-coverpkg`.

## Not done

- **Pair store and rule proof (workstream 11).** Still running, not merged:
  the hand ports split into stored pairs, `TestRuleAgainstPairs`, the
  substitution proof of the translator against every pair, and
  `fabric/transpile/GO-BUGS.md`. Until it lands, `port accept` stores no
  pair and briefs take their examples from the generated games.
- **The ratchet has never turned.** No refusal reason has been made a
  translator rule, so `agentNowMechanical` is 0 by construction and
  `--adopt` was exercised only on pig as a round trip (agent's report).
- **The second-agent review** of each accepted agent translation, which the
  proposal calls for, was not run on the two.
- **Surviving mutants per game** are `null` in the reports; mutants run per
  function in `port check`.
- **The brief writer's gaps** listed under "The agent step".
- The stub games under `boardgame-util/lib/stub/testdata/` and the GAMES
  repository, which is not on this machine.

## Unverified

- The exporter reads three unexported values by reflection (move-name
  sanitization policy, computed-property order, unexported config structs).
- A move whose only stock ancestor is `moves.Default` exports `Stock: null`,
  so its custom configuration has no place in the schema.
- Translator rules no checked-in game reaches at run time (maps other than
  werewolf's, labels, tagless switch, named results): golden text and the
  sandbox type check only. Integer overflow is checked for `*` only.
- Facade: `values` and `dynamic` on a generic component, sanitized states,
  board spaces, a plain enum move field; `atoi` past 2^53 reports a range
  error where Go would not.
- Constraints naming a chest constant; constraints on generic components.
- Fuzz driver branches no game triggered: an `Apply` error after `Legal`, a
  timer error, a seat rejection. The replay of `timer` and `seat`
  rejections is likewise untested.
- No game records an `unknown` predicate verdict; pig's end condition is
  not reached in the checked-in fuzz corpus; no tictactoe game there is won
  on a diagonal.
- Checkers' block coverage is 82 of 151 in the fuzz corpus, since play never
  starts; the rest is held only by the constructed positions.
- The full boot-error text when a game predicate's constructor fails.

## Decisions left for the owner

1. **Checkers.** The Go game cannot be played. Fix the Go and regenerate, or
   keep the port of the game as it is.
2. **Do generated games replace the hand ports?** All six now pass their
   conformance corpus and a clean fuzz run, which was the stated bar. It
   changes the policy identity of the memory and blackjack patterns.
3. **Three additive Go files in existing packages**: `moves/describe.go`,
   `moves/resolved_config.go`, `legal_plan_export.go`. Also: should the
   exporter's three reflective reads become exported accessors?
4. **The facade idiom.** Generated code reads like transliterated Go.
5. **What a header may contain**, if comments are part of a pattern's
   identity. Headers carry producer, function, source hash, and the first
   commit.
6. **Human review** of agent translations that feed legality: checkers'
   predicate is one.
7. **Acceptance run size.** Eight seeds per game are checked in. The
   transient runs here were 24 to 1,000 seeds at 150 versions, not 1,000 at
   default bounds.
8. **Refactor Go to be translatable?** Memory's refused function becomes
   mechanical if a package variable becomes a constant. An agent took it
   instead, in one try.
9. **AI players**, the exporter's `winners: null`, and integer overflow
   stand as the proposal left them.
10. **The Go findings** in `../transpile/status/go-descriptions.md` (sticky
    `max` in `CountAtMost` and `CountBetween`, `CountAll`'s doc comment) and
    the checkers bugs, none of which was changed.
