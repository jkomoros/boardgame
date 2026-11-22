# Workstream 2: exporter static half, subset checker, declarativeness lint

Base `d2cb55e4`. What `fabric/TRANSPILE-PROPOSAL.md` workstream 2 asked for,
as built and as run.

## What was built

| Path | What |
|---|---|
| `boardgame-util/lib/transpile/subset/check.go` | The subset checker. `New(Config)`, `(*Checker).Check(fn, node)`, `CheckKind(fn, node, kind, detail)`, `Nondeterminism(node)`, `QualifiedName(*types.Func)` |
| `boardgame-util/lib/transpile/subset/api.go` | The API table as data: `Table []Entry{Go, Facade, Kind, Source}`, `Lookup`, `GoNames`, `StockParts`, `FrameworkConstructors`, `FrameworkFuncValues` |
| `boardgame-util/lib/transpile/subset/gaps.go` | `EngineGaps`, the one gap table (facade names, `moves.<Type>`, `behaviors.<Type>`), `IsGap` |
| `boardgame-util/lib/transpile/subset/subset_test.go`, `testdata/fixture/fixture.go` | One fixture function per rule, each with a `want:` comment |
| `boardgame-util/lib/gamedef/static.go` | `Static(importPath)` (frozen signature, now implemented), `LoadStatic(pattern, dir) (*StaticReport, error)`, `CalleeFirst`, `StaticReport`, `MoveConfigStatic`, `FunctionRef`, `Finding`, bucket and added role constants |
| `boardgame-util/lib/gamedef/static_test.go`, `testdata/staticgame/` | A fixture game (with its generated `auto_reader.go`) covering every rule of the static half |
| `boardgame-util/lib/lint/declarative.go` | `Declarative(inputs, DeclarativeOptions)`, `DeclarativeFor(*StaticReport)`, the report types, `WriteDeclarativeText`, `MarshalDeclarativePackage`, `ParseGeneratedHeaders`, `StaleGenerated`, codes `BGLINT0101`..`0104`, `ExampleGames` |
| `boardgame-util/lib/lint/declarative_test.go`, `testdata/declarative/<game>.json`, `testdata/gen/declarative/main.go` | Baseline, outside-function, golden, max-refused and stale-header tests; goldens and their generator |
| `boardgame-util/cmd_lint.go` | `--declarative`, `--max-refused N`; the existing `--json` applies to it |

`hash.go` is unchanged (the freeze implemented it); `static_test.go` tests it.

## Exported interfaces

- `gamedef.Static(importPath string) (*StaticInfo, error)` fills every
  `StaticInfo` field: `Moves` (every configured type and every move type the
  game declares, `StockType` by a breadth-first walk of embedded fields
  confirmed with `types.LookupFieldOrMethod`, `Bodies` for `Apply`, `Legal`,
  `LegalCustom`, `DefaultsForState` whose declaring package is the game's,
  including methods promoted from an embedded game type), `Traits`,
  `Computed`, `Predicates`, `Counters` (empty, as v1 specifies), `SetUp`,
  `End` (FnRefs only), `CurrentPlayerIndex`, `Functions` (sorted by `Func`).
- `gamedef.LoadStatic(pattern, dir) (*StaticReport, error)`: the same plus
  `MoveConfigs`, `Buckets` (`subset`, `engineGap`, `outside`), `Order`
  (callee first), `NotCounted`, `NotPorted`, `Erased`, `Unreached`,
  `Findings`. Workstream 1 needs only `Static`.
- `subset.New(Config{Fset, Pkg, Info, Files, Dir, Erased}).Check(fnRef, *ast.FuncDecl | *ast.FuncLit) []refusal.Refusal`
  (sorted, deduplicated, never nil). `subset.Table`, `subset.EngineGaps` are
  the data the translator (workstream 10) should reuse.
- `lint.Declarative(inputs, DeclarativeOptions{BasePath, MaxRefused}) DeclarativeReport`.

## Measured baseline, beside the proposal's

`./scripts/go-local run ./boardgame-util lint --declarative --json ./examples/...`
(also `TestDeclarativeReproducesBaseline`):

| Game | Move configs | Data only | With game code | Declared specs | Functions | In subset | Outside | Engine feature | Not ported |
|---|---|---|---|---|---|---|---|---|---|
| memory | 7 (7) | 3 (3) | 4 (4) | 9 (9) | **17** (16) | **16** (15) | 1 (1) | 0 (0) | 11 (11) |
| blackjack | 23 (23) | 16 (16) | 7 (7) | 6 (6) | **23** (22) | **23** (22) | 0 (0) | 0 (0) | 1 (1) |
| pig | 6 (6) | 3 (3) | 3 (3) | 4 (4) | 10 (10) | 10 (10) | 0 (0) | 0 (0) | 1 (1) |
| tictactoe | 4 (4) | 3 (3) | 1 (1) | 0 (0) | 11 (11) | 11 (11) | 0 (0) | 0 (0) | **5** (6) |
| werewolf | 11 (11) | 4 (4) | 7 (7) | 0 (0) | 16 (16) | 16 (16) | 0 (0) | 0 (0) | 1 (1) |
| debuganimations | 13 (13) | 1 (1) | 12 (12) | 7 (7) | 24 (24) | 24 (24) | 0 (0) | 0 (0) | 1 (1) |
| checkers | 7 (7) | 4 (4) | 3 (3) | 5 (5) | 20 (20) | 19 (19) | 1 (1) | **8** (about 11) | 0 (0) |
| **Total** | 71 (71) | 34 (34) | 37 (37) | 31 (31) | **121** (119) | **119** (117) | 2 (2) | **8** (about 11) | **20** (21) |

Proposal values in parentheses; differences in bold. The first four columns
match exactly. The two functions outside the subset are the proposal's:
`memory.(*moveStartHideCardsTimer).Apply` (`package-var-read`
`memory.hideCardMoveName`) and
`checkers.(*gameDelegate).ConfigurePredicateConstructors`
(`predicate-constructor` `checkers.spaceIsBlack`).

### Each difference, and why the tool is right

1. **memory functions 17, in subset 16.** The computed property is two
   function bodies: the closure passed to `GlobalComputedBool`
   (`memory.(*gameDelegate).ConfigureComputedProperties.func1`) and the
   method it calls (`(*gameState).CurrentPlayerHasCardsToReveal`). The
   proposal makes a computed property's closure a translation unit of its
   own ("one function literal in a slot the exporter knows"), so both are
   counted; the hand count counted the property once.
2. **blackjack functions 23, in subset 23.** The same: `func1` and
   `(*playerState).HandValue` (which calls `handValue`).
3. **tictactoe not ported 5.** Not ported is `Diagram` and agent code:
   `Diagram` plus the four `Agent` methods. The sixth function the hand count
   probably included, `(*gameState).tokenValue`, is called by nothing; the
   report lists it under `unreached`.
4. **checkers engine features 8.** The proposal's 11 was an estimate ("the
   six move methods, the five token methods"). Measured:
   `(*moveCrownToken).Apply` (reads `ComponentIndex` of `moves.DefaultComponent`,
   not ported), `(*moveMoveToken).Apply` and `LegalCustom`
   (`swapComponentsByKey`, `componentAtKey`, `rangeValue`, `rangeToValue`,
   both color behaviors, and callees), `(*movePlaceToken).DefaultsForState`
   (`behaviors.ComponentColor`), and `(*token).Legal`, `AllNextSpaces`,
   `LegalCaptureSpaces`, `FreeNextSpaces` (`valueToRange`, `neighborsByKey`,
   `componentAtKey`, `ComponentColor`, or a callee that needs them).
   `(*movePlaceToken).Apply` and `LegalCustom` use only `MoveTo`,
   `MayMoveToSlot`, an enum value and `spaceIsBlack`; `(*token).Dynamic`
   uses only the component's dynamic values. None needs a missing feature.

## Test results

- `./scripts/go-local build ./...`: passes.
- `./scripts/go-local test -count=1` on `./boardgame-util/lib/gamedef/...`
  (6 tests), `./boardgame-util/lib/transpile/...` (subset: 5 tests; 45
  fixture functions), `./boardgame-util/lib/lint/...` (17 tests: the 10
  existing unchanged, 7 new), `./boardgame-util/lib/gamepkg/...` (unchanged),
  `./boardgame-util/lib/conformance/...`, `./boardgame-util`: all ok.
- `./scripts/go-local run ./boardgame-util conformance export --check`: "The
  conformance corpus is current".
- `go vet` on every touched package: clean. `gofmt -l`: clean.
- Node commands: skipped; no TypeScript changed.

## Judgment calls

1. **Buckets.** A function refused only for engine gaps (or for callees
   refused only for engine gaps) has `Verdict: refused` (its `Refusals` are
   not empty, as the frozen schema requires) and bucket `engineGap`, and is
   counted in the subset, as the proposal's table does (checkers: 19 in the
   subset, about 11 of them needing an engine feature).
2. **`--max-refused N`** counts functions with `Verdict: refused` across
   all packages (outside the subset or needing an engine feature), the
   literal sense of "refused". `--json` is the existing flag.
3. **Codes.** `BGLINT0101` is emitted for each function in the subset,
   `BGLINT0102` for each refused one (the message says outside or engine
   gap), `BGLINT0103` per nondeterministic construct anywhere in the package
   (memory's agent: two global `math/rand` calls), `BGLINT0104` per stale
   header in `server/static/src/engine/games/<game>/bodies.gen.ts` when that
   file exists (none does). Diagnostic `File`s are relative to the package.
4. **Data only** means no method declared in the game's hand-written files;
   methods in generated files (`auto_reader.go`, the codegen header, or the
   standard "Code generated" header) are ignored everywhere, as the
   proposal's count ignores generated Go.
5. **Counting.** Roots by slot: the delegate's set-up, end,
   `CurrentPlayerIndex`, `ReadyToStart`, `ProposeFixUpMove` and predicate
   methods; configured and declared move types' `Apply`, `Legal`,
   `LegalCustom`, `DefaultsForState`; the player state's turn hooks and
   `GameScore`; the game state's `PlayerOrder`; component values types'
   `Legal`; `moves/interfaces` callbacks (role `hook`); computed-property
   functions (literals or named) and `ValidCounter` literals. Everything they
   reach in the package is counted (role `helper`). Then `Diagram` and agent
   types' methods and what only they reach are not ported; the exporter's
   methods (the list in "How this was counted", plus `ConfigureAgents`,
   `ConfigureComputedProperties`, `LowScoreWins`, `NewDelegate`, `HelpText`)
   and what only they reach are not counted; the rest is unreached
   (memory's and debuganimations' misspelled `DefaultNumPlayeres`,
   tictactoe's `tokenValue`).
6. **Erasure** is by shape: a function whose results are all the game state,
   a player state or a slice of player states, and that calls nothing in the
   package. It is left out of `Functions` and `Calls`. A statement that only
   binds erased values does not count when deciding whether a stock
   `DefaultsForState` call is the first statement (tictactoe calls
   `concreteStates` first).
7. **Added roles** (additive, as the freeze allows): `hook`,
   `readyToStart`, `proposeFixUpMove`. None occurs in the examples.
8. **FnRef of a literal** is the enclosing declaration's FnRef plus
   `.funcN`, N counting the declaration's top-level literals in source
   order, as the Go toolchain names closures.
9. **Ordering.** `StaticInfo.Functions` stays sorted by `Func` (frozen);
   callee-first order is `StaticReport.Order`, `CalleeFirst(functions)`, and
   the lint's function list. `callee-refused` is added at the first
   reference to the refused callee and propagates to a fixed point.
10. **Predicate constructors and custom counters** get one refusal for their
    kind (detail: the predicate names) and are not examined further.
11. **`ValueFromString`** is `unknown-call` with detail
    `enum.ValueFromString (nondeterministic)`, and a `value-from-string`
    warning.
12. **Unported stock moves and behaviors** show as `engine-gap` with detail
    `moves.<Type>` or `behaviors.<Type>` wherever a field or method is
    selected through one (an embedded field on the path), and on a stock
    `Apply`/`Legal` call to one. Ported means mentioned in
    `server/static/src/engine/rules/*.ts`, plus `moves.FixUp` (a fix-up
    `moves.Default`, `MoveSpec.isFixUp`). `TestEngineGapsMatchStockTS` holds
    the table to that: 52 stock move types (abstract bases with `Legal` and
    no `Apply` included), 17 ported, 35 gaps; 18 behaviors, 9 ported, 9 gaps.
13. **Calls the examples need that the proposal's API table omits** are
    rows with `Source: "added: ..."`, each a direct analogue of a listed
    row: `ImmutableGameState()`/`GameState()` and a sub-state's
    `State()` (erased, like `concreteStates`), `EnumKey.Int()` (identity),
    `c.Deck().Name()` (the handle's deck), and
    `ContainingComponent().ImmutableInstance(state)` (a component method's
    own handle). Rows also list each method under the type that declares it
    (`(boardgame.State).Rand` beside `(boardgame.ImmutableState).Rand`,
    `(*graph.EnumGraph).NeighborsByKey`, `(*base.SubState).State`).
14. **Smaller rules the proposal leaves open:** a package variable of another
    package is `package-var-read`; a map literal is `unknown-type`;
    `len(map)`, `cap`, `new` and `panic` are `unknown-call`; `float32` and
    `complex` are `unknown-type`; a dereference is `pointer-local` (or
    `struct-copy` for a struct); an interface method call is
    `interface-dispatch` when the game declares the interface; a method call
    on another package's interface outside the table is `unknown-call`;
    `fmt` formats must be constant and use only `%d %s %v %q`. A chain row
    (`state.Game()`, `state.Manager()`) is accepted only as the receiver of
    another table call.
15. **Move configs** are found statically: each `auto.MustConfig`/`Config`
    call `ConfigureMoves` reaches (through in-package helpers), in source
    order, with `moves.DefaultRoundSetup` expanded to its three stock types
    (four with an option, `WithManualStart`, adding `CloseAllSeats`). Names
    are read only from literal `WithMoveName`/`WithMoveNameSuffix` options;
    the run-time half has the real names.

## What is unverified

- That the translator (workstream 10) accepts exactly the functions this
  checker passes: no translator exists. The checker is tested against its
  fixtures and the seven games only.
- That the added API rows are all implementable in the facade.
- `lib/dice.ts` and `lib/playingcards.ts` do not exist yet
  (`server/static/src/engine/lib/` is absent). Pig's `Roll` is not marked as
  a gap because the brief limits gaps to the table's "none (gap)" rows and
  unported stock moves and behaviors; the table's `Roll` row names a library
  facade, not a gap.
- `BGLINT0104` has only a unit test; no generated bodies file exists.
- The lint finds move configs statically. A game that builds configs in a
  loop, or passes them through a variable spread, would be miscounted; none
  of the seven does.

## Where the proposal was wrong about the code

- The baseline's function counts for memory and blackjack (closures),
  tictactoe's not-ported count, and checkers' engine-feature estimate:
  above.
- "Package `moves` declares about 50 stock move types; `rules/stock.ts`
  ports about 16": 52 and 17 (with `FixUp`), so 35 unported, against "about
  34" in risk 3.
- `legal.PredicateConstructor` is an alias of
  `boardgame.LegalPredicateConstructor`.
- The API table is keyed by the call as written; go/types reports the
  declaring type, so several rows needed a second spelling (13 above).

## Edits outside my files

None to existing files other than `cmd_lint.go` (mine). New files outside
the brief's list, each a fixture, golden or generator the brief asked for:
`boardgame-util/lib/gamedef/testdata/staticgame/{game.go,auto_reader.go}`
(`auto_reader.go` generated by `boardgame-util codegen`),
`boardgame-util/lib/lint/testdata/declarative/*.json`,
`boardgame-util/lib/lint/testdata/gen/declarative/main.go`,
`boardgame-util/lib/transpile/subset/testdata/fixture/fixture.go`.
