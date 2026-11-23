# engine

A TypeScript reproduction of the Go core's game semantics, held to the Go
implementation by known answers that real Go produced. A game defined here
runs with no Go: the engine creates it, applies proposed moves and the
fix-up moves that follow, and says what each viewer may see.

The plan this implements is `fabric/PROPOSAL.md`, "The engine: reproducing
the core in TypeScript". It is built in three layers:

| Layer | Directory | What it is | Status report |
|---|---|---|---|
| Core | this directory | The generator, ids, stacks, conservation, sanitization | `fabric/status/engine-core.md` |
| Legality | `legal/` | The declarative legal-predicate evaluator and plan | `fabric/status/engine-legal.md` |
| Rules | `rules/`, `games/` | The game-definition API, the engine step, timers, per-viewer output, the loader for exported definitions, and seven games generated from their Go | `fabric/status/engine-rules.md`, `fabric/transpile/status/` |

Start with [Defining a game](#defining-a-game) to write a game, or
[What the host does](#what-the-host-does) to run one.

## Constraints

The engine runs inside a sandboxed JavaScript handler that handles one
proposed move per invocation. Every file outside `testing/` and `*.test.ts`
is pure TypeScript:

- no DOM, no Node APIs, no `crypto`, no timers, no clock, no `Math.random`;
- no `BigInt`, no `TextEncoder`, no asynchrony;
- no dependencies: engine modules import only each other, always with the
  `.ts` extension.

`purity.test.ts` and `rules/purity.test.ts` enforce this by reading the
sources. Nothing is kept between calls: everything that persists is what the
host stores.

### No classes, and nothing but functions and data at module scope

The engine loads inside a Common Fabric pattern, whose sandbox checks every
module-scope statement before a module may run. Its rule
(labs `docs/specs/sandboxing/SES_SANDBOXING_SPEC.md`, section 4.2) is that a
module-scope binding is a function, or data it can freeze: no class, because
a class "creates mutable prototype/static state", no `let`, no record of
functions. That is deliberate, not a gap, so the engine is written to it:

| Instead of | The engine has |
|---|---|
| `class Stack { ... }` and `new Stack(...)` | `interface Stack` and a factory, `newStack(...)`, returning an object whose methods are its own properties. Calls read the same: `stack.moveComponent(i, dest, slot)` |
| A static method, `EngineState.load(...)` | A function, `loadEngineState(...)` |
| `x instanceof MergedStackView` | `isMergedStackView(x)`, over a `merged` field every stack view carries |
| `class StackError extends Error`, `throw new StackError(m)`, `e instanceof StackError` | `throw stackError(m)`: an ordinary `Error` whose `name` is `'StackError'`; `isStackError(e)` |
| `export const memory: GameSpec = { ... }` | `export function memory(): GameSpec { return { ... }; }`. A spec holds functions, so it is built when asked for |
| A table of functions, `{ '==': (a, b) => a === b }` | A `switch` in a function |
| `x as unknown as T` | One cast, or a typed helper. The pattern compiler rejects the double cast |

Methods of one of these objects use `this`, so call them on the object; do
not detach them. `private` no longer exists: an interface lists every member,
and the ones a caller should leave alone say so.

`sandbox.test.ts` enforces all of it by reading the sources.
`npx tsc --noEmit -p src/engine/tsconfig.sandbox.json` type-checks the engine
under the one extra compiler option a pattern is compiled with
(`exactOptionalPropertyTypes`). Neither needs labs. Loading in a real
pattern is checked from `fabric/patterns/`.

State is plain JSON in exactly the shape the Go server serializes, so a
recorded Go state loads directly, and what the engine serializes is what Go
would have.

## Layout

| File | What it is | Go source it ports |
|---|---|---|
| `u64.ts` | 64-bit integers as uint32 pairs; UTF-8 encoding | — |
| `fnv.ts` | 64-bit FNV-1 | `hash/fnv` `New64` |
| `rand.ts`, `rand-cooked.ts` | `math/rand` v1: the generator and `Int63`, `Uint32`, `Int31`, `Int31n`, `Int63n`, `Intn`, `Float64`, `Float32`, `Perm`, `Shuffle`; and the seed the core derives for a state | `math/rand`; `state.go` `(*state).Rand` |
| `sha1.ts` | SHA-1 | `crypto/sha1` |
| `gosort.ts` | `sort.Sort` (pdqsort), including its order among equal elements | `sort` |
| `ids.ts` | Component ids; the empty-slot and generic-component sentinels | `component.go` `ID`, `secretMoveCount` |
| `definition.ts` | `GameDefinition`: decks, enums, property schema and sanitization policies; struct-tag parsing | `struct_inflater.go`, `sanitization.go` `policyFromString` |
| `stack.ts` | `Stack` (growable and sized) and `MergedStackView` | `stack.go` |
| `constraints.ts` | Stack constraints from struct tags (`max`, `unique`, `same`, `maxdistinct`): parsing `PropDef.Constraints` and checking them | `stack_constraint.go`, `struct_inflater.go` `unpackStackStructTag`, `constraints/` |
| `state.ts` | `EngineState`: the state JSON plus identity, version, generator, stack handles, conservation | `state.go`, `stack_ownership.go`, `game_manager.go` `stateFromRecord` |
| `sanitize.ts` | Policies, groups, visibility facets, `sanitizedForPlayer`, `viewFor` | `sanitization.go`, `state_visibility.go`, `base/game_delegate.go` |
| `index.ts` | The public surface | |
| `legal/` | The legal-predicate evaluator; see `legal/README.md` | `legal/`, `legal_*.go` |
| `rules/types.ts` | The game-definition API: `GameSpec`, `MoveSpec` | a game's delegate, state structs and `ConfigureMoves` |
| `rules/compile.ts` | `compileGame`: derives the definition, moves, legal plans, schemas | `NewGameManager` |
| `rules/state.ts` | `RulesState`: a state as the rules see it; move legality; timers | `state.go`, `base/game_delegate.go`, `moves/default.go` |
| `rules/engine.ts` | `createGame`, `propose`: the engine step and fix-up loop | `game.go`, `move_input_bind.go` |
| `rules/timers.ts`, `rules/time.ts` | Timers as state; RFC 3339 instants | `timer.go` |
| `rules/view.ts` | `viewerOutput`: what one viewer is sent for one version | `game.go` `JSONForPlayer`, `move_sanitization.go`, `server/api` |
| `rules/progression.ts` | Ordered move progressions and tape matching | `moves/group_*.go`, `moves/count.go` |
| `rules/stock.ts` | Installers, behavior properties, and helpers for the reusable moves | `moves/`, `behaviors/` |
| `rules/binding.ts`, `rules/enums.ts` | The legal evaluator's state access, over the real state | — |
| `rules/sha256.ts` | SHA-256, for the schema fingerprints | `crypto/sha256` |
| `rules/data.ts` | `gameFromData`: a `GameSpec` from an exported definition (`GameData`) and a table of function bodies | `boardgame-util gamedef export` |
| `rules/stockRegistry.ts` | `stockMove(type, config)`: a stock move's parts from the configuration Go exports | `moves/` |
| `go/` | The facade translated Go calls: slices, strings, integer arithmetic, panics, and the state API in Go's shape | the Go standard library and core API |
| `lib/` | Ports of library packages: range enums, grid graphs, dice, playing cards | the packages `rules/stock-inventory.json` names |
| `games/<game>/` | `memory`, `blackjack`, `tictactoe`, `pig`, `debuganimations`, `werewolf`, `checkers`, each generated from its Go: `definition.gen.ts`, `bodies.gen.ts`, `index.ts` (see [From Go](#from-go-the-shipped-games)) | `examples/<game>` |
| `games/*.test.ts` | Every game against the conformance and fuzz corpora | |
| `testing/` | Test support: a port of `internal/patchtree`, a reader for golden history files, and the conformance runner | |
| `testdata/` | Known answers from Go, and the Go programs that produce them (`testdata/gen/`) | |

## Defining a game

A game is a function returning a `GameSpec` (`rules/types.ts`): plain data
plus functions. There are no base classes, decorators or reflection.
Everything a Go game spreads across its delegate, its state structs with
their tags and embedded behaviors, and `ConfigureMoves` is in one object.

There are two ways to get one, and the engine runs both the same way:

```ts
import { compileGame } from './rules/index.ts';
import { memory } from './games/memory/index.ts';

const game = compileGame(memory());   // once; throws a GameSpecError or a LegalSpecError on a bad spec
```

`compileGame` does what Go's `NewGameManager` does at boot: it derives the
state definition, resolves every move's fields, visibility and legal plan,
and builds the schemas a client is told about.

- **Generated from a Go game.** All seven games under `games/` are made this
  way, by `boardgame-util`, from `examples/<game>`. Nothing in them is
  written by hand; one function each in memory and checkers, which the
  translator refuses, was translated by a coding agent and checked against
  the recordings. See [From Go](#from-go-the-shipped-games).
- **Written by hand** against the API the rest of this section describes.
  The six games were first ported this way, move for move (the hand ports
  `games/<game>.ts`); the generated games replaced them once each matched
  Go on every recording there is (`fabric/transpile/status/replacement.md`).
  The hand ports' functions are kept, one per Go function, in
  `boardgame-util/lib/transpile/testdata/pairs/`. A hand-written spec is
  still the way to write a game that has no Go, and the excerpts below
  (memory, as its hand port wrote it) show that API.

### From Go: the shipped games

A generated game is three checked-in files under `games/<game>/`, each
headed `DO NOT EDIT`:

| File | What | Written by |
|---|---|---|
| `definition.gen.ts` | `definition()`: the game's exported definition (`GameData`: the state definition and chest, every move's configuration, plan, phases and progression, variants, computed properties, and the name of the Go function behind every behavior) as one literal | `boardgame-util transpile`, from `conformance/definitions/<game>.json`, which `boardgame-util gamedef export` writes from the running Go game |
| `bodies.gen.ts` | One TypeScript function per Go function the definition names, translated from the Go source against the Go-shaped facade in `go/`, each under a header naming the Go function and its source hash; and `bodies(ref)`, the table from those names to the functions | `boardgame-util transpile`; a function the translator refuses is translated by a coding agent and stored by `boardgame-util port accept` under an `@generated agent:` header |
| `index.ts` | `export function memory(): GameSpec { return gameFromData(definition(), bodies); }` | `boardgame-util transpile` |

So memory's state is data, as Go's struct tags declared it
(`definition.gen.ts`):

```ts
HiddenCards: {
  Type: 'sizedStack',
  Sanitize: {
    all: 'order',
  },
  Deck: 'cards',
  Size: 40,
},
```

and its Hide Cards move is its configuration and declared plan
(`Plan.Specs`: `proposerIsCurrentPlayer`, then the two preconditions from
`ConfigureMoves`, each with its message) plus one translated function,
`(*moveHideCards).Apply` (`bodies.gen.ts`):

```ts
export function memory_moveHideCards_Apply(state: RulesState, m: MoveInstance): string | null {
	state.timer('HideCardsTimer').cancel();
	{
		const $s2 = components(gameStack(state, 'VisibleCards'));
		for (let $i1 = 0; $i1 < lenOf($s2); $i1++) {
			let i: number = $i1;
			let c: Comp | null = $s2![$i1];
			if (c !== null) {
				{
					let err: string | null = moveTo(c, gameStack(state, 'HiddenCards'), i);
					if (err !== null) {
						return 'Couldn\'t move component: ' + nonNil(err);
					}
				}
			}
		}
	}
	return null;
}
```

A move built on a Go stock move (`moves.FinishTurn`, `moves.DealCountComponents`,
...) is built by `rules/stockRegistry.ts` from the configuration Go exports;
a translated `Apply` that calls its embedded stock move's `Apply` calls
`stockApply`. The order of the moves in `Moves`, which decides which fix-up
the engine applies first, is Go's registration order.

The pipeline, from the repository root (Go through `./scripts/go-local`):

```bash
./scripts/go-local run ./boardgame-util gamedef export                      # conformance/definitions/<game>.json from the running Go games
./scripts/go-local run ./boardgame-util lint --declarative ./examples/memory # which functions translate, and why any does not
./scripts/go-local run ./boardgame-util transpile ./examples/memory          # writes games/memory/
./scripts/go-local run ./boardgame-util transpile --check ./examples/...     # fails if a generated game is stale
./scripts/go-local run ./boardgame-util port --all --dry-run                # the per-game reports (fabric/transpile/REPORT.md): data, mechanical, agent
./scripts/go-local run ./boardgame-util port --all --no-game                # a brief for each refused function (fabric/transpile/briefs/)
./scripts/go-local run ./boardgame-util port check 'memory.(*moveStartHideCardsTimer).Apply' --file PATH                    # an agent's translation against the recordings
./scripts/go-local run ./boardgame-util port accept 'memory.(*moveStartHideCardsTimer).Apply' --file PATH --producer MODEL  # stores it in bodies.gen.ts
```

A generated game is held to Go by `games/<game>/<game>.gen.test.ts` (its
conformance documents and its checked-in fuzz documents, every version,
every viewer, every recorded rejection). The plan and its decisions are
`fabric/TRANSPILE-PROPOSAL.md` and `fabric/status/transpile-pipeline.md`.
After a Go game changes: `gamedef export`, then `transpile`, then the
game's tests; `conformance export --check` and `fuzz corpus --check` say
whether its recordings must be made again.

### Written by hand

The excerpts from here to the end of this section are memory as its hand
port wrote it against the API. The generated memory says the same things as
data and translated functions; the hand port's text is in the pair store.

#### Components

A deck is its components' static values, in deck order. Deck order matters:
component ids are derived from it.

```ts
decks: { cards: { components: newDeck() } },   // [{ Type: '🚴', CardSet: 'general' }, ...]
```

Give `props` when a value is an enum (blackjack's hand port did, for rank and
suit), with the enum itself under `componentEnums`.

#### State, and who may see it

Each property has a type and, optionally, a visibility policy written as the
Go `sanitize` struct tag. The policies are `visible`, `order`, `len`,
`nonempty` and `hidden`; a tag may name groups (`other:len`, `self:visible`).
A game-state property with no tag is visible to everyone; so is a
player-state property.

```ts
game: {
  props: {
    ...currentPlayerProps,                       // CurrentPlayer: whose turn it is
    CardSet: { Type: 'string' },
    NumCards: { Type: 'int' },
    // Everyone sees how many cards are face down and when they move, not which is which.
    HiddenCards: { Type: 'sizedStack', Deck: 'cards', Size: 40, Sanitize: 'order' },
    VisibleCards: { Type: 'sizedStack', Deck: 'cards', Size: 40 },
    // A view: each slot shows the face-up card if there is one, else the face-down one.
    Cards: { Type: 'mergedStack', Deck: 'cards', Overlap: true, Stacks: ['VisibleCards', 'HiddenCards'] },
    HideCardsTimer: { Type: 'timer' },
    UnusedCards: { Type: 'stack', Deck: 'cards' },
  },
},
player: {
  props: { ...seatProps, ...inactivePlayerProps, CardsLeftToReveal: { Type: 'int' }, WonCards: { Type: 'stack', Deck: 'cards' } },
  traits: {
    seat: true,
    inactive: true,
    gameScore: (state, player) => state.playerView(player, 'WonCards').numComponents(),
    turn: { turnDone, resetForTurnStart, resetForTurnEnd },
  },
},
```

Property types are `int`, `bool`, `string`, `playerIndex`, `enum`, the scalar
slices, `timer`, `stack` (growable, optionally capped by `MaxSize`),
`sizedStack` (fixed slots), `mergedStack` (a view over sibling stacks) and
`board`. Every property starts at its zero value; `initial` supplies others
(tictactoe's hand port gave each player a token value).

`traits` say what Go's embedded behaviors and methods say. `rules/stock.ts`
has the property sets to spread in:

| Go | Properties to spread | Trait to set |
|---|---|---|
| `behaviors.CurrentPlayerBehavior` | `currentPlayerProps` | none: a `CurrentPlayer` property is the current player |
| `behaviors.PhaseBehavior` | `phaseProps` (needs an enum named `phase`) | none |
| `behaviors.RoundRobin` | `roundRobinProps` | none |
| `behaviors.Seat` | `seatProps` | `player.traits.seat` |
| `behaviors.InactivePlayer` | `inactivePlayerProps` | `player.traits.inactive` |
| `behaviors.ScoreBehavior` | `scoreProps` | `player.traits.gameScore: scoreGameScore` |
| `behaviors.PlayerElimination` | `eliminationProps` | `player.traits.elimination` |
| `behaviors.PlayerRole` | `roleProps` (needs an enum named `role`) | `player.traits.role` |
| `behaviors.DrawDiscardPair` | your two stacks | `game.traits.drawDiscardPair: { draw, discard }` |
| `TurnDone`, `ResetForTurnStart`, `ResetForTurnEnd` | — | `player.traits.turn` |

Dynamic component values (a die's face, in pig) are declared per
deck under `dynamicComponentValues`.

#### What a function is handed

Every function in a spec receives a `RulesState` (`rules/state.ts`):

| Member | What it is |
|---|---|
| `state.g`, `state.players[i]` | The game and player sub-states as JSON. Read and write scalars here; an enum is its string |
| `state.gameStack(p)`, `state.playerStack(i, p)` | A `Stack` (see [Stacks](#stacks)) |
| `state.gameView(p)`, `state.playerView(i, p)` | Any stack, including a merged one, read-only |
| `state.component(deck, i)`, `state.dynamic(deck, i)` | A component's static and dynamic values |
| `state.currentPlayerIndex()`, `state.next(p)`, `state.previous(p)`, `state.ensureValid(p)`, `state.valid(p)` | Go's `PlayerIndex` arithmetic |
| `state.currentPhaseKey()`, `state.setPhase(key)` | The phase |
| `state.rand()` | The version's generator. Draw only in `apply` and set-up, in the order Go does |
| `state.timer(p)` | A timer property: `active()`, `start(ms, move)`, `cancel()` |
| `state.moveByName(name)` | A move with its defaults, for a timer to propose |
| `state.variant()`, `state.record` | The variant and the rest of the game record |

A function is given the state it should judge. Computed properties are given
each viewer's **sanitized** state, so use the read-only views there.

#### Moves

A move is a `MoveSpec`: a name, whether it is a fix-up, its fields, its
legality, and `apply`. In the source each one is returned from a function
(`function hideCards(): MoveSpec { return { ... }; }`), as the game is; the
excerpts below show the object.

```ts
const hideCards: MoveSpec = {
  name: 'Hide Cards',
  helpText: 'After the current player has revealed both cards ...',
  fixUp: false,                       // a player proposes it
  base: 'currentPlayer',              // has TargetPlayerIndex; only the current player may propose
  preconditions: [
    { ...propCompare('player.CardsLeftToReveal', '<=', 0), message: 'hide.cards_still_to_reveal' },
    { ...stackNotEmpty('game.VisibleCards'), message: 'hide.no_cards_to_hide' },
  ],
  apply(state) {
    state.timer('HideCardsTimer').cancel();
    const visible = state.gameStack('VisibleCards');
    const hidden = state.gameStack('HiddenCards');
    for (let i = 0; i < visible.len(); i++) {
      if (visible.componentAt(i) !== null) visible.moveComponent(i, hidden, i);
    }
  },
};
```

| Part of a move | Field | Notes |
|---|---|---|
| Name | `name`, `helpText` | The name is the client's animation key and appears in move records. Matched without regard to case, as in Go |
| Who may propose | `fixUp`, `base` | A fix-up is proposed only by the engine. `base: 'currentPlayer'` adds `TargetPlayerIndex` (defaulting to the current player) and the check that the proposer is that player. `base: 'anyPlayer'` adds `TargetPlayerIndex` set to whoever proposes |
| Input fields | `fields` | `{ type, sanitize?, input?, codec? }`. `input` is the disposition: `required` (the default; the proposer must supply it), `server-defaulted` (may supply it), `context-owned` (may not; the engine sets it), `unsupported` |
| Defaults | `defaults(state, move)` | Sets field values for a state: what a fix-up is applied with and what the move tray shows |
| Legal preconditions | `preconditions`, `suppress`, `contributes` | Specs from `legal/`, evaluated in order after the base checks. Each viewer receives every verdict |
| Custom check | `legalCustom(state, move, proposer)` | Runs last, only if the preconditions passed. Returns null or the reason |
| Imperative legality | `legal(state, move, proposer, base)` | For a move with no preconditions: call `base()` first (phase and progression), then your own checks |
| Apply | `apply(state, move)` | Mutates the copy it is given. Throw to refuse; nothing is saved |
| Argument visibility | `fields.X.sanitize` | Absent: only the admin sees the argument. `'self:visible'`: also the proposer |
| Name visibility | `nameSanitization`, `hiddenAnimationKey` | Who sees the name; everyone else gets the alias, or no move record at all if there is none |
| Choices | `choices` | Offers the proposer the legal values of one field (the occupied slots of a stack, the players, or an enum) |
| Placement | set by the installers | Which phases, and where in a phase's order |

A move has **declared legality** when it lists `preconditions`, has a
`legalCustom`, or sets `planEnabled`. A declared move's checks form a plan;
other moves are judged by their `legal` function alone and viewers get only
the verdict. Prefer declaring: a client can then show why a move is
unavailable, and evaluate most checks itself.

Memory's reveal shows a move with an input field, a default, a choice set and
a reusable part:

```ts
// "Move the card at CardIndex out of HiddenCards into the same slot of VisibleCards."
const reveal = moveComponentToSlot({ source: 'game.HiddenCards', destination: 'game.VisibleCards', slotField: 'CardIndex', sourceSlotField: 'CardIndex' });

const revealCard: MoveSpec = {
  name: 'Reveal Card',
  helpText: 'Reveals the card at the specified location',
  fixUp: false,
  ...reveal,                                    // base, the slot check, stack names
  fields: { CardIndex: { type: 'int' } },
  choices: { field: 'CardIndex', stack: { scope: 'game', property: 'HiddenCards' } },
  defaults(state, move) { /* the first card still hidden */ },
  suppress: ['mayMoveToSlot'],                  // re-authored below, so the game's own messages win
  preconditions: [
    { ...propAtLeast('player.CardsLeftToReveal', 1), message: 'reveal.no_cards_left' },
    revealableCardAt('game.HiddenCards', 'game.VisibleCards', 'move.CardIndex'),
    mayMoveToSameSlot('game.HiddenCards', 'game.VisibleCards', 'move.CardIndex'),
  ],
  apply(state, move) {
    reveal.apply(state, move);                  // the transfer
    state.players[state.ensureValid(state.g.CurrentPlayer as number)].CardsLeftToReveal--;
  },
};
```

and its two fix-ups show how the engine carries play forward. After every
move the engine applies the first fix-up move, in registration order, that
is legal for the admin, and repeats until none is:

```ts
const captureCards: MoveSpec = {
  name: 'Capture Cards', helpText: '...', fixUp: true,
  preconditions: [ /* two cards showing; their Type values equal */ ],
  apply(state) { /* move both to the current player's WonCards */ },
};

const startHideCardsTimer: MoveSpec = {
  name: 'Start Hide Cards Timer', helpText: '...', fixUp: true,
  preconditions: [ /* two cards showing; their Type values differ */ ],
  legalCustom: state => (state.timer('HideCardsTimer').active() ? 'the timer is already active' : null),
  apply(state) {
    state.timer('HideCardsTimer').start(4000, state.moveByName('Hide Cards'));
  },
};
```

A fix-up whose legality is too generous is applied forever; the engine stops
after 256 levels of recursion with `ERR_TOO_MANY_FIX_UPS`, having applied
them.

#### Stock moves and installers

`rules/stock.ts` has the reusable moves as functions. Each returns a
`MoveSpec`, or the parts of one to spread into yours and an `apply` to call.

| Helper | Go | What it does |
|---|---|---|
| `finishTurn()` | `moves.FinishTurn` | Fix-up: when the current player's turn is done, the next player's begins |
| `forceFinishTurn({ name })` | `moves.ForceFinishTurn` | The admin ends a turn that is not done |
| `startPhase({ phase, phaseName })` | `moves.StartPhase` | Fix-up: enters a phase |
| `noOp()` | `moves.NoOp` | Fix-up that does nothing |
| `dealCountComponents({ gameStack, playerStack })` | `moves.DealCountComponents` | Fix-up: one component to each player in turn, round robin |
| `shuffleDiscardIntoDraw()` | `moves.ShuffleDiscardIntoDraw` | Fix-up: when the draw stack is empty |
| `drawToPlayer({ gameStack, playerStack })` | `moves.DrawToPlayer` | Parts: the current player draws the top component |
| `moveComponentToSlot({ source, destination, slotField })` | `moves.MoveComponentToSlot` | Parts: a component into the slot the player chose |
| `anyPlayerLegal` with `base: 'anyPlayer'` | `moves.AnyPlayer` | A move any seated player makes for themself (werewolf's votes) |
| `seatPlayer()` | `moves.SeatPlayer` | Applied when the host injects a seat |
| `activateInactivePlayer()`, `inactivateEmptySeat()`, `waitForEnoughPlayers()` | the same names | Gathering players |
| `defaultRoundSetup()` | `moves.DefaultRoundSetup` | The usual three, as an ordered group |

Moves are installed in registration order. The installers place them:

```ts
moves: add(seatPlayer(), activateInactivePlayer(), revealCard, hideCards, finishTurn(), captureCards, startHideCardsTimer),
```

| Installer | Meaning |
|---|---|
| `add(...moves)` | Legal in every phase |
| `addForPhase(phase, ...moves)` | Legal only in that phase |
| `addOrderedForPhase(phase, ...items)` | Legal only in that phase and only in this order. An item is a move, or `optional(move)`, or `serial(...)`. The last must end the phase (`startPhase` or `noOp`) |
| `combine(...lists)` | Concatenates |

Blackjack's hand port used all four: its deal and its round clean-up are
ordered phases that run entirely as fix-ups.

#### Set-up, variants, computed properties and the end

```ts
variants: {
  numcards: { displayName: 'Number of Cards', description: '...', default: 'medium',
              values: { medium: { description: '...' }, small: { ... }, large: { ... } } },
},
setUp: {
  begin(state, variant) { /* size the layout from variant.numcards */ },
  distribute: state => state.gameStack('UnusedCards'),   // where each component starts
  finish(state) { /* shuffle, deal pairs into HiddenCards, shuffle again */ },
},
computed: [
  { name: 'CurrentPlayerHasCardsToReveal', scope: 'global', type: 'bool',
    compute: state => state.players[state.g.CurrentPlayer].CardsLeftToReveal > 0 },
],
end: { conditionMet: state => state.gameView('Cards').numComponents() === 0 },
```

- `setUp.begin` runs first, then every component of every deck is put where
  `distribute` says, then `setUp.finish`. Set-up draws from the version-0
  generator.
- `end.conditionMet` ends the game; the winners are the players with the
  highest score (the `gameScore` trait, or `end.playerScore`) who are neither
  inactive nor eliminated. `end.check` replaces that rule
  (tictactoe's hand port).
- A computed property is recomputed for each viewer from that viewer's
  sanitized state. In blackjack another player's `HandValue` counts their
  hole card as 0 until it is revealed.

## What the host does

The engine keeps nothing between calls. The host stores three things per
game and passes them back:

| Stored | Type | Notes |
|---|---|---|
| The game record | `GameRecord` | Id, **secret salt**, player count, variant, agents, whether players will be seated. Written once. The salt must never reach a client |
| The move history | `MoveRecord[]` | One per version, in order. Move progressions read it |
| The current state | `StateJSON` | The storage record, timers included. It is unsanitized: never release it |

The host also supplies, on every call, **`now`** in milliseconds since the
Unix epoch, and for a proposal **who is proposing**. The engine trusts both;
authenticating the proposer is the host's job.

```ts
import { compileGame, createGame, propose, viewerOutput, activeTimers } from './rules/index.ts';

const game = compileGame(memory());

// Create. versions[0] is version 0; any set-up fix-ups follow it.
const created = createGame(game, { id, secretSalt, numPlayers: 2, variant: { numcards: 'small' }, seating: false, now });
let record = created.record;                      // persist
let history = created.versions.flatMap(v => (v.move ? [v.move] : []));
let state = created.versions.at(-1).state;        // persist

// One input per invocation.
const result = propose(game, { record, history, state },
  { kind: 'propose', name: 'Reveal Card', input: { CardIndex: 3 }, proposer: 0, expectedVersion: history.length },
  { now });
if (result.error !== undefined) { /* refused: nothing changed; show result.error */ }
for (const version of result.versions) history.push(version.move);    // persist every version
if (result.versions.length > 0) state = result.versions.at(-1).state;
```

`propose` returns **every** version the input produced: the proposed move's
and each fix-up's, each with its stored state, its move record, and whether
the game was finished as of it. A client animates through them in order.
A refused input returns no versions and Go's error text. The engine never
modifies what it was handed. The one case with both versions and an error is
the fix-up cap, where Go has committed the versions before failing.

### Inputs

| Input | When |
|---|---|
| `{ kind: 'propose', name, input, proposer, expectedVersion? }` | A player (or the admin, `-2`) proposes a move, supplying only the fields a proposer supplies. With `expectedVersion`, it is refused if the state has moved on |
| `{ kind: 'seat', player }` | A person took a seat: `player`, or the next open seat if negative |
| `{ kind: 'timer', id? }` | A timer's deadline has passed |
| `{ kind: 'record', name, fields, proposer }` | A stored move is proposed again with all of its fields; for replays and tools |

### Seating

Whether players will ever be seated is decided at creation (`seating`) and
kept in the game record, not in the state. With it off, every seat is a
player from the start and the game begins at once. With it on, a game that
gathers (blackjack) waits, and each `seat` input makes the engine apply the
game's seating move and whatever fix-ups follow. A seat input that fills no
seat is refused.

### Timers

A timer is a deadline stored in the state. The engine has no clock and
schedules nothing:

1. When a move starts a timer, the stored state has an active timer record
   with `Deadline` set to the `now` of that call plus the duration.
2. After each call, the host asks `activeTimers(state)` for the active
   timers (earliest first, each with `deadlineMs`) and arranges to be woken.
3. When woken, the host sends `{ kind: 'timer', id }` with the current `now`.

A tick is validated against the state alone: the named timer must be active
with that id (a canceled, fired or restarted timer has none, so a late or
duplicate tick is refused as stale), and `now` must be at or past its
deadline. A valid tick applies the timer's move as the admin and then the
fix-ups. `force: true` skips the deadline check; it exists for tests.

### What a viewer is sent

```ts
const output = viewerOutput(game, { record, history, state, previousState, settled: true, now }, viewer);
```

`viewer` is a player index, `-1` for an observer, or `-2` for the admin.
**The admin's output is unsanitized; never release it.** The result has:

- `game`: `Name`, `ID`, `NumPlayers`, `Agents`, `Variant`, and, all as of
  this version, `Version`, `Finished`, `Winners`, `CurrentPlayerIndex`,
  `ActiveTimers` (timer id to milliseconds left at `now`) and `CurrentState`:
  the sanitized state with `Computed` and `Visibility`.
- `move`: the move record as this viewer may see it (`AnimationKey`,
  `Version`, visible `Properties`), or null.
- `legality`: for each move this viewer could be offered, whether it is legal
  for them now with default arguments, the reason if not, whether it is
  possible for anyone, the default arguments, and the precondition ledger.
- `projectedMoveChoices`: at a settled version, for a seated player, each
  declared choice set with which candidates are legal.

`history` and `state` are those of the version being viewed; `previousState`
is the state before it (move visibility groups are judged there); `settled`
is true for the last version an input produced.

Unlike Go's `JSONForPlayer`, nothing here comes from the head of the game: a
viewer replaying version 3 of a finished game is told it was not finished.

## The core, underneath

The rules layer is built on the core described in the rest of this file. A
game definition never needs most of it, except `Stack`.

### The state definition

`compileGame` turns a `GameSpec` into the core's `GameDefinition`
(`definition.ts`), which is JSON and is what Go discovers by reflection over
struct tags:

```ts
const def: GameDefinition = {
  Name: 'memory',
  Decks: { cards: { Size: 120, Components: [/* static values, by deck index */] } },
  Enums: {},
  GroupEnum: '',                 // the enum whose values are player groups, if any
  Game: { Props: {
    CurrentPlayer: { Type: 'playerIndex' },
    HiddenCards:  { Type: 'sizedStack', Deck: 'cards', Size: 40, Sanitize: 'order' },
    VisibleCards: { Type: 'sizedStack', Deck: 'cards', Size: 40 },
    Cards:        { Type: 'mergedStack', Deck: 'cards', Overlap: true, Stacks: ['VisibleCards', 'HiddenCards'] },
  } },
  Player: { Props: { WonCards: { Type: 'stack', Deck: 'cards' } } },
  DynamicComponentValues: {},    // deck name -> schema of per-component dynamic values
};
```

`Sanitize` takes the same string a Go `sanitize:"..."` struct tag takes, or
the parsed map (`{ other: 'len' }`). `testdata/gen/defs` derives this
definition from any Go game's live `GameManager`; `testdata/defs/` has the
results for the four ported games, and `rules/engine.test.ts` checks each
`GameSpec` against them.

### Using the core directly

```ts
import { loadEngineState, viewFor, OBSERVER_PLAYER_INDEX } from './index.ts';

const identity = { gameId, secretSalt };
const current = loadEngineState(def, identity, version, stateJSON);

const next = current.copy(version + 1);
next.gameStack('DrawStack').moveComponentToNextSlot(0, next.playerStack(0, 'Hand'));
next.gameStack('DiscardStack').shuffle();        // draws from next.rand()
if (next.conservationError() !== null) throw new Error('...');
const stored = next.toJSON();                    // what Go would persist

const forPlayer0 = viewFor(next, 0);
const forObservers = viewFor(next, OBSERVER_PLAYER_INDEX);
```

### State

- `loadEngineState(def, identity, version, json, options?)` copies the JSON,
  recomputes every id (recorded ids are ignored, as in Go), and throws if
  component conservation fails. `{ sanitized: true }` loads a per-viewer view
  instead: ids are kept as given and nothing can be modified.
- `state.copy(version?)` is Go's `state.copy(false)`. Passing a version is
  what `applyMove` does before applying a move.
- `state.rand()` is `state.Rand()`: seeded from game id + salt + version and
  memoized, so all draws while applying one move continue one sequence.
  **Anything that draws must draw in the order Go does.** Note that starting a
  timer draws 16 values for its id.
- `state.game`, `state.players[i]`, `state.dynamicValues(deck, index)` are the
  sub-states' JSON objects. Scalars, slices and enums (stored as their string
  values) are read and written there directly.
- `state.toJSON()` is the storage record: a copy, merged views recomputed,
  `SecretMoveCount` present only when non-empty.
- `state.conservationError()` returns Go's error text, or `null`.

### Stacks

`state.gameStack(prop)`, `state.playerStack(player, prop)`,
`state.dynamicStack(deck, index, prop)`, `state.gameBoardSpace(prop, space)`
(or `state.stack(ref)`) return a `Stack`: a handle over the JSON, not a copy.
`state.view(ref)` returns a read-only `StackView` for any stack property,
including merged ones.

Go's moves are methods on the component (`stack.ComponentAt(i).MoveTo(...)`).
Here they are methods on the stack that take the component's slot:

| Go | Engine |
|---|---|
| `c.MoveTo(dest, slot)` | `stack.moveComponent(i, dest, slot)` |
| `c.SecretMoveTo(dest, slot)` | `stack.secretMoveComponent(i, dest, slot)` |
| `c.MoveToFirstSlot/LastSlot/NextSlot(dest)` | `stack.moveComponentToFirstSlot/LastSlot/NextSlot(i, dest)` |
| `c.SlideToFirstSlot/LastSlot()` | `stack.slideToFirstSlot/LastSlot(i)` |
| `c.MayMoveTo(dest)` | `stack.mayMoveComponentTo(i, dest)` (error text or `null`) |
| `c.MayMoveToSlot(dest, slot)` | `stack.mayMoveComponentToSlot(i, dest, slot)` |
| `s.MoveAllTo(dest)`, `s.MoveCountTo(dest, n)` | `stack.moveAllTo(dest)`, `stack.moveCountTo(dest, n)` |
| `s.MayMoveAllTo`, `s.MayMoveCountTo` | `stack.mayMoveAllTo`, `stack.mayMoveCountTo` |
| `s.Shuffle()`, `s.PublicShuffle()` | `stack.shuffle()`, `stack.publicShuffle()` |
| `s.SwapComponents(i, j)` | `stack.swapComponents(i, j)` |
| `s.ComponentAtKey(k)`, `SwapComponentsByKey(i, j)`, `MaySwapComponentsByKey(i, j)` | `stack.componentAtKey(k)` (also on a merged view), `swapComponentsByKey`, `maySwapComponentsByKey` |
| `s.SortComponents(less)` | `stack.sortComponents(less)` (`less` gets deck indexes) |
| `s.ExpandSize/ContractSize/SetSize/SizeToFit` | `stack.expandSize/contractSize/setSize/sizeToFit` |

A failed operation throws a `StackError` (test with `isStackError`) with Go's
message and changes nothing.

A stack property's `Constraints` (the constraint strings of its struct tag,
`['same(Color)']`) are parsed when a state is built and checked wherever Go
checks them: in `MayMoveTo`, so by every move into the stack and every
`May...` form, never by swaps, slides, shuffles, sorts or set-up
distribution. A transfer of several components into a constrained stack is
simulated on a copy of the state, as Go does, so each check sees the earlier
insertions. A malformed constraint makes loading fail with Go's boot error.

A component is a `(deck name, deck index)` pair. `stack.componentAt(i)`
returns the deck index, `GENERIC_COMPONENT` (-2) in a sanitized stack, or
`null` where Go returns nil. Static values are `def.Decks[deck].Components[i]`.

### Views

- `viewFor(state, viewer, options?)` is the JSON a viewer receives as the
  current state (Go's `Game.JSONForPlayer(...).CurrentState`), with
  `Visibility`. `viewer` is a player index, `OBSERVER_PLAYER_INDEX`,
  `ANY_PLAYER_INDEX` (treated as the observer) or `ADMIN_PLAYER_INDEX`.
  **The admin view is the unsanitized state, `SecretMoveCount` included; never
  release it.**
- `sanitizedForPlayer(state, viewer)` returns the sanitized `EngineState`.
- `Computed` is produced by the `options.computed` callback, which is handed
  the sanitized state. `rules/view.ts` supplies it; the core alone omits it.
- `options.mergedStacksRevealDynamicValues` selects between the two answers Go
  can give in one corner; see the status document.

## Tests

`npm run test:unit` (Node 23 or newer, as for the rest of the suite) runs
everything. Each `*.test.ts` says at its top what it proves.

| Test | Proves |
|---|---|
| `games/conformance.test.ts` | Every document of `conformance/` (memory, blackjack, tictactoe, pig, debuganimations, werewolf), replayed on the generated game from its recorded id, salt, players, variant, seating and outside inputs alone: at every version the stored state, the move record, and every viewer's game object, sanitized move, legality and projected choices equal what Go recorded |
| `games/<game>/<game>.gen.test.ts`, `games/fuzz.test.ts` | Each generated game against its conformance documents and the checked-in fuzz corpus (`conformance/fuzz/`), every recorded rejection proposed again; with `BOARDGAME_FUZZ_DIR`, a transient `boardgame-util fuzz run` too |
| `rules/data.test.ts`, `rules/stockRegistry.test.ts` | The loader's contracts and its two hooks, and the stock registry, on the generated games' definitions and bodies |
| `rules/rejections.test.ts` | Go's error text and an untouched state for 47 refused proposals, and Go's outcome at the fix-up cap (2 cases) |
| `rules/binding.test.ts` | The legal corpus and every Go predicate verdict, through the real state |
| `rules/engine.test.ts` | Timers against a moving clock, seat inputs, purity of a step, and each game's state definition against the one Go derives |
| `rules/purity.test.ts`, `purity.test.ts` | The sandbox constraints |
| `sandbox.test.ts` | Nothing at module scope that a Common Fabric pattern's sandbox refuses: no class, no `let`, no record of functions, no double cast, no import that leaves the engine |
| `legal/message.test.ts` | The engine's legal-message renderer gives the same text as the client's |
| the core's and `legal/`'s tests | See the two status reports |

The expected values are never hand-derived. They come from Go:

```sh
server/static/src/engine/testdata/gen/regen.sh                  # the engine's vector files
./scripts/go-local run ./boardgame-util conformance export      # conformance/
```

The generators live under `testdata/gen/`; a directory named `testdata` is
skipped by `go build ./...` and `go test ./...`, so they are not part of the
repository's Go build. `testdata/gen/vecgame` is a game definition written
only to give Go one of every kind of property to serialize.

`testing/conformance.ts` is the conformance runner; its `passTable` prints
the per-document, per-field table, which `games/conformance.test.ts` emits as
a test diagnostic.
