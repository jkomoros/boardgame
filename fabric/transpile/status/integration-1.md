# Integration 1: engine-gap table and loader tests on real definitions

Base `e7c02e2e`. Two failures left by merging the static half, the stock
registry, the facade and the loader.

## Failure 1: the engine-gap table was stale

`subset.EngineGaps` listed as gaps six facade calls `engine/go/api.ts` now
exports and 18 stock move types and 2 behaviors that
`rules/stock-inventory.json` marks ported.

What changed:

- `boardgame-util/lib/transpile/subset/gaps.go`: the table is now what the
  inventory says. A stock move type is a gap exactly when the inventory does
  not have it both `ported` and with a `registryName`; a behavior exactly when
  it is not `ported`. 17 move gaps, 7 behavior gaps, no facade gaps.
- `TestEngineGapsMatchStockTS` (the "mentioned in rules/*.ts" heuristic) is
  replaced by `TestEngineGapsMatchStockInventory`, which reads the JSON and
  fails if: Go has a stock move type or behavior the inventory lacks, or the
  reverse; a kind disagrees; an entry is ported with no `verifiedBy`; a
  gap/non-gap disagrees with the rule above; or a facade name in the table is
  exported by `engine/go/*.ts` or `engine/lib/*.ts`. It logs
  `stock move types 52 (17 gaps), behaviors 18 (7 gaps)`.
- `TestTableIsWellFormed` no longer expects the six proposal "none (gap)"
  rows to be gaps; it checks they are still rows, and that any facade gap
  names a row's facade.
- Fixture (`subset/testdata/fixture/fixture.go`): `readFramework`
  (ValueToRange), `swapByKey` (was `engineGap`, SwapComponentsByKey) and
  `(*moveY).Apply` (DefaultComponent's ComponentIndex) now want `none`. The
  `engine-gap` reason stays covered by `engineGap` (a stock `Apply` call on
  embedded `moves.Increment`) and `readsUnportedBehavior` (a field read
  through embedded `behaviors.PlayerTeam`).
- `lib/gamedef/static_test.go`: the staticgame fixture was written against
  `componentAtKey` and `moves.DefaultComponent` being gaps. `loadFixture`
  restores those two as gaps for the fixture's load only (`t.Cleanup` removes
  them), so the fixture still covers the engineGap bucket, its callee
  propagation and an unported stock ancestor, independent of what the engine
  ports next. The fixture itself is unchanged.

### Judgment calls

- **RoundRobin and RoundRobinNumRounds** (inventory: ported as
  `roundRobinParts(hooks)`, `registryName: null`): **gap**. `gameFromData`
  builds a stock move only through `stockMove`, which returns undefined for
  them, so a game function that selects through one or calls its stock
  `Apply`/`Legal` has no move in the loaded game, however faithfully it is
  translated. It stops being a gap when the registry can build it (the
  registry would need the game's `ConditionMet`, `PlayerConditionMet` and
  `RoundRobinAction` as bodies). No example uses either directly.
- **HopAlongPath**: the inventory says not ported (needs `LocationBehavior`
  and shortest paths); it is a gap, although the TypeScript mentions it.
- **Configurations `stockMove` cannot build** (`WithRequireExplicitStart`,
  `WithManualStart`, `WithDrawDiscardPairField`, `DrawToPlayer` with inferred
  stacks, `MoveComponentToSlot` with a discovered slot field): not detected.
  The first three are statically detectable from the `auto.Config` options
  `lib/gamedef/static.go` already walks, but marking `StockPorted` false for
  them is a change in `static.go`, which I do not own, and no example uses
  them. Listed as known holes in the `EngineGaps` comment, with
  `PlayerColor`'s unported methods (refused anyway as `unknown-call`) and the
  `BeforeLeavePhaser`/`BeforeEnterPhaser` hooks.

### Measured table, before and after

`./scripts/go-local run ./boardgame-util lint --declarative ./examples/...`;
goldens regenerated with `./scripts/go-local run ./boardgame-util/lib/lint/testdata/gen/declarative`.
Only checkers changes; only the engine-feature column moves.

| Game | Functions | In subset | Outside | Engine feature before | after | Not ported |
|---|---|---|---|---|---|---|
| memory | 17 | 16 | 1 | 0 | 0 | 11 |
| blackjack | 23 | 23 | 0 | 0 | 0 | 1 |
| pig | 10 | 10 | 0 | 0 | 0 | 1 |
| tictactoe | 11 | 11 | 0 | 0 | 0 | 5 |
| werewolf | 16 | 16 | 0 | 0 | 0 | 1 |
| debuganimations | 24 | 24 | 0 | 0 | 0 | 1 |
| checkers | 20 | 19 | 1 | **8** | **0** | 0 |

No function is still blocked on an engine feature. The eight checkers
functions that were (`(*moveCrownToken).Apply`, `(*moveMoveToken).Apply` and
`LegalCustom`, `(*movePlaceToken).DefaultsForState`, `(*token).Legal`,
`AllNextSpaces`, `LegalCaptureSpaces`, `FreeNextSpaces`) needed
`componentAtKey`, `swapComponentsByKey`, `rangeValue`, `rangeToValue`,
`valueToRange`, `neighborsByKey`, `moves.DefaultComponent` and
`behaviors.ComponentColor`, all now present. The two functions outside the
subset are unchanged: `memory.(*moveStartHideCardsTimer).Apply`
(`package-var-read`) and `checkers.(*gameDelegate).ConfigurePredicateConstructors`
(`predicate-constructor`). `TestDeclarativeMaxRefused` now expects 1 refused
for checkers (was 9).

`conformance/definitions/checkers.json` regenerated (`gamedef export`): the
eight functions are `verdict: subset` with no refusals; `gamedef export
--check` reports current. The other six definitions are unchanged.

## Failure 2: three loader tests on real definitions

All three were in the test's adapter (`rules/data.test.ts`
`handPortBodies`), not in the loader, exporter or registry.

- **memory** `end.playerScore` for `memory.(*gameDelegate).PlayerScore`.
  Go declares `PlayerScore` as `return pState.(*playerState).GameScore()`;
  the hand port leaves `end.playerScore` unset and lets the engine fall back
  to the `gameScore` trait (`rules/engine.ts checkGameFinished`). Test body:
  the hand port's `player.traits.gameScore`.
- **debuganimations** `end.check` for
  `debuganimations.(*gameDelegate).CheckGameFinished`. Go returns
  `false, nil`; the hand port has no `end`, which the engine reads as never
  finished. Test body: `() => ({ finished: false, winners: [] })`.
- **werewolf** "two different hand-port functions for
  `werewolf.(*moveCastVote).Apply`". One Go type configured twice (Cast Vote,
  Cast Night Vote; likewise `moveResolveVotes` and
  `moveResolveVotesOnTimeout`), while the hand port builds each config from a
  factory (`castVote(options)`, `resolveVotesMove`, `resolveOnTimeout`), so
  each config has its own closure of the same function expression. The
  adapter now accepts a second hand-port function for a FnRef when its source
  text equals the first's, and keeps the first; the corpus then runs every
  config sharing the FnRef on that one body. In all three factories `apply`
  and `legal` read nothing from `options`.

The two test bodies are in a `TEST_BODIES` table keyed by FnRef, used only
when the hand port has no function for that role; the diagnostic line now
also lists the test bodies used.

Result (`node --test src/engine/rules/data.test.ts`), both modes for each:

| Game | Documents | Versions | Stand-in stock parts (bodies mode) | Test bodies |
|---|---|---|---|---|
| memory | 6 | 172 | 1 of 7: Reveal Card | PlayerScore |
| blackjack | 7 | 419 | 0 of 23 | none |
| tictactoe | 4 | 58 | 1 of 4: Place Token | none |
| pig | 3 | 283 | 0 of 6 | none |
| debuganimations | 1 | 16 | 0 of 13 | CheckGameFinished |
| werewolf | 1 | 13 | 0 of 11 | none |

Every move through `move:` passes for all six too.

**Stand-ins still needed (an engine gap to record):** both are
`moves.MoveComponentToSlot` configured without `WithSlotField`. Go discovers
the slot field at boot as the move's only int property
(`resolveSlotField`, `moves/move_component_to_slot.go`), so the exported
configuration has no `SlotField`, and `stockMove` throws
`MoveComponentToSlot needs configuration SlotField`. The fix belongs in
the registry (discover the single int field from `MoveData.Fields`, which
`stockMove` does not receive today) or the exporter (write a resolved slot
field, as `DrawToPlayer`'s `ResolvedDrawSource` does). Not fixed here: both
are other streams' files and the test passes through the stand-in.

## Test results

- `./scripts/go-local build ./...`: passes.
- Go tests: `lib/gamedef/...`, `lib/lint/...`, `lib/transpile/...`,
  `lib/conformance/...` all pass.
- `gamedef export --check`: current. `conformance export --check`: current.
- `npm run type-check`: passes. `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json`: passes.
- `npm run test:unit`: 1,253 tests, 1,253 pass, 0 fail, 0 skipped. (The
  brief's base was 1,253 with 1,247 pass and 3 fail; I did not re-run the
  base to account for the other 3.)
- `npm run test:layers`: 11 of 11.

## Unverified

- The rule "ported means `ported` and a `registryName`" trusts the
  inventory; nothing here checks that `stockMove` really builds every
  registered type for every configuration (see the known holes).
- The source-text equality for shared closures is a structural check; that
  the closures behave alike is shown only by the corpus.

## Edits outside my files

- `boardgame-util/lib/transpile/subset/testdata/fixture/fixture.go`: the
  subset tests' fixture, whose wants depend on the gap table (four wants
  changed, one function renamed, two functions and two types added).
- `conformance/definitions/checkers.json`: regenerated by the exporter.
