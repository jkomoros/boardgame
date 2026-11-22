# Phase 0: the freeze

What `fabric/TRANSPILE-PROPOSAL.md` "Phase 0: freeze" fixed, so that
workstreams 1 to 9 can start against stable names. Everything here is a
type, a schema or a stub. No existing behavior changed: the checked-in
conformance corpus is byte for byte the same, and so is every test result.

## Interfaces fixed

### Go

| File | What |
|---|---|
| `boardgame-util/lib/gamedef/schema.go` | `GameData` and every type under it: `Source`, `Initial`, `Traits` (`GameTraits`, `PlayerTraits`, `DrawDiscardPair`, `TurnTraits`), `ComputedData`, `PredicateData`, `SetUpData`, `EndData`, `AgentData`, `MoveData`, `StockData`, `FieldData`, `PlanData`, `MoveBodies`, `ProgressionData`, `CounterData`, `FunctionEntry`, `FnRef`. The definition: `Definition`, `DeckDef` (+ `GenericValues`, `ComponentProps`), `ComponentPropDef`, `EnumDef` (+ `Parents`, `Range`), `SubStateDef`, `PropDef` (+ `Constraints`). Constants: `SchemaVersion = 1`, `Exporter = "gamedef@1"`, `Base*`, `Progression*`, `Counter*`, `Role*`, `Verdict*`, `GameScoreFromScore` |
| `boardgame-util/lib/gamedef/static.go` | `StaticInfo`, `MoveStatic`; `Static(importPath string) (*StaticInfo, error)`, stubbed |
| `boardgame-util/lib/gamedef/export.go` | `Export(manager *boardgame.GameManager, static *StaticInfo) (*GameData, error)`, stubbed |
| `boardgame-util/lib/gamedef/hash.go` | `SourceHash(src []byte) string`, implemented: `"sha256:"` + hex SHA-256 of `go/format.Source(src)`; `SourceHashPrefix` |
| `boardgame-util/lib/transpile/refusal/refusal.go` | Leaf package `refusal`: `Refusal{Func, Reason, Detail, File, Line, Column}`, the 28 reason constants, `Reasons()`, `ValidReason(string) bool` |
| `boardgame-util/lib/transpile/refusal.go` | Package `transpile`: `type Refusal = refusal.Refusal`, `Reason*` constants, `Reasons()`, `ValidReason()` |
| `boardgame-util/lib/conformance/document.go` | `Document.Rejections []Rejection` (`json:"rejections,omitempty"`), `Rejection{AtVersion, Kind, Name, Input, Proposer, ExpectedVersion, Error}`, `SchemaVersionWithRejections = 2` |

Import direction: `gamedef` and `transpile` both import `transpile/refusal`;
neither imports the other.

### TypeScript (`server/static/src/engine/`, and `conformance/schema.ts`)

| File | What |
|---|---|
| `rules/data.ts` | `GameData` and every type under it, key for key with the Go JSON (`InitialData`, `TraitsData`, `GameTraitsData`, `PlayerTraitsData`, `ComputedData`, `PredicateData`, `SetUpData`, `EndData`, `MoveData`, `StockData`, `FieldData`, `PlanData`, `MoveBodies`, `ProgressionData`, `FunctionEntry`, `Refusal`, `FnRef`; `CounterData` re-exported from `progression.ts`). `Bodies`. `gameFromData(data, bodies): GameSpec`, throws |
| `rules/stockRegistry.ts` | `StockParts { spec?: Partial<MoveSpec>; apply?; legal?; defaults? }`; `stockMove(type, config): StockParts \| undefined`, throws |
| `rules/types.ts` | `MoveSpec.plan?: readonly LegalSpec[]`; `SubStateSpec.initialValues?: readonly Readonly<Record<string, unknown>>[]`; `DeckSpec.generic?: Readonly<Record<string, unknown>> \| null` |
| `rules/progression.ts` | `CounterData`; `Counter = ValidCounter \| CounterData`; `evalCounter(counter, currentCount, length)`; `ProgressionGroup.count` widened to `Counter` |
| `definition.ts` | `DeckDef.GenericValues?`, `DeckDef.ComponentProps?`, `ComponentPropDef`, `EnumDef.Range?`, `PropDef.Constraints?` |
| `go/api.ts` | Types `StackRef { state, at }`, `Comp { stack, slot, deck, index }`, `DynamicRef { deck, index }`, `Owner`, `GraphRef`. Functions `player`, `gameStack`, `playerStack`, `stackView`, `componentAt`, `first`, `last`, `components`, `componentAtKey`, `values`, `dynamic`, `containingStack`, `moveTo`, `secretMoveTo`, `moveToFirstSlot`, `moveToLastSlot`, `moveToNextSlot`, `slideToFirstSlot`, `slideToLastSlot`, `mayMoveTo`, `mayMoveToSlot`, `moveAllTo`, `moveCountTo`, `mayMoveAllTo`, `mayMoveCountTo`, `shuffle`, `publicShuffle`, `swapComponents`, `swapComponentsByKey`, `sortComponents`, `setSize`, `expandSize`, `contractSize`, `sizeToFit`, `variantValue`, `setEliminated`, `setPlayerInactive`, `setSeatFilled`, `enumKey`, `setEnumKey`, `enumEquals`, `rangeValue`, `rangeToValue`, `valueToRange`, `neighborsByKey`, `stockApply`, `legalError` |
| `go/num.ts` | `idiv`, `imod`, `imul` |
| `go/str.ts` | `join`, `repeat`, `split`, `toLower`, `trimSpace`, `hasPrefix`, `hasSuffix`, `contains`, `strCompare`, `utf8Len`, `itoa`, `atoi`, `sprintf` |
| `go/slice.ts` | `appendTo`, `lenOf`, `at`, `setAt`, `makeSlice`, `copyInto`, `subslice` |
| `go/panic.ts` | Constants `NIL_DEREF_TEXT`, `DIVIDE_BY_ZERO_TEXT`, `INDEX_OUT_OF_RANGE_PREFIX`, `SLICE_BOUNDS_PREFIX`, `NIL_MAP_WRITE_TEXT`; `nilDeref`, `divideByZero`, `indexOutOfRangeText`, `indexOutOfRange`, `sliceBoundsText`, `sliceBoundsOutOfRange` |
| `conformance/schema.ts` | `ConformanceDocument.schemaVersion: 1 \| 2`, `rejections?: Rejection[]`, `Rejection` |

Every facade body throws `go/<file>: <name> not implemented`; workstream 7
implements them.

## Judgment calls (where the proposal was silent)

1. **Schema version rule.** The proposal says `SchemaVersion` "becomes 2";
   it also says the corpus stays byte for byte. `SchemaVersion` stays 1 and
   `SchemaVersionWithRejections = 2` is added: a document is version 2
   exactly when it carries rejections. `rejections` is `omitempty`, so
   existing documents serialize unchanged.
2. **Rejection details.** `input` is `map[string]json.RawMessage` (the
   proposal's `Record<string, Json>`); `name` is always present, `""` for a
   timer or seat input that names no move.
3. **Refusal type location.** A leaf package `transpile/refusal`, aliased
   from `transpile`, so `gamedef` never imports `transpile`. The Go
   constants in `transpile` carry a `Reason` prefix (`ReasonMapRange`);
   in `refusal` they do not (`refusal.MapRange`).
4. **Refusal reasons.** The table's 28, nothing else. `select` and channels
   (listed as refused in "Statements" and "Types") have no row of their
   own; they are documented under `goroutine`. `ValueFromString` (refused
   under "Semantics that differ") likewise has no reason; workstream 10 must
   pick one (probably `unknown-call`) or the table must grow.
5. **`Definition` structs live in `schema.go`**, not `definition.go`, so
   workstream 1 can put the moved `DefinitionFor` code in `definition.go`
   without touching a frozen file. They are field-for-field vecgame's, with
   the additions `omitempty` so the projection without them is byte for
   byte today's `testdata/defs/*.def.json`. `EnumDef.Parents` is added to
   the Go side too (the TypeScript `EnumDef` had it; vecgame's export never
   emitted it).
6. **`ComponentProps`** is `map[prop]{Type, Enum?}`, the shape of
   `ValuePropDef`; the TypeScript type is `ComponentPropDef`.
7. **`Range`** is `[]int`, `RangeDimensions()`.
8. **Nullability.** Only `Constants`, `Variants`, `Stock`, `LegalPhases`,
   `Progression` and `GenericValues` may be null (or absent). The exporter
   must write every other map and slice non-nil (`{}`/`[]`). The TypeScript
   types say so.
9. **`ProgressionData` and `CounterData` have a `MarshalJSON`** that writes
   only `Kind` and the fields that kind uses, so `{"Kind":"atLeast","N":0}`
   and a serial group's `"Children":[]` survive. Unmarshalling is the
   default.
10. **`FnRef` format**: `pkg.(*T).M`, `pkg.(T).M` for a value receiver,
    `pkg.f`; the package name, not the import path.
11. **Roles** are fixed as constants (`apply`, `legal`, `legalCustom`,
    `defaults`, `computed`, `setUpBegin`, `distribute`, `setUpFinish`,
    `conditionMet`, `playerScore`, `checkGameFinished`,
    `currentPlayerIndex`, `playerOrder`, `gameScore`, `turnDone`,
    `resetForTurnStart`, `resetForTurnEnd`, `predicate`, `counter`,
    `component`, `helper`). Workstream 2 may add, not rename.
12. **`StaticInfo`** joins with the run-time half by `GoType`
    (`Moves map[string]MoveStatic`), plus maps for computed properties and
    predicates by name, and a `Counters` map (empty in v1) for custom
    counters. `Export` takes a `*boardgame.GameManager`, built as
    `testdata/gen/defs` and `lib/conformance` build one.
13. **`SourceHash` input** is one declaration without its doc comment; if
    it does not parse it is hashed as given (documented).
14. **`initialValues`** is an array indexed by player index (game: one
    entry at 0; dynamic values: deck index); a missing entry means all
    zero; `initial` wins over it if both are set.
15. **`CounterData` in the engine** includes `custom` (with `Count`, a
    FnRef) so `data.ts` can reuse it; `evalCounter` throws on a custom
    counter, which the loader must have replaced with its function.
    `matchTape` now calls `evalCounter`; for the function form that is the
    same call as before.
16. **Facade handles.** `StackRef` is `{ state, at }` where `at` is the
    engine's own `StackRef` location (sub-state, prop, space), so a handle
    is data and survives stack operations. It shares its name with
    `state.ts`'s `StackRef`, which `api.ts` imports as `EngineStackRef`.
    `Owner` (for enum properties) is `'game' | playerIndex |
    { dynamicOf } | { valuesOf }`, tagged because `Comp` is structurally a
    `DynamicRef`. `GraphRef` is `{ enum, dimensions, directions }`, so
    `neighborsByKey(graph, key)` needs no state; workstream 9 may refine the
    direction names.
17. **Facade extras** beyond the table: `stackView(s)` (for the "the same"
    length queries), `setAt`, `makeSlice`, `copyInto`, `subslice`,
    `divideByZero`, the slice-bounds and nil-map texts. `enumEquals` takes
    two owner/prop pairs; `setEnumKey` returns `string | null` because Go's
    `SetValue` returns an error. `containingStack` returns Go's three
    results as a tuple.
18. **Files.** `sprintf`, `itoa`, `atoi` are in `str.ts`; `legalError` and
    `stockApply` in `api.ts`.
19. **`stockMove`** throws (rather than returning undefined) so a premature
    caller fails loudly.

## Where the proposal and the code disagree

- "`SchemaVersion` becomes 2" against "the checked-in conformance corpus
  stays byte for byte": resolved by call 1.
- "`Parents` for a tree enum exists already": only in the TypeScript
  `EnumDef`; Go's `DefinitionFor` (vecgame) does not emit it.
- "`testdata/gen/vecgame/defs.go` (`DefinitionFor`), for five games":
  `testdata/defs/` holds four games (memory, blackjack, pig, tictactoe)
  plus vecgame.
- The API table writes `neighborsByKey(graph, k)`, while range-enum calls
  need the enum from the state; `GraphRef` therefore carries the
  dimensions (call 16).

## Acceptance as run

- `./scripts/go-local build ./...`: passes.
- `./scripts/go-local vet` on `gamedef`, `transpile`, `conformance`: clean.
- `./scripts/go-local test ./boardgame-util/lib/conformance/...`: ok (both
  packages).
- `./scripts/go-local run ./boardgame-util conformance export --check`:
  "The conformance corpus is current".
- `npm run type-check`: passes. `npx tsc --noEmit -p
  src/engine/tsconfig.sandbox.json`: passes.
- `npm run test:unit` (Node 24): 1168 pass, 0 fail, before and after.
- `npm run test:layers`: 11 pass.
