# Workstream 11: pair store and rule proof

Base `17168752`. What `fabric/TRANSPILE-PROPOSAL.md` workstream 11 asked for
("The ratchet and its metrics": stored pairs, turning a refusal reason into a
rule), plus the Go-bug register. One local commit; nothing pushed. The
per-game reports and `REPORT.md` are another agent's (workstream 12 brief),
not built here.

## What was built

| Path | What |
|---|---|
| `boardgame-util/lib/transpile/testdata/pairs/<game>/<name>/{source.go,accepted.ts,meta.json}` | The store: 98 pairs, producer `hand` |
| `boardgame-util/lib/transpile/testdata/pairs/index.ts` | Generated: imports every `accepted.ts`; `accepted(game, ref)`, `pairRefs(game)` |
| `boardgame-util/lib/transpile/testdata/pairs/tsconfig.json` | Type-checks the store under the engine's sandbox options |
| `boardgame-util/lib/transpile/pairs/` (`pairs.go`, `store.go`, `source.go`, `translate.go`, `proof.go`, `node.go`, `pairs_test.go`) | Reading, writing and validating the store; staleness; spans, block coverage, seeds; the translator's output as a substitutable module; the substitution proof; `ProveRule`; the Go tests below |
| `boardgame-util/lib/transpile/pairs/testdata/gen/main.go` | Rebuilds the store (`-check`, `-coverage`, `-profile FILE`, `-mutants FILE`, `-not-mutated WHY`) |
| `server/static/scripts/pairs-extract.mjs` | Splits the six hand ports into pairs (prints JSON; writes nothing) |
| `server/static/scripts/pairs-proof.mjs` | The TypeScript half of a proof: replays games with substitutions (driven by Go) |
| `server/static/scripts/pairs-mutants.mjs` | Body mutants of every resolved pair's `accepted.ts` against its game's corpus |
| `server/static/src/engine/games/pairs.testkit.test.ts` | Shared by the test and the two scripts: a game from its definition and the store, the corpus, one document's check (no test of its own) |
| `server/static/src/engine/games/pairs.test.ts` | The store is current against the hand ports; each game from the store passes its whole corpus |
| `fabric/transpile/GO-BUGS.md` | The register |
| `fabric/transpile/HAND-PORT-BUGS.md` | Appended: what the pairs and the substitution proof found |

### Rebuilding the store

```
BOARDGAME_NODE=$HOME/.nvm/versions/node/v24.22.3/bin/node \
  ./scripts/go-local run ./boardgame-util/lib/transpile/pairs/testdata/gen          # extract, write
./scripts/go-local run ./boardgame-util fuzz coverage --blocks --work DIR           # profile
./scripts/go-local run ./boardgame-util/lib/transpile/pairs/testdata/gen -profile DIR/fuzz-corpus.coverprofile
(cd server/static && node scripts/pairs-mutants.mjs --jobs 6 --json FILE)
./scripts/go-local run ./boardgame-util/lib/transpile/pairs/testdata/gen -mutants FILE -not-mutated "helper: ..."
```

The generator is the only writer. It runs `pairs-extract.mjs`, takes each
pair's Go source from the live manifest (`gamedef.LoadStatic`), checks it
hashes to the manifest's `SourceHash`, and keeps a measured field only while
what it measured is unchanged (blocks: the source hash; mutants: the
`accepted.ts` text, checked by SHA-256 against the mutant run).

### `meta.json`

```json
{ "func", "role", "sourceHash", "producer": "hand", "refusals": [ ...the manifest's, [] when accepted ],
  "extraction": "mechanical" | "wrapped" | "none", "note",
  "fuzz": { "seeds", "blocksCovered", "blocksTotal", "unreached", "mutantsKilled", "mutantsTotal", "survivors", "why" } }
```

`unreached`, `survivors` and `why` are additions to the proposal's fields
(listing what the counts summarize, and why a count is null).
`source.go` is the function's bytes exactly as `gamedef` hashes them, with
no newline added.

### Exported Go interface (`package pairs`)

`Load(root) ([]Pair, error)`; `(Pair).Validate()`; `Stale(pairs,
manifests) []Staleness`; `Manifests(root, games)`; `Spans(report)`;
`BlockCoverage(blocks, game, spans)`; `Seeds(root, game)`; `Extract(node,
root)`; `Build(...)`, `WriteStore(...)`, `UpdateMeta(...)`, `IndexText(...)`;
`Translate(root, game, refs, rewrite) map[FnRef]Translation`;
`Substitute(root, pairs, games, Options) []GameProof`; `SubstituteGame`;
`ProveRule(root, pairs, reason, detail, Options) RuleProof` (with
`Proven()`, `StillRefused()`); `RefusalsCarried(pairs)`; `Node()` (Node
from `BOARDGAME_NODE`, else `node` on PATH, 22.18 or later); `Signature`,
`TextuallyEqual`.

## The split

| Game | Manifest functions | Pairs | mechanical | wrapped | none | No pair |
|---|---|---|---|---|---|---|
| memory | 17 | 17 | 15 | 2 | 0 | 0 |
| blackjack | 23 | 22 | 22 | 0 | 0 | 1 |
| tictactoe | 11 | 9 | 7 | 2 | 0 | 2 |
| pig | 10 | 10 | 10 | 0 | 0 | 0 |
| debuganimations | 24 | 24 | 7 | 16 | 1 | 0 |
| werewolf | 16 | 16 | 12 | 4 | 0 | 0 |
| **total** | **101** | **98** | **73** | **24** | **1** | **3** |

How each kind is made (`pairs-extract.mjs` top comment): for every FnRef
the loader resolves, the hand port's function for that role is the one
`rules/data.test.ts`'s adapter uses; it is located in the hand port's source
by its `toString()` (Node strips types to spaces, so the text is found in
`module.stripTypeScriptTypes(source)` at the original offsets). Helpers have
no runtime slot; a table in the script names the hand-port declaration for
each, or says there is none (a new Go helper fails the extraction until it
gets an entry).

- **mechanical (73):** the function's body under the role's signature (the
  hand port's parameter names, the convention's types), with every
  module-scope declaration it reaches copied above it (found with the
  TypeScript checker) and its imports re-pointed at the engine.
- **wrapped (24):** 16 debuganimations move members that read a factory's
  `options` (`moveBetween`, `toggleThird`, `moveAll`): the factory is copied
  and the wrapper calls it with the configuration the game passes (found
  mechanically: the one call whose `name` is the move's) and calls the
  member. 8 hand-written wrappers, each with a note: `memory PlayerScore`
  (the gameScore trait the engine falls back to), `memory
  CurrentPlayerHasCardsToReveal` (inlined into the computed property),
  `tictactoe checkRunWon`, `checkGameFinished` (object to tuple; no length
  check), `werewolf finishVoteResolution`, `resolveVotes`, `startVoteTimer`
  (throw where Go returns an error; the wrapper returns null),
  `populateFellowWolves` (ignores the slice it is passed).
- **none (1):** `debuganimations.(*gameDelegate).CheckGameFinished`: the
  hand port has no `end`; the body is the engine default written out
  (`return [false, null]`).
- **No pair (3):** `blackjack.handValue` (Go's takes the hand's components;
  the hand port's `handValue(state, player)` is `(*playerState).HandValue`
  with it inlined), `tictactoe.(*gameState).tokenValueAtIndex` (inlined in
  `boardTokens`), `tictactoe.rowColToIndex` (inlined as `boardResult`'s
  local `at`).

The one Go method serving two configs (werewolf's `moveCastVote`,
`moveResolveVotes`, `moveResolveVotesOnTimeout`) is one closure expression
in the hand port, copied once; the note says so.

### Staleness (decided)

`TestStore` fails if a pair's `source.go` does not hash to its
`meta.sourceHash` (a corrupt pair). `TestStale` compares each pair with the
live manifest: a `hand` pair whose Go changed (or vanished, or changed role)
is **logged, never a failure**, because the hand port is not regenerated
from Go and its pair says what the hand port does; rebuilding the store
after re-checking the hand port clears it. Any other producer's stale pair
fails. Separately, `pairs.test.ts` fails if a hand port changed and the
store was not rebuilt (it re-runs the extraction and compares each
`accepted.ts` and the extraction fields).

### Where the store is type-checked (decided)

Outside the engine's sandbox project, on purpose: the store is Go test data
under `boardgame-util/`, the sandbox test refuses imports that leave the
engine, and the test project's `rootDir` is `server/static` (a static
import of the store from a test fails with TS6059). So
`pairs.testkit.test.ts` loads `index.ts` by a computed specifier, and the
store has its own `tsconfig.json` extending `tsconfig.sandbox.json`
(`exactOptionalPropertyTypes`), run by the Go test `TestStoreTypeChecks`
(99 files: 98 `accepted.ts` and `index.ts`). `sandbox.test.ts` and the
purity tests are untouched and pass. The extracted text comes from hand
ports the sandbox test already checks; the store's own module scope holds
only constants, functions and types copied from them.

## Proof that the pairs are faithful (`games/pairs.test.ts`)

Each game built by `gameFromData` from `conformance/definitions/<game>.json`
with a bodies table made entirely of the stored `accepted.ts` functions,
replayed against every conformance document and every checked-in fuzz
document (rejections re-proposed):

| Game | Documents | Pairs | FnRefs served by pairs | Not resolved (helpers) | Other bodies |
|---|---|---|---|---|---|
| memory | 14 | 17 | 16 | 1 | `stock:` Reveal Card (below) |
| blackjack | 15 | 22 | 21 | 1 | none |
| tictactoe | 12 | 9 | 7 | 2 | `stock:` Place Token (below) |
| pig | 11 | 10 | 9 | 1 | none |
| debuganimations | 9 | 24 | 24 | 0 | none |
| werewolf | 9 | 16 | 10 | 6 | the hand port's `setUp.distribute` for the empty FnRef `""` (Go's werewolf has no `DistributeComponentToStarterStack`, so nothing to pair) |

All 70 documents pass: 2,629 versions, 97 rejections. Every FnRef the
loader resolves is served by a pair (memory's refused
`(*moveStartHideCardsTimer).Apply` included); the test fails on any other
fallback.

**The stock stand-in.** memory's Reveal Card and tictactoe's Place Token are
`MoveComponentToSlot` without `WithSlotField`, which `stockRegistry.ts`
cannot build yet. They go through the same `stock:` hook
`rules/data.test.ts` uses, but with the **registry's own parts** built with
the slot field Go discovers at boot (the move's only int field,
`resolveSlotField`), not the hand port's parts: a translated `apply` calls
`stockApply`, which needs the stock `apply` that hand-port parts do not have
(see "Substitution proof"). When the registry learns to discover the field,
the hook goes unused and the test still passes (it allows these two, at
most).

## Fuzz fields

**Blocks** (`fuzz coverage --blocks --work DIR`, then `gen -profile`): every
pair measured. A block counts for the innermost manifest function holding it,
so `ConfigureComputedProperties.func1` is measured alone.

| Game | Pairs with every block reached | Blocks reached |
|---|---|---|
| memory | 11 of 17 | 79 of 91 |
| blackjack | 13 of 22 | 81 of 94 |
| tictactoe | 5 of 9 | 63 of 69 |
| pig | 8 of 10 | 36 of 38 |
| debuganimations | 11 of 24 | 135 of 154 |
| werewolf | 10 of 16 | 143 of 151 |
| **total** | **58 of 98** | **537 of 597** |

Most unreached blocks are error returns (`if err != nil { return err }`)
the corpus never takes; `memory PlayerScore` is 0 of 1 (no fuzz game ends).
Each pair lists its unreached blocks in `meta.json`.

**Mutants (unfinished).** `pairs-mutants.mjs` mutates only the exported
function's body (comparison flipped, integer literal shifted, statement
deleted, one at a time), builds the game from the store with that one
function replaced, and replays the corpus to the first failing document.
What was run and recorded:

| Game | Pairs measured | Mutants killed | Command |
|---|---|---|---|
| pig | 9 of 10 (1 helper) | 33 of 40 | `pairs-mutants.mjs --game pig --jobs 4` |
| tictactoe | 7 of 9 (2 helpers) | 32 of 35 | `pairs-mutants.mjs --game tictactoe --game werewolf --game debuganimations --jobs 3` |
| werewolf | 10 of 16 (6 helpers) | 109 of 124 | same run |
| debuganimations | 24 of 24 | 68 of 89 | same run |
| memory | 0 of 17 | null | started, stopped unfinished at the coordinator's request |
| blackjack | 0 of 22 | null | started, stopped unfinished at the coordinator's request |
| **total** | **50 of 98** | **242 of 288** | |

The 48 unmeasured pairs record `mutantsKilled`/`mutantsTotal` null with the
reason in `fuzz.why`: 11 are helpers (never resolved by the loader, so a
mutant cannot change the game: they will stay null), 37 are memory and
blackjack pairs whose run was stopped. Those runs are slow because a
surviving mutant costs a whole-corpus replay (blackjack's computed
`HandValue` alone has 31 mutants); the first attempt over all games was
killed after 40 minutes with one worker still in blackjack. To finish:
`(cd server/static && node scripts/pairs-mutants.mjs --game memory --game
blackjack --jobs 6 --json FILE)`, merge with the recorded runs, then `gen
-mutants FILE`. Each pair's survivors are listed in its `meta.json`
(`survivors`); most are statement deletions the corpus cannot see (an
initial value the definition already sets, a `return false` turned into an
implicit `undefined`, the hand port's redundant resets).

## `TestRuleAgainstPairs`

For every (reason, detail) any stored pair carried, `ProveRule` checks
(1) the translator now accepts each such pair's Go source; (2) with the
translator's output (the function plus every in-package callee, as one
module importing the facade by absolute path) substituted for that pair's
`accepted.ts`, the game's conformance and fuzz corpora pass (a helper is
substituted through every stored caller whose translation bundles it); (3)
the pair's recorded block coverage is complete, else it does not count. A
rule is proven when every carrying pair passes 1 and 2 and at least one
counts. Textual equality is reported.

Today one refusal is carried: `package-var-read memory.hideCardMoveName`
by `memory.(*moveStartHideCardsTimer).Apply` (blocks complete). The
translator still refuses it, so the test logs "not yet a rule"; when the
translator accepts it, the test requires the proof.

## The general substitution proof (`TestSubstitutionProof`)

For every stored pair whose Go the translator accepts, its translation
substituted alone for `accepted.ts`, the rest of the game from the store,
the game's whole corpus replayed (stopping at the first failing document):

| Game | Pairs | Translated | Substituted | Passing | Helpers (exercised through callers) | Textually equal | Failing |
|---|---|---|---|---|---|---|---|
| memory | 17 | 16 | 15 | 15 | 1 | 0 | 0 |
| blackjack | 22 | 22 | 21 | 21 | 1 | 1 | 0 |
| tictactoe | 9 | 9 | 7 | 7 | 2 | 0 | 0 |
| pig | 10 | 10 | 9 | 9 | 1 | 0 | 0 |
| debuganimations | 24 | 24 | 24 | 24 | 0 | 0 | 0 |
| werewolf | 16 | 16 | 10 | 10 | 6 | 0 | 0 |
| **total** | **98** | **97** | **86** | **86** | **11** | **1** | **0** |

Not translated: `memory.(*moveStartHideCardsTimer).Apply`
(`package-var-read memory.hideCardMoveName`). Textually equal: one blackjack pair
(named in the test's log; the run that logs it by name was not repeated). Baselines (the store alone): memory 14 documents / 441
versions / 18 rejections, blackjack 15 / 946 / 9, tictactoe 12 / 192 / 5,
pig 11 / 696 / 25, debuganimations 9 / 184 / 22, werewolf 9 / 170 / 18.

**First run: 2 failures, root cause the test harness, now fixed.** The
translations of `memory.(*moveRevealCard).Apply` and
`tictactoe.(*movePlaceToken).Apply` failed at version 1: "move Reveal Card
embeds no stock move with an Apply". The stand-in stock parts were the hand
port's, which carry no stock `apply`. Fixed in `pairs.testkit.test.ts` as
described above; recorded in `HAND-PORT-BUGS.md`. No translator bug and no
hand-port bug was found: 86 of 86 pass.

## The broken rule (`TestBrokenRuleFailsProof`)

The translator's output for `blackjack.handValue` is changed in memory
(`currentValue >= 11` to `currentValue > 11`: the "fix" of the ten-plus-ace
bug); the four blackjack pairs whose translation bundles it (computed
`HandValue`, Accumulate Scores, Current Player Hit's Apply and LegalCustom)
are substituted, and each must fail. All four fail (e.g. computed:
`view.game ... HandValue: got 21, want 11` in
`conformance/blackjack/golden-82ef86840a83b9ed` v39). 6 s.

## Test results

`npm run test:unit` (run as its command line under Node 24.22.3): **1,442 of 1,442 pass**, 0 skipped, 26.5 s. `games/pairs.test.ts` adds 83 tests (about 12 s on its own; no test over 1 s except the extraction check, 0.5 s).

| Command | Outcome |
|---|---|
| `./scripts/go-local build ./...` | passes |
| `go vet` of `lib/transpile/pairs/...` and `pairs/testdata/gen`; `gofmt -l` | clean |
| `BOARDGAME_NODE=<node 24> go test -short -v ./boardgame-util/lib/transpile/pairs/` | ok, 9.8 s: TestStore (98 pairs), TestStale, TestIndexCurrent, TestHelperSignatures, TestStoreTypeChecks, TestFuzzFieldsMeasured, TestRuleAgainstPairs, TestBrokenRuleFailsProof pass; TestSubstitutionProof skipped by `-short` |
| `TestSubstitutionProof` without `-short` | passed, 124.5 s (last run before the final edits to the test kit; not re-run after it stopped importing the hand ports, at the coordinator's request to stop long runs) |
| `go test ./boardgame-util/lib/conformance/... ./boardgame-util/lib/transpile/ ./boardgame-util/lib/transpile/subset/` | all ok |
| `conformance export --check` | current |
| `npm run type-check` (both projects), `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | pass |
| store tsconfig (`TestStoreTypeChecks`) | passes (99 files) |
| `npm run test:layers` | 11 of 11 |

Without `BOARDGAME_NODE` and with the default Node 18 on PATH, the four
Node-driven Go tests skip and say why.

## Judgment calls

1. **The store's helpers have the translator's signature**, not a role
   convention (data.ts: helpers are called directly "with whatever signature
   their translation has"); `TestHelperSignatures` compares parameter and
   result types.
2. **A helper pair is never substituted alone** (the loader never resolves a
   helper, and each caller's `accepted.ts` carries its own copy of the hand
   port's function). Its translation is exercised inside every substituted
   caller's module; a rule about a helper is proven through those callers.
   Helper pairs get no mutant counts, for the same reason (`why` says so).
3. **Translations are bundled** with every in-package callee (callee
   first, unexported) so each substitution is one self-contained module; a
   substitution therefore tests the function and the helpers it reaches.
4. **The Go test drives Node.** `TestRuleAgainstPairs`,
   `TestSubstitutionProof`, `TestBrokenRuleFailsProof` and
   `TestStoreTypeChecks` need Node 22.18+ (`BOARDGAME_NODE`, else `node` on
   PATH); with the default Node 18 they skip with that message. The other
   Go tests need no Node. `TestSubstitutionProof` takes about 125 s (six
   Node processes, one per game) and skips under `-short`.
5. **`stock:` with the discovered slot field** (above), instead of the hand
   port's stock parts.
6. **The extraction runs at unit-test time** to prove the store current
   (about 0.5 s; it builds one TypeScript program over the six hand ports).

## Unfinished (exact state)

- Mutants for memory (17 pairs) and blackjack (22): not measured; null with
  reason. Everything else in `fuzz` is measured.
- `TestSubstitutionProof` was last run (all 86 passing) before the kit's
  hand-port imports were removed; the change only replaced the stand-ins the
  run never used (werewolf's `""` now gets an inline body with the base
  delegate's error text instead of the hand port's identical throw), and
  `pairs.test.ts` and `TestBrokenRuleFailsProof` were re-run after it.
- The per-game reports and `REPORT.md` (workstream 11's row in the proposal)
  were not built: the brief assigned them to another agent.

## For the hand ports' removal

The owner has ruled that the six hand ports are replaced by the generated
games. The store is their record: no `accepted.ts`, `index.ts`,
`pairs.testkit.test.ts`, `pairs-proof.mjs` or `pairs-mutants.mjs` imports a
hand port (checked: no store file imports `games/<game>.ts`; the kit's
stand-ins are inline). Only the extraction depends on them:
`pairs-extract.mjs` and the store generator's default mode read the hand
ports, as a **one-time step**. After the removal, `pairs.test.ts`'s
currency test skips with that message, and rebuilding the store means
running `gen -profile` / `gen -mutants` only (the extraction cannot run
again; `gen` with no flags would fail, which is the intended signal).

## What is unverified

- Helper pairs' `accepted.ts` behavior is proven only by type and signature,
  not by any replay: no substitution can reach it.
- 40 of 98 pairs have an unreached Go block; a translation that differed only
  there would pass. `ProveRule` does not count such a pair.
- The wrapped pairs' factories and the declarations copied beside every
  pair are not mutated (only the exported function's body).
- `ExperimentalWarning: stripTypeScriptTypes` is printed once by the unit
  test that runs the extraction.
- The checkers rows of `GO-BUGS.md` are marked fixed in Go at `2256c0d3` on
  the coordinator's word; that commit is not in this base and nothing was
  run against it.

## Where the proposal was wrong about the code

- "Each of their about 99 functions": the six manifests list 101; 98 have a
  hand-port counterpart.
- "A pair is stored ... `accepted.ts` the accepted TypeScript, verbatim": a
  hand port's function is a closure inside a factory, so verbatim text is
  not self-contained; 24 pairs need a wrapper and 1 has no hand code.
- A hand-port pair cannot always be swapped for a translated function with
  no change to the stock side: the hand port inlines the stock apply its
  move embeds, so the stock parts the existing loader tests use do not
  serve a translated body.

## Edits outside my files

None. (`pairs.testkit.test.ts`, `pairs-proof.mjs` and `pairs-mutants.mjs` are
new files on the TypeScript side of the proof.)
