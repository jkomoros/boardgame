# Hand-port bugs found by the fuzz replay

Every difference between a hand port (or the TypeScript engine) and Go that
the fuzz replay (`server/static/src/engine/testing/fuzz.ts`,
`games/fuzz.test.ts`) found, with its divergence report, root cause, and the
fix or the reason it is left. Workstream 5 of `fabric/TRANSPILE-PROPOSAL.md`.

## Checked-in corpus (`conformance/fuzz/`, 48 documents)

**No difference.** All 48 documents of the six hand ports pass on the first
run: every version's state, move record and status, every viewer's game
object, sanitized move, legality and projected choices, and all 97 recorded
rejections (Go's exact error text, no versions returned, state and history
untouched). Nothing was fixed, and `games/fuzz.test.ts` lists no expected
failure (`EXPECTED_FAILURES` is empty).

## Transient run

**No difference.** Seeds 100-111 of each of the six games, max versions
150 (`boardgame-util fuzz run --game G --seeds 100..111 --max-versions 150
--out DIR`, written outside the repository), replayed with
`BOARDGAME_FUZZ_DIR=DIR`: 72 of 72 documents pass, 545 rejections equal.

## How to add an entry

When a document fails, `runFuzzDocument` returns a `report` (the test prints
it as JSON). Record it here with the root cause; if the fix is outside the
files that may be edited, add the document's name and the report's `field`
to `EXPECTED_FAILURES` in `games/fuzz.test.ts`, which then requires it to keep
failing on that field and fails when it starts passing.

## Generated games against the hand ports (Phase 4, `status/games.md`)

### memory: Capture Cards throws where Go discards `MoveToNextSlot`'s error

**Not observable; read, not run.** No recorded document, checked-in or
transient, differs, and none can: the difference needs a failing move.

- Go, `examples/memory/moves.go` `(*moveCaptureCards).Apply`:
  `game.VisibleCards.ComponentAt(i).MoveToNextSlot(p.WonCards)` as a bare
  statement; the error is dropped and the loop continues; `Apply` returns nil.
- Generated, `games/memory/bodies.gen.ts` `memory_moveCaptureCards_Apply`:
  `moveToNextSlot(componentAt(...), playerStack(state, p, 'WonCards'));` as a
  bare statement. The facade's `moveToNextSlot` returns Go's error text
  (`go/api.ts` `moveToEnd` through `stackCall`; vectored against
  `c.MoveToNextSlot(dst)` in `testdata/gen/facadevec/stacks.go`), so the
  generated body drops it exactly as Go does.
- Hand port, `games/memory.ts` Capture Cards `apply`:
  `visible.moveComponentToNextSlot(i, won)`, the engine stack call, which
  throws the stack error; the move would fail where Go's succeeds.
- Reachability: `WonCards` is a growable `boardgame.Stack` (`state.go:31`,
  no size or constraint), and the component always exists at slot `i` (the
  loop's snapshot says so), so no move to it fails today. Left as is: the
  hand ports are not edited (Phase 4 brief), and the generated game is the
  faithful one.

## Pair store and substitution proof (workstream 11)

The six hand ports were split into 98 stored pairs
(`boardgame-util/lib/transpile/testdata/pairs/`) and each pair's translator
output was substituted, one at a time, for its `accepted.ts` in a game
built from the definition and the store, then replayed against the game's
whole conformance and fuzz corpus (`TestSubstitutionProof`,
`boardgame-util/lib/transpile/pairs`). See `status/pairs.md`.

**No hand-port difference from Go was observed.** Every one of the 86
substituted translations (all pairs the loader resolves whose Go the
translator accepts) passes every document; the 11 translated helpers pass
inside their callers. The store itself (all 98 accepted functions together)
passes all 70 documents of the six games.

**One harness failure, fixed in the harness.** On the first run the
translations of `memory.(*moveRevealCard).Apply` and
`tictactoe.(*movePlaceToken).Apply` failed at version 1 of the first
document: `the engine refused the recorded propose: The move's apply
function returned an error:move Reveal Card embeds no stock move with an
Apply` (likewise Place Token). Root cause: neither the translator nor the
hand port. Both moves are `moves.MoveComponentToSlot` without
`WithSlotField`, which `rules/stockRegistry.ts` cannot build from the
exported configuration (the gap in `status/integration-1.md`), so the
tests give them stock parts through the `stock:` hook. `rules/data.test.ts`
builds those parts from the hand port, whose `apply` already calls its own
`reveal().apply` / `place().apply`, so the parts carry no stock `apply`; the
translated body calls Go's `m.MoveComponentToSlot.Apply(state)` as
`stockApply(state, m)`, which needs one. `games/pairs.testkit.test.ts` now
builds those two moves with the registry's own `MoveComponentToSlot` and the
slot field Go discovers at boot (the move's only int field,
`moves/move_component_to_slot.go resolveSlotField`). A generated memory or
tictactoe needs the same: the registry (or the exporter) has to supply the
discovered slot field.

**Differences read while splitting, which no document shows** (each is in
the pair's `meta.json` note):

- `tictactoe.checkRunWon` and `checkGameFinished`: Go first returns Empty
  (and `false, Empty`) when the run or board has the wrong length; the hand
  port's `runWinner` and `boardResult` do not check. Every call passes three
  and nine tokens.
- `werewolf.finishVoteResolution`, `resolveVotes`, `startVoteTimer`: Go
  returns errors; the hand port throws them (the move fails with the same
  text either way).
- `werewolf.populateFellowWolves`: Go walks the slice it is passed, the
  hand port every player; the one caller passes every player.
- `memory.(*moveCaptureCards).Apply` discards `MoveToNextSlot`'s error; the
  hand port would throw (known; `GO-BUGS.md` G4).

## Closed: the hand ports are replaced

The six hand ports (`server/static/src/engine/games/{memory,blackjack,
tictactoe,pig,debuganimations,werewolf}.ts`) were removed when the owner
ruled that a generated game is the one in use wherever it is fully
semantically equivalent to its hand port, and all six were
(`status/replacement.md`: the structural comparison, the rejection vectors,
the engine tests, and a fresh transient fuzz run of 1,316 documents
replayed against both, with no difference). Nothing in this register was
ever a difference a recording showed. The read-only differences above are
moot for play: each was the hand port departing from Go where no document
reaches, and the generated game follows Go in every one of them.

This register stays as the record of what the fuzz replay and the pair
store found while the hand ports existed. Their functions are kept, one
per Go function, in `boardgame-util/lib/transpile/testdata/pairs/`. A
difference the fuzz replay finds from now on is a generated game's (or the
engine's) and is recorded here the same way, with `EXPECTED_FAILURES` in
`games/fuzz.test.ts`.
