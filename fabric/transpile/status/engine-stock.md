# Workstream 9: stock moves, behaviors, enums, graphs

Branch `worktree-agent-a75a7505c1c0d3500`, based on `d2cb55e4`. Nothing
pushed. What `fabric/TRANSPILE-PROPOSAL.md` workstream 9 asks for: framework
code from Go's `moves/`, `behaviors/`, `enum/`, `enum/graph/`,
`components/dice` and `components/playingcards`, ported to the TypeScript
engine and held to vectors from real Go, so generated games have something
to call.

## What was built

| File | What |
|---|---|
| `server/static/src/engine/rules/stock.ts` | Every existing helper exposes its parts. New: `stockParts(spec)`, `baseLegal`, `StockMoveParts`; the bases `defaultMove`, `fixUp`, `fixUpMulti`, `currentPlayerMove`, `done`; `defaultComponent`; `closeEmptySeat`, `activateEmptySeat`, `activateFilledSeat`, `closeAllSeats`; `shuffleStack`; `roundRobinParts(hooks)`, `RoundRobinHooks`, `roundRobinAllPlayersMet`; `dealComponents({kind})` for all eight deal and collect types; `moveComponents({kind})` for all four move-components types; `playerColorProps`, `componentColorProps` |
| `server/static/src/engine/rules/stockRegistry.ts` | `stockMove(type, config)`, a `switch` over the Go type name (with or without `moves.`), reading `CustomConfiguration()` keys fully qualified as Go stores them (`github.com/jkomoros/boardgame/moves.TargetCount`) or bare. `CONFIG_PREFIX`, `configValue` |
| `server/static/src/engine/rules/stock-inventory.json` | Every stock move type, behavior, hook and library in Go: `{ goType, kind, ported, registryName, verifiedBy, notes }`. 75 entries; the table below is generated from it |
| `server/static/src/engine/lib/enumrange.ts` | `enum.RangeEnum`: `rangeEnumDef` (AddRange), `rangeDimensions`, `valueToRange`, `rangeToValue`, `rangeValue`, `setRangeValueError`, `enumValueFromString`, `enumValid`, `enumString`, `ILLEGAL_VALUE` |
| `server/static/src/engine/lib/graph.ts` | `enum/graph` grids: `gridConnectedness(def, directions)`, `neighborsByKey`, `connectedByKey`, `parseDirection`, `graphMatchesEnum`; `EnumGraph` is data |
| `server/static/src/engine/lib/dice.ts` | `components/dice`: `basicDie`, `defaultDie`, `dieMin`, `dieMax`, `rollWith(rand, values, dynamic)`, `roll(state, deck, index)`, `dieDynamicProps` |
| `server/static/src/engine/lib/playingcards.ts` | `components/playingcards`: `newDeck`, `newDeckMulti`, `suitEnum`, `rankEnum`, `cardEnums`, `card`, `genericCard`, `rankKey`, `suitKey`, `cardString`, the `SUIT_*` and `RANK_*` keys |
| `testdata/gen/libvec/` | Generator of `testdata/lib.vectors.json` |
| `testdata/gen/stockgame/`, `testdata/gen/seatgame/` | Two small real Go games (with codegen'd `auto_reader.go`) using every newly ported stock move and behavior |
| `testdata/gen/stockvec/` | Plays five scenarios of those games on the Go engine and writes `testdata/stock/*.json.gz` (conformance documents), `index.json` and `fixups.json` (every fix-up's Legal text for the admin, at every version) |
| `lib/lib.test.ts`, `lib/purity.test.ts`, `rules/stock.test.ts`, `rules/stockRegistry.test.ts` | The tests below |

### How a game uses the parts

A stock helper returns a `MoveSpec` (or, for `drawToPlayer` and
`moveComponentToSlot`, `MoveParts`, which now carry `legal` too).
`stockParts(spec)` gives `{ spec, apply, legal, defaults }`, `legal` always
present (`baseLegal`, i.e. `base()`, where Go's Legal is `Default.Legal`).
`stockMove(type, config)` returns the same shape as `StockParts`; for a
base whose Apply the game must supply (`Default`, `FixUp`, `FixUpMulti`,
`CurrentPlayer`, `AnyPlayer`, `DefaultComponent`) there is no `apply`. A game
move that overrides Apply calls `parts.apply(state, move)` first, as its Go
calls `m.<Stock>.Apply(state)`; one that overrides Legal calls
`parts.legal(state, move, proposer, base)`.

`moves.DefaultComponent` needs the components' `Legal` (interfaces
`LegalComponent`): it is the new `DeckSpec.componentLegal(state, index,
legalType)`, `legalType` being the configured key or `null` where Go
passes nil.

For the facade (workstream 7): `rangeValue`, `rangeToValue`, `valueToRange`
are `lib/enumrange.ts` over `state.game.enums.get(name).def`;
`neighborsByKey(graph, k)` is `neighborsByKey(gridConnectedness(rangeEnumDef(graph.dimensions), graph.directions), k)`.
`parseDirection` accepts `graph.DirectionDown`, `DirectionDown` or `down`.

## Test results

All run on this branch with Go 1.27.4 (`./scripts/go-local`) and Node
24.22.3.

| Check | Result |
|---|---|
| `./scripts/go-local build ./...` | passes |
| `./scripts/go-local test ./boardgame-util/lib/conformance/...` | ok (both packages) |
| `./scripts/go-local run ./boardgame-util conformance export --check` | "The conformance corpus is current" |
| `./scripts/go-local vet` on the four new generator packages | clean |
| `libvec` and `stockvec` rerun | byte-identical output |
| `npm run type-check` | passes |
| `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | passes |
| `npm run test:unit` | 1,198 of 1,198 pass (1,168 at base + 30 new: lib 10, lib purity 3, stock 8, registry 9) |
| `npm run test:layers` | 11 of 11 |

What the new tests prove:

| Test | Proves |
|---|---|
| `lib/lib.test.ts` (10) | Against `lib.vectors.json`: six range enums' values, default, dimensions and Go's marshalled form; 510 ValueToRange/RangeToValue/SetRangeValue/ValueFromString cases (out-of-grid, wrong arity, `+1`, `-0`, `007`, padded strings); AddRange's four errors; grid connectedness for 14 graphs (every Direction filter alone, checkers' two, a 1x1, 1x5, and the 1-D and 3-D errors): every node's neighbors and 1,000+ ConnectedByKey pairs; BasicDie/Min/Max; 125 rolls on five seeds; four playing-card decks with their strings, generic card and enums. Also `games/blackjack.ts`'s inlined deck and enums equal `lib/playingcards.ts`, and `games/pig.ts`'s Roll Dice and `lib/dice.ts`'s `roll` leave the same die on eight games |
| `lib/purity.test.ts` (3) | `lib/` keeps the rules layer's purity rules (the sandbox test already scans it) |
| `rules/stock.test.ts` (8) | The five `testdata/stock` documents replayed with stock moves built by `stockMove` from Go's own configuration keys: every field, every version, every viewer. 154 versions, 715 viewer-versions. Then every fix-up's Legal text at every version from 1 on (over 2,000 comparisons) |
| `rules/stockRegistry.test.ts` (9) | Every hand-port move that is a Go stock move used as is (13 in blackjack, including both deals on the generic `dealComponents`; at least one in each of the six) rebuilt from `stockMove` parts: all 22 corpus documents still pass every field. The registry's lookups and refusals. Every move the inventory calls ported resolves in the registry, every one it calls unported does not, and every ported entry names its proof |

Pass table of `rules/stock.test.ts`:

| document | versions | viewers | header | state | move | status | view.game | view.move | view.legality | view.choices | result |
|---|---|---|---|---|---|---|---|---|---|---|---|
| stockgame-full | 71 | 4 | ok 1 | ok 71 | ok 71 | ok 71 | ok 284 | ok 284 | ok 284 | ok 284 | PASS |
| stockgame-three | 51 | 5 | ok 1 | ok 51 | ok 51 | ok 51 | ok 255 | ok 255 | ok 255 | ok 255 | PASS |
| seatgame-seated | 17 | 6 | ok 1 | ok 17 | ok 17 | ok 17 | ok 102 | ok 102 | ok 102 | ok 102 | PASS |
| seatgame-slow | 14 | 5 | ok 1 | ok 14 | ok 14 | ok 14 | ok 70 | ok 70 | ok 70 | ok 70 | PASS |
| seatgame-unseated | 1 | 4 | ok 1 | ok 1 | ok 1 | ok 1 | ok 4 | ok 4 | ok 4 | ok 4 | PASS |

Is the replay honest? Three mutations were made one at a time and
reverted: `DealComponentsUntilGameCountLeft`'s `<=` made `<` fails both
stockgame documents on state, move, status and the views (the replay stops
at a missing fix-up); `ColorValue` suppressed fails `view.game` on all 539
stockgame viewer-versions; and the first run, before the deal and
move-components fallback help texts were written, failed the header on
exactly that field.

The six hand ports are unchanged; `games/conformance.test.ts` still passes
all 22 documents (part of the unit count).

## Inventory: every stock move type and behavior in Go

| Go | Kind | Ported | TypeScript | Verified by | Notes |
|---|---|---|---|---|---|
| `moves.Default` | move | yes | `defaultMove` | base of every corpus move: conformance/ via games/conformance.test.ts; rules/stock.test.ts: Done and CloseAllSeats, whose Legal is Default.Legal plus their own checks | The game supplies Apply; the registry returns no apply part. |
| `moves.FixUp` | move | yes | `fixUp` | rules/stock.test.ts: testdata/stock/seatgame-*.json.gz (Begin, Reopen, Finish built from stockMove('moves.FixUp')); testdata/stock/fixups.json | The game supplies Apply. |
| `moves.FixUpMulti` | move | yes | `fixUpMulti` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz (every deal, collect, move-components and DefaultComponent move is one) | The game supplies Apply. |
| `moves.CurrentPlayer` | move | yes | `currentPlayerMove` | conformance/ (every hand port); rules/stock.test.ts: Place Card and Call Cleanup built from stockMove('moves.CurrentPlayer') | The game supplies Apply. |
| `moves.AnyPlayer` | move | yes | `anyPlayerLegal` | conformance/werewolf | Registry spec has base 'anyPlayer' and legal anyPlayerLegal; the game supplies Apply. |
| `moves.AdminPlayer` | move | **no** | — | — | Needs behaviors.GameAdministrator, not ported. |
| `moves.Done` | move | yes | `done` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz (Pass) |  |
| `moves.NoOp` | move | yes | `noOp` | rules/stock.test.ts: testdata/stock/stockgame-full.json.gz; testdata/stock/fixups.json |  |
| `moves.FinishTurn` | move | yes | `finishTurn` | conformance/ (memory, blackjack, tictactoe, pig); rules/stockRegistry.test.ts (corpus with registry parts); testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.ForceFinishTurn` | move | yes | `forceFinishTurn` | conformance/blackjack; rules/stockRegistry.test.ts (corpus with registry parts); rules/rejections.test.ts |  |
| `moves.StartPhase` | move | yes | `startPhase` | conformance/ (blackjack, werewolf); rules/stockRegistry.test.ts (corpus with registry parts); testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json | BeforeLeavePhase/BeforeEnterPhase hooks are not called (see the hook entries); a tree-enum leaf check at boot is not ported. |
| `moves.SeatPlayer` | move | yes | `seatPlayer` | conformance/ (all six games); rules/stockRegistry.test.ts (corpus with registry parts); testdata/stock/seatgame-*.json.gz | Does not mark the first seated player game admin (behaviors.GameAdministrator is not ported). |
| `moves.ActivateInactivePlayer` | move | yes | `activateInactivePlayer` | conformance/ (memory, blackjack, tictactoe, pig, werewolf); rules/stockRegistry.test.ts (corpus with registry parts) |  |
| `moves.InactivateEmptySeat` | move | yes | `inactivateEmptySeat` | conformance/ (blackjack, werewolf); rules/stockRegistry.test.ts (corpus with registry parts); testdata/stock/seatgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.WaitForEnoughPlayers` | move | yes | `waitForEnoughPlayers` | conformance/ (blackjack, werewolf); rules/stockRegistry.test.ts (corpus with registry parts) | WithRequireExplicitStart (DefaultRoundSetup's WithManualStart) is not ported; stockMove returns undefined for it. |
| `moves.CloseEmptySeat` | move | yes | `closeEmptySeat` | rules/stock.test.ts: testdata/stock/seatgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.ActivateEmptySeat` | move | yes | `activateEmptySeat` | rules/stock.test.ts: testdata/stock/seatgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.ActivateFilledSeat` | move | yes | `activateFilledSeat` | rules/stock.test.ts: testdata/stock/seatgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.CloseAllSeats` | move | yes | `closeAllSeats` | rules/stock.test.ts: testdata/stock/seatgame-*.json.gz (accepted, and refused below target in viewer legality) | WithRequireAdmin with a GameAdministrator player state is not ported (the behavior is not). |
| `moves.DealCountComponents` | move | yes | `dealComponents` | conformance/blackjack (dealCountComponents, and dealComponents via rules/stockRegistry.test.ts); testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json | The hand ports keep dealCountComponents; the registry uses dealComponents({kind}). Both pass blackjack's corpus. |
| `moves.DealComponentsUntilPlayerCountReached` | move | yes | `dealComponents` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.DealComponentsUntilGameCountLeft` | move | yes | `dealComponents` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.DealAllComponents` | move | yes | `dealComponents` | rules/stock.test.ts: testdata/stock/stockgame-full.json.gz; testdata/stock/fixups.json | Go's fallback name and help text swap the two stack names; reproduced. |
| `moves.CollectCountComponents` | move | yes | `dealComponents` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.CollectComponentsUntilPlayerCountLeft` | move | yes | `dealComponents` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.CollectComponentsUntilGameCountReached` | move | yes | `dealComponents` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.CollectAllComponents` | move | yes | `dealComponents` | rules/stock.test.ts: testdata/stock/stockgame-full.json.gz; testdata/stock/fixups.json |  |
| `moves.MoveCountComponents` | move | yes | `moveComponents` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json | Only game-stack paths are exercised; player paths (player.X, players[move.F].X) are resolved but unproven. |
| `moves.MoveComponentsUntilCountReached` | move | yes | `moveComponents` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.MoveComponentsUntilCountLeft` | move | yes | `moveComponents` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.MoveAllComponents` | move | yes | `moveComponents` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz; testdata/stock/fixups.json |  |
| `moves.ShuffleStack` | move | yes | `shuffleStack` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz (the shuffled order is in every later state); testdata/stock/fixups.json |  |
| `moves.ShuffleDiscardIntoDraw` | move | yes | `shuffleDiscardIntoDraw` | conformance/blackjack (scenario-reshuffle); rules/stockRegistry.test.ts (corpus with registry parts) | WithDrawDiscardPairField (a named pair) is not ported; stockMove returns undefined for it. |
| `moves.DrawToPlayer` | move | yes | `drawToPlayer` | conformance/blackjack (Current Player Hit embeds it) | Registry reads ResolvedDrawSource/ResolvedDrawDestination or GameProperty/PlayerPropety; the DrawDiscardPair and single-stack inference Go does at boot is the exporter's to resolve. |
| `moves.MoveComponentToSlot` | move | yes | `moveComponentToSlot` | conformance/ (memory Reveal Card, tictactoe Place Token) |  |
| `moves.DefaultComponent` | move | yes | `defaultComponent` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz (Crown Card without, Mark Corner with a legal type); testdata/stock/fixups.json (its Legal error texts at every version) | The components' Legal is DeckSpec.componentLegal; the game supplies Apply. Checkers' Crown Token is one. |
| `moves.RoundRobin` | move | yes | — | rules/stock.test.ts: every deal and collect move is roundRobinParts | Abstract: needs the game's ConditionMet, PlayerConditionMet and RoundRobinAction. Ported as roundRobinParts(hooks); not in the registry. |
| `moves.RoundRobinNumRounds` | move | yes | — | rules/stock.test.ts: DealCountComponents and CollectCountComponents | Abstract, as RoundRobin; its ConditionMet is in dealComponents. |
| `moves.ApplyUntil` | move | **no** | — | — | Abstract: needs the game's ConditionMet. Its Legal is inside moveComponents and roundRobinParts, not exposed alone. |
| `moves.ApplyUntilCount` | move | **no** | — | — | Abstract: needs Count/TargetCount. |
| `moves.ApplyCountTimes` | move | **no** | — | — | Abstract; countMovesApplied is inside moveComponents (MoveCountComponents). |
| `moves.AllPlayersSubmitted` | move | **no** | — | — | Needs behaviors.PlayerSubmission. |
| `moves.ResetAllPlayerSubmissions` | move | **no** | — | — | Needs behaviors.PlayerSubmission. |
| `moves.Increment` | move | **no** | — | — |  |
| `moves.AdvanceToken` | move | **no** | — | — | Needs behaviors.LocationBehavior. |
| `moves.HopAlongPath` | move | **no** | — | — | Needs behaviors.LocationBehavior and graph shortest paths. |
| `moves.MoveOnGraph` | move | **no** | — | — | Needs behaviors.LocationBehavior and graph shortest paths. |
| `moves.ReplenishMarket` | move | **no** | — | — | Needs behaviors.FaceUpMarket. |
| `moves.SelectColor` | move | **no** | — | — | The PlayerColor behavior is ported; the gathering move is not. |
| `moves.SelectRole` | move | **no** | — | — |  |
| `moves.SelectTeam` | move | **no** | — | — | Needs behaviors.PlayerTeam. |
| `moves.RecordCurrentPlayerChoice` | move | **no** | — | — | Move choice recording (move_choice_recording.go) is not ported. |
| `moves.DefaultRoundSetup` | helper | yes | `defaultRoundSetup` | conformance/blackjack | A function returning a progression group, not a move type. WithManualStart is not ported. |
| `behaviors.CurrentPlayerBehavior` | behavior | yes | `currentPlayerProps` | conformance/ (all six games); testdata/stock/* |  |
| `behaviors.PhaseBehavior` | behavior | yes | `phaseProps` | conformance/ (blackjack, werewolf); testdata/stock/* |  |
| `behaviors.RoundRobin` | behavior | yes | `roundRobinProps` | conformance/blackjack; testdata/stock/stockgame-*.json.gz |  |
| `behaviors.Seat` | behavior | yes | `seatProps` | conformance/ (all six games); testdata/stock/seatgame-*.json.gz | Player trait seat. |
| `behaviors.InactivePlayer` | behavior | yes | `inactivePlayerProps` | conformance/ (all six games); testdata/stock/seatgame-*.json.gz | Player trait inactive. |
| `behaviors.ScoreBehavior` | behavior | yes | `scoreProps` | conformance/ (pig, blackjack) | Player trait gameScore: scoreGameScore. |
| `behaviors.PlayerElimination` | behavior | yes | `eliminationProps` | conformance/ (pig, blackjack) | Player trait elimination. |
| `behaviors.PlayerRole` | behavior | yes | `roleProps` | conformance/werewolf | Player trait role; computed RoleValue and AvailableRoles. |
| `behaviors.DrawDiscardPair` | behavior | yes | `game.traits.drawDiscardPair` | conformance/blackjack | Only the anonymous pair; named pairs are not ported. |
| `behaviors.PlayerColor` | behavior | yes | `playerColorProps` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz (computed Color, ColorValue, AvailableColors in every view) | Player trait color. OwnsToken, Token, TokenFromDeck and TokenSpaceIndex are not ported. |
| `behaviors.ComponentColor` | behavior | yes | `componentColorProps` | rules/stock.test.ts: testdata/stock/stockgame-*.json.gz | DeckSpec.props. |
| `behaviors.PlayerOrderBehavior` | behavior | **no** | — | — | GameTraits.playerOrder and the computed PlayerOrder exist in the rules layer but nothing verifies them against Go, and there is no property set. |
| `behaviors.PlayerTeam` | behavior | **no** | — | — | Includes the computed TeamValue and AvailableTeams. |
| `behaviors.GameAdministrator` | behavior | **no** | — | — | Includes SeatPlayer marking the first seated player admin and the computed IsGameAdmin. |
| `behaviors.FaceUpMarket` | behavior | **no** | — | — |  |
| `behaviors.LocationBehavior` | behavior | **no** | — | — |  |
| `behaviors.MoveBudget` | behavior | **no** | — | — |  |
| `behaviors.PlayerSubmission` | behavior | **no** | — | — |  |
| `interfaces.BeforeLeavePhaser` | hook | **no** | — | — | StartPhase.Apply does not call BeforeLeavePhase. |
| `interfaces.BeforeEnterPhaser` | hook | **no** | — | — | StartPhase.Apply does not call BeforeEnterPhase. |
| `moves.RepeatFromProp` | hook | **no** | — | — | A progression counter read from a property; rules/progression.ts (not this stream). |
| `interfaces.LegalComponent` | hook | yes | `DeckSpec.componentLegal` | testdata/stock/fixups.json (DefaultComponent's verdicts) |  |
| `enum.RangeEnum` | library | yes | `lib/enumrange.ts` | lib/lib.test.ts: testdata/lib.vectors.json | AddRange, RangeDimensions, ValueToRange, RangeToValue, SetRangeValue, ValueFromString. |
| `graph.NewGridConnectedness` | library | yes | `lib/graph.ts gridConnectedness` | lib/lib.test.ts: testdata/lib.vectors.json (every Direction filter, checkers' two graphs, the errors) | Also NewEnumGridConnectedness, NeighborsByKey (ascending; Go's map order is unspecified), ConnectedByKey. |
| `graph.ShortestPath` | library | **no** | — | — | Also Distance, edge weights, hand-built graphs (graph.New, AddEdge). Go's tie-break among equal paths follows map order. |
| `components/dice` | library | yes | `lib/dice.ts` | lib/lib.test.ts: testdata/lib.vectors.json; lib/lib.test.ts: equal to games/pig.ts's inlined roll | BasicDie, DefaultDie, Min, Max, DynamicValue.Roll (SelectedFace, Value, RollCount). |
| `components/playingcards` | library | yes | `lib/playingcards.ts` | lib/lib.test.ts: testdata/lib.vectors.json; lib/lib.test.ts: equal to games/blackjack.ts's inlined deck and enums | NewDeck, NewDeckMulti, the Suit and Rank enums, generic values, Card.String. |

Counts: moves 37 ported, 15 not; behaviors 11 ported, 7 not; libraries 4
ported, 1 not (graph shortest paths); hooks 1 ported (LegalComponent), 3 not.

## Checkers

Everything checkers uses from this stream exists and is proven:
`SeatPlayer`, `ActivateInactivePlayer`, `FixUpMulti`, `StartPhase`,
`DefaultComponent` (with `WithSourceProperty`), `FinishTurn`,
`CurrentPlayer`; `CurrentPlayerBehavior`, `PhaseBehavior`, `Seat`,
`InactivePlayer`, `PlayerColor`, `ComponentColor`; `MustAddRange("spaces",
8, 8)`, `RangeValue`, `RangeToValue`, `ValueToRange`;
`MustNewEnumGridConnectedness` with checkers' two filter pairs and
`NeighborsByKey` (`lib.vectors.json` has checkers' exact two graphs).

What checkers still lacks, none of it this stream's:

- **A game-registered predicate** (`checkers.spaceIsBlack`, via
  `ConfigurePredicateConstructors`): `GameSpec` has no hook for one, so
  `compileGame` cannot build Move Token's plan. Workstream 6 (types,
  compile) or the loader.
- **Stack constraints** on `CapturedTokens` (`same(Color)`) and the by-key
  stack calls (`SwapComponentsByKey`, `ComponentAtKey`): workstream 8.
- **`ContainingImmutableStack`** in `token.Legal`: the facade's
  `containingStack` (workstream 7).
- The facade's `rangeValue`, `rangeToValue`, `valueToRange`,
  `neighborsByKey` bodies still throw (workstream 7); they are thin calls of
  `lib/` (above).

## Judgment calls

1. **Vectors as conformance documents.** Rather than a bespoke format, the
   new stock moves are proven the way the hand ports are: a real Go game is
   played and exported with `boardgame-util/lib/conformance`, and
   `testing/conformance.ts` replays it. That compares states, move records,
   every viewer's view, legality with Go's texts, choices and the header
   (names, help texts, schemas, fingerprints). A conformance document never
   shows a fix-up's error text, so `fixups.json` adds it.
2. **Two Go games, not one per move.** `stockgame` and `seatgame` are in
   separate packages because Go's moves-with-no-progression cache is keyed
   by game name in a package variable (engine-rules bug 8), and two
   delegates with one name would share it. `ActivateEmptySeat` and
   `InactivateEmptySeat` are in different phases because Go refuses to boot
   with both in one scope.
3. **Fix-up legality skips version 0.** Go's `MoveRecords(upToVersion < 1)`
   is the whole history (engine-rules divergence 2), so at version 0 Go
   judges a progression against the finished game; `stockgame-three` v0
   showed it. Versions 1 on are compared.
4. **`dealComponents` beside `dealCountComponents`.** The hand ports keep the
   latter; the registry uses the generic one for all eight types, and
   `stockRegistry.test.ts` runs blackjack's corpus on it.
5. **Configuration values.** The exporter (workstream 1) has not fixed how an
   `enum.Enum` in a configuration (`WithPhaseToStart`'s, `WithLegalType`'s)
   serializes. The registry accepts a name string or an object with
   `Name` or `Values`; the phase name only feeds the fallback move name,
   which `MoveData.Name` overrides anyway.
6. **Unproven helpers were removed**, not shipped: `PlayerColor.OwnsToken`
   and `CSSColorForKey` were written and then deleted, since nothing ran
   them against Go.
7. **Neighbor order.** Go's `Neighbors` ranges over a map, so its order is
   unspecified; `lib/graph.ts` returns ascending and the vectors compare
   sorted lists. Every Go caller in the repository uses the result as a set.
8. **`ILLEGAL_VALUE`** is `9223372036854775807` written out, which a double
   rounds to 2^63; the vectors parse to the same number.
9. **Move-components moves contribute no stack-constraints check**: Go calls
   `legalWithoutStackConstraints`, so the spec names no source or
   destination property (the header matched with that).

## Unverified

- Player-scoped stack paths (`player.X`, `players[move.F].X`) in
  `ShuffleStack`, `DefaultComponent` and the move-components moves: only
  game stacks are exercised.
- `CloseAllSeats` with `WithRequireAdmin`, and `DefaultComponent` with a
  `LegalType` but no enum: written to Go's code, not run.
- Round robins whose `PlayerConditionMet` skips every player for a full
  circuit (the `counter > len` branch): read, not run.
- `ActivateEmptySeat`/`ActivateFilledSeat`'s "Player state didn't implement
  interfaces.PlayerInactiver" path.
- Dice with no faces (`Intn(0)` panics in Go; the engine's `intn(0)` text is
  not compared).

## Where the proposal was wrong about the code

- "Of 18 behaviors, 9 are ported": the rules layer had 10 (`DrawDiscardPair`
  as a trait), now 12 with the two color behaviors; `PlayerOrderBehavior` has
  a trait but no proof, so it is listed unported.
- "Package `moves` declares about 50 stock move types": 52 exported move
  types, plus `DefaultRoundSetup`, a function.
- The `Stock.Type` doc in `rules/data.ts` says the bare name
  (`"StartPhase"`); the brief says `moves.StartPhase`. `stockMove` accepts
  both.
- The proposal's workstream table gives `rules/stockRegistry.ts` to
  workstream 6; the brief gave it here. Only the registry body changed;
  `StockParts` and `stockMove`'s signature are as frozen.
- Color behaviors were said to need "their computed values": `Color` (CSS)
  was already in `view.ts`; `ColorValue` and `AvailableColors` were not.

## Edits outside my files

All additive:

- `server/static/src/engine/rules/types.ts`: `DeckSpec.componentLegal?`
  (interfaces.LegalComponent, for DefaultComponent) and
  `PlayerTraits.color?` (behaviors.PlayerColor).
- `server/static/src/engine/rules/view.ts`: with `traits.color`, the
  framework's computed `AvailableColors` (global) and `ColorValue` (per
  player, where the viewer may see the value), as
  `base.GameDelegate.FrameworkComputed*Properties` adds them. Nothing changes
  for a game without the trait.
- `server/static/src/engine/testdata/gen/regen.sh`: `libvec stockvec` added
  to the loop, and their comment lines.
