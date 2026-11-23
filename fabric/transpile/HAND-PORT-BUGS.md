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
