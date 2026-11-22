# Workstream 10: the translator

Base `e7c02e2e`. What `fabric/TRANSPILE-PROPOSAL.md` workstream 10 ("The
translator") asked for, as built and as run. One local commit; nothing pushed.

## What was built

| Path | What |
|---|---|
| `boardgame-util/lib/transpile/translate.go` | `Load(pattern, dir) (*Package, error)`: the static report (`gamedef.LoadStatic`) plus the package's syntax and types; `(*Package).Translate() []Function` (manifest order, callee first); `TryRefused(ref)`; `Package.AssumeGapsClosed`; `Version` (the `transpile@N` producer, 1) |
| `boardgame-util/lib/transpile/emit.go` | The emitter: role signatures (the body calling convention), statements, assignment, names, imports |
| `boardgame-util/lib/transpile/expr.go` | Expressions: constant folding, the types table, the API table, the semantics table |
| `boardgame-util/lib/transpile/kinds.go` | How each Go type is represented in TypeScript (`kind`), `tsType`, zero values |
| `boardgame-util/lib/transpile/framework.go` | The package-variable rule: framework enums and grid graphs rebuilt from their constructors' constants; never-written constant-initialized variables folded |
| `boardgame-util/lib/transpile/mangle.go` | `Mangle(FnRef)`, `Unmangle(name)` |
| `boardgame-util/lib/transpile/files.go` | `(*Package).Generate(definitionJSON, existing, commit) (*Output, error)`: the three files; header parsing (`first:` kept); `ModuleRoot`, `GamesDir`, `DefinitionsDir`, `FileNames` |
| `boardgame-util/lib/transpile/transpile_test.go`, `testdata/emitgame/` (+ codegen'd `auto_reader.go`), `testdata/emitgame.bodies.golden` | Go tests (below) |
| `boardgame-util/cmd_transpile.go`, two lines in `cmd_base.go` | `boardgame-util transpile [--check] [--dry-run] [--stats] [--out DIR] [--commit SHA] [--assume-gaps-closed] PKG...` |
| `server/static/src/engine/games/pig/{definition.gen.ts,bodies.gen.ts,index.ts}`, `pig.gen.test.ts` | Pig, generated, and its test |
| `server/static/src/engine/games/blackjack/…`, `games/debuganimations/…` (same four files each) | Two more generated games, checked in because they pass their whole corpus and fuzz documents with no work (the brief allows it) |
| `server/static/src/engine/games/generated.testkit.test.ts` | `generatedGameSuite(name, make, {corpus, fuzz})`, the shared test each `<game>.gen.test.ts` calls (no test of its own) |

### Unit and output

One exported function per manifest function with verdict `subset`, in
callee-first order (`StaticReport.Order`), each under its header:

```
// @generated transpile@1 go:pig.(*moveRollDice).Apply sha256:<source hash> first:<commit>
export function pig_moveRollDice_Apply(state: RulesState, m: MoveInstance): string | null { … }
```

then `export function bodies(ref: string): ((...args: never[]) => unknown) | undefined`,
a `switch` over every non-helper FnRef. Refused functions are listed (FnRef,
reason, detail, position) in the file's leading comment and never emitted.

**Mangling** (`mangle.go`, reversible, tested both ways): parts joined by
one `_`, an underscore inside an identifier doubled: `pkg.fn` → `pkg_fn`,
`pkg.(*T).M` → `pkg_T_M`, `pkg.(T).M` → `pkg_T$v_M`, a trailing `.funcN` →
`$N`. An identifier that starts or ends with `_` is refused (ambiguous next
to a separator).

**Headers.** `sha256:` is always the function's current `SourceHash` (the
declarative lint's stale check, BGLINT0104, reads it). `first:` is the
commit the existing file's header names when the function's code is
byte-identical to the block on disk, else the commit given (`--commit`,
default `git rev-parse --short=8 HEAD`). So `transpile` twice, or at a
later commit with unchanged bodies, writes identical files.

**definition.gen.ts** is `export function definition(): GameData { return
{…}; }`, the JSON of `conformance/definitions/<game>.json` written as a
TypeScript literal in the document's key order (no JSON import).
**index.ts** is `export function <game>Generated(): GameSpec { return
gameFromData(definition(), bodies); }`.

### Signatures by role

Exactly the convention at the top of `rules/data.ts`. Every function takes
`state: RulesState` first; Go parameters of type `State`/`ImmutableState`
are that one state; the delegate and `*gameState` receivers have no value.

| Role | TypeScript |
|---|---|
| apply, setUpBegin, setUpFinish, turnDone, resetForTurnStart/End | `(state, …) => string \| null` (the loader throws a returned string where the convention says THROWS) |
| legal | `(state, m, proposer, base) => string \| null`; `m.<Stock>.Legal(state, proposer)` is `base()` |
| legalCustom | `(state, m, proposer) => string \| LegalVerdict \| null`; `return legal.Errorf(key, nil)` is `legalError(key)` |
| defaults | `(state, m) => void`; a leading `m.<Stock>.DefaultsForState(state)` is dropped |
| distribute | `(state, deck, index) => StackRef`; Go's `c boardgame.Component` is `const c: DynamicRef = { deck, index }` when used; `return s, nil` returns, `return nil, err` throws |
| checkGameFinished | `(state) => [boolean, number[] \| null]` |
| computed | `(state)` or `(state, player)` |
| component (`(*token).Legal`) | `(state, t: DynamicRef, legalType: number \| null) => string \| null` (DeckSpec.componentLegal's shape; the body may not read legalType). No loader slot consumes it yet |
| helper and every other role | `(state, [receiver], params…)` with Go's results (none → `void`, one → its type, several → a tuple) |

### Types, statements, expressions

Types: `int`/`PlayerIndex`/`EnumKey`/`float64` number; `error` `string | null`;
`*playerState` and player sub-states the index; `*gameState` no value; a
move the `MoveInstance` (`m.X` is `m.fields.X`); component values and dynamic
values structs a `DynamicRef` handle (`values(state, h).F`,
`dynamic(state, h).F`); stacks `StackRef` (`StackRef | null` for a local
declared without a value or ever assigned nil, read through `nonNil`);
`ComponentInstance` `Comp | null`; timers `state.timer('P')`; enum
properties have no value, only an owner (`'game'`, a player,
`{ valuesOf }`, `{ dynamicOf }`, `{ fieldsOf: m }`) for `enumKey`,
`setEnumKey`, `enumEquals`, `rangeValue`; slices `T[] | null`; maps
`Map<K, V>`; multiple results tuples; framework enum variables their name
(`'spaces'`), grid graph variables a `GraphRef` literal rebuilt from
`MustNewEnumGridConnectedness`'s arguments.

`concreteStates` (and every erased function) is not emitted: `game, players
:= concreteStates(state)` becomes `let players: number[] | null =
playerIndexes(state);` and `game.X` reads go through `state.g`. Locals get
unique names per function (Go's shadowing never reaches TypeScript); a name
that is a TypeScript keyword, `state`/`base`/`deck`/`index` or any facade
name is suffixed (`player` → `player_`, `first` → `first_`).

Statements: blocks, `:=`/`=`/tuple and compound assignment, `++`/`--`,
`var`, `if` with an initializer (a block), the three `for` forms (`let` in
the head is Go 1.22's per-iteration variable), `range` over a slice, an
integer, `Components()` or `PlayerStates()` (a hidden index over a snapshot,
so assigning the loop variable never changes the iteration), `switch` with
and without a tag (as `switch (true)`), several values per case, `default`,
labels, `break`/`continue`, `return`. Expressions: every constant folded by
go/types (`iota`, `dim*dim`, `time.Duration` to whole milliseconds,
`enum.IllegalValue`); `/` `%` `*` on integers through `idiv`, `imod`,
`imul`; string `<` through `strCompare`, `len(s)` through `utf8Len`; slice
indexing through `at`/`setAt`; `append` (self form) through `appendTo`;
`make`, `copy`, `delete`, `min`, `max`; comma-ok map reads; `nil`
comparisons, including `c.Values() == nil` (`valuesAreNil`); conversions
among the integer types; type assertions to the game's own types (erased).

## Facade additions (edits outside my files)

All additive, in `server/static/src/engine/go/`:

| Function | Why | Verified by |
|---|---|---|
| `panic.ts` `nonNil(x)` | Go's nil-dereference where Go dereferences a possibly-nil stack, move, error | pig, blackjack, debuganimations corpora (err.Error(), timer moves) |
| `api.ts` `playerIndexes(state)` | concreteStates' `players`, `PlayerStates()` ranged over | the three corpora |
| `api.ts` `valuesOf`, `dynamicOf` | a held `c.Values().(*T)` / `c.DynamicValues().(*T)` handle, with Go's checks at the call | blackjack (cards), pig (die) |
| `api.ts` `diceRoll(state, c)` | `(*dice.DynamicValue).Roll(state.Rand())` (inlined: go/purity.test.ts forbids importing `lib/`) | pig's 11 documents (every roll) |
| `api.ts` `valuesAreNil(state, c)` | `c.Values() == nil` | type check only (checkers, gaps closed) |
| `api.ts` `Owner` `{ fieldsOf: MoveInstance }` | an enum field of a move (`m.TargetIndex.Value()`) | type check only (checkers) |
| `api.ts` `values`, `Owner.valuesOf`, `containingStack` accept a `DynamicRef` | a distribute's `Component`, a values method's own handle | blackjack's distribute (corpus); containingStack: type check only |

None changes an existing function's behavior for its existing argument
types. No Go vector was added for the three "type check only" rows.

## Test results

| Command | Outcome |
|---|---|
| `./scripts/go-local build ./...` | passes |
| `./scripts/go-local test ./boardgame-util/lib/transpile/` | ok: TestMangleRoundTrip, TestGolden, TestGoldenRules (42 rule strings), TestDeterministicAndFirstPreserved, TestHeadersMatchLint, TestExamplesAgreeWithLint, TestCheckedInCurrent |
| `./scripts/go-local test ./boardgame-util/lib/transpile/subset` | FAIL: TestEngineGapsMatchStockTS, pre-existing on the base (another agent's) |
| `./scripts/go-local test` of `lib/conformance/...`, `lib/gamedef/...`, `lib/lint/...`, `./boardgame-util` | all ok (the lint's stale-header check now reads the three checked-in bodies files and passes) |
| `./scripts/go-local run ./boardgame-util conformance export --check` | "The conformance corpus is current" |
| `go vet` of `lib/transpile/...` and `./boardgame-util`, `gofmt -l` of my files | clean |
| `boardgame-util transpile --check ./examples/pig ./examples/blackjack ./examples/debuganimations` | "The generated games are current" |
| two runs of `transpile --out <scratch>` over `./examples/...` | byte-identical (`diff -r`) |
| `npm run type-check` | passes |
| `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | passes (with the three checked-in games; also with all seven generated, and with checkers generated `--assume-gaps-closed`, temporarily in the tree) |
| `npm run test:unit` | 1,298 tests: 1,295 pass, 3 fail. The 3 are the known-red loader tests on this base (`rules/data.test.ts`: memory, debuganimations, werewolf "exported definition"). Base was 1,253; +45 (pig 14, blackjack 18, debuganimations 12, the testkit file 1) |
| `npm run test:layers` | 11 of 11 |

### Pig, end to end

`transpile ./examples/pig` translates all 10 manifest functions, refuses
none. `games/pig/pig.gen.test.ts` (through `generatedGameSuite`):

| documents | versions | viewer-versions | rejections | fields |
|---|---|---|---|---|
| 3 corpus + 8 fuzz = 11 | 696 | 4,023 | 25 (propose: "it's not your turn", stale version) | header, state, move, status, view.game, view.move, view.legality, view.choices: 0 mismatches |

Per document: golden-67672ddc7e8a71ad 228 versions × 4 viewers,
golden-81137f3f9f5d40ba 28 × 4, scenario-short 27 × 4, fuzz seed-0000 51 × 8,
0001 52 × 8, 0002 51 × 5, 0003 52 × 5, 0004 51 × 8, 0005 52 × 8, 0006 52 × 7,
0007 52 × 7. No difference from Go was found, so nothing was fixed in the
loader, facade, registry or definition for pig.

Is the test honest? Two hand mutations of the generated bodies, one at a
time, each restored by regenerating: `value === 1` → `value === 2` in Count
Die fails 12 of 14 tests; TurnDone's second error returning null fails 12 of
14. The rejection replay asserts the error text, no versions, an unchanged
state and that every rejection was reached.

The rejections are replayed locally (`generated.testkit.test.ts`);
`testing/fuzz.ts` did not exist on this base, and the coordinator will
switch it to the shared helper.

### Every game

`transpile --dry-run ./examples/...` and `--out <scratch>` (not checked in
for the four games below), then the sandbox type check with the output
copied into `games/<game>/` and removed afterwards:

| Game | Translated | Refused (reason, detail) | Type-checks | Corpus |
|---|---|---|---|---|
| pig | 10 | 0 | yes | 11 of 11 documents pass (checked in) |
| blackjack | 23 | 0 | yes | 7 corpus + 8 fuzz pass: 946 versions, 5,888 viewer-versions, 9 rejections (checked in) |
| debuganimations | 24 | 0 | yes | 1 corpus + 8 fuzz pass: 184 versions, 736 viewer-versions, 22 rejections (checked in) |
| memory | 16 | 1: `(*moveStartHideCardsTimer).Apply`, package-var-read `memory.hideCardMoveName` | yes | does not load: `stockMove: MoveComponentToSlot needs configuration SlotField` (the definition's Reveal Card config has `SourceSlotField`) |
| tictactoe | 11 | 0 | yes | does not load: the same `SlotField` error (Place Token's config has no slot field key) |
| werewolf | 16 | 0 | yes | does not load: `gameFromData: no body for distribute` (the definition's `SetUp.Distribute` is `""`; werewolf has no deck) |
| checkers | 11 | 9: predicate-constructor `checkers.spaceIsBlack` (1); engine-gap `moves.DefaultComponent`, `behaviors.ComponentColor`, `behaviors.PlayerColor`, `componentAtKey`, `swapComponentsByKey`, `neighborsByKey`, `rangeValue`, `rangeToValue`, `valueToRange` (and callee-refused on those) (8) | yes | does not load: predicates are not supported in v1 |
| **total** | **111** | **10** | | |

The accepted and refused sets equal the lint's verdicts for all seven
(TestExamplesAgreeWithLint: 121 manifest functions, 111 `subset`, 10
`refused`; the emitter refuses none the checker accepts). The load failures
are in the exporter/stock-registry/loader contract, not in the translated
bodies; they are other workstreams' (not chased, as the brief says).
Debuganimations' generated game passes although its hand-port loader test
is red on this base, which suggests that red test's cause is in the hand
port's adaptation to the convention, not in the definition.

`--stats` also reports what the emitter alone makes of each checker-refused
function. With `--assume-gaps-closed` checkers translates 19 of 20 (all but
the predicate constructor) and the output type-checks, so when the gap table
(`subset/gaps.go`, being shrunk by another agent) loses those entries the
agreement test should keep holding. That was checked by type checking only.

## Judgment calls

1. **The translator's verdict is the checker's.** A function the manifest
   refuses is refused with the checker's refusals; the emitter runs only on
   `subset` functions, and an emitter failure there is reported as its own
   refusal (`EmitterRefused`, detail prefixed `emitter:`) so the agreement
   test names it. Equality of the sets is therefore a test of the emitter
   against the checker, not of two independent checkers.
2. **`sha256:` follows the source; `first:` follows the body.** The
   proposal's "a block is rewritten only when its body would change" taken
   literally would leave a stale hash under an unchanged body, which the
   lint flags. `first:` is the commit at generation time (HEAD), since the
   commit that adds the block does not exist yet.
3. **Error results are returned, not thrown,** in every role whose Go result
   is `error`: the loader throws a returned string where the convention says
   THROWS, so one shape serves helpers and roles. Distribute is the
   exception (its convention returns the stack).
4. **Locals are `let` with a declared type**, never `const`; names are made
   unique per function, so Go shadowing and TDZ never arise.
5. **`variant[k]`** (BeginSetUp's parameter) is `variantValue(state, k)`, the
   state's own variant, which is the same record during set-up.
6. **Values and dynamic structs are handles** (`DynamicRef`): a field read
   is `values(state, h).F` each time, matching Go's pointer to the live
   struct; a write to static values is refused.
7. **`players[i]`** is `at(players, i)` over `playerIndexes(state)`, so an
   out-of-range index throws Go's own text.
8. **Generated blackjack and debuganimations are checked in** beside pig:
   they pass everything with no change. Memory, tictactoe, werewolf and
   checkers are not, since their definitions do not load.
9. **`--assume-gaps-closed`** exists for measurement and refuses to write
   into the engine unless `--out` is given.

## What is unverified

- The facade additions marked "type check only" above, and every emitter
  rule no checked-in game reaches at run time (the golden fixture is
  textual): maps, labels, tagless switches, range over int, named results,
  `legalError`, the component role, framework variables, range enums and
  graphs. Each has golden text and passes the sandbox type check.
- Integer overflow: `+`/`-` are native (the proposal leaves the check to
  the save); only `*` goes through `imul`.
- A Go `switch` case expression with side effects, and Go's evaluation order
  in tuple assignments with index targets (temporaries are used, but no
  example game exercises an index target there).

## Where the proposal was wrong about the code

- `Components()` "a handle whose component has since moved fails": it does
  not (the facade report found this; the translator relies on identity).
- The proposal's role list has no slot for a component's `Legal` (role
  `component`); the loader has none either.
- A `boardgame.Component` (distribute's parameter) is not a
  `ComponentInstance`: it has no stack, so the facade needed `DynamicRef`
  variants of `values` and `containingStack`.
- The facade may not import `lib/` (`go/purity.test.ts`), so the dice row of
  the API table cannot be `lib/dice.ts` directly from facade code.
- Werewolf's definition carries `SetUp.Distribute: ""`, which the loader
  treats as a FnRef.

## Edits outside my files

- `server/static/src/engine/go/api.ts`, `go/panic.ts`: the additions above.
- `server/static/src/engine/games/blackjack/`, `games/debuganimations/`,
  `games/generated.testkit.test.ts`: new files beside the owned `games/pig/`.
