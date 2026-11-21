# engine/legal

The TypeScript evaluator for boardgame's declarative move legality. It
reproduces the Go implementation (`legal/`, and the `legal_*.go` files at the
repository root), which stays the reference. Given the same state, move and
proposer, it returns the same verdict Go does: the same outcome, the same
template key and bindings, and the same text.

Pure TypeScript with no dependencies, no DOM and no Node APIs outside the
`*.test.ts` files. Import from `index.ts`.

## What it answers

| Question | Function |
|---|---|
| Is this move legal, and if not, why? | `evaluateLegalPlan`, `legalPlanError` |
| What does each precondition say? | `evaluateLegalPlan(plan, ctx, true)` |
| What does one predicate say? | `resolveLegalSpecs`, then `evalLegalPredicate` |
| What text does a failure render as? | `renderLegalVerdict`, `renderLegalMessage` |
| Could this viewer have computed that verdict? | `legalReadEvaluable`, `legalEntryEvaluable`, `buildPreconditionLedger` |

## Using it

A move's legality is a plan: an ordered list of predicates built once per move
type, then evaluated for each question.

```ts
import {
  buildLegalPlan, contributedPreconditions, defaultLegalRegistry, evaluateLegalPlan,
  legalErrorf, mergeLegalTemplates, propAtLeast, revealableCardAt, mayMoveToSameSlot,
} from './index.ts';

const templates = mergeLegalTemplates({ 'reveal.no_cards_left': 'You have no cards left to reveal this turn' });

const plan = buildLegalPlan({
  moveName: 'Reveal Card',
  // The base checks the move type contributes. They run first.
  contributed: contributedPreconditions({ moveName: 'Reveal Card', currentPlayer: true }),
  // The author's specs. They run next, in this order.
  authored: [
    { ...propAtLeast('player.CardsLeftToReveal', 1), message: 'reveal.no_cards_left' },
    revealableCardAt('game.HiddenCards', 'game.VisibleCards', 'move.CardIndex'),
    mayMoveToSameSlot('game.HiddenCards', 'game.VisibleCards', 'move.CardIndex'),
  ],
  // Optional imperative residue. It runs last, only if everything above passed.
  custom: ctx => (somethingTheCatalogCannotSay(ctx) ? legalErrorf('reveal.custom', { n: 2 }) : null),
  registry: defaultLegalRegistry(),
  exampleState, moveReader, templates,   // optional; they turn typos into errors here
});

const { verdict } = evaluateLegalPlan(plan, { state, move, proposer, framework });
// verdict.outcome is 'pass', 'fail' or 'unknown'.
```

`buildLegalPlan` throws `LegalSpecError` for anything Go's `NewGameManager`
rejects: an unknown predicate, bad arguments, a nested `any`, a path that does
not exist or has the wrong type, a suppression that matches nothing, a
template key missing from the table, a template body naming a binding the
predicate does not supply. The message is Go's.

### Order

1. Contributed specs, in the order given. `contributedPreconditions` produces
   Go's order: `inPhase`, `inProgression`, `stackConstraints`,
   `proposerIsCurrentPlayer`, each only if configured. Names listed in
   `suppressions` are dropped.
2. Authored specs, in declaration order. They are never suppressed.
3. The `custom` function.

Evaluation stops at the first verdict that is not a pass. An unknown stops it
just as a fail does: it means legality could not be confirmed. With
`fullLedger` true, every predicate is evaluated and reported, and the overall
verdict is still the first non-pass.

Nothing is reordered by cost, and there is no cache. Go memoizes
field-independent verdicts per state version; that is a speed-up with no
effect on the answer for a real state.

### Verdicts and text

```ts
interface LegalVerdict {
  outcome: 'pass' | 'fail' | 'unknown';
  message?: { template: string; bindings?: Record<string, string | number | boolean> };
  reason?: string;   // why it is unknown
}
```

A failure carries a template key and bindings, not a string.
`renderLegalVerdict(verdict, table)` gives the text `move.Legal()` returns in
Go. `table` has the shape the chest ships as `LegalTemplates`;
`defaultLegalTemplates()` is the built-in English and `mergeLegalTemplates`
overlays a game's own entries.

`message.ts` holds the one TypeScript mirror of Go's `RenderLegalMessage`.
The client used to carry a copy at `src/legal/renderLegalMessage.ts`; it had
no importer in the client and could not load in a pattern, so it was removed
and its tests are `render-message.test.ts` here.

### The base checks whose logic lives elsewhere

Phases and move progressions belong to a later engine task. Two predicates
take them as an input, through `LegalContext.framework`:

```ts
interface LegalFramework {
  // The current phase: its key, its display name, and its self-inclusive
  // ancestor keys when the phase enum is a tree (null otherwise).
  currentPhase(state: LegalState): LegalPhase | null;
  // The move-tape match for the named move: null if it is next in the
  // progression, a string (the error text) if not, { unknown } if undecidable.
  moveInProgression?(moveName: string, state: LegalState, proposer: number): LegalHostCheck;
}
```

`inPhase` does its own comparison against `currentPhase` and produces Go's
`Move is not legal in phase X` text. `inProgression` is entirely the host's
answer. With no framework supplied, both are unknown, never pass.
`stackConstraints` and `proposerIsCurrentPlayer` are implemented in full.

### Client evaluability

A predicate declares each property it reads and the facet it needs: `values`,
`count`, `occupancy`, `order` or `nonempty`. A viewer's state JSON carries a
`Visibility` block listing, per property, the facets that viewer can trust.

```ts
const view = { viewer, visibility: state.Visibility, currentPlayerIndex };
legalReadEvaluable(view, { path: 'game.DrawStack', facet: 'count' });
buildPreconditionLedger(view, entries);   // the wire ledger, for that viewer
```

`facetSurvives(policy, facet)` is Go's table. `legalEntryEvaluable` is the
server's rule: the generic client catalog knows the predicate, it has a wire
form, and every read survives. `buildPreconditionLedger` withholds bindings
from an entry the viewer could not have computed, leaving only its template
key.

## The state-access interface

The evaluator never sees a concrete state, stack or enum. It reads through the
interfaces in `access.ts`, and does no stack, deck or enum arithmetic of its
own. Each member names the Go method it stands for.

| Interface | Stands for | Members |
|---|---|---|
| `LegalState` | `ImmutableState` plus three delegate answers | `game()`, `players()`, `currentPlayerIndex()`, `playerMayBeActive(i)`, `customPlayerOrder()`, `playerIsInactive(i)` |
| `LegalMove` | `Move` | `reader()` |
| `LegalPropertyReader` | `PropertyReader` | `props()`, `prop(name)` |
| `LegalStack` | `ImmutableStack` | `len()`, `numComponents()`, `componentAt(i)`, `first()`, `mayMoveAllTo(dst)`, `mayMoveCountTo(dst, n)`, `maySwapComponents(i, j)` |
| `LegalComponent` | `ImmutableComponentInstance` | `values()`, `mayMoveTo(dst)`, `mayMoveToSlot(dst, slot)` |
| `LegalBoard` | `ImmutableBoard` | `enum()`, `spaceAt(i)`, `spaceAtKey(key)` |
| `LegalEnum`, `LegalEnumValue` | `enum.Enum`, `enum.ImmutableVal` | `name`, `valueFromString(s)`, `ancestors?(key)`; `enum`, `value`, `string` |
| `LegalChest` | `ComponentChest` | `enums()` |

`props()` maps each property name to its `PropertyType`, spelled as Go's
`String()` form (`'TypeInt'`, `'TypeStack'`, ...). `prop(name)` returns a
number for `TypeInt` and `TypePlayerIndex`, a boolean, a string, a
`LegalEnumValue` or null, a `LegalStack` or null, or a `LegalBoard` or null.

Path resolution over these is in `path.ts`:

| Path | Resolves against |
|---|---|
| `game.X` | `state.game()` |
| `player.X` | `state.players()[state.currentPlayerIndex()]` |
| `proposer.X` | `state.players()[ctx.proposer]` |
| `move.X` | `ctx.move.reader()` |
| `players[move.F].X` | the player a `TypePlayerIndex` move field names |
| `players[*].X` | nothing; only `allActivePlayers` iterates it |
| `game.B[move.F]` | the stack at space `F` of board `B` (also on the player kinds) |

## Binding it

Implement the interfaces over the real state types. What the contract
requires:

- **Same answers as Go.** A `may*` method returns null when Go returns a nil
  error and otherwise Go's `err.Error()` text, which becomes the `detail`
  binding and so the text a player sees. Stack constraints, slot capacity and
  deck checks all live behind these methods.
- **`componentAt` never throws.** An empty slot and an out-of-range index both
  return null. `ImmutableComponentAtKey(key)` is `componentAt(key)`.
- **Enum identity is the name.** Two values are of the same enum when
  `enum.name` matches. `valueFromString` must follow Go's
  `Enum.ValueFromString`: trim and lower-case both sides, then accept the
  decimal key itself.
- **`prop` may throw** to report a failed read. The message becomes part of an
  unknown verdict's reason.
- **`playerIsInactive`** is `behaviors.PlayerIsInactive`, which
  `allActivePlayers` uses to skip players. **`playerMayBeActive`** is the
  delegate's `PlayerMayBeActive`, which the proposer checks use. They are
  different questions.
- **Readers are views.** Nothing here mutates or retains what it is handed.

`testing/fixtures.ts` is a complete binding, written for tests: a reader over
a plain object, an enum over a key table, and a stack that models growable,
sized and merged stacks without constraints. `bindWireFixture` builds all of
it from a `state.StorageRecord()`, the chest JSON and the type tables a state
record cannot carry. Use it as the worked example. It is not the engine's
stack implementation, and it throws on a constrained destination rather than
guess.

## Files

| File | Go counterpart | Holds |
|---|---|---|
| `types.ts` | `legal_types.go` | Verdict, spec, facet, read; the special player indexes |
| `access.ts` | none | The state-access interface |
| `path.ts` | `legal_path.go`, `legal/path_resolve.go` | Grammar, resolution, boot validation of paths |
| `playerIndex.ts` | `state.go` (`PlayerIndex`) | `Valid`, `Next`, `EnsureValid`, `Equivalent` |
| `predicate.ts` | `legal_predicate.go` | Predicate type, `evalLegalPredicate`, `any`, `resolveLegalSpecs` |
| `catalogCompare.ts` | `legal/catalog_compare.go` | `propAtLeast`, `propCompare`, `playerBool`, `propEquals`, `propNotEquals` |
| `catalogStack.ts` | `legal/catalog_stack.go`, `catalog_count.go` | Occupancy, the `mayMove*` and `maySwap*` family, counts |
| `catalogPlayers.ts` | `legal/catalog_players.go`, `catalog_behaviors.go` | `allActivePlayers`, proposer checks, `playerBoolAt` and selectors |
| `catalogPurpose.ts` | `legal/catalog_purpose.go`, `catalog_component_scalar.go` | `revealableCardAt`, component value comparisons |
| `catalogFramework.ts` | `legal/catalog_framework.go`, `legal_framework.go`, `moves/catalog_framework.go` | `inPhase`, `inProgression`, `stackConstraints` |
| `catalog.ts` | `legal.DefaultConstructors()` | The default registry |
| `templates.ts` | `legal/templates.go` | Template keys and default text |
| `message.ts` | `legal_error.go` | Rendering a verdict |
| `plan.ts` | `legal_plan.go` | Assembly and ordered evaluation |
| `evaluable.ts` | `legal_evaluable.go`, `state_visibility.go`, `server/api` | Facet survival and the per-viewer ledger |
| `goformat.ts` | `strconv`, `fmt` | `Atoi`, `%q`, `%v` of a string slice |

Each catalog file also exports the spec builders Go has (`propAtLeast(...)`,
`any(...)`, `playerHasSubmitted(proposer())`, ...), which return plain
`LegalSpec` objects in Go's JSON shape.

## Tests

`npm run test:unit` runs them with the rest of the client's unit tests.

| File | Checks |
|---|---|
| `conformance.test.ts` | Every file in `legal/testdata/conformance/`, read in place |
| `vectors.test.ts` | About 3,950 Go verdicts in full, predicate metadata, constructor errors |
| `plan.test.ts` | About 1,850 Go plan evaluations, 30 Go boot errors, and hand-written ordering tests |
| `evaluable.test.ts` | Go's facet table and `LegalReadEvaluable` for every path kind and viewer |
| `path.test.ts` | Indexed board paths and boot validation, on hand-written fixtures |
| `predicate.test.ts` | Fail-closed evaluation, `any`, player indexes, Go formatting, rendering |

### Test vectors

`testdata/*.json` is generated from the real Go implementation by the program
in `testdata/gen`. From the repository root:

```sh
./scripts/go-local run ./server/static/src/engine/legal/testdata/gen
```

It imports the repository's packages and changes none of them. Games are
created with a fixed id and salt, so the output is byte-stable: after
regenerating, `git diff` on `testdata/` shows exactly what changed in Go's
behavior.

| File | From |
|---|---|
| `fixtures.json` | The corpus's named fixtures, restated from `legal/conformance_test.go` and checked against every corpus case before anything is written; plus the extra arrangements the vectors use |
| `vectors.json` | Each catalog constructor's `Evaluate`, called directly, as the corpus harness does |
| `plans.json` | `GameManager.LegalEvaluateLedger`, `LegalEvaluatePlan`, `LegalRenderVerdict` and `move.Legal`, on the example games' moves and on synthetic moves added to a wrapped example game |
| `evaluable.json` | `boardgame.LegalReadEvaluable` and each viewer's `Visibility` block, with one property's sanitization policy overridden at a time |

The corpus names its fixtures but does not contain them; they exist only as Go
test code. `fixtures.json` is what lets the corpus run here.

## Not ported

- **The verdict memo** (`legal_memo.go`) and the phase index
  (`legal_index.go`). Performance only.
- **Component-field boot validation** (`RequiredComponentFields`). It walks
  every component of a stack's deck, which the access interface does not
  expose. The metadata is carried on the predicate; nothing checks it yet.
- **The reachability probe and the base-type allowlist.** They guard Go's
  embedded `Legal()` overrides, which have no equivalent here.

Known differences from Go, and questions for whoever integrates this, are in
`fabric/status/engine-legal.md`.
