# Porting games from Go to the TypeScript engine by machine: execution proposal

A game is ported by four tools and a ratchet. A Go **exporter** writes
everything about a game that is already data as a definition the TypeScript
engine loads. A **translator** turns the small imperative Go functions that
remain into TypeScript over a restricted subset, and refuses the rest with a
machine-readable reason. A **coding agent** translates each refused function,
one function per task. A **differential fuzzer** plays seeded random games in
Go, records them in the conformance format, and replays them in TypeScript; a
port is accepted when that run is clean and its coverage report has no
unexplained gap. Every agent translation is kept, and the commonest refusal
reasons become translator rules. Measured on the seven games in this
repository, 34 of 71 configured moves need no code at all, and of the 119
game-written functions that reach engine output, 117 fall inside the subset
proposed here. So on these games the agent has almost nothing to do. What
stands between a Go game and a TypeScript port is mostly something the agreed
design does not name: framework code (stock moves, behaviors, enum and stack
features) that the TypeScript engine has not ported yet. Checkers, the one
game not yet ported, is blocked on six such features and on one function no
translator will take.

## Baseline

Seven games exist in this repository, all under `examples/`. Six are ported
by hand; checkers is not.

| Game | Move configs | Data only | With game code | Declared precondition specs | Functions reaching engine output | In the v1 subset | Outside it | Need an engine feature TypeScript lacks | Not ported (diagram, agents) |
|---|---|---|---|---|---|---|---|---|---|
| memory | 7 | 3 | 4 | 9 | 16 | 15 | 1 | 0 | 11 |
| blackjack | 23 | 16 | 7 | 6 | 22 | 22 | 0 | 0 | 1 |
| pig | 6 | 3 | 3 | 4 | 10 | 10 | 0 | 0 | 1 |
| tictactoe | 4 | 3 | 1 | 0 | 11 | 11 | 0 | 0 | 6 |
| werewolf | 11 | 4 | 7 | 0 | 16 | 16 | 0 | 0 | 1 |
| debuganimations | 13 | 1 | 12 | 7 | 24 | 24 | 0 | 0 | 1 |
| checkers (unported) | 7 | 4 | 3 | 5 | 20 | 19 | 1 | about 11 (estimate) | 0 |
| **Total** | **71** | **34** | **37** | **31** | **119** | **117** | **2** | **about 11** | **21** |

### How this was counted

Every count is from reading the non-generated, non-test Go of each game
(`main.go`, `moves.go`, `state.go`, `components.go`, `group_enum.go`,
`agent.go`). Nothing was run to produce it.

- **Move configs** are the entries `ConfigureMoves` installs, with
  `moves.DefaultRoundSetup` counted as the three it expands to
  (`moves/seat_player.go`).
- **Data only** means the move's Go type is a stock type from package `moves`,
  or a game type that embeds one and declares no method. Its name, help text,
  phase, progression, preconditions and options are all configuration.
- **Declared precondition specs** are the top-level `legal.Spec` values passed
  to `moves.WithLegalPreconditions`.
- **Functions reaching engine output** are game-written function and method
  bodies whose result can change a stored state, a move record, a legality
  verdict or a viewer's output: `Apply`, `Legal`, `LegalCustom`,
  `DefaultsForState`, the turn hooks, `GameScore`, computed properties,
  `BeginSetUp`, `DistributeComponentToStarterStack`, `FinishSetUp`, the end
  tests, `CurrentPlayerIndex`, component methods, a predicate constructor, and
  the helpers those call. `concreteStates` is not counted: the translator
  erases it (see "Types").
- **Not counted as functions**, because the exporter gets their result by
  calling them on the live manager: `Name`, `DisplayName`, `Description`, the
  player-count methods, `Variants`, `ConfigureMoves`, `ConfigureDecks`,
  `ConfigureConstants`, `ConfigureLegalTemplates`, `ConfigureEnums`, the three
  state constructors, and the twelve `HelpText` methods in debuganimations.
- **In the v1 subset** is a judgment made by reading each body against the
  subset defined below. It is an estimate until the subset checker exists and
  is run; the lint in this proposal replaces these columns with measured ones.
- **Need an engine feature** counts functions the translator could translate
  but whose calls have no TypeScript counterpart yet. For checkers it is an
  estimate: the six move methods, the five `token` methods.
- **Not ported** is `Diagram` (the exporter drops it as a normalization) and
  agent code, which the TypeScript engine does not run
  (`rules/types.ts`, `GameRecord.agents`).

### The functions outside the subset, and the hard ones inside it

| Function | Why it is hard | Bucket |
|---|---|---|
| `memory.(*moveStartHideCardsTimer).Apply` | Reads `hideCardMoveName`, a package variable assigned inside `ConfigureMoves` from `hideCardConfig.Name()`. Its value is not a constant and not reachable by reflection | Outside (`package-var-read`). One line of Go (a string constant) would bring it in |
| `checkers.(*gameDelegate).ConfigurePredicateConstructors` | Returns a `legal.Predicate` holding an `Evaluate` closure over `legal.Context` | Outside (`predicate-constructor`) |
| `checkers.(*token).AllNextSpaces`, `LegalCaptureSpaces`, `Legal`; `(*moveMoveToken).Apply` | Read `graphUpward`, `graphDownward` and `spacesEnum`, package variables built by framework constructors; use range enums, grid graphs, `ComponentAtKey`, `SwapComponentsByKey`, `ContainingImmutableStack` | In the subset by the framework-variable rule; blocked on engine features |
| `blackjack.handValue` | Runs on each viewer's sanitized state, where a hidden card is the deck's generic component. Go reads the generic values; the hand port special-cases index -2. Carries the ten-plus-ace bug | In the subset once decks export generic values |
| `blackjack.(*moveRevealHiddenCard).LegalCustom` | Calls a method on `ImmutableFirst()` without a nil check. The full ledger runs it after an earlier check failed, so the panic text is recorded | In the subset with nil-dereference emulation |
| `memory.(*moveCaptureCards).Apply`, `(*moveHideCards).Apply` | Range over `Components()`, a snapshot taken before the loop, while moving components out of the same stack. Capture discards the error from `MoveToNextSlot` | In the subset; both semantics are listed under "Semantics" |
| `werewolf.resolveVotes` | A `map[PlayerIndex]int` that is indexed and never ranged | In the subset; ranging over it would be refused |
| `werewolf.populateFellowWolves` | Assigns `nil` to a slice property and appends to it. Go stores `null` for nil and `[]` for empty | In the subset with nil-aware slices |
| `werewolf.finishVoteResolution` | Calls `state.Manager().Delegate().CheckGameFinished(state)`: dynamic dispatch that resolves to the game's own method | In the subset by a table entry |
| `werewolf.(*moveBeginGame).Apply` | Calls the embedded stock `StartPhase.Apply`, then draws `state.Rand().Perm`; holds a slice of `*playerState` | In the subset; needs every stock helper to expose its parts |
| `pig.(*moveRollDice).Apply` | Calls `dice.DynamicValue.Roll`, a method in another package | In the subset once `components/dice` is ported as a library |
| `memory.(*gameDelegate).FinishSetUp` | The longest body: two shuffles, a search loop, four component moves with wrapped errors | In the subset |

### What the hand ports needed that the Go game source does not show

Read by comparing each `games/*.ts` with its Go original.

1. **Code from other packages.** Pig's roll is `components/dice` inlined,
   with the `SelectedFace` and `RollCount` fields of a type the game never
   names. Blackjack restates the `playingcards` deck order and rank enum.
   Every game restates what its embedded behaviors contribute as properties.
2. **Where a function runs.** Computed properties run on the sanitized
   state (`state.go:1273`), so `handValue` must handle a generic component.
3. **Panics that are part of the output.** `games/blackjack.ts` throws Go's
   `runtime error: invalid memory address or nil pointer dereference` text
   by hand.
4. **Constants a pattern module may not compute.** `TOTAL_DIM` is written as
   `9`, not `DIM * DIM`, because a pattern module may hold only functions and
   frozen data at module scope.
5. **Enums are strings in the stored state and integers in Go.**
   `p.Role.Value() == roleWerewolf` became a string comparison;
   `game.Phase.Value()` became `state.currentPhaseKey()`.
6. **Enum display strings** come from generated `auto_enum.go`, not from the
   game's hand-written source.
7. **Sub-state pointers became indexes.** A Go `*playerState` is, in every
   hand port, a player index.
8. **`null` against `[]`.** `FellowWolves` is left `null` unless it has
   members, to match Go's nil slice.
9. **Which state a new move is bound to.** `state.Game().MoveByName` binds to
   the committed state and `MoveByNameForState` to the one being built; the
   ports use `state.moveByName` and `state.newMove` accordingly.
10. **Values the Go gets by accident.** Memory and debuganimations declare
    `DefaultNumPlayeres` (misspelled), so the base delegate's default
    applies. A hand port has to notice; the exporter cannot get it wrong.

One hand port departs from Go in a way no recording shows (read, not run):
`memory.(*moveCaptureCards).Apply` discards the error from `MoveToNextSlot`
and continues, where `games/memory.ts` lets the stack operation throw. The
destination is an uncapped stack, so nothing reaches it today.

## Where the agreed design has to change

Each item is something the code contradicts. The rest of the proposal
assumes the correction.

1. **The exporter cannot be only the running binary.** Whether a move's
   `Apply` is the stock one or the game's own is not visible at run time: a
   method promoted from an embedded struct is compiled as a wrapper on the
   outer type, so reflection reports the same thing for both (from knowledge
   of the compiler; not run here). The exporter is therefore two halves that
   are joined: a run-time half over the live `GameManager`, and a static half
   over `go/packages` and `go/types`, which `boardgame-util/lib/gamepkg`
   already loads for the existing lint.
2. **Progressions are not data on either side.** In Go the group types
   (`serial`, `repeat`, `parallelCount`) are unexported and a count is a
   closure (`moves.ValidCounter`). In TypeScript `ProgressionGroup.count` is
   also a function (`rules/progression.ts:13`). The exporter needs a new,
   additive description function inside package `moves`, and the engine needs
   counters as data.
3. **The assembled legal plan is unexported.** The engine-legal report names
   `LegalPlan.specs` as the list a client needs; it is manager state. The
   exporter needs a read-only accessor in the core, in a new file, as
   `server/api/conformance_export.go` did.
4. **There is a fourth bucket: engine gap.** A function can be perfectly
   translatable and still have no port, because it calls framework code the
   engine lacks. Package `moves` declares about 50 stock move types;
   `rules/stock.ts` ports about 16, bases included. Of 18 behaviors, 9 are ported. Stack
   constraints, range enums, grid graphs, enum-keyed stack access and boards
   keyed by an enum are absent. Neither the translator nor an agent with one
   function in view can fix that. The metric gets a fourth number, and
   closing gaps is its own workstream.
5. **Stock moves are not only configuration.** They call the game through
   interfaces: `TurnDone`, `ResetForTurnStart`, `GameScore`, `ConditionMet`,
   `RoundRobinAction`, a component's `Legal` (`moves/interfaces`). Those
   bodies are translation units. And a game move that overrides `Apply`
   usually calls the embedded stock `Apply` first, so every stock helper must
   expose its apply, legality and defaults as callable parts. Two do today
   (`drawToPlayer`, `moveComponentToSlot`).
6. **Custom legal predicates are a fifth kind of imperative body**, beside
   apply, defaults, the end test, computed properties and set-up.
7. **A header naming the Go commit makes generated code churn.** The commit
   changes on every commit; the function does not. Each header names a hash
   of the function's Go source and the producer; the commit is recorded once,
   when the block is first written or its body changes. A block is rewritten
   only when its body would change.
8. **The ratchet cannot rewrite accepted code freely.** A new rule that can
   now translate an agent-written function will usually produce different
   text with the same behavior, which changes a pattern's policy identity for
   nothing. The ratchet metric is therefore measured in a dry run; the
   checked-in text changes only when someone adopts it.
9. **"No access to the rest of the game" needs one exception.** Most refused
   or agent-bound functions call helpers in the same package. The brief
   carries the TypeScript signatures of those callees, and functions are
   handled callee first.
10. **A clean fuzz run is not enough to accept a port.** A run that never
    reaches a branch says nothing about it, and the conformance format
    records no refused proposal at all. Acceptance is a clean run plus a
    coverage report with every gap either closed or signed off, and the
    format gains refusals.
11. **The framework's own agents are thinner than assumed.** There is one
    real `boardgame.Agent` (memory); it draws from the global `math/rand`
    (`examples/memory/agent.go:162`, `:258`), so its games cannot be
    regenerated. Tictactoe's `Agent.ProposeMove` returns nil. The seeded
    interface is `bots.Policy` (`bots/policy.go`), with one implementation
    (`examples/tictactoe/bot`). The fuzzer's main driver has to be its own
    random-legal policy.
12. **The translator should not target the hand-port idiom.** The engine's
    stack API takes a slot and throws; Go's takes a component and returns an
    error. Translating between the two shapes per call site is where a
    mechanical tool would go wrong (the discarded error above is an
    example). The translator targets a thin facade with Go's shapes, written
    once over the existing engine.

## The pipeline, by file

New Go, all additive:

| Path | What |
|---|---|
| `boardgame-util/lib/gamedef/` | The definition: schema structs, run-time export, static pass, source hashing |
| `boardgame-util/cmd_gamedef.go` | `boardgame-util gamedef export [--check] PKG...` |
| `moves/describe.go` | `DescribeProgression`, `DescribeCounter`: read-only descriptions of unexported group types |
| `legal_plan_export.go` (core) | `(*GameManager).LegalPlanSpecs(moveName) []LegalSpec` |
| `boardgame-util/lib/transpile/` | The subset checker, the API table, the emitter, refusals; `testdata/pairs/` |
| `boardgame-util/cmd_transpile.go` | `boardgame-util transpile [--check] [--dry-run] PKG...` |
| `boardgame-util/lib/fuzz/` | The driver, legal-move enumeration, coverage |
| `boardgame-util/cmd_fuzz.go` | `boardgame-util fuzz run`, `fuzz corpus [--check]` |
| `boardgame-util/lib/lint/declarative.go` | The declarativeness report; `cmd_lint.go` gains `--declarative` |
| `boardgame-util/cmd_port.go` | `boardgame-util port PKG`: export, translate, list refusals, fuzz, report |

New TypeScript, under `server/static/src/engine/`:

| Path | What |
|---|---|
| `rules/data.ts` | `GameData` types and `gameFromData(data, bodies): GameSpec` |
| `rules/stockRegistry.ts` | Stock move by name, as a `switch` |
| `go/` | The Go-shaped facade: `api.ts`, `num.ts`, `str.ts`, `slice.ts`, `panic.ts` |
| `lib/` | Ports of `components/dice`, `components/playingcards`, range and tree enums, grid graphs |
| `games/<game>/definition.gen.ts`, `bodies.gen.ts`, `index.ts` | A generated game |
| `testing/fuzz.ts`, `games/fuzz.test.ts` | The fuzz replay and its coverage |

Records of the ratchet live in `fabric/transpile/` (reports, the bug
register) and `boardgame-util/lib/transpile/testdata/pairs/` (stored pairs).

## The exported definition

### Schema

One JSON document per game, `GameData`. Keys are capitalized like the state
JSON and the existing definition.

```ts
interface GameData {
  SchemaVersion: 1;
  Source: { Package: string; Exporter: string };
  Name: string; DisplayName: string; Description: string;
  MinNumPlayers: number; MaxNumPlayers: number; DefaultNumPlayers: number;

  /** Exactly what testdata/gen/defs emits today, plus the three additions below. */
  Definition: GameDefinition;
  Constants: Record<string, unknown> | null;
  ComponentEnums: Record<string, EnumDef>;
  PhaseEnum: string; GroupEnum: string;

  /** Constructor output: non-zero starting values. */
  Initial: {
    Game: Record<string, unknown>;
    Players: Record<string, unknown>[];              // one per index, 0 .. MaxNumPlayers-1
    Dynamic: Record<string, Record<string, unknown>[]>;  // deck -> one per component
  };
  Traits: {
    Game: { DrawDiscardPair?: { Draw: string; Discard: string }; RoundRobin?: true; PlayerOrder?: FnRef };
    Player: { Seat?: true; Inactive?: true; Elimination?: true; Role?: true; Color?: true;
              GameScore?: FnRef | 'score'; Turn?: { TurnDone: FnRef; ResetForTurnStart: FnRef; ResetForTurnEnd: FnRef } };
  };
  Variants: unknown;                 // the conformance header's variantConfig
  Computed: { Name: string; Scope: 'global' | 'player'; Type: string; Compute: FnRef }[];
  LegalTemplates: Record<string, string>;
  Predicates: { Name: string; Construct: FnRef }[];
  Moves: MoveData[];
  SetUp: { Begin?: FnRef; Distribute: FnRef; Finish?: FnRef };
  End: { ConditionMet?: FnRef; PlayerScore?: FnRef; LowScoreWins?: true; Check?: FnRef };
  CurrentPlayerIndex?: FnRef;
  Agents: { Name: string; DisplayName: string }[];
  Functions: FunctionEntry[];
}

interface MoveData {
  Name: string; HelpText: string; IsFixUp: boolean;
  GoType: string;                                    // "blackjack.moveCurrentPlayerHit"
  /** The nearest stock ancestor and its options, or null for moves.Default alone. */
  Stock: { Type: string; Config: Record<string, unknown> } | null;
  Base: 'default' | 'currentPlayer' | 'anyPlayer';
  Fields: Record<string, { Type: string; Enum?: string; Sanitize?: string; Input: string; Codec: string }>;
  NameSanitization?: string; HiddenAnimationKey?: string;
  LegalPhases: number[] | null;
  Progression: ProgressionData | null;
  AllowMultipleInProgression: boolean;
  /** The plan Go assembled: base checks, contributions, authored specs, suppressions applied. */
  Plan: { Declared: boolean; Specs: LegalSpec[]; HasCustom: boolean };
  Choices?: unknown;                                 // the header's projection entry
  SeatsPlayer?: true; GatheringStart?: true; EndsPhaseProgression?: true;
  /** Present only for a method a game type declares. */
  Bodies: { Apply?: FnRef; Legal?: FnRef; LegalCustom?: FnRef; Defaults?: FnRef };
}

type ProgressionData =
  | { Kind: 'move'; Name: string }
  | { Kind: 'serial'; Children: ProgressionData[] }
  | { Kind: 'repeat'; Count: CounterData; Child: ProgressionData }
  | { Kind: 'parallel'; Count: CounterData; Children: ProgressionData[] };
type CounterData =
  | { Kind: 'any' | 'all' | 'infinite' }
  | { Kind: 'atLeast' | 'atMost' | 'exactly'; N: number }
  | { Kind: 'between'; Min: number; Max: number }
  | { Kind: 'custom'; Count: FnRef };

/** A function by its Go identity: "blackjack.(*moveCurrentPlayerHit).Apply". */
type FnRef = string;

interface FunctionEntry {
  Func: FnRef; Role: string;                          // 'apply', 'legal', 'computed', 'helper', ...
  File: string; Line: number;
  SourceHash: string;                                 // SHA-256 of the function's gofmt-ed source
  Calls: FnRef[];                                     // in-package callees
  Verdict: 'subset' | 'refused';
  Refusals: Refusal[];
}
```

Three additions to `GameDefinition`:

- `Decks[name].GenericValues`: what `Deck.SetGenericValues` set, or null.
- `Decks[name].ComponentProps`: each component value's type and enum, which
  `DeckSpec.props` carries by hand today.
- `Enums[name].Range`: the dimensions of a range enum (`Parents` for a tree
  enum exists already).

And one to `PropDef`: `Constraints`, the constraint strings from the stack
struct tag (`same(Color)`).

### How it relates to what exists

| Part of `GameData` | Already emitted by | Change |
|---|---|---|
| `Definition` | `testdata/gen/vecgame/defs.go` (`DefinitionFor`), for five games | The code moves to `lib/gamedef`; `gen/defs` calls it. `testdata/defs/*.def.json` stay as a projection, and a Go test keeps them equal |
| Move names, help text, fix-up flag, fields; input and choice schemas and fingerprints; computed descriptors; `Variants`; decks, enums, constants, templates | The header of every conformance document (`lib/conformance/export.go`, `header`) | None in the exporter. The runner's `header` comparison already checks a compiled game against these; a generated game must pass it |
| `Initial`, `Traits`, `Stock`, `Plan`, `Progression`, `LegalPhases`, `Bodies`, `Functions`, `GenericValues` | Nothing | New |

How each new part is obtained:

- **`Initial`**: call `GameStateConstructor()`, `PlayerStateConstructor(i)`
  for each index up to the maximum, and `DynamicComponentValuesConstructor`
  for each deck; read every property through the generated reader; keep the
  non-zero ones. This covers tictactoe's per-index token value and
  werewolf's `-1` votes.
- **`Stock` and `Bodies`**: the static half walks each move type's embedding
  chain with `types.LookupFieldOrMethod`. A method whose declaring package is
  the game is a body; the first embedded type declared in package `moves` is
  the stock ancestor. `Config` is the move config's `CustomConfiguration()`,
  with keys as stored and values serialized by kind (string, int, bool, enum
  key, `LegalSpec` list, input field list).
- **`Plan`**: from the new core accessor. The loader uses it verbatim.
- **`Progression`**: from `moves.DescribeProgression`. A counter is
  described by fitting it to the seven stock families: the function is
  evaluated on the grid `currentCount` 0 to 64 at the group's real length,
  and a family is accepted only if it reproduces every result and error
  text. No fit means `custom`, a function to translate.
- **`Traits`**: the static half reads which behaviors the state structs
  embed and which interface methods the game declares.

The exporter is deterministic: maps are emitted with sorted keys, moves in
registration order. `gamedef export --check` fails when a checked-in
definition differs, as `conformance export --check` does.

### What the engine needs in order to load it

A hand port today expresses these as functions, which a data file cannot
hold (read from `rules/types.ts`): `SubStateSpec.initial`,
`GameTraits.playerOrder`, the three `TurnHooks`, `PlayerTraits.gameScore`,
`ComputedSpec.compute`, `MoveSpec.defaults`, `legalCustom`, `contributes`,
`legal` and `apply`, the counters inside `ProgressionGroup`, `setUp.begin`,
`distribute` and `finish`, `end.conditionMet`, `playerScore` and `check`,
`currentPlayerIndex`, `readyToStart`. A hand port also shares private
helpers between moves and spreads a stock helper's parts into a move.

The loader keeps `GameSpec` as the engine's one input and builds it:

```ts
// rules/data.ts
export type Bodies = (ref: string) => ((...args: never[]) => unknown) | undefined;
export function gameFromData(data: GameData, bodies: Bodies): GameSpec;
```

`bodies` is a function containing a `switch` over `FnRef`, because a
pattern module may not hold a record of functions. Changes to existing
files, all additive:

| File | Change |
|---|---|
| `rules/types.ts` | `MoveSpec.plan?: readonly LegalSpec[]` (used verbatim; `contributes` and `suppress` are then ignored). `SubStateSpec.initialValues?`. `DeckSpec.generic?` |
| `rules/progression.ts` | A counter may be `CounterData`; `matchTape` evaluates it with a `switch` |
| `rules/compile.ts` | Honor `plan`, `initialValues`, `generic` |
| `rules/state.ts` | `state.component(deck, GENERIC_COMPONENT)` returns the deck's generic values |
| `rules/stock.ts` | Every helper returns its parts (`apply`, `legal`, `defaults`) as well as a spec |
| `definition.ts` | `GenericValues`, `ComponentProps`, `EnumDef.Range`, `PropDef.Constraints` |

A new test, `rules/data.test.ts`, proves the loader on the six hand ports:
for each, the `GameSpec` built from its exported definition with the hand
port's own functions as bodies compiles to the same header, plans and
definition as the hand port.

The definition is checked in as `games/<game>/definition.gen.ts`, a
function returning the literal, with the JSON beside the Go corpus in
`conformance/definitions/<game>.json`. Whether a pattern can import JSON
directly was not determined; a TypeScript literal is known to work.

## The translator

`boardgame-util/lib/transpile` takes a loaded package and the function
manifest and, for each function, either emits TypeScript or returns
refusals. It never emits a partial function.

### Unit and output

The unit is one function declaration, or one function literal in a slot
the exporter knows (a computed property's closure). Each becomes one
exported TypeScript function in `bodies.gen.ts`, taking `state` first:

```ts
// @generated transpile@3 go:blackjack.(*moveCurrentPlayerHit).Apply sha256:9f2c… first:f0646494
export function blackjack_moveCurrentPlayerHit_Apply(state: RulesState, m: MoveInstance): void { … }
```

Constant expressions are folded with `go/types` (`DIM * DIM` becomes `9`,
`iota` constants become literals), which also satisfies the rule against
computing at module scope.

### Types

| Go | TypeScript |
|---|---|
| `int` | `number`, checked to be a safe integer when stored |
| `bool`, `string`, `float64` | `boolean`, `string`, `number` |
| `boardgame.PlayerIndex` | `number` |
| `enum.EnumKey`, and an enum value read from state | `number` (the key); the facade converts to and from the stored string |
| `*gameState` (the game's game sub-state) | no value: property access goes through `state` |
| `*playerState` | `number`, the player's index |
| A dynamic component values struct | a handle `{ deck, index }` |
| `boardgame.Stack`, `SizedStack`, `ImmutableStack`, `MergedStack` | the facade's `StackRef` |
| `boardgame.ComponentInstance`, immutable or not | `Comp \| null`: `{ stack, slot, deck, index }` |
| `boardgame.Timer` | the engine's `TimerHandle` |
| `boardgame.Move` | `MoveInstance` |
| `error` | `string \| null` |
| `[]T` for the scalar types above and `[]*playerState` | `T[] \| null`, nil-aware |
| `map[K]V` with K an integer or string type | `Map<K, V>`, index and assign only |
| Multiple results | a tuple array, destructured at the call |
| `time.Duration` constant | whole milliseconds |

`concreteStates(state)` and any function whose body only type-asserts
sub-states is erased: its results are the game handle and the list of
player indexes.

Refused types: fixed arrays, struct values that are copied or mutated,
pointers to locals, sized integers (`int8`, `uint32`, ...), `float32`,
`complex`, channels, function values other than a `less` callback or a
literal passed straight to a known call, interfaces outside the table,
type parameters.

### Statements

Allowed: short and plain variable declarations; assignment, including
tuple assignment and compound operators; `++` and `--`; `if` with an
initializer; `for` in its three-clause, condition-only and infinite forms;
`range` over a slice, an integer, or one of the table's snapshot calls
(`Components()`, `PlayerStates()`); `switch` on a value or with no tag,
including multi-value cases and `default`, without `fallthrough`; `break`
and `continue`, labeled or not; `return`; expression statements that are
calls; blocks.

Refused: `go`, `select`, `defer`, `goto`, `fallthrough`, type switches,
`range` over a map, a channel or a string, and assignment to a package
variable.

### Expressions

Allowed: literals; identifiers for locals, parameters, constants and
framework-typed package variables (below); arithmetic and comparison on
integers; `==` and `!=` on strings, booleans and `nil`; `&&`, `||`, `!`;
string concatenation; indexing and `len` of slices; `make`, `append` in
self-assignment form, `copy`, `delete`, `min`, `max`; calls to in-package
functions and methods; calls in the API table; conversions among `int`,
`PlayerIndex` and `EnumKey`; type assertions to the game's own sub-state,
component and move types (erased); composite literals of slices.

A package variable may be read when it is a constant-initialized value, or
a framework object built by a recognized constructor with constant
arguments: `enums.MustAdd("phase", …)`, `enums.MustAddRange("spaces", 8, 8)`,
`graph.MustNewEnumGridConnectedness(spacesEnum, graph.DirectionDown, …)`.
The translator replaces the read with a facade call that looks the object
up by name or rebuilds it from the exported enum.

### The API table

Go names are from `stack.go`, `component.go`, `state.go`, `timer.go`,
`enum/` and `behaviors/`; engine names from `rules/state.ts` and
`stack.ts`. The facade (`go/api.ts`) is the third column's shape over the
second's implementation.

| Go | Engine today | Facade call the translator emits |
|---|---|---|
| `game.X` (scalar), `players[i].X` | `state.g.X`, `state.players[i].X` | `state.g.X`, `player(state, i).X` (bounds-checked) |
| `game.X` (stack), `p.X` (stack) | `state.gameStack('X')`, `state.playerStack(i, 'X')`, or `gameView`/`playerView` for a merged stack | `gameStack(state, 'X')`, `playerStack(state, i, 'X')` |
| `len(state.PlayerStates())` | `state.numPlayers()` | the same |
| `state.CurrentPlayerIndex()`, `state.CurrentPlayer()` | `state.currentPlayerIndex()` | the same; the player is its index |
| `p.EnsureValid(state)`, `Next`, `Previous`, `Valid`, `WithinBounds` | `state.ensureValid(p)`, `next`, `previous`, `valid`, `withinBounds` | the same |
| `state.Rand().Intn(n)`, `Perm(n)`, `Shuffle`, `Int63n`, `Float64` | `state.rand().intn(n)`, `perm(n)`, ... | the same |
| `state.Game().Variant()[k]` | `state.variant()[k]` | `variantValue(state, k)` (missing key gives `''`) |
| `state.Game().MoveByName(n)` | `state.moveByName(n)` | the same |
| `state.Game().MoveByNameForState(n, state)` | `state.newMove(n)` | the same |
| `state.Manager().Delegate().M(state)` | none | a direct call of the game's own `M` |
| `behaviors.PlayerIsInactive(p)`, `PlayerIsEliminated(p)` | `state.playerIsInactive(i)`, `playerIsEliminated(i)` | the same |
| `p.SetEliminated()`, `SetPlayerInactive()`, `SetSeatFilled()` | property writes | `setEliminated(state, i)`, ... |
| `game.Phase.Value()`, `game.SetCurrentPhase(k)`, `game.Phase.SetValue(k)` | `state.currentPhaseKey()`, `state.setPhase(k)` | the same |
| `v.Value()`, `v.SetValue(k)`, `v.Equals(o)` on an enum property | the stored string, through `rules/enums.ts` | `enumKey(state, owner, 'Prop')`, `setEnumKey(...)`, `enumEquals(...)` |
| `v.RangeValue()`, `e.RangeToValue(...)`, `e.ValueToRange(k)` | none (gap) | `rangeValue`, `rangeToValue`, `valueToRange` |
| `g.NeighborsByKey(k)` | none (gap) | `neighborsByKey(graph, k)` |
| `s.Len()`, `NumComponents()`, `SlotsRemaining()`, `MaxSize()` | `len()`, `numComponents()`, `slotsRemaining()`, `maxSize()` | the same |
| `s.ComponentAt(i)`, `ImmutableComponentAt(i)` | `componentAt(i)`: a deck index, null, or -2 | `componentAt(s, i): Comp \| null` |
| `s.First()`, `Last()`, and the `Immutable` forms | `firstComponentIndex()`, `lastComponentIndex()` | `first(s)`, `last(s)` |
| `s.Components()`, `ImmutableComponents()` | `indexes()` | `components(s)`: a snapshot array |
| `s.ComponentAtKey(k)`, `SwapComponentsByKey(i, j)` | none (gap) | `componentAtKey`, `swapComponentsByKey` |
| `c.Values().(*T).F` | `state.component(deck, index).F` | `values(state, c).F` (generic values for a generic component; panics on nil) |
| `c.DynamicValues().(*T).F` | `state.dynamic(deck, index).F` | `dynamic(state, c).F` |
| `c.MoveTo(dest, slot)` | `stack.moveComponent(i, dest, slot)`, throws | `moveTo(c, dest, slot): string \| null` |
| `c.SecretMoveTo(dest, slot)` | `stack.secretMoveComponent(i, dest, slot)` | `secretMoveTo(c, dest, slot)` |
| `c.MoveToFirstSlot(d)`, `MoveToLastSlot`, `MoveToNextSlot` | `moveComponentToFirstSlot(i, d)`, ... | `moveToFirstSlot(c, d)`, ... |
| `c.SlideToFirstSlot()`, `SlideToLastSlot()` | `slideToFirstSlot(i)`, `slideToLastSlot(i)` | `slideToFirstSlot(c)`, `slideToLastSlot(c)` |
| `c.MayMoveTo(d)`, `MayMoveToSlot(d, slot)` | `mayMoveComponentTo(i, d)`, `moveComponentError(i, d, slot)` | `mayMoveTo(c, d)`, `mayMoveToSlot(c, d, slot)` |
| `c.ContainingImmutableStack()` | `EngineState.locateComponent` | `containingStack(state, c)`: a tuple |
| `s.MoveAllTo(d)`, `MoveCountTo(d, n)`, and the `May` forms | `moveAllTo`, `moveCountTo`, `mayMoveAllTo`, `mayMoveCountTo` | the same names, returning `string \| null` |
| `s.Shuffle()`, `PublicShuffle()`, `SwapComponents(i, j)` | `shuffle()`, `publicShuffle()`, `swapComponents(i, j)` | the same names, returning `string \| null` |
| `s.SortComponents(less)` | `sortComponents(less)`, `less` over deck indexes | `sortComponents(s, less)`, `less` over `Comp` |
| `s.SetSize(n)`, `ExpandSize`, `ContractSize`, `SizeToFit` | the same in camel case | the same, returning `string \| null` |
| `t.Active()`, `t.Start(d, move)`, `t.Cancel()` | `state.timer('P').active()`, `start(ms, move)`, `cancel()` | the same |
| `m.TargetPlayerIndex`, `m.F` | `move.fields.TargetPlayerIndex`, `move.fields.F` | the same |
| `m.<Stock>.Apply(state)` | the stock helper's `apply` | `stockApply(state, m)`, resolved by the loader |
| `m.<Stock>.Legal(state, proposer)` | the `base` argument of `legal` | `base()` |
| `m.<Stock>.DefaultsForState(state)` | run before `defaults` by the engine | dropped, when it is the first statement; refused otherwise |
| `errors.New(s)`, `err.Error()` | a string | the string |
| `fmt.Errorf`, `fmt.Sprintf` with `%d %s %v %q` | `legal/goformat.ts` | `sprintf(...)` |
| `legal.Errorf(key, bindings)` | a `LegalVerdict` | `legalError(key, bindings)` |
| `strconv.Itoa`, `strconv.Atoi` | none | `itoa(n)`, `atoi(s)`: a tuple with Go's error text |
| `strings.Join`, `Repeat`, `Split`, `ToLower`, `TrimSpace`, `HasPrefix`, `HasSuffix`, `Contains` | none | `str.ts` |
| `dice.(*DynamicValue).Roll(r)`, `playingcards` values | inlined by hand in `games/pig.ts`, `games/blackjack.ts` | `lib/dice.ts`, `lib/playingcards.ts` |

A call that is in the table but whose engine column says "none (gap)" is
refused as `engine-gap` until the engine has it. A call that is not in the
table is refused as `unknown-call`.

### Semantics that differ

| Go | Handling |
|---|---|
| Integer division truncates toward zero; `%` takes the dividend's sign; division by zero panics | `idiv(a, b)` is `Math.trunc(a / b)` and throws Go's `runtime error: integer divide by zero`. `%` is the same in both languages; `imod` adds the zero check |
| `int` is 64 bits and wraps | Not emulated. Every integer written to state or to a move field is checked with `Number.isSafeInteger` when the version is saved, and the engine refuses the move with its own text. `*` goes through `imul`, which checks at once. A game that relies on wraparound is out of scope and the lint says so |
| Map iteration order is random | `range` over a map is refused (`map-range`). Go code whose result depends on the order has no single answer for a port to match |
| `nil` | A nil component is `null`. Calling a method on it goes through the facade, which throws Go's nil-dereference text. Go code that compares to `nil` first translates to a `null` comparison. A panic inside a declared plan becomes an unknown ledger entry with that text, as in Go; outside one it is a refusal, where Go takes the process down (engine-rules divergence 5) |
| Index out of range panics | Slice and player indexing by a non-constant goes through `at(xs, i)` or `player(state, i)`, which throw `runtime error: index out of range [i] with length n` |
| A nil slice serializes as `null`, an empty one as `[]` | `var x []T` and `x = nil` are `null`; `append(null, v)` is `[v]`; `make([]T, 0)` and `[]T{}` are `[]`. `len(null)` is 0 and ranging over it does nothing |
| Slices alias their backing array | `append` is allowed only as `x = append(x, ...)`. A slice expression `a[i:j]` is allowed only where the result is read and never written or appended to. Anything else is refused (`slice-alias`) |
| `Components()` is a snapshot | `components(s)` returns an array of handles taken before the loop. A handle whose component has since moved fails as Go's does, with the engine's "not public information" text (engine-core divergence 3 is the same mechanism) |
| A stack operation returns an error and changes nothing on failure | Facade operations return `string \| null`. A discarded result is discarded; `if err := ...; err != nil { return ... }` is the same shape in TypeScript |
| An `error` result, by role | In `Apply` and set-up, returning a non-nil error becomes `throw`; the engine discards the state. In `Legal`, `LegalCustom` and `TurnDone` it is the returned string. In a helper it stays a returned `string \| null` |
| `sort.Sort`, `sort.Slice` are unstable but deterministic | `gosort.ts` ports `sort.Sort`, including the order among equal elements. `sort.Slice` is generated from the same template in the standard library; that its order among equals matches was not checked and needs a vector before it is mapped. `sort.Stable` and `sort.SliceStable` map to `Array.prototype.sort`, which is stable. `less` must return what Go's returns for every pair |
| Strings compare by UTF-8 bytes; `len` counts bytes | `<` on strings goes through `strCompare`, which orders by code point (the same order as UTF-8 bytes; JavaScript's own `<` differs for characters outside the basic plane, which memory's cards are). `len(s)` is `utf8Len(s)`. Indexing or slicing a string is refused (`string-bytes`) |
| `state.Rand()` is the only legal source | `state.rand()`, in Go's call order. Go evaluates call arguments left to right, as JavaScript does. Any other `*rand.Rand` or the package-level functions are refused (`foreign-rand`); the existing lint already flags the import. Starting a timer draws 16 values (engine-core divergence 13); the facade's `start` does |
| Enum values are integers | Stored as strings; the facade converts. `ValueFromString` on names that normalize alike has no single answer (engine-legal finding 9) and is refused |
| `time.Duration` is nanoseconds | Folded to milliseconds; a constant that is not a whole number of milliseconds is refused (`sub-ms-duration`) |
| Loop variables are per iteration (the module is on Go 1.26) | `let` in a `for` head has the same meaning |
| Wall clock, I/O, logging | `time.Now`, `log`, `os`, `encoding/json` are refused (`io`) |

### Refusals

A refusal is `{ Func, Reason, Detail, File, Line, Column }`. `Reason` is
one of a closed list; `Detail` carries the specific name, which is what
the ratchet counts.

| Reason | Meaning |
|---|---|
| `unknown-call` | A call outside the package with no table entry. Detail: the qualified name |
| `engine-gap` | The table knows the call; the engine lacks it. Detail: the facade name |
| `unknown-type` | A type outside the table |
| `package-var-read`, `package-var-write` | A package variable that is neither constant nor a recognized framework object; or any write |
| `map-range`, `slice-alias`, `string-bytes` | As above |
| `struct-copy`, `pointer-local`, `sized-int`, `fixed-array` | Value semantics the subset does not model |
| `interface-dispatch`, `type-switch`, `closure-escape`, `generic` | Dynamic features |
| `goroutine`, `defer`, `goto`, `fallthrough`, `reflection`, `unsafe` | Control flow and runtime features |
| `foreign-rand`, `io`, `sub-ms-duration` | As above |
| `predicate-constructor`, `custom-counter` | Kinds of body with no v1 rule |
| `callee-refused` | A callee in the same package was refused. Detail: its name |

## The agent loop

A task is one refused function. The orchestrator (`boardgame-util port`)
writes the brief; nothing in it is composed by hand.

### The brief

```text
Translate one Go function to TypeScript. Its behavior must be identical,
including behavior that looks like a bug: do not fix, simplify or guard
anything the Go does not guard. If the Go would panic, the TypeScript must
throw the same text (see "Panics").

FUNCTION
  <FnRef>, role <role>, from <file>:<line> at Go source hash <hash>

GO SOURCE
  <the function, verbatim>

WHY THE TRANSLATOR REFUSED IT
  <each refusal: reason, detail, the source line>

TYPES IT TOUCHES
  <each Go type the function names: its definition, verbatim, with struct tags>
  <for each state property it reads or writes: the PropDef from the definition>

FUNCTIONS IT CALLS (already translated; call them, do not rewrite them)
  <each in-package callee: its Go signature and its TypeScript signature>

THE TYPESCRIPT API YOU MAY USE
  <engine/go/api.ts, num.ts, str.ts, slice.ts, panic.ts: exported signatures with doc comments>
  <the API table rows for every Go call in this function>

SEMANTICS
  <the "Semantics that differ" table>

EXAMPLES
  <three stored pairs: Go source and accepted TypeScript, chosen by nearest refusal reason, then by role>

OUTPUT
  One file, <path>, containing exactly:
    export function <ts name>(<signature>): <result> { ... }
  Imports only from ../../go/*.ts and ../../lib/*.ts, with the .ts extension.
  No classes, no module-scope let, no objects of functions, no double casts,
  no Math.random, Date, BigInt, async, or anything outside the engine.

HOW TO CHECK
  boardgame-util port check <FnRef>
  It type-checks your file, runs the sandbox and purity checks on it, and
  replays the game's fuzz corpus with your function in place. It prints
  PASS, or the first difference: version, viewer, field, the JSON path, what
  Go recorded and what you produced, and the arguments your function was
  last called with.
```

### What the agent may read and run

It may read the brief and its own output file. It may run
`boardgame-util port check` for its function. It does not get the game's
other Go source, the hand ports, or the rest of the generated game: a
translation that needs them is using knowledge the Go function does not
contain, and that knowledge belongs in the brief's inputs (the definition,
the callee signatures), where the next agent gets it too.

### Validation and storage

1. Static: the file has one exported function with the required signature;
   `sandbox.test.ts` and `rules/purity.test.ts` rules pass on it; it
   type-checks under `tsconfig.sandbox.json`.
2. Differential: the game's fuzz corpus passes with the function in place.
3. Coverage: the Go coverage of this function in that corpus is recorded.
   A function with unreached blocks is accepted only as "unproven", listed
   in the game's report, and gets a directed scenario (below).
4. Mutation: the mutations under "The fuzzer" are applied to the function;
   surviving mutants are listed in the report.

An accepted translation is stored as a pair (see "The ratchet") and copied
into `bodies.gen.ts` with the header
`// @generated agent:<model id> brief:<sha256> go:<FnRef> sha256:<hash> first:<commit>`.

### Determinism and review

The checked-in text is the artifact; nothing regenerates it. A function is
sent to an agent again only when its Go source hash changes or its fuzz
run stops passing. Each accepted translation is read by a second agent in
a fresh context, given the Go, the TypeScript and the facade signatures,
and asked only to list behaviors that differ; any item it lists becomes a
directed scenario that must pass in both engines before acceptance. A
function that feeds a computed property or a legality verdict is also read
by a person, since those reach what a viewer may learn. Whether that
second rule is wanted is the owner's call (decision 6).

## The differential fuzzer

### Driver

`boardgame-util/lib/fuzz` runs one game from one seed on the real engine
and returns a history plus the refusals it met. It then calls
`conformance.FromHistory`, unchanged, to produce the document: the same
replay, capture and normalization as the checked-in corpus, so every
version is recorded for every viewer and the exporter re-verifies the
history against the engine while exporting.

A run is described by `Config{ Game, Seed, Policy, MaxProposals,
MaxVersions, RefusalRate }`. From the seed the driver derives, by SHA-256
with its own prefixes (the scheme `scenario.Run` uses): the game id, the
secret salt, the generator for the driver's choices, the player count
within the game's bounds, a value for each variant key, and whether
players are seated. The manual clock is fixed at `conformance.Clock`.

Policies:

| Policy | What it does |
|---|---|
| `random-legal` | At each settled version, enumerates the legal actions and picks one uniformly |
| `bot:<name>` | A registered `bots.Policy` decides for the current actor; `random-legal` fills in when it declines |
| `agent` | The game's `boardgame.Agent` proposes. Not reproducible (it draws from global `math/rand`); allowed in a transient run, never in the checked-in corpus |

### Enumerating legal moves in Go

At a settled version, for each actor (every seated player, and the admin):

1. Take each non-fix-up move from `game.Moves()`, bound to the current
   state with its defaults.
2. Build candidate inputs from `boardgame.BuildMoveInputSchema`: only
   fields a proposer supplies (`required`, `server-defaulted`). A field's
   domain is: both values for a boolean; every player index; every key for
   an enum; for an integer, the candidates of its declared choice
   projection (`api.ProjectMoveChoicesForViewer`) if it has one, otherwise
   the default and every index from 0 to one past the longest stack in the
   state; for a string, the default only (reported as uncovered).
3. Take the product, capped at 4,096 candidates per move; above the cap,
   sample that many with the seeded generator.
4. Keep the candidates for which `game.MoveWithInput` succeeds and
   `move.Legal(state, actor)` returns nil.

To those are added: firing the next timer, when one is active
(`ForceNextTimerWithError`); and seating a player, in a seated run with an
open seat. With probability `RefusalRate` (default one step in eight) the
driver instead proposes a candidate that failed step 4, or a legal move by
the wrong actor or at a stale version, and records the refusal: the
proposal, Go's error text, and that state and version did not change. A
legal proposal whose `Apply` returns an error is recorded the same way.

### Bounds

A run ends when the game is finished, when no action is available, or
when a count is reached: `MaxProposals` (default 300, refusals included)
or `MaxVersions` (default 600). There is no wall-clock limit anywhere. The
engine's own fix-up cap bounds a runaway chain, and a run that hits it is
recorded as the fix-up-cap case the rejection vectors already cover.

### Format

The conformance document gains one optional top-level field, and
`SchemaVersion` becomes 2:

```ts
rejections?: {
  atVersion: number;
  kind: 'propose' | 'timer' | 'seat';
  name: string; input: Record<string, Json>; proposer: number; expectedVersion: number;
  error: string;
}[];
```

Existing documents stay valid as version 1 and are not regenerated. The
TypeScript side reuses `runDocument` (`testing/conformance.ts`) as it is
for versions and views; `testing/fuzz.ts` adds the replay of `rejections`
between versions, asserting Go's text and an untouched state, which
`rules/rejections.test.ts` does today from a separate vector file.

Two corpora:

- **Checked in**: `conformance/fuzz/<game>/seed-NNNN.json.gz`, eight seeds
  per game, with `fuzz corpus --check` as the staleness test. About the
  size of the existing corpus per game.
- **Transient**: `boardgame-util fuzz run --seeds 0..999 --out DIR` writes
  to a directory outside the repository; `games/fuzz.test.ts` reads the
  directory named by `BOARDGAME_FUZZ_DIR`. This is the acceptance run.

### The divergence report

`firstDifference` already yields a JSON path. The report for a failing
document is:

```json
{ "document": "checkers/seed-0412", "version": 37, "viewer": 1,
  "field": "view.game", "path": "CurrentState.Players[0].CapturedTokens.Indexes[2]",
  "want": 14, "got": 9,
  "move": { "name": "Move Token", "proposer": 1, "trigger": "propose" },
  "lastOutsideInput": { "atVersion": 36, "name": "Move Token", "input": { "TokenIndexToMove": 41, "SpaceIndex": 27 } },
  "bodiesExecuted": ["checkers.(*moveMoveToken).Apply", "checkers.(*token).LegalCaptureSpaces"],
  "repro": "boardgame-util fuzz run --game checkers --seed 412 --max-versions 37 --out DIR" }
```

Fields are compared in the runner's order (state, move, status, then each
viewer's game, move, legality and choices), so the first difference is the
most upstream one. A run from the same seed truncated at the failing
version is a prefix of the full run, which gives the minimal document
without a separate shrinker. `bodiesExecuted` comes from a counter the
loader wraps around each body in test builds.

### Coverage

Four measures, all per game, all in the report:

1. **Moves.** For each move type: times applied; and for a non-fix-up
   move, times proposed legally and times refused. Read from the documents.
2. **Predicates.** For each precondition of each declared plan: whether a
   pass, a fail and an unknown verdict were recorded. Read from the ledgers
   the documents already carry.
3. **Go blocks.** The fuzz binary is built with `go build -cover` and run
   with `GOCOVERDIR` set; `go tool covdata` gives per-function block
   coverage of the game package. This is the "branches no recording
   exercised" list, with no instrumentation of our own. (That this build
   works under `./scripts/go-local` was not tried.)
4. **Visibility.** For each property with a policy other than visible:
   whether some viewer other than the owner was shown it while it held a
   non-zero value. Read from the documents.

A gap is closed by a directed scenario (`scenario.Spec`, in
`lib/conformance/corpus/scenarios.go`, as today) or signed off by name in
the game's report.

### Mutations

The engine-rules session made 20 single-line mutations to the hand ports
and the engine, one at a time, and each made the corpus fail on the right
field: an ace comparison, a swapped fix-up order, a visibility policy, the
timer id's draw count, computing from the unsanitized state, a dropped
field tag, a dropped alias, an extra random draw, one word of an error
string, empty choice sets, a round-robin offset, a seated player left
active, repeatable moves in a progression, the any-player target, role
visibility, a vote's group visibility, a tree enum default, a public
shuffle made secret, and a timer's move bound to the wrong state. That was
done once, by hand. The pipeline makes it routine for generated bodies:
for each function, flip each comparison, shift each integer literal by
one, and delete each statement, one at a time, and require the game's fuzz
corpus to fail. Survivors are reported per function. This is the measure
of whether the fuzz corpus is adequate for that function; block coverage
alone is not.

## The ratchet and its metrics

### Stored pairs

```text
boardgame-util/lib/transpile/testdata/pairs/<game>/<ts function name>/
  source.go        the function, verbatim, at the hash in meta.json
  accepted.ts      the accepted TypeScript, verbatim
  meta.json        { func, role, sourceHash, producer, refusals: [...],
                     fuzz: { seeds: [...], blocksCovered, blocksTotal, mutantsKilled, mutantsTotal } }
```

A pair is stored for every agent translation. The six hand ports seed the
store on day one: each of their about 99 functions is split into this
layout with producer `hand`, giving the translator a reference corpus
before any agent runs.

### Turning a refusal reason into a rule

`boardgame-util transpile --stats` lists refusals by reason and detail
across all games. A rule is proposed for the top entry. It is proven when,
for every stored pair that carried that refusal:

1. the translator now accepts the Go source;
2. with the translator's output substituted for `accepted.ts`, the game's
   checked-in fuzz corpus and conformance documents pass; and
3. the pair's recorded block coverage is complete. A pair with unreached
   blocks does not count as proof and is listed as such.

Textual equality with `accepted.ts` is reported and not required. This is
a Go test (`TestRuleAgainstPairs`) plus the TypeScript replay; both run on
every change to `lib/transpile`.

### After each rule

`boardgame-util port --all --dry-run` re-runs every game and rewrites only
the reports. A function that was agent-translated and is now mechanically
translatable is counted as mechanical in the dry-run metric and keeps its
checked-in text. `boardgame-util port PKG --adopt FnRef` replaces the text
on purpose, after the same three checks.

### The per-game report

`fabric/transpile/reports/<game>.json`, and a table in
`fabric/transpile/REPORT.md`:

| Field | Meaning |
|---|---|
| `movesData`, `movesWithCode` | Move configs that are data only, and the rest |
| `functions.data` | Bodies the exporter evaluated (constructors, counters fitted to a family) |
| `functions.mechanical` | Translated by the translator, checked in |
| `functions.agent` | Translated by an agent, checked in |
| `functions.agentNowMechanical` | Of those, how many a dry run now translates |
| `functions.engineGap` | Refused only as `engine-gap`; each named |
| `functions.notPorted` | Diagram and agents |
| `coverage` | The four measures, and surviving mutants |
| `stale` | Functions whose Go source hash differs from the header |

The one number to watch per game is `functions.agent -
functions.agentNowMechanical`. It should fall.

## The declarativeness lint

`boardgame-util lint --declarative [--json] PKG...`, in
`boardgame-util/lib/lint/declarative.go`, beside the existing preflight
and using the same loaded package. It is the static half of the exporter
and the subset checker, run without writing anything. It reports and does
not fail, unless `--max-refused N` is given.

Per package it prints the baseline row of this document, measured:

- each move config, as `data`, or with the methods the game declares;
- each move whose legality is an imperative `Legal` override, with the
  observation that its checks are invisible to viewers (werewolf has six
  such configs and no declared precondition);
- each function reaching engine output with its verdict, and for a refusal
  the reason, detail and position;
- constructs with no deterministic answer, as warnings whatever the
  verdict: `range` over a map, global `math/rand`, `ValueFromString`;
- totals: the share of moves that are data, the share of legality that is
  declared, and the count in each bucket.

New diagnostic codes continue the existing series: `BGLINT0101` (function
needs translation), `BGLINT0102` (function refused), `BGLINT0103`
(nondeterministic construct), `BGLINT0104` (generated TypeScript is stale
for this function).

## Work breakdown

Twelve workstreams. Each owns its files outright; no file has two writers.
A freeze at the start fixes the interfaces the streams share.

### Phase 0: freeze (one agent, first)

Lands type-only and schema-only changes, then everyone else starts:
`GameData` Go structs and TypeScript types; the additions to
`rules/types.ts` and `definition.ts`; conformance `SchemaVersion` 2 with
`rejections`; the facade's exported signatures with bodies that throw; the
refusal list as `lib/transpile/refusal.go`. Acceptance: `go build ./...`
and the TypeScript type check pass; no test changes result.

### Phase 1: exporter, fuzzer and engine gaps (all parallel after the freeze)

| # | Workstream | Owns | Depends on | Acceptance |
|---|---|---|---|---|
| 1 | Exporter, run-time half | `lib/gamedef/export.go`, `definition.go`, `cmd_gamedef.go`, `testdata/gen/defs/main.go`, `testdata/gen/vecgame/defs.go`, `conformance/definitions/` | 3 | `gamedef export` writes all seven games; `--check` passes; the five existing `testdata/defs/*.def.json` are reproduced byte for byte, as the projection without the new fields |
| 2 | Exporter, static half, and the lint | `lib/gamedef/static.go`, `hash.go`, `lib/lint/declarative.go`, `cmd_lint.go` | freeze | For all seven games the manifest lists every function this document's table counts; `lint --declarative --json` reproduces the table's first four columns exactly, and any difference in the others is explained in the report |
| 3 | Go descriptions | `moves/describe.go`, `legal_plan_export.go`, their tests | freeze | For every move of every example game, `LegalPlanSpecs` equals the names and arguments in the recorded ledgers; every blackjack, werewolf and checkers progression is described, with counters fitted. No existing test changes |
| 4 | Fuzz driver | `lib/fuzz/`, `cmd_fuzz.go`, `lib/conformance/document.go` (the one field), `conformance/schema.ts`, `conformance/fuzz/` | freeze | Eight seeds for each of the six ported games are generated and reproduce under `--check`; every move type of each game is applied at least once across the eight, or named; a refusal is recorded in each game |
| 5 | Fuzz replay and coverage | `testing/fuzz.ts`, `games/fuzz.test.ts`, `lib/fuzz/coverage.go`, the mutation harness | 4 | The six hand ports pass their checked-in fuzz corpus, or each difference is filed as a hand-port bug with its report. The 20 historical mutations each fail the fuzz corpus |
| 6 | Loader | `rules/data.ts`, `rules/data.test.ts`, `rules/stockRegistry.ts`, `rules/types.ts`, `rules/compile.ts`, `rules/progression.ts`, `definition.ts` | freeze; 1 for real data | `rules/data.test.ts`: each hand port rebuilt from its exported definition, with its own functions as bodies, passes the whole conformance corpus |
| 7 | Facade | `go/api.ts`, `num.ts`, `str.ts`, `slice.ts`, `panic.ts`, and their tests | freeze | Vector tests from Go (a new generator beside `testdata/gen/`) for `idiv`, `atoi`, `strCompare`, nil and index panics' text, nil-aware append; `sandbox.test.ts` and the purity tests pass on `go/` |
| 8 | Stack constraints and enum-keyed access | `stack.ts`, `state.ts`, `constraints.ts`, `testdata/gen/stackvec/` | freeze | Stack vectors extended with a constrained destination and `ByKey` operations match Go; tictactoe's and checkers' constraints load from the definition |
| 9 | Stock moves, behaviors, enums, graphs | `rules/stock.ts`, `rules/enums.ts`, `lib/enumrange.ts`, `lib/graph.ts`, `lib/dice.ts`, `lib/playingcards.ts` | freeze | Every stock helper exposes its parts; `moves.DefaultComponent`, the color behaviors, range enums and grid connectedness match Go vectors; the six hand ports still pass conformance |

Streams 1 to 5 are the agreed first step. With them alone a port is "run
the exporter, write the functions by hand against the loader, run the
fuzzer", and the lint has measured how much imperative code each game
holds.

### Phase 2: translator and agent loop

| # | Workstream | Owns | Depends on | Acceptance |
|---|---|---|---|---|
| 10 | Translator | `lib/transpile/` except `testdata/pairs/`, `cmd_transpile.go` | 2, 7 | Pig translates with no refusal and its generated game passes pig's conformance and fuzz corpora. Across the six ported games, the functions accepted and refused match the lint's verdicts |
| 11 | Pair store and rule proof | `lib/transpile/testdata/pairs/`, `TestRuleAgainstPairs`, `fabric/transpile/reports/`, `REPORT.md` | 5, 10 | The hand ports are split into pairs; the report exists for seven games; a deliberately broken rule fails the proof |
| 12 | Orchestrator and agent harness | `cmd_port.go`, `lib/port/` (brief writer, `port check`), `games/<game>/` generated output | 5, 6, 10 | `port check` on memory's refused function prints PASS for the hand port's version and the right first difference for a mutated one; a brief is byte-identical across two runs |

### Phase 3: checkers, end to end

Checkers is the first game through the whole pipeline. It is the only
game in this repository without a port, so the choice is forced; it is
also the right one. It has the richest declared legality of any example,
it is the one game whose moves go into a constrained stack, and it needs
every gap in streams 8 and 9, a custom predicate (the one function bound
for an agent), and dynamic component values. It is not in the conformance
corpus, so the fuzzer is its only oracle, which is the situation every
later game will be in.

Pig goes through first as a shakedown, in stream 10's acceptance: it is
the smallest game, the hand port is the known answer, and it exercises the
generator, a library call and dynamic values.

Order for checkers: exporter and lint (definition, measured row); fuzz in
Go alone, to learn whether a full game can be played (see risk 1);
directed scenarios for what the fuzzer does not reach; translate; the
agent for the predicate constructor; accept on a clean 1,000-seed run with
the coverage report signed.

### Phase 4: the rest, and the ratchet

Re-derive the six hand ports through the pipeline, in this order, each
compared against its hand port and its existing conformance documents:
tictactoe (tuple results, string slices), memory (timers, snapshot loops,
the package-variable refusal), blackjack (progressions as data, the nil
panic, generic values), debuganimations (volume: 24 functions),
werewolf (maps, nil slices, a permutation, groups, no declared legality).
Whether the generated versions then replace the hand ports is decision 3.
After that: the generated stub games under
`boardgame-util/lib/stub/testdata/`, then the GAMES repository.

### What can start now

The freeze, and after it streams 1 to 9 at once. Stream 4 can begin before
the freeze on everything but the `rejections` field.

## Risks and open questions

### Risks

1. **Checkers may not be playable in Go.** Read, not run:
   `(*moveMoveToken).Apply` computes the "middle" space as finish minus
   start, not their midpoint (`examples/checkers/moves.go`), and passes it
   to `RangeToValue`. For a move toward a lower row or column an index is
   negative, which should give `IllegalValue` and an error from `Apply`
   after `Legal` passed; for other moves it names an unrelated space and
   captures whatever opposing token is there. The tests cover legality
   only. `spaceIsBlack` is `index % 2 == 0`, which on an eight-wide board
   marks columns, not a checkerboard. If the fuzzer confirms this, a
   faithful port is a port of a broken game.
2. **The examples are not the workload.** 117 of 119 functions in the
   subset says the examples are small, not that real games are. The GAMES
   repository is not on this machine. The lint is built first so that the
   number can be measured there before more is invested in the translator.
3. **Engine gaps dominate.** About 34 stock move types and 9 behaviors are
   unported. Each is hand work with its own vectors. The pipeline's
   through-put on a new game is bounded by this, not by translation.
4. **Coverage of sanitized-state behavior.** A computed property or
   precondition that differs only for a viewer who is not the owner needs
   that viewer and that state in a recording. The visibility measure
   reports it; nothing forces it.
5. **Fuzz corpus size.** Every viewer at every version is about 2.5 MB of
   JSON per document today. A thousand-seed run is transient for that
   reason; if it is too slow on the TypeScript side, views can be compared
   at settled versions only for the transient run.
6. **Identity churn.** Any change to a generated game's text changes the
   policy identity of a pattern that imports it and strands tables in
   flight. The header and adoption rules above limit that to real changes;
   they do not remove it.
7. **The agent's isolation can be defeated by the check command**, which
   reveals the game's behavior through differences. That is intended: it
   is the same information a failing test gives.

### What must not change in existing Go

No existing function changes behavior. The Go added is new packages under
`boardgame-util`, one new file in `moves`, one new file in the core, and
one optional field in the conformance document. The checked-in conformance
corpus stays byte for byte (`conformance export --check` passes),
`LegalCatalogVersion` and both schema fingerprints are unchanged, and no
move, property or struct is renamed. Fixing any Go bug below is a separate
change, approved by the owner, that regenerates the corpus and the
affected ports together.

### Go bugs the hand ports reproduced, and how the pipeline treats them

The rule is the same for all: the port does what Go does. What differs is
which tool carries the behavior.

| Bug (source) | Carried by |
|---|---|
| Blackjack scores ten plus ace as 11 (`examples/blackjack/state.go:96`); pig's chosen starter never gets the first turn; blackjack's turn order shifts each round; memory's canceled and fired timers end differently | The translator, translating the game's Go as written. The agent brief forbids fixes. The mutation run confirms the corpus pins the ace comparison |
| `ForceFinishTurn` ignores its phase; `AnyPlayer.Legal` mutates the move; `DealCountComponents` can index -1; the no-winner case for negative scores; the fix-up cap leaves 258 versions; move names matched without case; `MoveWithInput` names moves two ways | The engine's stock moves and step, already reproduced in `rules/`. New stock ports (stream 9) follow the same rule and get vectors |
| Sanitization of dynamic values depends on map order and can panic; a non-visible policy on a merged stack panics; `ContractSize` leaves a stale index; an empty concatenation serializes `null`; sorting an empty stack leaves `null`; unstable but deterministic sort; `IDsLastSeen` drops version 0; `len` on a sized stack; a hidden player index becomes 0; player-state group membership | The engine core, already reproduced or documented as unmatched (engine-core divergences 1 to 12). The lint warns when a game has the shape that makes divergence 1 reachable: a visible merged stack over hidden parts of a deck with dynamic values |
| The full ledger runs `LegalCustom` after a failed check, so it can panic; `any` ignores an admin policy; an out-of-range negative player index panics; `playerBool` and `playerBoolAt` bind differently | The legal evaluator, already reproduced. The facade's panic texts keep a translated `LegalCustom` matching the recorded unknown entries |
| The exporter writes `winners: null` at unsettled versions; the choice fingerprint of an empty schema hashes `null` | The runner's one undeclared normalization, and the engine. Unchanged; the one-line exporter fix remains open (decision 8) |
| `enum.ValueFromString` and input-field error order depend on map order | No tool can: Go has no single answer. Refused by the translator and flagged by the lint |
| New in this reading: checkers' middle-space arithmetic and `spaceIsBlack`; memory's discarded capture error; the misspelled `DefaultNumPlayeres` in memory and debuganimations | Checkers: risk 1 and decision 2. Capture: the facade's error-return shape reproduces Go, and the hand port is corrected when memory is re-derived. The misspelling: the exporter reads the value Go actually uses |

A register of these, one row per bug with its status, is kept in
`fabric/transpile/GO-BUGS.md` so that a later fix in Go is matched by a
regeneration and not rediscovered as a divergence.

### Decisions for the owner

1. **Two additive files in existing packages**: `moves/describe.go` and a
   core `legal_plan_export.go`. The alternative is reflection over
   unexported fields, which cannot read a counter at all.
2. **Checkers**: if the fuzzer confirms risk 1, fix the Go first and port
   the fixed game, or port it as it is.
3. **Do generated games replace the six hand ports?** Replacing them
   proves the pipeline and removes duplicate code; it changes the policy
   identity of the memory and blackjack patterns.
4. **Generated code targets a Go-shaped facade**, so it reads like
   transliterated Go, not like the hand ports. The gain is a translator
   with no judgment in it.
5. **What a header may contain.** Whether comments are part of a pattern
   module's identity was not determined. If they are, the header's fields
   are part of the policy and should be the minimum: producer, function,
   source hash.
6. **Human review** of agent translations that feed computed properties
   or legality.
7. **AI players** (`boardgame.Agent`, `bots.Policy`) are not ported by
   this plan. Memory's agent is ten functions using JSON and global
   randomness.
8. **The exporter's `winners: null`.** Fixing it rewrites every checked-in
   document.
9. **How many fuzz seeds are checked in** (eight per game proposed), and
   how many an acceptance run needs (1,000 proposed).
10. **Refactoring Go to be translatable.** Memory's refused function
    becomes translatable by replacing a variable with a constant. Whether
    the lint should push game authors that way, or the Go is left alone
    and an agent takes it.
11. **Integer overflow**: refuse at save, as proposed, or emulate 64-bit
    wraparound with the existing `u64.ts`, which is slower and makes every
    arithmetic expression a call.

### What could not be determined by reading

- The contents of the GAMES repository; it is not on this machine.
- Whether checkers' `Apply` fails as it reads; nothing was run.
- That reflection cannot tell a promoted method from a declared one; this
  is from knowledge of the compiler.
- Whether `sort.Slice` orders equal elements exactly as the ported
  `sort.Sort` does.
- Whether a pattern module's identity includes comments, and whether a
  pattern may import JSON.
- Whether a `-cover` build of the fuzz binary works under
  `./scripts/go-local`, and what Node's coverage reports for generated
  TypeScript; no Node command could be run in this session.
- The size and replay time of a fuzz document for a game with many
  players.
- Whether fitting a counter to the stock families by evaluation is exact
  for every stock counter at every length; the error texts read so far
  (`moves/count.go`) are shared between families, so the fit relies on the
  results over the grid and needs its own test.

## Considered and not pursued: compiling Go to WebAssembly

Compiling the Go game and core to WebAssembly would remove translation
altogether. It was set aside because it puts Go at run time, and because
the pattern sandbox's support for WebAssembly was not checked. Reading the
code gives no reason to reopen it and adds three against: the core is
built on goroutines and channels (a proposal is a channel receive,
`<-game.ProposeMove`), and Go's WebAssembly runtime needs a host event
loop, timers and a random source at start-up, all of which the engine's
contract forbids in a handler that runs one synchronous step; the module's
identity would be a hash of a compiler's output, which changes with the
toolchain and cannot be read by a reviewer; and it does nothing for
legality evaluated in the client. Its one remaining use, as a test oracle
in a browser, is covered by running the native Go engine in the fuzzer.
