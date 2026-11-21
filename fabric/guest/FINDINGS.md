# Findings: what the client assumed about the webapp

Written while building the guest in this directory against
`server/static/src` at commit `34bfd655`, without editing it. File and line
references are to that commit. "Bundled" means the file is in the guest's
import closure (132 modules; the list is in `dist/bundle-report.json` after a
build).

## What has changed since: where each finding stands

Everything below this section is the original record, written against
`34bfd655`. Stream A has since made most of the changes it asked for, and the
guest was rebuilt on the result (`fabric/status/client-packages.md`). File and
line references below are to the old tree: the client now lives in
`server/static/packages/elements` and `server/static/packages/renderer-kit`.

| # | Then | Now |
|---|---|---|
| F1 | Prototype patch over the path-computed `import()` | **Gone.** `RendererHost.loadRendererModule`; the guest's resolves for a bundled game. No `import(` is left in the bundle |
| F2 | `api.ts` redirected to a refusing shim; `gameId` left empty | **Gone.** The kit has no HTTP import; board preview is `RendererHost.boardPreview`, which the guest does not supply. `gameId` is passed through |
| F3 | `companionInfo` set to steer around cookie and storage reads | **Gone.** `RendererHost.resolveSurface`; the guest's returns the solo surface |
| F4 | The same redirect, for the Table base's server calls; `API_HOST` declared | **Gone.** `RendererHost.tableActions`, absent in the guest; the declaration is deleted |
| F5 | `socketActive = true` | **Still a shim**, one assignment. The property is the kit's word for "live" |
| F6 | Fake move forms to carry legality | **Gone.** `boardgame-render-game.moveLegality` takes legality by move name |
| F7 | Player info mounted by hand with `rendererLoaded = true` | **Gone.** `mountPlayerInfo()` with the host |
| F8 | `expandGameStateSnapshot` imported from the Redux selector module, dragging in `reselect` | **Gone.** `@boardgame/elements/state/expand.js`; `reselect` is no longer a dependency |
| F9 | No art or fonts cross the protocol | Unchanged. Token art now has a host resolver (`RendererHost.resolveTokenArt`); the guest supplies none, so a token with no art of its own draws its generated solid or nothing |
| F10, F11 | Theme tokens copied from the webapp's stylesheet by path | The stylesheet is in the kit: `@boardgame/renderer-kit/theme/md3-theme.css`. The list of custom properties a host must set is still not written down |
| F12 | `Sets` patched back on to projected choices | **Gone.** The kit reads an absent `Sets` as empty |
| F14 | Side-effect registration of every element | Unchanged; seam S6 |
| F17, F18 | Initial entrance animation; legacy overlap hook | Unchanged |

The animation counters (`window.__bgAnimTestHooks`) are no longer published by
importing the client. Both guest entry points call `installAnimTestHooks()`,
because the guest's tests and the pattern's two-user script read them.

## Result

`boardgame-render-game` mounts unmodified in a sandboxed, opaque-origin
iframe and plays Memory with Memory's unmodified renderer, the real animator,
effect layer and animation gate. No fallback to the fixture path was needed.

It took one run-time patch (F1), one build-time module redirect (F2, F4),
four property values chosen to steer around webapp code (F2, F3, F5, F6), and
about 630 lines of host (`src/guest-host.ts`) replacing the state manager and
the game view.

## For stream A: the changes that would remove each shim

Ordered by how much guest code each removes.

| # | Change to `server/static/src` | Removes |
|---|---|---|
| F1 | `boardgame-render-game` takes a renderer loader (or an already-registered tag) as a property | The prototype patch |
| F2 | `boardgame-render-game` takes the batch legality preview as a transport, like its other three | The `api.ts` redirect for render-game; the empty `gameId` |
| F4 | `BoardgameTableViewBase` takes table actions as a transport and stops importing `api.ts` and `util.ts` | The rest of the `api.ts` redirect; the `API_HOST` declaration |
| F3 | `boardgame-render-game` takes the surface as a property instead of asking `companion-surface` | The `companionInfo` value; one throwing `document.cookie` read |
| F8 | `expandGameStateSnapshot` moves out of `selectors.ts` | The `reselect` dependency and 6.6 KB |
| F12 | The per-entry wire decoders in `types/server-response.ts` are exported | The `Sets` patch-up, and the same bug in the scenario fixture |
| F6 | `boardgame-render-game` takes legality as `{ [move]: legality }` | The fake move forms |
| F11, F10 | A documented list of the custom properties a host must set, with defaults inside the kit | Guesswork; the `--animation-length` rule |
| F5 | `socketActive` renamed or defaulted to "live" | One assignment |
| F7 | A host-owned mount for player info | Nothing today; the guest already mounts it directly |
| F14 | Layer 2 has no side-effect registration of unused elements | About 235 KB of the 590 KB bundle |

## Findings

### F1. The renderer is loaded by a path-computed dynamic `import()`

- **Where.** `components/boardgame-render-game.ts:1060` builds
  `../../game-src/${gameName}/boardgame-render-game-${gameName}${suffix}.ts`;
  `:1108-1110` imports it with `/* @vite-ignore */`.
  `components/boardgame-player-roster.ts:340` does the same for player info.
- **In a guest.** There is nothing to fetch. The call would be a network
  request against the document's base URL, and it survives into the bundle
  as the only `import(` in it.
- **Shim.** `src/guest-host.ts` replaces
  `BoardgameRenderGame.prototype._loadRendererModule` with a function that
  resolves at once. The method is TypeScript-private, so this is a cast.
  Renderers are registered by static imports in `src/games.ts`.
- **Fix.** A `rendererLoader` property (S2), defaulting to today's import.
  With it the dynamic `import()` also leaves the guest bundle.

### F2. `boardgame-render-game` calls the HTTP API for board legality preview

- **Where.** `boardgame-render-game.ts:10` imports `movePreviewBatch` from
  `../api.js`; `:969` calls it from `_refreshPreview`. The request is skipped
  when `gameId` is empty (`:943`) or the renderer's `previewSpec()` returns
  nothing (`:945-955`). Memory's returns nothing.
- **Shim.** Two layers. The guest leaves `gameId` empty, so the call cannot
  be reached. And the Vite plugin `guest-redirect-server-api` redirects every
  import that resolves to `server/static/src/api.ts` to `src/shims/api.ts`,
  whose functions refuse and post an `error` to the host. No test saw one.
- **Cost of the empty id.** `gameId` also seeds effects
  (`:1144`) and keys transitions (`:824`). With it empty, effect seeds
  depend on version only. Harmless for one game per guest.
- **Fix.** Inject it. `targetPreviewTransport` already exists as a property
  (`:209-210`) and carries the same question; `_refreshPreview` could use it.

### F3. Choosing the surface reads location, storage and cookies

- **Where.** `boardgame-render-game.ts:16` and `:1122-1127` call
  `surfaceForGame`. `utils/companion-surface.ts:26` reads
  `window.location.search`; `:31` reads `window.sessionStorage` inside a
  `try`; `:39` reads `document.cookie` **outside** it.
- **In a guest.** In an opaque origin both storage and `document.cookie`
  throw `SecurityError`. The storage read is caught. The cookie read is not,
  so with a non-empty `gameId` the renderer would never load.
- **Shim.** The guest sets `companionInfo = { CompanionMode: false }`, which
  makes `surfaceForGame` return on its first line (`:25`). The empty `gameId`
  from F2 would also skip the storage and cookie block (`:29`).
- **Fix.** The host passes the surface. Independently, `:39` belongs inside
  the `try`: the webapp itself breaks today if cookies are blocked.

### F4. The renderer kit imports the server and an `index.html` global

- **Where.** `components/boardgame-table-view-base.ts:7` imports
  `apiHttpPost` and `buildGameUrl` from `../api.js`, and `:9` imports
  `apiPath` from `../util.js`. `util.ts:17` reads the global `API_HOST`,
  which `index.html:60` defines in an inline script. `:287` and `:294` read
  `window.location.origin` and `.host` to build a join link.
- **Why it is in every guest.** `client.ts:15-16` re-exports the Table and
  Hand bases, every generated `_game_renderer.ts` imports them, and
  `boardgame-render-game.ts:37-38` imports them for `instanceof` checks
  (`:576-582`, `:624`). So a solo-only game still bundles both (26 KB before
  minification) and their server imports.
- **Shim.** The same `api.ts` redirect. `API_HOST` is declared in
  `src/vite-env.d.ts` only so the imported source type-checks; it is never
  defined, and a Table renderer would throw `ReferenceError` on first render.
  This guest mounts only the solo surface.
- **Fix.** S4: table actions as a transport; the join link as a host input.

### F5. "Socket active" is the only way to say "live"

- **Where.** `boardgame-render-game.ts:276-277`. When false the move
  transport is withheld (`:603`, `:1190`) and a reconnecting banner renders
  (`:1283-1295`). `:289` and `:425-427` also read `navigator.onLine` and
  listen for `online`/`offline` on `window`.
- **Shim.** The guest sets `socketActive = true` and never changes it.
- **Fix.** Rename to a host-neutral `live`, and let the host supply the
  banner or suppress it. `navigator.onLine` is meaningless in a guest.

### F6. Legality arrives only inside HTTP move forms

- **Where.** `boardgame-render-game.ts:292-293` (`moveForms`), reduced to
  the legality map by `_deriveLegality` at `:1029-1041`.
- **Shim.** The guest rebuilds a `MoveForm`-shaped array with the four fields
  that function reads, and casts it.
- **Fix.** Accept the map directly; the webapp can keep deriving it.

### F7. Player info: the mount element is already host-neutral

- **Where.** `components/boardgame-render-player-info.ts` has no store and no
  server call. In the webapp only `boardgame-player-roster-item.ts:262`
  mounts it, and the Redux-bound roster does the module load
  (`boardgame-player-roster.ts:325-340`).
- **In the guest.** Mounted directly, one per seat, with
  `rendererLoaded = true`. It worked first time, and shows Memory's
  "Won Cards" count updating as the game plays.
- **Not provided.** The roster's chip presentation event
  (`player-chip-presentation-changed`), seat state, and join controls. The
  guest draws a label and a colour underline itself.
- **Fix.** S7 is smaller than the proposal sized it: what is missing is a
  roster shell without Redux, not the mount.

### F8. State expansion lives in the Redux selector module

- **Where.** `selectors.ts:260` (`expandGameStateSnapshot`), in a file that
  begins `import { createSelector } from 'reselect'` (`:1`) and builds
  selectors at import. `testing/scenario-fixture.ts:3` imports it from there
  too. `reselect` is not in `server/static/package.json`; only
  `reselect-tools` is (`:41`), which happens to depend on it.
- **Shim.** The guest declares `reselect` itself and bundles 6.6 KB of
  selectors it never calls.
- **Fix.** S3, as planned.

### F9. Fonts come from Google, by a `<link>` in `index.html`

- **Where.** `index.html:99-102` loads Crimson Text, Source Sans 3 and
  Material Symbols Outlined. `styles/md3-theme.css:62-132` names the first two
  in its type tokens; `:186-187` names the third for `md-icon`. Memory's own
  renderer asks for `var(--md-sys-typescale-display-small-font, 'Crimson
  Text', serif)`.
- **In the guest.** All three are omitted. Text falls back to local serif and
  system sans stacks set in `src/guest-main.ts`. Memory's card faces are
  emoji, drawn by the system emoji font, so they look right without any.
  Nothing bundled renders an `md-icon`.
- **Not done.** No font is inlined, and the cost of inlining the two text
  families was not measured.
- **Fix.** S5: fonts become a documented host input with these fallbacks
  written into the theme itself.

### F10. Theme tokens come from a stylesheet linked by `index.html`

- **Where.** `index.html:105` links `/src/styles/md3-theme.css`. Components
  read `--md-sys-color-*` with literal fallbacks, so they render without it,
  in fallback colours.
- **Shim.** `src/guest-main.ts` imports the file with Vite's `?inline` and
  injects it in a `<style>` element (12.8 KB). This is the one thing a
  no-inline-style policy blocks outright; see "Content Security Policy".
- **Fix.** Ship the tokens as a constructable stylesheet export, which no
  policy blocks, or pass them through the protocol.

### F11. Two custom properties are set by `boardgame-game-view`

- **Where.** `boardgame-game-view.ts:120` sets `--animation-length: 0.5s` on
  itself. `:1402-1406` sets `--player-N-color` per seat.
- **In a guest.** Without the first, `boardgame-animatable-item.ts:166-167`
  falls back to 250 ms and every animation runs at double speed with no
  error. This was found by measuring, not by a failure.
- **Shim.** `src/guest-main.ts` sets `--animation-length` on the guest root;
  `src/guest-host.ts` sets the seat colours from `init.players`.
- **Fix.** Defaults belong in `boardgame-render-game`, which already owns
  the override (`:727-733`).

### F12. A wire shape is only normalized by the whole-response decoder

- **Where.** Go omits `Sets` from projected move choices when it is empty.
  `types/server-response.ts:168-229` (`decodeProjectedMoveChoices`) restores
  it, but is not exported; only the whole-response decoders at `:522` and
  `:556` are. `moves/projected-choices.ts:306` then iterates `wire.Sets`.
- **In a guest.** Handing the Go JSON straight to `projectedMoveChoicesWire`
  logs `[projected-choices] rejected invalid snapshot: TypeError: e.Sets is
  not iterable` and drops the choices.
- **Existing bug this exposes.** `testing/scenario-fixture.ts:62` passes the
  same unnormalized object. The checked-in
  `examples/memory/client/scenario-replay.json` has such an entry at version
  3 for both players, so the scenario review page hits this today.
- **Shim.** `src/guest-host.ts` adds `Sets: []` when absent.
- **Fix.** Export per-entry decoders and use them in both places. This is
  also the right boundary for validating what a pattern sends.

### F13. Ambient globals in the bundled closure

A list for seam S14. None needed a shim; all work in an opaque origin.

| Global | Where |
|---|---|
| `window.__bgAnimTestHooks`, assigned at import | `utils/anim-test-hooks.ts:53` |
| `document.getElementById` | `boardgame-component-animator.ts:552` |
| `getComputedStyle` | animator `:272, :501, :591, :919, :1117`; stack `:1369`; card `:737-746`; component `:334, :363`; animatable item `:166`; token `:611`; die `:1508, :1699`; `moves/projected-choice-consumption.ts:96` |
| `window.matchMedia('(prefers-reduced-motion: reduce)')` | animator `:947, :1180`; effect layer `:384, :413, :518, :562`; animatable item `:344`; token `:621`; die `:1407` |
| `ResizeObserver`, `MutationObserver` | stack `:1344`; spatial board `:352`; board viewport `:131`; market `:169`; projected choices `:66`; `projected-choice-consumption.ts:55, :59` |
| `requestAnimationFrame`, `setTimeout`, `Date.now`, `performance.now` | render-game `:368-370, :927, :1016, :1227`; anim hooks `:34` |
| `navigator.onLine`, `online`/`offline` on `window` | render-game `:289, :425-427` |
| `navigator.vibrate`, `navigator.userActivation` | hand view base `:104-105` |
| `fetch`, `document.baseURI` | `boardgame-spatial-board.ts:550, :503` (SVG board art by URL) |
| `window.location`, `sessionStorage`, `localStorage`, `document.cookie`, `crypto` | `utils/companion-surface.ts` (F3) |
| `API_HOST` | `util.ts:17` (F4) |

The guest reads the test hook from its own tests, so it is useful as an
opt-in; it should not be installed on `window` by an import.

### F14. Build: flags, resolution and registration

- **Compiler flags are found by directory.** The source needs
  `experimentalDecorators` and `useDefineForClassFields: false`
  (`server/static/tsconfig.json`). Vite takes them from the nearest
  `tsconfig.json` to each file, and files under `examples/*/client` have
  none above them. The webapp is saved by reaching them through the
  `game-src` symlinks with `preserveSymlinks`. The guest sets the flags for
  every file with `esbuild.tsconfigRaw` and uses the same symlinks.
- **A renderer cannot be imported from where it lives.** Its
  `../../src/client.js` resolves only through
  `server/static/game-src/<game>`. The guest aliases `@game-src` there.
- **Bare imports resolve next to the importer.** `server/static` has no
  `node_modules` in a fresh worktree, so `lit` would not resolve. The guest
  uses `resolve.dedupe`, which also guarantees one Lit copy.
- **Registration is an import side effect, and `client.ts` registers
  everything.** `client.ts:245-255` and its transitive imports register 39
  element modules. Memory uses about a dozen. See "Bundle size".
- **Fix.** S1 and S6. An explicit `define()` entry point per element family
  would let a guest include only what its games use.

### F15. Redux and store types in layer 2

Type-only, so nothing reached the bundle, but each is an import across the
layer boundary that a package split would break.

- `boardgame-render-game.ts:216, :550, :564`: `import('../types/store').CompanionInfo`.
- `boardgame-render-game.ts:9`: `MoveLegalityInfo` from `../selectors.js`.
- `boardgame-render-game.ts:154` and `boardgame-effect-layer.ts:36`:
  `VersionAnimationContext` from `companion-sync.js`.
- `testing/scenario-fixture.ts:5`: `GameChest` from `../types/store.js`.
- `legal/previewLegality.ts:6`: `MovePreviewBatchResult` from `../api.js`.

The guest's own casts (`as unknown as BoardgameRenderGame['chest']` and
similar, six of them) exist because the chest, state and choices types are
the store's, not wire types. S8 would remove them.

### F16. `Game.JSONForPlayer` describes the head, not the state it is given

Found while writing `tools/export-replay`. Go, not client.

- `game.go:266-287`: `Version`, `Finished`, `Winners` and `ActiveTimers` are
  read from the game, whatever `state` was passed. Asked for version 3 of a
  finished 29-version game, it answers `Version: 29, Finished: true`.
- So the webapp's intermediate bundles carry the head's version and outcome.
  The client compensates by trusting `Move.Version` and the bundle order.
- `Game.MoveRecords` (`game.go:469-497`) returned no record for the final
  fix-up after a run in this tool. The exporter reads storage instead. Not
  investigated further.
- **Consequence for the protocol.** A `state` entry carries its own
  `version`, `outcome` and `timers`, and an engine that produces entries must
  compute them per version. Stream C's exporter rebuilds timers from stored
  state for the same reason.

### F17. The first state has no completion signal

- **Where.** `boardgame-render-game.ts:769-794`: a state installed over
  `null` opens no gate cycle. `:1223-1228` closes the gate once, a frame
  later, only when the renderer is instantiated after state was set.
- **Observed.** Memory's first render does animate, for about 500 ms. Those
  animations register with the unopened gate, and when the last settles it
  fires `all-animations-done` carrying the current cycle id. So a signal does
  arrive, but only if something happened to animate.
- **In the guest.** The first version is released after two animation frames
  and reported as `how: "initial"`. A version queued behind it interrupts the
  entrance animation, as it would in the webapp.
- **Fix.** `boardgame-render-game` should always report the first state as a
  cycle, so a host has one rule.

### F18. The webapp's queue installs immediately when it is empty

- **Where.** `boardgame-game-state-manager.ts:1126-1131`: a bundle enqueued
  into an empty queue is scheduled at once, even if the previous bundle is
  still animating. `render-game` then cuts the old cycle over (`:755-757`).
- **In the guest.** Stricter: a version is installed only after the previous
  one is released. Two things the manager does are not reproduced: the
  legacy `animationOverlap` timer (`:1097-1116`) and companion-synchronized
  install times (`:981-1042`). `animationLength`, including the negative
  "skip this version" result, and `motion-cycle-release` are honoured.

### F19. Input mutation (seam S10), measured

`boardgame-component-stack.ts:1158` and `:1222` assign `item.ID` onto the
components they are given. Deep-freezing the expanded state before handing
it to `render-game`, as a Fabric cell value would be, produced 11 uncaught
`TypeError: Cannot add property ID, object is not extensible` in five
versions. The board still rendered. `postMessage` hands the guest private
copies, so the guest itself is unaffected; a trusted component reading a
cell directly would not be.

### F20. Document-relative assets

- `boardgame-token.ts:503`: token art defaults to
  `src/assets/token_${type}.svg`, relative to the document. In a guest that
  is a request that fails. Memory has no tokens.
- `boardgame-spatial-board.ts:503, :550`: board art is fetched by URL.
- `boardgame-card.ts` takes art through `artLayerStyle`
  (`component-art.ts`), which accepts any URL including `data:`. Memory has
  no card art.
- No asset is inlined in this bundle because Memory needs none. The protocol
  has no field for art yet.

## Bundle size

One file, `dist/guest.html`: **589,701 bytes** (576 KiB), **162 KB gzipped**.
That is 14% of the 4 MiB limit. There is no source map, no second file and no
asset.

Sizes below are per module after tree-shaking and before minification
(948 KB in total), scaled to the minified size.

| Part | Share | Approx. minified |
|---|---|---|
| Cards, stacks, animator, motion kernel | 25.0% | 144 KB |
| Boards, markets, zones, drafts, inspector | 18.1% | 104 KB |
| Dice: element, physics simulation, geometry | 14.4% | 83 KB |
| `render-game`, effect layer, renderer bases | 13.3% | 77 KB |
| Status, timers, surface, outcome and other small elements | 9.5% | 55 KB |
| Tokens and 3D solids | 5.0% | 29 KB |
| Move actions and legality | 4.2% | 24 KB |
| Companion Table and Hand bases | 2.9% | 17 KB |
| Guest host, protocol, shims | 2.4% | 14 KB |
| Lit | 2.4% | 14 KB |
| Theme CSS | 1.4% | 8 KB |
| Memory's renderer and generated code | 1.0% | 6 KB |
| `reselect` and the selector module | 1.0% | 6 KB |

- **What dominates.** The five largest modules are `boardgame-die.ts`,
  `boardgame-component-animator.ts`, `boardgame-component-stack.ts`,
  `boardgame-render-game.ts` and `boardgame-spatial-board.ts`.
- **What Memory does not use.** Dice, boards and markets, tokens and solids,
  and the companion bases: about 40% of the bundle, roughly 235 KB minified.
  They are present because `client.ts` registers them as a side effect and
  every renderer imports `client.ts` (F14).
- **What a second game costs.** Its renderer and generated files. Memory's
  are 6 KB. The shared 580 KB is paid once.
- **Headroom.** At this size the limit is not a constraint. Splitting
  registration would matter for the trusted set, where bundle size is review
  cost, more than for the guest.

Regenerate the numbers with `npm run build`; the per-module list is written
to `dist/bundle-report.json`.

## Content Security Policy

Measured, with `tools/serve.mjs` adding a policy to the guest document
(`harness.html?csp=network` and `?csp=strict`). Both are in the test suite.

| Policy | Result |
|---|---|
| `default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; img-src data:; font-src data:` | Plays with no violation. Confirms there is no network access of any kind |
| `default-src 'none'; script-src 'sha256-…'; img-src data:; font-src data:` (no inline style at all) | Plays and animates. 39 style violations in four versions, none for script |

What a stricter policy would break:

- **Inline script.** The bundle is one classic inline `<script>`. A policy
  must allow it by `'unsafe-inline'`, by hash (tested), or by nonce.
  Alternatively `dist/guest.js` is the same code as a file, for a host that
  prefers `script-src` by URL.
- **`eval` and `new Function`.** None in the bundle. `'unsafe-eval'` is not
  needed.
- **Dynamic `import()`.** One, from F1, never executed. Under any policy
  here it would be blocked if it ran.
- **Inline `<style>`.** One: the theme (F10). Blocked by a policy without
  `style-src 'unsafe-inline'` or a hash. Without it the page uses fallback
  colours and the browser's default font.
- **Inline `style` attributes.** 38 violations: Lit templates that bind
  `style="..."`, for example `boardgame-card.ts:906-910` and `:950`.
  Component styles themselves are constructable stylesheets, which a policy
  does not block, so layout and animation survive; per-instance values (card
  aspect, messy rotation, seat colours on badges) are lost. A hash cannot
  allow an attribute. Converting those bindings to `styleMap` or to custom
  properties set through the CSSOM would make the kit work under a
  no-inline-style policy.
- **Web Animations.** Not governed by CSP. All motion is WAAPI.
- **Workers, WebAssembly, frames, forms.** None used.

Not measured: Trusted Types, and whatever policy `cf-iframe` actually
applies. That needs a run inside Common Fabric.
