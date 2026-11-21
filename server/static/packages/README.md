# Client packages

The web client is three source trees. Two of them are packages here; the
third is the webapp in `../src`.

| Package | Layer | Holds |
|---|---|---|
| `@boardgame/elements` (`elements/`) | 1 | Presentational custom elements (card, token, die, stack, boards, status and layout elements), the motion system, state expansion, and the wire, state and move-intent types they take |
| `@boardgame/renderer-kit` (`renderer-kit/`) | 2 | What a game renderer is written against (`client.js`), the base renderer classes, drafts, effects, `boardgame-render-game`, `boardgame-render-player-info`, and the renderer-host contract (`renderer-host/renderer-host.js`) |

`@boardgame/renderer-kit` depends on `@boardgame/elements`. Neither depends
on the webapp, Redux, the HTTP API or anything else a particular host is
built from. `lit` is a peer dependency of both.

The layers and their rules are described in `../src/ARCHITECTURE.md` and
enforced by `npm run test:layers`.

## Using a package

Each `package.json` lists every public module under `exports`, explicitly;
there are no wildcard exports. Import a module by its listed subpath, with a
`.js` extension:

```ts
import '@boardgame/renderer-kit/components/boardgame-render-game.js';
import type { RendererHost } from '@boardgame/renderer-kit/renderer-host/renderer-host.js';
import { expandGameStateSnapshot } from '@boardgame/elements/state/expand.js';
```

A subpath that is not listed does not resolve. To make a module public, add
it to `exports`; `npm run test:layers` fails on any cross-package import that
is not exported, and on an export that names a missing file.

**The packages ship TypeScript source, not compiled JavaScript.** A consumer
must compile them with `experimentalDecorators: true` and
`useDefineForClassFields: false`. With other settings Lit's reactive
properties break silently. Shipping compiled output with an explicit
`define()` entry point is seam S6 of `fabric/PROPOSAL.md` and is not done.

**Importing an element module registers its custom element** as a side
effect, as before.

### From another directory (the Fabric guest)

Depend on the packages with `file:` dependencies and resolve them through
`exports` like any other package:

```json
"dependencies": {
  "@boardgame/elements": "file:../../server/static/packages/elements",
  "@boardgame/renderer-kit": "file:../../server/static/packages/renderer-kit"
}
```

`fabric/guest` does exactly this.

### From the webapp

The webapp does **not** install its own packages into `node_modules`. The Go
build (`boardgame-util build`/`serve`) keeps one `node_modules` in a per-user
cache, shared by every checkout, and a `file:` link from that cache would
point at whichever checkout wrote it last. So inside `server/static` the
package names are resolved from the same `exports` maps by
`scripts/packages.mjs`:

| Consumer | How it resolves `@boardgame/*` |
|---|---|
| Vite (`vite.config.ts`, and `tests/renderer/facade-production.mjs`) | `packageAliases()`: one exact alias per exported runtime subpath |
| `node --test` (`npm run test:unit`) | `scripts/package-resolve.mjs`, a resolve hook over the same map |
| TypeScript (`tsconfig.json`) | `paths`, which is a wildcard and cannot read `exports`; the layer test checks every import against `exports` instead |
| `boardgame-util check-client` | Declarations for `src/` and `packages/` are staged together, and each game's project maps the package names to the staged declarations |

The Go build assembles its build directory by linking every top-level entry
of `server/static`, so `packages/` and `scripts/` arrive there without any
special handling, and the `node_modules` cache is untouched.

## Old paths that still work

Nothing that named a file under `src/` before the packages existed had to
change:

| Specifier | Used by | Kept working by |
|---|---|---|
| `../../src/client.js` | Every game client, the code generator's templates, the client checker's policy | `src/client.ts`, which re-exports `@boardgame/renderer-kit/client.js` |
| `../../src/types/boardgame-types.js` | Generated `_types.ts` in every game client | `src/types/boardgame-types.ts`, which re-exports `@boardgame/elements/types/boardgame-types.js` |
| `../../src/testing/renderer-fixture.js` | Game clients' fixtures | The file never moved; it is test support in `src/testing/` |
| `'/src/<path>.ts'` strings evaluated in the page | The Playwright suites, in several hundred places | The `boardgame-legacy-src-paths` plugin in `vite.config.ts`: a `/src/<path>` that no longer exists but exists in a package is served as a re-export of the package's file, so the page gets the same module instance the app uses |
| `/src/assets/token_<type>.svg` | The webapp's token art, some tests | The files never moved |

Files keep the path they had under `src/` inside their package's `src/`
(`src/motion/timing.ts` is `elements/src/motion/timing.ts`), which is what
lets the last-but-one row work by rule instead of by table.

The legacy-path plugin is test-path indirection only. It bypasses `exports`,
and no source file may rely on it.
