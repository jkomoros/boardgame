# Expanded Client Contract

## Four distinct layers

Renderer code now names four different representations instead of allowing
them to blur together:

1. The sanitized server wire payload is the JSON received for one viewer.
2. The expanded renderer snapshot resolves stack indexes through the chest and
   adds component metadata, dynamic values, and timer projections.
3. Generated `MoveInputs` are native creator-owned proposal values.
4. Form-encoded move strings are internal transport and never a creator type.

TypeScript describes the shape of the sanitized projection. It cannot prove
that the projection contains unsanitized truth. Privacy remains a server
sanitization invariant and must be tested with per-viewer matrices.

## Components and stacks

A stack position has three intentionally separate states:

- `VisibleComponent<Values, DynamicValues>` has required `Index`, `Values`,
  `Deck`, `GameName`, and `ID`; `DynamicValues` is optional because the deck may
  not define it or sanitization may omit it.
- `OpaqueComponent` is the exact empty object `{}` used for an occupied but
  hidden component.
- `null` is an empty fixed-stack position.

`isVisibleComponent()` is the supported narrowing guard. Checking truthiness
only distinguishes `null`; it deliberately does not reveal whether an occupied
slot is visible. Expanded stacks and boards contain readonly arrays of those
entries. Dynamic-value stacks nested inside dynamic component state remain raw
server stacks; top-level game/player boards are expanded by the selector and
are typed as `ExpandedBoard`.

Generated deck catalogs preserve decks with static values, dynamic-only values,
and no custom values. Each catalog entry is a readonly array of
`CatalogComponent<Values>`: exactly the chest's static `{ Index, Values }`
shape. Expanded stack components separately add `Deck`, `GameName`, `ID`, and
optional dynamic values. A no-static-values deck uses an exact empty `Values`
record rather than pretending arbitrary properties exist.

The wire state's top-level `Components` field is different: it contains only
per-index dynamic values for decks that define them. Generation therefore emits
a separate `DynamicComponentValues` map and binds it as the fifth
`FullGameState` parameter. Confusing these two shapes is a type error.

## Readonly snapshots and computed values

The renderer snapshot is deeply readonly. Generated state fields and slices are
readonly, as are expanded stack metadata and component values. A new snapshot
arrives when the host installs a version; creators do not mutate it locally.

Computed properties are declared alongside their evaluators through typed Go
constructors. The generator reads those declarations—not an observed example
snapshot—and emits closed, game-specific `GameComputed` and `PlayerComputed`
interfaces. There is no unknown index signature or declaration-merging step:
undeclared keys and misspelled accesses fail TypeScript compilation. Framework
keys are included automatically, while game keys are always present with the
declared primitive, slice, player-index, or enum-literal type.

Exact proposal typing rejects missing and extra fields even through aliased and
spread objects. TypeScript's `number` does not distinguish integers from
fractions, so finite/integer validation remains a mandatory runtime check at
the proposal boundary rather than a misleading compile-time claim.

## Bound renderer

`boardgame-util emit-types` refreshes the complete dependency set and emits
`_game_renderer.ts`. Its `GameClientContract` type slots bind the exact state
(including computed and dynamic-component values), resolved component catalog,
move names, and creator inputs. Its abstract `GameRenderer` consumes those
slots and installs the runtime schema and fingerprint. The generated module
also owns the exact custom-element tag, so a creator writes only:

```ts
import { GameRenderer, registerGameRenderer } from './_game_renderer.js';

@registerGameRenderer
export class BoardgameRenderGameExample extends GameRenderer {
  // render and interaction code
}
```

The tag cannot be mistyped or copied from another game. Registered renderer
classes are exported so the strict `noUnusedLocals` gate and direct renderer
tests agree that the class is part of the package API.

The local generated module imports the single framework facade, so source,
assembled development, and production builds share the same Lit and base-class
identity.

Renderer base classes have no widening generic defaults, and proposal dispatch
has no implicit stringify-only compatibility path. A renderer without generated
schema metadata fails loudly and directs the creator to `GameRenderer`; stale or
missing server fingerprints fail before an event is dispatched.
