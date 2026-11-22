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
