# Gaps 2: inferred stock configuration, ComponentLegal, facade holes

Base `ca4492a0` (branch `claude/transpile-pipeline-e23924`); one local
commit. Nothing pushed, no PR. Closes gap A (integration-1), gaps B and C
(checkers-engine).

## Gap A: configuration Go infers at boot

### Where the fix is, and why

`moves.MoveComponentToSlot` without `WithSlotField` re-derives its slot
field (the move's only int property, `resolveSlotField`) every time
`slotPlanFor` runs; it never stores the answer, so `CustomConfiguration()`
has nothing for the exporter to copy. `DrawToPlayer` differs: its
`ValidConfiguration` writes `ResolvedDrawSource`/`ResolvedDrawDestination`
into the configuration itself, which is why the exporter already carried it.

Go's resolution is reachable without changing behavior through one new,
read-only accessor, so the fix exports what Go resolved and the registry
does not re-implement the inference:

- `moves/resolved_config.go` (new file): `moves.ResolvedSlotField(move)
  (string, error)`, a wrapper over the unexported `resolveSlotField`.
- `lib/gamedef/export.go`: `ResolvedSlotFieldKey`
  (`github.com/jkomoros/boardgame/moves.ResolvedSlotField`) and
  `addResolvedConfig`, which writes that key into `Stock.Config` for every
  move whose stock type is `MoveComponentToSlot` (configured or discovered,
  as DrawToPlayer always records its pair).
- `rules/stockRegistry.ts`: `MoveComponentToSlot` reads `ResolvedSlotField`,
  else `SlotField`.

Result: memory Reveal Card `CardIndex`, tictactoe Place Token `Slot`.
`TestResolvedSlotField` checks both, that the field is the one Go's own
contributed `mayMoveToSlot`/`mayMoveFirstToSlot` precondition reads, and
that no other stock type gets the key.

### Audit of the other stock types (every `ValidConfiguration` in `moves/`)

| Stock type | Inference at boot | Status |
|---|---|---|
| `MoveComponentToSlot` | slot field = only int property | fixed (above) |
| `MoveComponentToSlot` `WithSourceSlotField` | none (named) | already exported |
| `DrawToPlayer` | draw stack from `DrawDiscardPair`, player stack = only stack | Go stores it (`ResolvedDraw*`); already read |
| `ShuffleDiscardIntoDraw` | the game state's embedded `DrawDiscardPair` | a trait (`Traits.Game.DrawDiscardPair`), already exported; `WithDrawDiscardPairField` (a named pair) is an unported feature, not an inference: `stockMove` returns undefined |
| `WaitForEnoughPlayers` `WithRequireExplicitStart` | none | unported feature (`stockMove` returns undefined); no example uses it |
| `CloseAllSeats`, `WaitForEnoughPlayers` `TargetCount` | defaults to `MinNumPlayers` | resolved at run time from the delegate, as the registry's parts do; not configuration |
| `DefaultComponent` | `SourceStack`/`LegalType` through the concrete move's methods | config when configured; a game overriding the methods supplies bodies, not configuration |
| Deal/Collect/Move-components, `ShuffleStack` | `GameStack`/`PlayerStack`/`SourceStack` overrides | same as above: an override is a method, not inferred configuration |
| `StartPhase` | `PhaseToStart` | configured key; an override is a method |
| `ForceFinishTurn`, `FinishTurn`, seat moves, `ActivateInactivePlayer` | validation only | nothing to export |

Not fixed here and still listed in `subset.EngineGaps`' comment (the
translator's file, untouched): `WithRequireExplicitStart`/`WithManualStart`
and `WithDrawDiscardPairField` are unported features rather than missing
data.

### Acceptance

`node --test src/engine/rules/data.test.ts`, 16 of 16. Exported definitions,
bodies mode:

| Game | Documents | Versions | Stand-ins |
|---|---|---|---|
| memory | 6 | 172 | 0 of 7 |
| blackjack | 7 | 419 | 0 of 23 |
| tictactoe | 4 | 58 | 0 of 4 |
| pig | 3 | 283 | 0 of 6 |
| debuganimations | 1 | 16 | 0 of 13 |
| werewolf | 1 | 13 | 0 of 11 |

All six also pass through `move:`. The exported-definition test now asserts
the stand-in list is empty. The hand-written `rules/testdata/blackjack.gamedata.json`
fixture (not an export; its `Stock.Config` objects are empty by design)
still shows 7 of 23 through `stock:`; it is not one of the six definitions
and was not changed.

## Gap B: `GameData.ComponentLegal`

- `lib/gamedef/static.go`: `StaticInfo.ComponentLegal map[string]FnRef`,
  values Go type (`checkers.token`, MoveStatic.GoType's spelling) to the
  declared `Legal` FnRef, filled where `roots` already gives a component
  values type's `Legal` role `component`. Which deck holds which values
  type is not known statically, so:
- `lib/gamedef/export.go`: `componentLegal` joins it per deck by each
  component's reflected values type, only where the values implement
  `interfaces.LegalComponent`. A deck whose values implement it with no
  static `Legal`, or with two different ones, is an Export error.
- `conformance/definitions/checkers.json`: `"ComponentLegal": {"Tokens":
  "checkers.(*token).Legal"}`; the function was already in `Functions`
  with role `component`, verdict `subset`. The other six definitions get no
  `ComponentLegal` (omitempty).

Tests: `TestComponentLegal` (export_test.go; also the error when the
static half lacks the method), `TestStaticComponentLegal` (static_test.go).

`rules/checkers.test.ts`: every slot is now named by `checkers.json` with
the stand-in's spelling, so `patchRefs` patches nothing; the first test
now asserts `patched` is `[]`. Slots checked: the 6 move bodies, GameScore,
the three Turn hooks, the predicate's Construct, ComponentLegal.Tokens,
SetUp.Distribute, End.ConditionMet, End.PlayerScore. None needed a reason.

## Gap C: facade holes

`go/api.ts` (additive):

1. `Owner` gains `{ readonly move: MoveInstance }`. `enumKey`,
   `setEnumKey`, `enumEquals`, `rangeValue` work on a move's own enum
   fields: the enum is the compiled move's field spec (`findMove`), the
   value is the string in `m.fields`, which is how both the engine and Go's
   move record (`json.Marshal(move)`, inflated by `json.Unmarshal` in
   storage.go) hold it. Go-generated vector:
   `testdata/gen/facadevec/moves.go` writes
   `testdata/facade-moves.vectors.json` from real Go checkers' Move Token
   `SpaceIndex` (`enum.RangeVal`): the new move's value; 12 `SetValue`s in
   sequence (in range, -1, 64, 100: Go's `That value is not valid for this
   enum`, value unchanged) with the stored string, `Value()` and
   `RangeValue()` after each; four records unmarshalled into fresh moves.
   The two existing facade vector files are byte-identical after the run.
2. `componentInstance(state, deck, index): Comp`, Go's
   `v.ContainingComponent().ImmutableInstance(state)`/`Instance(state)`:
   the handle located where the component is now; for one the state cannot
   locate (generic, or stale after ContractSize) `slot` is -1 and `stack` is
   some stack of the state (only its state is used), so operations fail with
   Go's texts. Throws for an unknown deck or index.

`subset/api.go`, `added:` rows audited (one row edited):

| Row | Facade | Exists |
|---|---|---|
| `ImmutableGameState`/`GameState` | game handle (erased) | no call emitted |
| sub-state `State()` | state (erased) | no call emitted |
| `(enum.EnumKey).Int` | identity | no call emitted |
| `(boardgame.Component).Deck` | `c.deck` | `Comp.deck` |
| `(*boardgame.Deck).Name` | deck name | `Comp.deck` (with the row above) |
| `ContainingComponent` | chain only | receiver of the next row |
| `ImmutableInstance`/`Instance` | was "component handle", no function | **now `componentInstance`**, row's Facade renamed |

`rules/checkers.test.ts` stand-ins now use the facade: `fieldKey`/
`setFieldKey` are gone (`enumKey`/`setEnumKey` with `{ move: m }`), Move
Token's Apply uses `rangeValue` on the move fields (Go's
`m.X.RangeValue()`; was `valueToRange` of the key), `(*token).Legal` and
`(*token).Dynamic` take their receiver from `componentInstance`. Two new
tests: the move-field vectors, and `componentInstance` on every occupied
square of a vector position. Results, all zero differences:

| What | Result |
|---|---|
| seed-0000, 0001, 0006 | 26/30/26 versions, every field ok, PASS; the refusal refused again |
| constructed positions | 44 positions, 35,456 Move Token Legal texts (109 legal), 4,702 ledgers, 2,904 Crown Token Legal, 1,459 applies (567 ok, 11 captures), 0 failures |
| checkers.test.ts | 7 of 7 |

## Results, as run

| Command | Outcome |
|---|---|
| `./scripts/go-local build ./...` | passes |
| `go vet` (moves, gamedef, subset, facadevec), `gofmt -l` on my files | clean |
| Go tests `./moves/... . lib/gamedef/... lib/transpile/... lib/lint/... lib/gamepkg/... lib/conformance/... ./boardgame-util/` | all ok |
| `gamedef export --check` | "The game definitions are current" |
| `conformance export --check`, `fuzz corpus --check` | "The conformance corpus is current"; "The fuzz corpus is current (48 documents)" |
| `npm run type-check`; sandbox `tsc` | pass; pass |
| `npm run test:unit` | 1,315 of 1,315 (base 1,313, plus 2 checkers tests) |
| `npm run test:layers` | 11 of 11 |

Lint goldens: no verdict changed (`lib/lint` tests pass unregenerated).

## Judgment calls

1. An accessor in `moves/` rather than re-deriving the rule in the
   exporter or registry: the brief prefers exporting what Go resolved; this
   is the one additive, read-only Go change (`moves.ResolvedSlotField`).
2. The key is written for every MoveComponentToSlot, configured or not,
   mirroring DrawToPlayer; it sits under the moves prefix like Go's own
   `ResolvedDraw*` keys, though Go never stores it (documented on
   `ResolvedSlotFieldKey`).
3. `StaticInfo.ComponentLegal` is keyed by values type, not deck; the join
   is in Export, which errors rather than omits on a mismatch.
4. The `{ move }` owner's vector test lives in `rules/checkers.test.ts`,
   which already builds the real compiled checkers game the owner needs.

## Unverified

- A plain (non-range) enum move field: no example has one; the code path
  is the same apart from `rangeValue`.
- `componentInstance` for a lost or generic component: written to Go's
  shape, not vectored.
- A deck mixing values types: the error is written, not exercised.

## Edits outside my files

- `moves/resolved_config.go` (new, additive accessor).
- `server/static/src/engine/testdata/gen/facadevec/main.go` and `regen.sh`:
  one line each for the new vector file (facadevec is mine; regen.sh a
  comment line).
