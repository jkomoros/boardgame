# Client renderer authoring baseline

**Status:** implementation evidence for Task 1; the facade described here is
experimental, not yet a compatibility promise.

## What game creators receive today

Generated state is an expanded client projection, not a typed copy of the Go
struct. `FullGameState` has `Game`, `Players`, optional computed values, and an
optional component catalogue. Expanded stacks carry deck/name metadata,
indexes, stable IDs, shuffle metadata, optional size bounds, and a
`Components` array. A component entry may be `null` or a partial object because
hidden components are serialized as `{}`. Timers expose `ID`, `IsTimer`,
`TimeLeft`, and `originalTimeLeft`.

Representative generated game projections:

| Game | Game state that drives rendering | Player state that drives rendering | Main authoring pressure |
| --- | --- | --- | --- |
| Pig | current player, die stack with static faces and dynamic selected face/value, target score | score, round score, done/die-counted and seat flags | zero-input actions on a die and button |
| Tic-tac-toe | phase and nine-slot token stack | token value, unused tokens, tokens remaining | per-cell target legality and keyboard grid |
| Memory | hidden/visible/combined card stacks, hide timer, current player | won cards and cards left to reveal | actionable cards, delayed result visibility, timer UI |
| Blackjack | phase, draw/discard/unused stacks, round-robin and round counters | visible/hidden/combined hands, score, stood/busted flags | card zones plus distinct solo/table/hand surfaces |
| Werewolf | day/night/gathering phase and round | private role, vote, eliminated and seat flags | privacy-safe simultaneous choices and readiness |
| Murder Mr. Monroe | map/wing/light flags, current actors, card stacks, token locations and paths | room/location/path, hand, movement and draw state | typed artwork geometry and explicit piece projection |

These declarations are useful evidence but are not yet fully honest contracts:
`Component` is intentionally partial, several current renderers cast through
`any`, and board stacks have a separate raw-versus-expanded distinction.

## Import and interaction inventory

Every renderer currently reaches through framework internals. Common imports
are `BoardgameBaseGameRenderer`, Lit's `html`/`css`, component registration
modules, generated move names/state, and Material controls. Advanced renderers
also import preview-legality types, Lit decorators/directives, companion base
classes, or the spatial board.

Current action paths are stringly `propose-move` attributes, `data-arg-*`
attributes, `componentAttrs` forwarding, and occasional hand-authored
`CustomEvent("propose-move")`. Legality is usually duplicated through
`isMoveCurrentlyLegal`; Checkers and Tic-tac-toe additionally implement
`previewSpec`. Deck/card views still rely on global custom-element/template
registration. These are legacy mechanisms to migrate, not facade candidates.

Classification for the first facade spike:

| Classification | Symbols/mechanisms |
| --- | --- |
| Candidate facade | `BoardgameBaseGameRenderer`, `html`, `css`, declarative registration of a small proven primitive set |
| Experimental | the facade module itself and primitive registration through it |
| Legacy migration surface | `propose-move`, `data-arg-*`, `componentAttrs`, deck moustache templates, direct custom proposal events |
| Internal | reducers/store, state manager, render-game gate, admin controls, selectors, raw API helpers, concrete primitive classes and lifecycle methods |

No Polymer/moustache machinery, Redux implementation, admin API, or concrete
primitive class is exported merely for convenience.

## Facade spelling decision

The initial spelling is:

```ts
import { BoardgameBaseGameRenderer, html, css } from '../../src/client.js';
```

This is not aesthetically final, but it has one valuable property: from the
assembled `static/game-src/<game>` topology, it resolves identically in Vite
development and production without package-manager aliases or a game-specific
path. Generated `_types.ts` already relies on the same topology.

`@boardgame/client` is deferred. It would currently require coordinated Vite,
TypeScript, test-runner, and external-package aliasing, while Node's native
resolver would still need a package/export map. Adopting the prettier spelling
before those environments agree would create a facade that works only in some
tools—the exact foot gun this task is meant to remove.

## Evidence required before stabilization

- Framework TypeScript type-check and Vite production build.
- Pig loading through the facade in the ordinary source/dev server.
- A Go topology test proving `../../src/client.js` from assembled
  `game-src/<game>` reaches the copied/symlinked framework source.
- The self-started renderer Playwright shard, which exercises
  `boardgame-util serve`, its assembled tree, Vite development, and same-origin
  API proxy.
- A production `boardgame-util build static --prod` smoke check.
- One external game assembled against the same facade before the module is
  described as stable.

The next tasks may add typed action and composition exports, but only after
their own fixtures prove those contracts. Deep imports remain available for
legacy renderers during migration; they are not endorsed as public API.

## Task 1 verification record

The spike passed all of the required resolution environments:

- `go test ./boardgame-util/lib/path ./boardgame-util/lib/build/static`
  proves relative resolution from an absolute assembled temp directory.
- `npm run type-check` proves the source renderer and facade declarations.
- `npm run build` proves the framework production build, and
  `npm run test:facade-production` assembles framework source, Pig, and
  dependencies under a temporary production root and makes Vite statically
  resolve and bundle Pig through the facade (rather than relying on the app's
  intentionally runtime-dynamic, `@vite-ignore` loader).
- `npm run test:renderer` starts `boardgame-util serve` on allocated ports,
  creates a real Pig game, and asserts the Pig renderer and facade-registered
  die both mount without failed renderer/facade requests.
- `boardgame-util build static --prod <absolute-temp-dir>` type-checks the
  assembled game sources and completes the Vite production bundle.
- A copied Murder Mr. Monroe external renderer, changed only to import the
  facade (plus one isolated pre-existing `unknown`-to-boolean correction),
  passes a standalone strict TypeScript compile in that assembled production
  tree. The source external repo was not modified in this framework branch.

The production proof exposed and fixed two `boardgame-util` infrastructure
defects: hand-rolled relative paths broke across canonicalized roots such as
macOS `/var` → `/private/var`, and symlinked production `index.html` files made
Vite emit an illegal outside-root asset name. Assembly now canonicalizes
existing roots, uses `filepath.Rel`, supports absolute destinations, copies
production top-level files automatically, and excludes framework unit tests
from the assembled game-source type-check.
