# Exporter, run-time half (workstream 1): status

Workstream 1 of `fabric/TRANSPILE-PROPOSAL.md`. Based on `d2cb55e4`; one
local commit. Nothing pushed, no PR. No TypeScript changed.

## What was built

- `boardgame-util/lib/gamedef/definition.go`: `DefinitionFor(manager)`, moved
  from `testdata/gen/vecgame/defs.go`, now also emitting
  `Decks[].GenericValues`, `Decks[].ComponentProps`, `Enums[].Parents`,
  `Enums[].Range`, `PropDef.Constraints`, and `PropDef.Enum` for a board
  keyed by an enum. `(*Definition).WithoutAdditions()` removes exactly those,
  giving the old shape. Also `PolicyName`, `TypeName`.
- `boardgame-util/lib/gamedef/export.go`: `Export(manager, static)` fills
  every run-time part of `GameData`; `Marshal(data)` is the checked-in form
  (tab indent, no HTML escaping, trailing newline).
- `server/static/src/engine/testdata/gen/vecgame/defs.go`: its types are now
  aliases of gamedef's; `DefinitionFor` is `gamedef.DefinitionFor(...)
  .WithoutAdditions()`; new `DefFiles()` derives every `testdata/defs/*.def.json`
  (used by `gen/defs` and by the gamedef test). `sanitizevec` compiles
  unchanged against the aliases.
- `server/static/src/engine/testdata/gen/defs/main.go`: writes `DefFiles()`.
  Regenerating leaves all five `testdata/defs/*.def.json` byte-identical.
- `boardgame-util/cmd_gamedef.go`: `boardgame-util gamedef export [--check]
  [--root DIR] [PKG...]`. PKG is a game name or import path of one of the
  seven example games (the run-time half needs a compiled-in delegate);
  default all seven. Calls `gamedef.Static(importPath)`; an error containing
  "not implemented" is tolerated (note on stderr, run-time parts only), any
  other error fails. `--check` reports missing, differing and (with no PKG)
  extra files under `conformance/definitions/`.
- `conformance/definitions/<game>.json` for memory, blackjack, pig,
  tictactoe, werewolf, debuganimations, checkers (run-time parts only, since
  `Static` is still the stub).

## Exported interfaces (package gamedef)

```go
func DefinitionFor(manager *boardgame.GameManager) (*Definition, error)
func (d *Definition) WithoutAdditions() *Definition
func Export(manager *boardgame.GameManager, static *StaticInfo) (*GameData, error) // frozen signature
func Marshal(data *GameData) ([]byte, error)
func PolicyName(p boardgame.Policy) string
func TypeName(t boardgame.PropertyType) string
```

`vecgame`: `DefinitionFor` (unchanged signature), `PolicyName`, type aliases
`Definition`, `DeckDef`, `EnumDef`, `SubStateDef`, `PropDef`; new `DefFile`,
`DefFiles()`.

`schema.go` was not changed.

## How each part is obtained

| Part | Source |
|---|---|
| Names, player counts | `manager.Delegate()` through the interface, as the manager and the conformance header call them (so a misspelled `DefaultNumPlayeres` is not used) |
| `Definition` | `DefinitionFor` |
| `Constants` | chest `ConstantNames`/`Constant`; null when none |
| `ComponentEnums` | enums of component values and dynamic values whose enum is not the chest set's own |
| `PhaseEnum`, `GroupEnum` | delegate |
| `Initial` | `GameStateConstructor()`, `PlayerStateConstructor(i)` for i in 0..Max-1, `DynamicComponentValuesConstructor(deck)` called once per component; every scalar/slice prop read through the generated reader; kept when non-zero (number != 0, true, string != "", enum value set (by its string), slice non-nil). Stacks, boards, timers never kept. Decks whose constructor returns nil are absent |
| `Variants` | `manager.Variants()`, canonical JSON (as the header) |
| `Computed` | declaration order from `ConfigureComputedProperties()` (the unexported `descriptor` read by reflection), descriptor from `ComputedPropertyDescriptors()`, Type by `TypeName` |
| `LegalTemplates` | the delegate's own `ConfigureLegalTemplates()` (not the merged table; the TS loader merges defaults as `compileGame` does) |
| `Predicates` | names in order from `ConfigurePredicateConstructors()`; `Construct` from static |
| `Moves[]` | `manager.ExampleMoves()` in registration order. `GoType` = reflected type (`moves.StartPhase`, `blackjack.moveCurrentPlayerHit`). `Base` = currentPlayer / anyPlayer when `moves.CurrentPlayer` / `moves.AnyPlayer` is embedded at any depth. `Fields` = every `ReadSetter().Props()` field except `TargetPlayerIndex` when Base supplies it, with Input/Codec from `boardgame.ResolveMoveInputFields` (fix-ups too), Sanitize from the struct tag. `NameSanitization` = the `policy` field of the core's unexported name-sanitization config, by reflection; `HiddenAnimationKey` = `AnimationKeys()[1]`. `LegalPhases` = `InstalledLegalPhases()`. `Progression` = `DescribeProgression(InstalledProgression(config))`. `Plan` = `LegalPlanSpecs`. `Choices` = the move's `BuildMoveChoiceProjectionSchema` entry, canonical. `SeatsPlayer`, `GatheringStart` from the interfaces; `EndsPhaseProgression` = has `PhaseToStart` or embeds/is `moves.NoOp` (the installers.go rule) |
| `Stock` | `Type` at run time: the move's own type when package moves declares it, else the first embedded type (depth first, field order) package moves declares; null when that is `moves.Default`. `Config` = `CustomConfiguration()` with all keys as stored, values: progression group -> ProgressionData, enum -> name, enum value -> string, string/bool/int (enum keys) as is, types with no unexported field -> encoding/json (so LegalSpec keeps lowercase keys, MoveInputField its Go names), otherwise struct -> object of Go field names (unexported included), slice -> list, pointer/interface -> target or null; funcs/chans are an error |
| `Traits` (run-time part) | anonymous `behaviors.*` fields of the game and player state structs: DrawDiscardPair (with its `draw`/`discard` tags), RoundRobin, Seat, InactivePlayer, PlayerElimination, PlayerRole, PlayerColor; GameScore = `"score"` when ScoreBehavior is embedded |
| `End.LowScoreWins` | `LowScoreWins()` on the delegate |
| `Agents` | `manager.Agents()` |

Join with `*StaticInfo` (when non-nil): `static.Moves[GoType]` must exist
(else error); its `StockType` must equal the run-time one (else error) and
its `Bodies` are used; `Computed[].Compute`, `Predicates[].Construct`,
`SetUp`, `End` FnRefs, `CurrentPlayerIndex`, `Functions` (re-sorted by
Func), and a custom counter's `Count` (key `"<move name>"` plus `"/<child
index>"` per step from the root) come from static. Traits: booleans are OR'ed,
but if the static half sets any boolean trait or a DrawDiscardPair, a
boolean that differs is an error; differing DrawDiscardPairs are an error;
a static GameScore FnRef replaces `"score"` (reflection cannot tell a
declared method from a promoted one); PlayerOrder and Turn come from static
only.

## Tests (`boardgame-util/lib/gamedef/export_test.go`, all pass)

- `TestTestdataDefsCurrent`: `vecgame.DefFiles()` equals the five checked-in
  `testdata/defs/*.def.json` byte for byte, and no extra file exists.
- `TestDefinitionsCurrent`: the seven `conformance/definitions/*.json` equal
  a fresh export (the `--check` rule; calls `Static` the same tolerant way).
- `TestExportDeterministic`: two exports from fresh managers are
  byte-identical, for all seven; the output round-trips into `GameData`.
- `TestNonNullCollections`: the nullability rule for the top-level and
  per-move collections.
- `TestAgreesWithConformanceHeader`, for the six corpus games against each
  game's shortest checked-in document: game name/display name/player
  counts; phase and group enums; move table in order (name, help text,
  fix-up, gathering start, animation keys, namePublic iff no
  NameSanitization, fields with type and enum); the input schema (set of
  non-fix-up moves, every field's wireType, disposition, codec, enum, enum
  values); choice projections; computed descriptors; variant config; chest
  decks (every component's values), enums (values, default, parents,
  dimensions vs Range), constants, and templates (every exported template
  in the merged table with the same text; every other merged key is a
  `legal.` default). Compared per game (moves / fields / input fields /
  choice projections / computed / components / enums / game templates):
  memory 7/5/3/1/1/120/0/6, blackjack 23/8/2/0/1/52/1/6, pig
  6/5/2/0/0/1/0/4, tictactoe 4/4/2/0/0/10/1/1, werewolf 11/7/4/2/0/0/3/0,
  debuganimations 13/3/2/0/0/92/1/6.
- `TestInitial`: tictactoe players 0 and 1 are exactly `{TokenValue: X|O,
  TokensToPlaceThisTurn: 1}`; every werewolf player (7) is exactly
  `{DayVote: -1, NightVote: -1}`; pig game `{TargetScore: 100}` and one die
  `{Value: 1}`.
- `TestDefaultNumPlayersAsGoUsesIt`: memory and debuganimations still declare
  `DefaultNumPlayeres`, and the export equals base.GameDelegate's default and
  the interface call.
- `TestCheckers`: `spaces` Range `[8,8]` with 64 values; `CapturedTokens`
  constraints `["same(Color)"]`, `UnusedTokens` none; `Spaces` a sized stack
  of 64 = spaces cardinality = `BOARD_SIZE`; Tokens dynamic values
  (`Crowned` bool) and one empty initial entry per component; component prop
  `Color` enum; Move Token's two RangeVal fields are enum `spaces` with codec
  `integer`, declared plan. Tictactoe `UnusedTokens` constraints
  `["same(Value)"]`.
- `TestDefinitionAdditions`, `TestMoves`: blackjack component enums/props,
  generic values (blackjack, memory), blackjack traits, debuganimations tree
  enum parents and a deck with no props; blackjack Begin Game stock config
  (StartPhase 1, StartPhaseEnum "phase"), Current Player Hit
  (currentPlayer, DrawToPlayer, HasCustom, phases [2]), 15 blackjack moves
  with progressions (the number workstream 3 counted), tictactoe name
  sanitization and field sanitize tag, werewolf Cast Vote anyPlayer.
- `TestStaticJoin`: with a synthetic `StaticInfo` agreeing with the run-time
  half, bodies/computed/set-up/functions are joined; a missing move entry, a
  disagreeing stock type, a disagreeing DrawDiscardPair and a disagreeing
  RoundRobin each make Export fail; a declared GameScore replaces "score".
- `TestWithoutAdditionsLeavesOriginal`.

## Where a part of GameData could not be obtained at run time

- Everything that is a FnRef: `MoveData.Bodies`, `Functions`,
  `Computed[].Compute`, `Predicates[].Construct`, `SetUp` (so
  `SetUp.Distribute`, which is not omitempty, is `""`), `End.ConditionMet`/
  `PlayerScore`/`Check`, `CurrentPlayerIndex`, `Traits.Game.PlayerOrder`, a
  declared `GameScore`, `Traits.Player.Turn` (the interface is detectable,
  but the type requires three FnRefs, so it is left nil), a custom counter's
  `Count`. All wait for `Static`.
- `Stock.Config` for a move whose only stock ancestor is `moves.Default`:
  the schema has `Stock: null` there, so its configuration has nowhere to
  go. This is 12 of debuganimations' 13 moves; their phases, plan and help
  text are still in the other MoveData fields.
- `Computed[].Type` for an enum computed property: `ComputedType` has no
  enum and `ComputedData` no enum name; it would be written `"enum"` with the
  enum lost. No example has one.
- `Initial`: stacks, boards and timers a constructor might set are not
  exported (no example sets one).

## Judgment calls

1. `LegalTemplates` is the game's own `ConfigureLegalTemplates()`, as the
   schema comment says, not the merged table the header's chest carries.
2. `Stock.Type` for a stock type configured directly (`moves.StartPhase`) is
   that type itself, not its first embedded type. The static half must use
   the same rule or Export fails on every such move (33 moves across the
   seven games); the coordinator should check `static.go` against it.
3. `Fields` leaves out `TargetPlayerIndex` exactly when Base is currentPlayer
   or anyPlayer (what `compile.ts`' `moveFields` adds back). A stock fix-up
   with its own `TargetPlayerIndex` (SeatPlayer, ActivateInactivePlayer)
   keeps it, with Input "required" as `ResolveMoveInputFields` says.
4. `Stock.Config` keeps every key, including ones also split out into
   MoveData (progression, phases, help text, preconditions).
5. `Initial` keeps an enum value whenever the constructor set one, and a
   slice whenever it is non-nil (empty included), since Go stores `[]` and
   `null` differently.
6. Reflection reads three unexported core/moves values rather than adding
   accessors outside my files: the name-sanitization policy string, a
   ComputedProperty's descriptor (for declaration order; falls back to the
   sorted order if the shape changes), and unexported config structs in
   `Stock.Config` (`moveChoiceProjectionDeclaration` with `projection`/`err`,
   the sanitization config with `policy`/`hiddenAnimationKey`/`err`).
7. Positional PKG accepts names or import paths of the seven compiled-in
   games only; any other package cannot be exported without a delegate.
8. "Board keyed by an enum": no example has a `boardgame.Board` with an
   `enum:` tag (checkers' board is a sized stack addressed by the `spaces`
   range enum). `PropDef.Enum` is emitted for such a board and stripped by
   `WithoutAdditions`; untested on a real game.

## Where the proposal was wrong about the code

- "Five games" in `testdata/defs`: four example games plus vecgame (freeze
  noted this too).
- The misspelled `DefaultNumPlayeres` returns 2 in both games, which is also
  the base default, so the exported value is right either way; the test pins
  that the interface value is used but cannot tell the two apart.
- `EnumDef.Parents` is an addition to the Go `DefinitionFor` (the chest JSON
  already carried it); `WithoutAdditions` removes it with the others.
- The checkers "board" is a `SizedStack`, not a `Board`.

## Unverified

- Nothing is checked against a real static half (`Static` is the stub). The
  join is tested only with a synthetic `StaticInfo`.
- The `Plan`, `Progression` and `Stock.Config` contents are checked only
  where noted above; the plan content is workstream 3's accessor, used
  verbatim.
- Checkers has no conformance corpus, so its export is checked only by the
  assertions in `TestCheckers`.

## Edits outside my files

- `boardgame-util/cmd_base.go`: two lines registering `gamedefCmd` (as the
  brief allows).

## Acceptance as run

- `./scripts/go-local build ./...`: passes.
- `./scripts/go-local vet` on `boardgame-util`, `lib/gamedef`, `gen/...`: clean; `gofmt -l` on my files: none.
- `./scripts/go-local run ./boardgame-util gamedef export`: wrote 7 files (a note per game that the static half is not implemented); `--check`: "The game definitions are current".
- `./scripts/go-local run ./server/static/src/engine/testdata/gen/defs`: the five `testdata/defs/*.def.json` byte-identical (`git status` clean for them).
- `./scripts/go-local run ./boardgame-util conformance export --check`: "The conformance corpus is current".
- `./scripts/go-local test -count=1` on `lib/conformance/...` (both ok), `lib/gamedef` (ok, 14 top-level tests), `./boardgame-util/` (ok), `gen/...` (no test files).
- Node commands skipped: no TypeScript changed.
