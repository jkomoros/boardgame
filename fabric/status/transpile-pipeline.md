# Status: transpile pipeline (in progress)

Branch `claude/transpile-pipeline-e23924`, local only. Nothing pushed, no
pull request, `master` and `fabric-factoring` untouched. The plan is
[`../TRANSPILE-PROPOSAL.md`](../TRANSPILE-PROPOSAL.md). A coordinator session
assigned each workstream to an Opus sub-agent in its own worktree, read each
diff, re-ran its tests and merged it. Each stream's own report is under
[`../transpile/status/`](../transpile/status/); the numbers below are from the
coordinator's re-runs unless a line says "agent's report".

This file is rewritten as phases land. Sections marked PENDING are not done.

## Environment

- Go 1.27.4 through `./scripts/go-local`. Node v24.22.3; the shell's default
  Node is v18, under which `npm run test:unit` prints "bad option" and
  reports one passing test. Every Node command here was run with
  `PATH=$HOME/.nvm/versions/node/v24.22.3/bin:$PATH`.
- Baseline on `a11c3883`: Go build passes; Go tests of
  `lib/conformance/...`, `lib/lint`, `lib/gamepkg`, `moves`, root pass;
  type check passes; 1,168 unit tests; 11 layer tests.

## Merged so far

| Stream | Merge | What | Coordinator re-run |
|---|---|---|---|
| 3 Go descriptions | `c6953217` | `moves.DescribeProgression`, `DescribeCounter`, `InstalledProgression`, `NewCounterFromDescription`; `(*GameManager).LegalPlanSpecs` | build, vet, `moves` and root tests pass |
| 0 Freeze | `d2cb55e4` | `GameData` schema (Go and TypeScript), refusal list, facade signatures, `rejections`, additive engine fields | build, vet, corpus current, 1,168 unit, 11 layer |
| 1 Exporter, run-time half | `50539d60` | `gamedef.Export`, `gamedef export [--check]`, seven definitions in `conformance/definitions/` | build, vet, `--check` current, `testdata/defs` unchanged after regeneration, `gamedef` tests pass |
| 6 Loader | `ab708abc` | `gameFromData`, body calling convention, `MoveSpec.plan`, `initialValues`, generic values | 1,184 unit (6 skipped, see below) |
| 8 Constraints, by-key | `53a66ca7` | `constraints.ts` (max, unique, same, maxdistinct), `componentAtKey`, `swapComponentsByKey`, `mayMoveComponentToSlot` | 1,192 unit |
| 4 Fuzz driver | `09f9e6a9` | `lib/fuzz`, `fuzz run`, `fuzz corpus [--check]`, 48 documents in `conformance/fuzz/` | build, vet, `fuzz corpus --check` current, `lib/fuzz` tests pass (41 s), no sleep or timeout in the driver (grep) |
| 7 Facade | `21be23d7`, `cc6a81d6` | `engine/go/` implemented against Go vectors; `stockApply` wired to the loader by the coordinator | 1,223 unit (1,217 pass, 6 skipped), 11 layer |

Running: 2 (static half and lint), 9 (stock moves, behaviors, enums, graphs,
libraries), 5 (fuzz replay, coverage, mutations).

## Checkers in Go alone

Asked: does `Apply` fail after `Legal` passes? Answer: it is never reached.
**No move can be made in a game of checkers as the Go stands.** In 200 seeds
(agent's report) and 20 seeds re-run by the coordinator, `Move Token`,
`Crown Token` and `Finish Turn` are applied zero times; every run ends at the
first playing position with no legal action. The agent tried all 4,096
inputs for player 0 at that position: none is legal. Causes, from its report
(not re-derived by the coordinator): `spaceIsBlack` is `index % 2 == 0`, so
tokens sit in even columns and every diagonal neighbor is an odd column;
each color's graph points toward its own back edge; both players' `Color` is
red in the stored state, so the black tokens belong to nobody. The
finish-minus-start arithmetic in `(*moveMoveToken).Apply` (the proposal's
risk 1) cannot be observed, because no proposal is ever legal. No Go was
changed.

Consequence for the plan: a faithful port of checkers is a port of a game
with no legal move. The fuzzer can check set-up, seating, the 24 placements
and refusals, and nothing after. `Apply`, the capture arithmetic, crowning
and the custom predicate can only be exercised from a constructed position.

## Decisions made by the coordinator

1. Existing conformance documents stay `schemaVersion` 1. A document is
   version 2 only when it carries `rejections`. The proposal's "SchemaVersion
   becomes 2" would have rewritten the checked-in corpus.
2. `LegalPlanSpecs` returns `(specs, declared, hasCustom, ok)`: a declared
   plan can have zero specs, so a bare slice cannot carry `Plan`.
3. `RepeatFromProp` and any foreign group are described as `unknown`, with a
   detail, never approximated.
4. A counter is fitted at the length its group passes it and on the grid
   0 to 64 only; `any` at length 1 is described as `all`.
5. The exporter treats a stock type configured directly as its own stock
   ancestor.
6. A definition exported without the static half is skipped by the loader
   test with a message, not failed and not passed.
7. Seeds for werewolf are the first eight that derive a seated game, and for
   debuganimations the first eight unseated, chosen by the derived set-up and
   never by the run's outcome. The fuzz driver proposes no admin move,
   because `FromHistory` replays a non-seat admin move as a timer.

## Where the proposal was wrong about the code

- Checkers has no enum-keyed board: `Spaces` is a sized stack indexed by the
  range enum's keys.
- Go has four stack-constraint kinds, and a multi-component transfer into a
  constrained stack is validated on a copy of the state.
- A `ComponentInstance` is the component's identity, so a snapshot handle
  whose component has moved still works. Only a component the index lost
  (after `ContractSize`) fails as "not public information".
- `NeighborsByKey` has no Go order; it is map iteration order. The facade
  returns ascending.
- `testdata/defs` holds four games and vecgame, not five games. Tree-enum
  `Parents` was not emitted by Go's `DefinitionFor` before this work.
- Memory's and debuganimations' misspelled `DefaultNumPlayeres` returns 2,
  the base default, so the misspelling changes nothing observable.
- The ledger in a conformance document exists only for a move offered in a
  view, so fix-up plans cannot be compared with it.
- Example games are not safe for concurrent managers (a delegate memoizes
  its name in a package variable), so fuzz runs are sequential.

## Unverified so far

- The loader against real exported definitions: the checked-in definitions
  hold run-time parts only until the static half is joined. Proven only on
  hand-written fixtures for pig and blackjack.
- No move has been built through the real stock registry.
- The exporter reads three unexported values by reflection (move-name
  sanitization policy, computed-property order, unexported config structs).
- A move whose only stock ancestor is `moves.Default` exports `Stock: null`,
  so its custom configuration has no place in the schema (12 of
  debuganimations' 13 moves).
- Facade: `values` and `dynamic` on a generic component, sanitized states,
  board spaces; `atoi` past 2^53 reports a range error where Go would not.
- Constraints naming a chest constant; constraints on generic components.
- Fuzz driver branches no game triggered: an `Apply` error after `Legal`, a
  timer error, a seat rejection.

## PENDING

Static half and lint; stock registry; fuzz replay, coverage and mutations;
translator; pair store and reports; orchestrator and agent harness; pig
through the pipeline; checkers; the other five games; final suite counts;
decisions for the owner.
