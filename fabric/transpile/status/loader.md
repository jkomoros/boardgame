# Loader (workstream 6): status

Workstream 6 of `fabric/TRANSPILE-PROPOSAL.md`: the TypeScript engine
builds a `GameSpec` from an exported `GameData` plus a function table.
Based on `d2cb55e4`; one local commit. Nothing pushed, no PR.

## What was built

- `rules/data.ts`: `gameFromData(data, bodies)`, fully implemented, plus
  `callStockApply`, `moveOverrideRef`, `stockOverrideRef`, and the "Body
  calling convention" section (repeated below).
- `rules/compile.ts`: honors `MoveSpec.plan` (verbatim; `preconditions`,
  `contributes`, `suppress` and the derived base checks are then ignored),
  `SubStateSpec.initialValues` (deep-copied; `initial` wins), and
  `DeckSpec.generic` (written to `DeckDef.GenericValues` only when set, so
  hand ports' definitions are unchanged).
- `rules/state.ts`: `state.component(deck, GENERIC_COMPONENT)` returns the
  deck's generic values, `{}` when none were set (Go's `Values()` is then
  nil; `{}` is the engine's existing answer for a valueless component).
- `rules/types.ts`: doc updates for the three frozen fields; new
  `MoveSpec.stock?: MoveStockParts` and the `MoveStockParts` type (the
  embedded stock move's `apply`, `legal`, `defaults`), which the loader sets
  and the facade's `stockApply` reads through `callStockApply`.
- `definition.ts`: doc update for `GenericValues`.
- `rules/testdata/pig.gamedata.json`, `rules/testdata/blackjack.gamedata.json`:
  hand-made stand-in `GameData` (see below).
- `rules/testdata/gen/movefacts/main.go` and its output
  `rules/testdata/movefacts/<game>.json` (all six games): per move, what the
  live Go manager reports: name, help text, fix-up, `InstalledLegalPhases`,
  `DescribeProgression`, `AllowMultipleInProgression`, `LegalPlanSpecs`.
  Deterministic (re-run, byte-identical). Run with
  `./scripts/go-local run ./server/static/src/engine/rules/testdata/gen/movefacts`.
- `rules/data.test.ts`: the proof (below).

## Exported interfaces (rules/data.ts)

```ts
export type Bodies = (ref: string) => ((...args: never[]) => unknown) | undefined;
export function gameFromData(data: GameData, bodies: Bodies): GameSpec;
export function moveOverrideRef(moveName: string): string;   // 'move:' + name
export function stockOverrideRef(moveName: string): string;  // 'stock:' + name
export function callStockApply(state: RulesState, m: MoveInstance): string | null;
```

`rules/types.ts` adds `MoveSpec.stock?: MoveStockParts` and
`interface MoveStockParts { apply?; legal?; defaults? }` (same signatures as
`StockParts` in `rules/stockRegistry.ts`, without `spec`).

For workstream 7 (facade): `stockApply(state, m)` in `go/api.ts` should be
`return callStockApply(state, m);` (it finds the move by name, case
insensitively, runs `spec.stock.apply`, returns a thrown `Error`'s message
or null, and throws if the move has no stock apply). `go/api.ts` may import
`../rules/data.ts`; `rules/` may not import `go/` (rules/purity.test.ts's
import pattern), which is why the loader detects a facade `StackRef`
structurally.

## Body calling convention (verbatim from rules/data.ts)

gameFromData resolves every FnRef the data names through `bodies(ref)`
and calls the function it gets with exactly the arguments below. A
translated body (bodies.gen.ts), an agent-written one, and a hand port's
own function all follow it. `state` is always the RulesState of
rules/state.ts (the facade's functions take it first); `m` is the
MoveInstance being judged or applied (`m.fields.X` is Go's `m.X`); a
player is its index (Go's `*playerState`, `ImmutableSubState`,
`PlayerIndex`).

Go's error results follow the proposal's "An error result, by role":
where the signature below says THROWS, a Go `return err` with a non-nil
err is `throw new Error(text)`; for convenience the loader also accepts
a returned non-null string there and throws it, so `return text` and
`return null` (Go's `return nil`) work too. Where it says a `string |
null` result, the string IS the error and nothing is thrown. A body that
throws anywhere is a Go panic, and the engine treats it as it treats a
throw from a hand port.

| Role (FunctionEntry.Role) | Where the FnRef is | Signature |
|---|---|---|
| apply | MoveData.Bodies.Apply | `(state: RulesState, m: MoveInstance) => void \| string \| null`, THROWS |
| legal | MoveData.Bodies.Legal | `(state: RulesState, m: MoveInstance, proposer: number, base: () => string \| null) => string \| null` |
| legalCustom | MoveData.Bodies.LegalCustom | `(state: RulesState, m: MoveInstance, proposer: number) => string \| null \| LegalVerdict` |
| defaults | MoveData.Bodies.Defaults | `(state: RulesState, m: MoveInstance) => void` |
| computed (global) | Computed[i].Compute, Scope 'global' | `(state: RulesState) => unknown` (the value) |
| computed (player) | Computed[i].Compute, Scope 'player' | `(state: RulesState, player: number) => unknown` |
| setUpBegin | SetUp.Begin | `(state: RulesState, variant: Readonly<Record<string, string>>) => void \| string \| null`, THROWS |
| distribute | SetUp.Distribute | `(state: RulesState, deck: string, index: number) => Stack \| { state: RulesState; at: StackRef }`, THROWS |
| setUpFinish | SetUp.Finish | `(state: RulesState) => void \| string \| null`, THROWS |
| conditionMet | End.ConditionMet | `(state: RulesState) => boolean` |
| playerScore | End.PlayerScore | `(state: RulesState, player: number) => number` |
| checkGameFinished | End.Check | `(state: RulesState) => [finished: boolean, winners: number[] \| null] \| { finished: boolean; winners: number[] }` |
| currentPlayerIndex | CurrentPlayerIndex | `(state: RulesState) => number` |
| turnDone | Traits.Player.Turn.TurnDone | `(state: RulesState, player: number) => string \| null` |
| resetForTurnStart | Traits.Player.Turn.ResetForTurnStart | `(state: RulesState, player: number) => void \| string \| null`, THROWS |
| resetForTurnEnd | Traits.Player.Turn.ResetForTurnEnd | `(state: RulesState, player: number) => void \| string \| null`, THROWS |
| gameScore | Traits.Player.GameScore (not 'score') | `(state: RulesState, player: number) => number` |
| playerOrder | Traits.Game.PlayerOrder | `(state: RulesState) => readonly number[] \| null` |
| counter | CounterData { Kind: 'custom', Count } | `(currentCount: number, length: number) => string \| null` (Go's ValidCounter) |
| helper | FunctionEntry only | Never resolved by the loader: bodies call helpers directly, with whatever signature their translation has (`state` first) |
| predicate | Predicates[i].Construct | Not supported in v1: a non-empty Predicates is a GameSpecError |

Details that the table cannot carry:

- legal: `base` is the embedded stock move's Legal (Go's
  `m.<Stock>.Legal(state, proposer)`), which for moves.Default and the
  bases that add nothing is the move's base checks (phase, progression,
  current player, the plan when one is declared). A body that never
  calls it skips them, as in Go.
- apply: Go's `m.<Stock>.Apply(state)` is the facade's `stockApply(state,
  m)`, which returns the stock apply's error text or null. The loader
  stores the stock parts on the MoveSpec (`MoveSpec.stock`), and
  `callStockApply(state, m)` is what the facade calls.
- defaults: the stock move's DefaultsForState always runs first (the
  translator drops `m.<Stock>.DefaultsForState(state)` as the first
  statement), then the body.
- distribute: the result may be an engine Stack or a facade StackRef
  (`{ state, at }`, at an engine StackRef location); the loader turns the
  latter into the stack of the state it was handed.
- checkGameFinished: Go's two results as a tuple, the translator's shape
  for multiple results; the object form is accepted too. A null winners
  list is empty.
- Bodies are called with `this` undefined; a body must not rely on it.

### Hooks beside FnRefs

`bodies` may also answer two keys that are not FnRefs, for moves the
engine cannot build from a stock type yet, and so that a generated game
can sit beside hand-ported moves:

- `move:<move name>` (moveOverrideRef): `() => MoveSpec`, a whole move,
  typically a hand port's. The loader takes the move's BEHAVIOR from it
  (`apply`, `legal`, `defaults`, `legalCustom`, `sourceProperty`,
  `destinationProperty`, `contributes`) and its CONFIGURATION from the
  data (name, help text, fix-up, base, fields, legal phases,
  progression, plan, flags, name sanitization, choices). The move's
  Bodies FnRefs and Stock are not resolved.
- `stock:<move name>` (stockOverrideRef): `() => StockParts`, used in
  place of `stockMove(Stock.Type, Stock.Config)` for that move; its
  Bodies are resolved as usual. With this hook a move whose stock type is
  not ported yet still runs the game's own bodies.

## How a move is built

Configuration always comes from `MoveData`: name, help text, `IsFixUp`,
`Base`, `Fields` (Input and Codec carried; an empty Codec is inferred),
`LegalPhases` (null or empty: none), `Progression` (each `move` leaf's
`allowMultiple` from the named move's `AllowMultipleInProgression`;
counters stay `CounterData`, `custom` resolved to its body),
`AllowMultipleInProgression`, `planEnabled = Plan.Declared`, `plan =
Plan.Specs` only when `Plan.Declared`, name sanitization, hidden animation
key, `Choices` (the header entry turned back into a `ChoiceSpec`; for
`enum-values`, `exclude` is the enum's values not among `candidateValues`),
`SeatsPlayer`, `GatheringStart`, `EndsPhaseProgression`.

Behavior: a `move:` override if present; else the stock parts (`stock:`
hook, else `stockMove`, else `{}` for `Stock: null`), with the game's
`Bodies` layered on as the convention says. Errors (all `GameSpecError`):
a missing body (message names the role and the FnRef), an unported stock
type (`stockMove` returns undefined), a move with no Apply from any source,
`Plan.HasCustom` disagreeing with whether a LegalCustom exists, a
LegalCustom without a declared plan, unknown field type/disposition/codec,
base, computed type/scope, progression kind, choice source, SchemaVersion
other than 1, non-empty `Predicates`, a `PhaseEnum` other than
`phase`/`Phase` or a `GroupEnum` other than `group` (compileGame finds them
by those names), and an `Initial.Dynamic` deck without dynamic values. A
`stockMove` that throws (today's stub) propagates its own error.

The rest of `GameData`: `Definition.Decks` → `decks` (components padded to
`Size`, `ComponentProps` → `props`, `GenericValues` → `generic`),
`Definition.Enums` → `enums`, `ComponentEnums`, `Constants`,
`Definition.Game/Player/DynamicComponentValues.Props` → props verbatim (Go's
map-form `Sanitize`), `Initial` → `initialValues`, `Traits` (GameScore
`'score'` → `scoreGameScore`), `Variants` (the header's variantConfig) →
`variants`, `Computed`, `LegalTemplates`, `SetUp`, `End` (only set when one
of its parts is), `CurrentPlayerIndex`, `DisplayName`. Ignored, as the
engine has no use for them: `Source`, `Description`, `GoType`,
`Traits.Game.RoundRobin` (its properties are in the Definition),
`Traits.Player.Color` (no engine trait), `Agents`, `Functions`.

## The proof (rules/data.test.ts) and results

For each GameData, with the hand port's own functions as bodies (mapped by
the role the data names each FnRef in), twice: once resolving `Bodies`
(moves whose stock type `stockMove` cannot build get the hand port's
stock-side parts through `stock:`), and once with every move through the
`move:` override. Each pass must (1) compile to the same summary as the
hand port, compared with `firstDifference`: display name, player counts,
normalized definition (as rules/engine.test.ts normalizes, without the
GameData additions), phase enum, merged templates, constants, variant
config, computed descriptors, move input schema and fingerprint, choice
schema and fingerprint, and per move name, help text, fix-up, base, fields,
field and name policies, declared, plan specs, custom tail, legal phases,
progression (counters compared by their results on 0..16), seat and
gathering flags; and (2) pass every conformance document of the game
through `runDocument`, all eight fields, every version.

Fixture checks against Go: each fixture's moves (name, help text, fix-up,
legal phases, progression, AllowMultipleInProgression, plan) deep-equal
`movefacts/<game>.json`; its `Definition` minus `GenericValues` and
`ComponentProps` deep-equals `testdata/defs/<game>.def.json`.

Real definitions: one subtest per hand port reads
`conformance/definitions/<game>.json`; all six are absent here, so all six
are skipped with "conformance/definitions/<game>.json is absent: the Go
exporter has not written it yet". When present they run the same proof.

Plus unit tests: a missing body is a GameSpecError naming the FnRef; an
unbuildable stock type without a stand-in throws; HasCustom disagreement is
a GameSpecError; generic values, component props and initial values come
through (`state.component('cards', -2)` is blackjack's generic card, pig's
generic die is `{}`, pig's empty state has TargetScore 100 and die Value
1); `callStockApply` returns null / the thrown text / throws with no stock
apply; a returned string from an apply-like body is thrown and a check
tuple `[true, null]` reads as `{ finished: true, winners: [] }`.

Outcomes, as run:

| Command | Outcome |
|---|---|
| `node --test src/engine/rules/data.test.ts` | 16 tests: 10 pass, 6 skipped (the absent definitions), 0 fail |
| pig fixture, both passes | 3 documents, 283 versions, all fields pass; 6 of 6 moves via `stock:` stand-ins in the bodies pass |
| blackjack fixture, both passes | 7 documents, 419 versions, all fields pass; 23 of 23 moves via `stock:` stand-ins in the bodies pass |
| `npm run test:unit` (Node 24) | 1184 tests: 1178 pass, 6 skipped, 0 fail (1168 before + 16 new) |
| `npm run type-check` | passes (both tsc runs) |
| `npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` | passes |
| `npm run test:layers` | 11 of 11 |
| `./scripts/go-local build ./...` | passes |
| `./scripts/go-local test ./boardgame-util/lib/conformance/...` | ok (both packages) |
| `./scripts/go-local run ./boardgame-util conformance export --check` | "The conformance corpus is current" |
| `go vet`, `gofmt -l` on the movefacts generator | clean |

The Node commands were run as `node` (v24.22.3, by absolute path) with the
exact arguments of the package scripts. sandbox.test.ts and both purity
tests are part of `test:unit` and pass with `rules/data.ts` in scope.

## The fixtures are hand-made stand-ins

`rules/testdata/pig.gamedata.json` was written by hand.
`rules/testdata/blackjack.gamedata.json` was assembled once by a scratch
script (not checked in) from `testdata/defs/blackjack.def.json`,
`movefacts/blackjack.json`, the corpus header's variantConfig, and a
hand-written per-move table. Neither is exporter output.

Schema fields they exercise: SchemaVersion, Source, Name, DisplayName,
Description, the three player counts; Definition with Decks (Components,
GenericValues (blackjack), ComponentProps (both)), Enums, GroupEnum '',
Game/Player/DynamicComponentValues (pig) Props in Go's map-Sanitize form;
Constants null; ComponentEnums (blackjack Suit, Rank); PhaseEnum ('phase',
''); Initial.Game (pig TargetScore), Initial.Players (all empty),
Initial.Dynamic (pig die Value); Traits.Game.DrawDiscardPair and
RoundRobin (blackjack); Traits.Player Seat, Inactive, Elimination,
GameScore 'score', Turn; Variants (blackjack maxrounds, pig {}); Computed
(blackjack HandValue, player scope); LegalTemplates; Predicates [];
Moves with Stock (12 stock types), Base (default, currentPlayer), Fields
(TargetPlayerIndex), LegalPhases, Progression (serial, repeat with
`between` counters, move leaves with and without allowMultiple),
AllowMultipleInProgression, Plan (declared with and without HasCustom, and
undeclared), SeatsPlayer, EndsPhaseProgression, Bodies Apply / Legal /
LegalCustom; SetUp Begin, Distribute, Finish; End.ConditionMet; Agents [];
Functions [].

Not exercised by any fixture: NameSanitization, HiddenAnimationKey,
Choices, GatheringStart, Bodies.Defaults, Traits.Player Role and Color,
Traits.Game.PlayerOrder, GameScore as a FnRef, End.PlayerScore,
LowScoreWins, End.Check (only by a unit test), CurrentPlayerIndex,
`parallel` groups, `custom` counters, non-null Constants, GroupEnum,
EnumDef.Parents/Range, PropDef.Constraints, Fields with Enum/Sanitize.
Memory, tictactoe, werewolf and debuganimations cover most of these once
their exported definitions exist.

What the fixtures cannot prove: `Stock.Config` is `{}` everywhere (the
config keys are the exporter's to define, and no stock parts are built from
it yet); the TargetPlayerIndex fields' `Input`/`Codec` on fix-ups are
unverifiable (fix-ups are not in the input schema); FnRef spellings for
closures (`...ConfigureComputedProperties.func1`) are guesses, and the
loader does not care.

## Judgment calls

1. **Bodies keep the engine's own signatures**, so a hand port's function
   is a body as it is, and the loader adapts only where Go's shape differs:
   an error result may be thrown or returned (THROWS roles), CheckGameFinished
   may be a tuple, distribute may return a facade StackRef.
2. **Stock parts live on the MoveSpec** (`MoveSpec.stock`), not in a context
   argument, because the frozen facade signature is `stockApply(state, m)`.
3. **`plan` only for declared moves.** Go's `LegalPlanSpecs` returns no
   specs for an undeclared move, whose Default.Legal checks run
   imperatively; the TS engine derives those from legalPhases/progression/
   base/stack properties, so the loader leaves `plan` unset there.
4. **Two hooks**, `move:` (whole move, behavior only) and `stock:` (stock
   parts only), rather than one. The brief asked for the first; the second
   is what let pig's and blackjack's own Apply/Legal/LegalCustom run through
   `Bodies` while every stock type is still a stub.
5. **Stock defaults run before a Defaults body**, always (the proposal's
   "run before `defaults` by the engine").
6. **`ComponentProps` is not written back into the compiled definition**
   (it feeds `DeckSpec.props`, where the engine reads it). Writing it would
   change blackjack's compiled `def.Decks`, which rules/engine.test.ts
   compares byte for byte with Go's DefinitionFor. `GenericValues` is
   written only when set for the same reason.
7. **Generic component with no generic values** is `{}`, the engine's
   existing answer, not a nil to dereference; the facade's `values()` is
   where Go's nil panic should be reproduced.

## Unverified

- No real exported definition exists yet; the six subtests skip.
- `stockMove` is still the stub, so no move was built from a real
  `StockParts`; every move in the bodies pass used a `stock:` stand-in made
  from the hand port. Once stock parts exist, the same test uses them
  automatically (it probes `stockMove` per move).
- Facade-shaped bodies (StackRef from distribute, string errors) are
  covered only by small unit tests, not by a corpus run.

## Where the proposal was wrong about the code

- `MoveSpec.plan` "used verbatim" cannot apply to undeclared moves: Go has
  no plan for them (call 3).
- The proposal's table lists `rules/stockRegistry.ts` under workstream 6;
  the brief moved it to the stock-move agents, and the loader codes against
  the frozen `stockMove`.
- `readyToStart` (listed among the spec's functions) has no GameData slot;
  the loader never sets it. No hand port uses it.

## Edits outside my files

None. New files only under `rules/testdata/` (owned) and
`rules/testdata/gen/movefacts/` and `rules/testdata/movefacts/` (inside the
owned `rules/testdata/`).
