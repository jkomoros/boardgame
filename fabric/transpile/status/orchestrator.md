# Workstream 12: orchestrator and agent harness

`fabric/TRANSPILE-PROPOSAL.md` workstream 12 ("The agent loop" in full, "The
ratchet and its metrics" from "After each rule" on), plus the per-game report.
Base `17168752`. One local commit; nothing pushed.

## What was built

| Path | What |
|---|---|
| `boardgame-util/cmd_port.go`, two lines in `cmd_base.go` | `boardgame-util port [--dry-run] [--no-game] [--check] [--all] [--adopt FnRef] PKG...`, `port check`, `port accept`, `port brief` |
| `boardgame-util/lib/port/game.go` | `Load(root, importPath, definition, commit) (*Game, error)`: the translator's package, its `Generate` output (keeping agent blocks), the definition, the recordings (`Documents`) |
| `lib/port/source.go` | The game's syntax and types for the brief: function source, declarations (game and framework, verbatim, struct tags and doc comments), what a function names, package-variable assignments with the defining statements of their right-hand sides, Go calls |
| `lib/port/apidoc.go` | Exported declarations, with doc comments, of an engine TypeScript module (signatures for functions, whole interfaces/types/constants) |
| `lib/port/examples.go` | `ExampleSource` (`Examples()`, `Name()`): `PairStore` (the proposal's `testdata/pairs/` layout) and `GeneratedExamples` (checked-in `bodies.gen.ts` blocks paired with their Go source, hash-checked); `DefaultExamples` picks the store when it holds a pair |
| `lib/port/brief.go` | `(*Game).WriteBrief(ref, ExampleSource) (*Brief, error)`, `(*Game).Signature(ref)`, `OutputPath`, `BriefPath` |
| `lib/port/check.go` | `(*Game).Check`, `Accept`, `Adopt`, `Replay`, `WithCandidate`, `FunctionBlocks`; `PairWriter` (`FilePairWriter`, `ErrNoPairStore`) |
| `lib/port/report.go` | `(*Game).BuildReport`, `MarshalReport`, `TableMarkdown` |
| `lib/port/node.go` | `Node.Run(mode, job, out)`: the Node half as a child process, waited for |
| `server/static/scripts/port-check.mjs` | The TypeScript half: `replay` and `check` (static, types, differential with the last call's arguments, body mutants) over a generated game served from memory |
| `server/static/src/engine/testing/sourceRules.ts` | The sandbox and purity rules as functions (factored from `sandbox.test.ts` and `rules/purity.test.ts`, which now call them) |
| `boardgame-util/lib/transpile/files.go` | `Generate` keeps agent blocks; `AgentHeader`, `Blocks`, `WithoutBlock`, `WithAgentBlock`, `AgentImportable`, `(*Package).Signature`; `Output.Agent` |
| `lib/transpile/agent_test.go` | Tests of the above |
| `lib/port/*_test.go`, `lib/port/testdata/*.ts` | Brief determinism and content, the TS scanner, end-to-end checks |
| `fabric/transpile/briefs/memory/…`, `briefs/checkers/…` | The two briefs, written by `port --all --no-game` |
| `fabric/transpile/reports/<game>.json` (7), `fabric/transpile/REPORT.md` | Written by `port --all --dry-run` |

### `port PKG`

For each game, in order: exports the definition in memory (`exportGamedef`,
the same function `gamedef export` uses) and records whether the checked-in
one is current; translates (`transpile.Load` + `Generate`, keeping agent
blocks); prints each refusal; writes a brief for each function refused for a
reason other than an engine gap, callee first; replays the generated game on
the TypeScript side (the checked-in one, else the translator's output served
from memory) against the game's fuzz corpus and conformance documents; builds
and writes the report. Without `--dry-run`/`--no-game` it also writes the
definition when stale and the generated game when it differs. `--dry-run`
writes only reports; `--no-game` writes briefs and reports; `--check` writes
nothing and fails on a stale report or table. `--all` is the seven games and
also writes `REPORT.md`. Block coverage needs a Go profile: `--profile FILE`,
else one is built with `fuzz.BuildBlockProfile` (about 30 s here).

`--adopt FnRef`: the three checks of "Turning a refusal reason into a rule":
the translator now translates the function; with its output in place
(`WithoutBlock` then `Generate`) every document replays; the function's Go
blocks are all reached. Only then is the file written.

### The agent header and `Generate`

`// @generated agent:<model> brief:<64 hex> go:<FnRef> sha256:<hash> first:<commit>`.
`Generate` keeps such a block verbatim while the function's source hash
equals the header's, whether or not the translator can now translate it (the
ratchet's rule: text changes only on adopt), puts it in the `bodies` switch
(not for helpers, as for translator blocks), and recovers its imports from
the identifiers it uses (`AgentImportable`: the emitter's facade names plus
`passVerdict`, `failT`, `unknownVerdict` and the predicate types). A block
whose hash no longer matches is dropped and the function is translated or
refused afresh. The lint's stale-header regex already reads agent headers.

### The brief

Every section of the proposal's template, composed from sources only:
FUNCTION (manifest entry); GO SOURCE (verbatim span gamedef hashes); WHY
(each refusal with its source line); TYPES IT TOUCHES (the game's types
verbatim; framework types verbatim unless the Types table or the API table
covers them; the PropDef of every game/player/move/component field the
function selects; package variables with every assignment and the defining
statements of what is assigned; game and framework constants with values;
the definition's predicate entries the function constructs; the game's moves
by name and Go type); FUNCTIONS IT CALLS (Go signature, TypeScript signature
from the block, the import line); THE TYPESCRIPT API (importable names,
API-table rows for each Go call or "not in the API table", stock parts of an
embedded stock move, RulesState/TimerHandle/MoveInstance/LegalVerdict
verbatim, for role predicate the predicate and verdict declarations, then
every exported declaration of `go/api.ts`, `num.ts`, `str.ts`, `slice.ts`,
`panic.ts` and `lib/*.ts` with doc comments); SEMANTICS (the proposal's
"Types" and "Semantics that differ" sections, verbatim); EXAMPLES (three
pairs, never from the same game, nearest by shared refusal reasons, then role,
then shared Go selector calls, then shorter source); OUTPUT (path
`server/static/src/engine/games/<game>/<ts name>.ts`, the signature, the
`rules/data.ts` convention row and detail for the role, the import rules);
HOW TO CHECK. Its SHA-256 is the header's `brief:`.

### `port check`

1. Static (Node): imports and exactly one exported function, the name, the
   signature (parameter and result types compared with the required line),
   imports only names `AgentImportable` maps to that module or named callees
   from `./bodies.gen.ts`, `sandboxProblems`, `platformUses`,
   `loadsDynamically`.
2. Types: the candidate, placed virtually at its OUTPUT path, and the
   generated game with it in place, type-checked with
   `tsconfig.sandbox.json`'s options (compiler API, in-memory host).
3. Preconditions: every other manifest function has a block; the game has
   recordings; the game loads.
4. Differential: `runFuzzDocument` on every fuzz and conformance document,
   stopping at the first failure; the divergence report plus the arguments
   the function was last called with (a recording copy of the bodies module,
   replayed to the failing version; `RulesState` is shown as `<RulesState>`).
5. Coverage: the function's blocks in the fuzz-corpus profile; unreached →
   `PASS (unproven)`.
6. Mutation: `bodyMutants` on the assembled `bodies.gen.ts`, each mutant a
   separate module graph (query-suffixed virtual URLs) over the shared
   engine, run until the first failing document.

`port accept FnRef --file PATH --producer MODEL` runs the check, writes the
three generated files with the agent block, and stores the pair through
`PairWriter` when `boardgame-util/lib/transpile/testdata/pairs` exists (it
does not on this base; accept says "No pair stored").

## Results (all run)

### Harness acceptance

| Run | Outcome |
|---|---|
| `port check 'pig.(*moveCountDie).Apply'` on a rewritten correct version (`lib/port/testdata/pig_moveCountDie_Apply.ts`) | static ok, types ok, loads, 11 of 11 documents, 8 of 8 Go blocks, 7 of 8 mutants killed (survivor: deleting `return null;`, equivalent since the loader treats `undefined` as `null`), **PASS** |
| same, `value === 1` → `value === 2` | **FAIL**: `pig/seed-0000`, version 1, field `state`, path `Players[3].Eliminated`, Go `true`, produced `false`, move Count Die (setup), bodies executed, last call `["<RulesState>",{"name":"Count Die","fields":{"TargetPlayerIndex":3},…}]`, repro `fuzz run --game pig --seed 0 --max-proposals 40 --max-versions 1` |
| `port check blackjack.handValue` (a helper called by a computed property) on an independently written version | 15 of 15 documents, 12 of 12 Go blocks, 30 of 31 mutants killed (survivor `aces--`, the same survivor workstream 5 found in the hand port), **PASS** |
| same, `value >= 11` → `value > 11` (the historical ace mutation) | **FAIL**: `blackjack/seed-0000`, version 34, field `state`, path `Players[0].Stood`, Go `false`, produced `true`, move Current Player Hit by 0, last call: the hand's three components |
| `port check 'memory.(*moveStartHideCardsTimer).Apply'` on `state.timer('HideCardsTimer').start(4000, nonNil(state.moveByName('Hide Cards')))` | static ok, types ok, **PRECONDITION FAILED: the generated game does not load: stockMove: MoveComponentToSlot needs configuration SlotField** |
| static failures (a module-scope statement, a second function, a forbidden import, a wrong signature, `Math.random`, a type error) | each FAIL with the problem named (`TestCheckStaticAndTypes`) |
| `port accept 'pig.(*moveCountDie).Apply' --producer claude-opus-5-5`, then `transpile --check`, `port pig --dry-run`, `node --test games/pig/pig.gen.test.ts`, then `port pig --adopt` | accepted with the agent header; `transpile --check` still current (block kept); report agent 1, agentNowMechanical 1, mechanical 9; pig's generated suite 14 of 14; adopt passed its three checks and restored `bodies.gen.ts` byte for byte (`git diff` empty) |

**Memory is pending the stock-registry gap fix** (another agent's): the
check cannot get past loading memory's generated game. `TestCheckMemory`
asserts the static and type steps pass now and skips at that precondition;
after the fix it asserts PASS for the correct version and a first difference
for two mutants (the wrong move, `'Reveal Card'`; a 5000 ms duration). The
coordinator's rerun after the merge turns it into the real acceptance.

### Briefs

`TestBriefDeterministic`: memory's and checkers' briefs, each from two fresh
loads, are byte-identical and carry all ten sections; no example is from the
game itself. `TestMemoryBriefCarriesTheValue` checks the value chain is in the
brief: `var hideCardMoveName string`, its assignment in `ConfigureMoves`,
`hideCardConfig := auto.MustConfig(new(moveHideCards), …)`, the move list
(`"Hide Cards": memory.moveHideCards`), `game.HideCardsTimer`'s PropDef,
`const hideCardsDuration = 4 * time.Second` = 4000000000, the API rows for
`MoveByName` and `Timer.Start`, `TimerHandle.start(durationMs, move)`, the
signature, the apply convention row, the check command.

**The checkers dry exercise.** I wrote `checkers_gameDelegate_ConfigurePredicateConstructors`
from its brief alone (`lib/port/testdata/`). It passes static and types, and
with it in place checkers' generated game loads (it does not without it:
"no body for predicate"). The check stops at "checkers has no recordings",
so nothing about its behavior is verified. Found missing while writing it,
and fixed in the writer: the values of framework constants it reads
(`boardgame.LegalCostTrivial = 0`, `LegalFacetValues = 0`, `TypeEnum = 5`),
and which definition predicate the function constructs (Go returns a list;
the translation is one constructor). Still missing:

1. No mapping from a framework constant's Go value to its TypeScript value:
   Go `LegalFacetValues` is `0` and `TypeEnum` is `5`; TypeScript's are the
   strings `'values'` and `'TypeEnum'`. The brief shows both types
   (`LegalFacet`, `PropertyType`) but an agent must infer the pairing by name.
2. The API table has no rows for package `legal`: `legal.FailT`,
   `PassVerdict`, `UnknownVerdict`, `(LegalContext).ResolvePath`,
   `legal.PropPath`. The brief says "not in the API table"; their
   counterparts are only in the predicate declarations it quotes.
3. `enum.ImmutableVal` is labelled "its methods are the API table's rows"
   (`enumKey(state, owner, prop)`), which is wrong inside a predicate: there
   the value is `resolvePath`'s `{ enum, value, string }`.
4. `EmittedBindings: map[string][]string{template: nil}`: nothing says
   whether a nil slice inside it is `[]` or absent in `GamePredicate`
   (its doc covers only an absent map).
5. A failing `val.(enum.ImmutableVal)` assertion: the brief does not say what
   a non-enum value looks like in TypeScript, only "by type".

### The per-game report (`port --all --dry-run`)

| Game | Units | Data | Mechanical | Agent | Engine gap |
|---|---|---|---|---|---|
| memory | 20 | 3 (15.0%) | 16 (80.0%) | 1 (5.0%) | 0 (0.0%) |
| blackjack | 39 | 16 (41.0%) | 23 (59.0%) | 0 (0.0%) | 0 (0.0%) |
| pig | 13 | 3 (23.1%) | 10 (76.9%) | 0 (0.0%) | 0 (0.0%) |
| tictactoe | 14 | 3 (21.4%) | 11 (78.6%) | 0 (0.0%) | 0 (0.0%) |
| werewolf | 20 | 4 (20.0%) | 16 (80.0%) | 0 (0.0%) | 0 (0.0%) |
| debuganimations | 25 | 1 (4.0%) | 24 (96.0%) | 0 (0.0%) | 0 (0.0%) |
| checkers | 24 | 4 (16.7%) | 19 (79.2%) | 1 (4.2%) | 0 (0.0%) |
| **Total** | **155** | **34 (21.9%)** | **119 (76.8%)** | **2 (1.3%)** | **0 (0.0%)** |

| Game | movesData | movesWithCode | functions | data | mechanical | translatable | agent | agentNowMechanical | engineGap | boundForAgent | notPorted | stale | TS replay |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| memory | 3 | 4 | 17 | 14 | null | 16 | 0 | 0 | 0 | 1 | 11 | null | does not load (SlotField) |
| blackjack | 16 | 7 | 23 | 13 | 23 | 23 | 0 | 0 | 0 | 0 | 1 | 0 | 15 of 15 |
| pig | 3 | 3 | 10 | 12 | 10 | 10 | 0 | 0 | 0 | 0 | 1 | 0 | 11 of 11 |
| tictactoe | 3 | 1 | 11 | 14 | null | 11 | 0 | 0 | 0 | 0 | 5 | null | does not load (SlotField) |
| werewolf | 4 | 7 | 16 | 11 | null | 16 | 0 | 0 | 0 | 0 | 1 | null | does not load (no body for distribute) |
| debuganimations | 1 | 12 | 24 | 23 | 24 | 24 | 0 | 0 | 0 | 0 | 1 | 0 | 9 of 9 |
| checkers | 4 | 3 | 20 | 14 | null | 19 | 0 | 0 | 0 | 1 | 0 | null | does not load (no body for predicate) |

Go blocks reached over the fuzz corpus, manifest functions only: memory 79
of 91, blackjack 93 of 106, pig 36 of 38, tictactoe 67 of 73, werewolf 143 of
151, debuganimations 135 of 154; checkers has no corpus (coverage null).
Surviving mutants are null for every game, with the reason in `nulls`.
`REPORT.md` says how each bucket is counted.

`port --all --dry-run --check` passes, with the profile given and with a
freshly built one (48 s); `git status` after `port --all --dry-run` shows no
change under `server/static/src/engine/games/` or `conformance/`.

### Checks before reporting

| Command | Outcome |
|---|---|
| `./scripts/go-local build ./...` | passes |
| `BOARDGAME_NODE=<node 24> ./scripts/go-local test -count=1 ./boardgame-util/...` | 24 packages ok, none failing (lib/fuzz 72 s, lib/port 37 s, lib/transpile 22 s, lib/transpile/subset ok) |
| the same `lib/port` tests with the default Node 18 | the four Node tests skip with a message; the rest pass |
| `go vet` of `./boardgame-util`, `lib/port`, `lib/transpile`; `gofmt -l` of my files | clean |
| `conformance export --check` | "The conformance corpus is current" |
| `transpile --check ./examples/pig ./examples/blackjack ./examples/debuganimations` | "The generated games are current" |
| `npm run type-check` | passes |
| `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | passes |
| `npm run test:unit` | 1,358 of 1,358 (a first run while the Go tests ran in parallel failed two 10 s time budgets, `rules/data.test.ts` blackjack fixture and `stockRegistry.test.ts` blackjack; rerun alone: all pass) |
| `npm run test:layers` | 11 of 11 |

## Judgment calls

1. **The headline's units are data-only move configs plus manifest
   functions.** A move config with game code is not a unit: its
   configuration is data and its code is its functions; counting both would
   count the code twice. Data-only configs whose stock type the engine lacks
   are engine gap. Functions refused for any other reason count as agent
   (bound for one) before an agent has written them.
2. **functions.mechanical counts checked-in translator blocks**, so it is
   null for the four games with no generated game checked in;
   `functions.translatable` is the dry-run count. `functions.agent` is 0
   there (no agent block exists), not null.
3. **functions.data** is gamedef's `NotCounted` (the functions the exporter
   evaluates on the live manager). No example game has a game-written
   counter; every progression counter is a stock one.
4. **Surviving mutants are null** in every report: a mutant run over every
   function of a game is minutes long, and `--check` would have to repeat it.
   `port check` runs them per function.
5. **Briefs only for functions bound for an agent**, not for engine-gap ones
   (no agent can close a gap).
6. **Examples never come from the target's game**, since the agent may not
   see the rest of its game.
7. **The candidate is checked in place** of the translator's block (or
   inserted where none exists), as `accept` will write it, so a helper's
   in-module callers call it; the producer is `candidate` until accepted.
8. **Coverage is over the fuzz corpus only**: the profile comes from
   regenerating it (`fuzz corpus --check`); the conformance documents are
   replayed but not regenerated in Go.
9. **Zero recordings is a precondition failure**, never a pass.
10. Commit attribution: COORDINATION.md asks for a `Claude Fable 5.2` line; I
    am Claude Opus 5.5 and used that name in the trailer.

## Where the proposal was wrong about the code

- OUTPUT's "imports only from `../../go/*.ts` and `../../lib/*.ts`" cannot
  hold: every signature names `RulesState` (`rules/state.ts`) and
  `MoveInstance` (`rules/types.ts`), a legal body returns `LegalVerdict`
  (`legal/types.ts`), a predicate needs `rules/types.ts` and the verdict
  functions, and a function with in-package callees must import them from
  `./bodies.gen.ts` (item 9). Conversely `lib/*.ts` names cannot be imported
  by a block of `bodies.gen.ts`: the translator rebuilds that file's imports
  by name, and `lib/` names collide with facade names (`neighborsByKey`,
  `rangeToValue`, `valueToRange`, `rangeValue`); `go/purity.test.ts` also
  keeps the facade from importing `lib/`. The brief lists `lib/*.ts` for
  reference and allows the facade's names.
- The emitter has no signature for roles `predicate` and `counter`; the
  brief takes them from `rules/data.ts`.
- `brief:<sha256>` is written as 64 hex digits, without the `sha256:` prefix
  the source hash carries.
- Memory's refused function cannot be checked on this base: its generated
  game does not load (stock registry, `SlotField`).

## Unverified

- Memory's PASS and first difference (pending the gap fix; the test is in
  place).
- The predicate translation's behavior (checkers has no recordings).
- `PairStore` and `FilePairWriter` against a real pair store (the directory
  does not exist on this base); `Adopt`'s refusal paths (exercised: the
  success path on pig).
- The `rejection` branch of the last-call capture (no mutant tried made a
  rejection the first difference).

## Edits outside my files

- `server/static/src/engine/sandbox.test.ts`, `server/static/src/engine/rules/purity.test.ts`:
  their rules moved, unchanged, into the new `server/static/src/engine/testing/sourceRules.ts`,
  which they call (the brief: "factor the smallest shared helper"); test
  names and counts unchanged.
- `boardgame-util/lib/transpile/agent_test.go` (new test file for the
  `files.go` change).
