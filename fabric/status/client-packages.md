# Client packages, the guest on the real host, blackjack in the guest

Branch `claude/fabric-a-client-layers-cdc3bc`. Written 2026-11-21. Follows
[client-layers.md](client-layers.md). Scope: stream A step 4 of
[PROPOSAL.md](../PROPOSAL.md) (seams S1 and part of S6 and S8), then stream B
step 2 redone on the result, then blackjack in the guest.

Nothing pushed, no PR, `master` untouched, nothing installed. Paths owned and
changed: `server/static/**` except `server/static/src/engine/**`,
`fabric/guest/**`, this file, and one Go file outside those paths that the
task named (`boardgame-util/lib/build/static/check_client.go`, see "Go").
Nothing under `src/engine`, `fabric/patterns`, `fabric/declarative` or
`fabric/patterns-declarative` was edited.

**Toolchain for every result here:** Node v24.22.3, Go 1.27.4 with no
`GOTOOLCHAIN`, the repository's own commands. No shim, no transpile step.

## Result

- Layers 1 and 2 are two packages with their own `package.json` and explicit
  `exports`: `@boardgame/elements` and `@boardgame/renderer-kit`, under
  `server/static/packages/`. The webapp reaches them only by name, through
  exported subpaths. The layer test checks that boundary and its allow-list
  is still empty.
- The guest depends on those two packages and nothing else of the client. It
  implements `RendererHost`; one shim is left of the original set.
- The guest bundles Memory and Blackjack, picks the renderer from `init`'s
  `gameName`, and plays a 47-version blackjack game recorded on the Go engine.
- Webapp: every non-browser gate passes, and the renderer suite matches the
  merged baseline. The e2e suite fails everything the merged tip fails plus
  three tests in the final run; each was chased, and the one that reproduces
  alone fails as often before the packages as after ("Commands and counts").

## Read this first: three things the merge changed

Found by running the baseline on the merged tip (`b3c609ce`) before any work.

1. **`npm run test:layers` failed at the merged tip** (7 pass, 1 fail).
   `src/engine/legal/message.ts` imports `src/legal/renderLegalMessage.ts`,
   which the layer table had in the kit, so the engine imported upward at
   runtime. Fixed in `f6ed1d39`, inside my paths: nothing in the client
   imports that module, the engine is its only user, so the table now assigns
   it to the engine layer and it declares its one type itself instead of
   importing it from the wire types. It still sits at `src/legal/`, outside
   `src/engine/`, because moving it means editing the engine's import. **For
   the engine's owner:** move it under `src/engine/` when convenient.
2. **The guest's own suite failed at the merged tip** (19 pass, 2 fail). My
   earlier change made the animation counters (`window.__bgAnimTestHooks`)
   opt-in; two guest tests read them. This was a regression my stream caused
   in the guest stream's code, visible only after the merge. Fixed in
   `099b964c`: both guest entry points call `installAnimTestHooks()`. The
   pattern's `tools/play-two-users.mjs` reads the same counters with a
   fallback to `{}`, so it did not fail, but between the merge and that
   commit its animation numbers would have read as zero.
3. **The renderer suite's "one known failure" is not deterministic.** At the
   merged tip it had two: `scenario-replay.spec.ts:3` (Memory, the one
   recorded in client-layers.md as failing in every run) and
   `scenario-replay.spec.ts:30` (Werewolf). Run by itself four times over,
   that spec file passes 12 of 12. Both tests fail only inside the full
   parallel run, and `:30` fails in some full runs and not others. So the
   honest baseline is "`scenario-replay.spec.ts` fails one or two of its
   tests under the full parallel run", not "one known failure".

## Commits

| Commit | Step | What |
|---|---|---|
| `f6ed1d39` | Merge fix | The legal message renderer assigned to the engine layer (item 1 above) |
| `06077c74` | 1 | Package extraction: 172 files moved, resolution tooling, compatibility entry points, layer test extended, `check-client` staging |
| `099b964c` | 2 | The guest rebuilt on the packages and `RendererHost`; four small kit and elements changes it needed |
| `beb4a6db` | 3 | Blackjack in the guest, played from a conformance document |
| (this commit) | Docs | This file |

`fabric-factoring` (`b3c609ce`) was merged in first; it was a fast-forward.

## Step 1: the packages

| Package | Directory | Layer | Source files | Exported subpaths |
|---|---|---|---|---|
| `@boardgame/elements` | `server/static/packages/elements` | 1 | 98 | 73 |
| `@boardgame/renderer-kit` | `server/static/packages/renderer-kit` | 2 | 20 | 11 |

56 unit-test files moved with the modules they test. The webapp host stays in
`server/static/src`. Every file kept its path under `src/` inside its
package's `src/`, so `git log --follow` works and one rule maps old paths to
new ones. Full description: `server/static/packages/README.md`.

### How the names resolve

The proposal expected this: "Packages in the build cache: not supported
today; if it fails, packages stay as aliases". The Go build keeps one
`node_modules` in a per-user cache shared by every checkout
(`lib/build/static/node.go` copies `package.json` there and runs `npm ci`).
A `file:` link from that cache would point at whichever checkout wrote it
last, so **the webapp does not install its own packages**. Instead
`server/static/scripts/packages.mjs` reads each package's `exports` and
produces:

| Consumer | Mechanism |
|---|---|
| Vite: `vite.config.ts`, `tests/renderer/facade-production.mjs` | One exact alias per exported runtime subpath. An unexported subpath does not resolve |
| `node --test` | `scripts/package-resolve.mjs`, a resolve hook over the same map, added to `npm run test:unit` |
| TypeScript | `paths` in `tsconfig.json`. This is a wildcard and cannot read `exports`; the layer test covers the gap |
| A host outside `server/static` (the guest) | `file:` dependencies and `exports`, the ordinary way. This is the real test that the packages stand alone |

The build cache itself needed no change: the Go build assembles its build
directory by linking every top-level entry of `server/static`, so `packages/`
and `scripts/` arrive there. Checked by running
`boardgame-util build static --prod <dir>` to completion.

### Where each old specifier is kept working

| Specifier | Who relies on it | Where the compatibility lives |
|---|---|---|
| `../../src/client.js` | Every game client, the generator's templates, the checker's creator policy | `server/static/src/client.ts`: `export * from '@boardgame/renderer-kit/client.js'` |
| `../../src/types/boardgame-types.js` | Generated `_types.ts` | `server/static/src/types/boardgame-types.ts`: re-exports the elements' file |
| `../../src/testing/renderer-fixture.js` | Game clients' fixtures | Not moved; `src/testing/` is test support |
| `'/src/<path>.ts'` evaluated in the page | 369 occurrences in 56 Playwright files, naming 71 distinct modules, 53 of which moved | The `boardgame-legacy-src-paths` plugin in `vite.config.ts`: a `/src/<path>` that moved is served as a re-export of the package's file, so the page gets the module instance the app uses. No stub files exist on disk |
| `/src/assets/token_<type>.svg` | Webapp token art, tests | Not moved |
| `*.compile.ts` beside each module | The type-level contract tests | Not moved; their imports now name the packages |

No file under `examples/`, no generator template and no existing Playwright
string was changed. One link changed: `index.html` loads the theme stylesheet
from `/packages/renderer-kit/theme/md3-theme.css`.

### The layer test

`npm run test:layers`, 11 tests, was 8. Still an empty allow-list. New:

- every cross-package import is to an exported subpath, and a types-only
  export is never imported at runtime;
- a package that imports another declares it as a dependency;
- no relative import crosses a package edge in either direction;
- a package holds exactly the files of its layer, and no file of that layer
  is left in `src/`;
- every `exports` entry names a file that exists;
- a fixture tree that commits each of those breaches, so the checks cannot
  pass by seeing nothing.

Test support (`*.compile.ts`, `src/testing`, `src/examples`) and the two
compatibility files are exempt from the exports check but not from
resolution.

### Go

`boardgame-util/lib/build/static/check_client.go` is outside the paths I was
given, and the task named it ("handle ... the build cache as the proposal
describes"; the proposal's S1 names `check_client.go`). It staged framework
declarations with `rootDir` set to `src/`. It now stages `src/` and
`packages/` together and writes `paths` into each game's project so package
names resolve to the staged declarations and not to framework source.

**One thing went wrong on the way, and was caught.** Before that change, I
ran the old `check-client` once against the moved tree. TypeScript reported
the files as outside `rootDir` and still emitted, writing 112 `.d.ts` files
beside the package sources. The layer test failed on them as unassigned
files, which is how they were found. I reproduced it with the exact old
command to be sure of the cause, then deleted only those files (each an
untracked `.d.ts` with a `.ts` sibling). The current code does not do it:
checked after `check-client`, `tsc`, both npm client checks, the Go tests and
a `serve` start.

## Step 2: the guest on the real host

| | Before (merged tip) | After step 2 | After step 3 |
|---|---|---|---|
| `dist/guest.html`, the one-file standalone guest | 591,783 bytes | 589,998 | 595,192 |
| `fabric/patterns/guest/guest.js`, the `cf-iframe` entry | not built by me; the pattern's status gives 1.08 MB | 1,085,032 | 1,091,076 |
| Share of the 4 MiB limit (standalone) | 14.1% | 14.1% | 14.2% |
| Modules in the bundle from `server/static/src` | most of it | 0 | 0 |
| `import(` in the bundle | 1 | 0 | 0 |

The standalone guest is one file. The `cf-iframe` entry is one script plus a
9-line document, as the pattern stream built it; both are under the limit.
The bundle barely shrank because the shims were small: the weight is the
elements (730 KB before minification) and the kit (162 KB). Shrinking it is
S6 (no side-effect registration of unused elements), not done.

**What the guest's host is now.** `GuestHost` builds a `RendererHost`: a state
queue, the timer service, `resolveSurface` returning the solo surface, and
loaders that succeed for a bundled game. It supplies no board preview and no
table actions. Player info is mounted with the kit's `mountPlayerInfo()`.

**Shims deleted:** the prototype patch over the renderer loader (F1); the
`api.ts` redirect plugin and `src/shims/api.ts` (F2, F4); the `companionInfo`
and empty-`gameId` steering (F2, F3); the hand-mounted player info (F7); state
expansion imported from the Redux selector module, and `reselect` with it
(F8); the fake move forms (F6); the projected-choice `Sets` patch (F12); the
`API_HOST` declaration. `fabric/guest/FINDINGS.md` has a table of where each
finding stands.

**Shim left:** `socketActive = true` (F5), one assignment.

**Kit and elements changes this needed.** Each is used by the webapp too.

| Change | Why |
|---|---|
| `types/client-move.ts` (the wire move decoder) moved from the host into the elements | The guest needed it and must not import the host |
| The theme stylesheet moved into the kit as `theme/md3-theme.css` | The guest was importing it from the webapp by path |
| `boardgame-render-game.moveLegality`, legality by move name | A host with no HTTP move forms had to fake them |
| Projected choices read an absent `Sets` as empty | Go omits the empty field; every host was patching it |

**Game clients in the guest.** A game's client directory is linked under the
guest's own `game-src/<game>`, so its `../../src/client.js` lands on the
guest's `src/client.ts`, which re-exports the kit. The guest no longer uses
`server/static/game-src`, which exists only after `boardgame-util serve` has
run in that checkout.

### The bridge protocol

**Nothing changed on the wire.** `fabric/guest/protocol.ts` is untouched, and
`src/fabric-bridge.ts` is untouched. Two behaviors behind it changed:

| | Before | Now |
|---|---|---|
| `init.gameId` | Accepted and ignored | Passed to the renderer host, where it seeds effects and keys transitions. The Fabric bridge does not send one, so nothing changes for the pattern |
| The `error` message "the client tried to reach the server (...)" | Could be posted by the API shim | Cannot occur: there is no server code in the bundle to refuse |

`PROTOCOL.md` is updated for both.

**Not verified: the guest inside Common Fabric.** `npm run build:fabric`
succeeds and produces the script the pattern deploys. I did not deploy it or
run the pattern's tests; that needs a labs checkout and local servers, which
are the pattern session's. To build it I copied their vendored Fabric client
(`vendor/commonfabric-guest.js`, labs `4815052651`, git-ignored) from their
worktree. **The pattern session should rebuild and rerun its two-user script
against this branch before relying on it.**

## Step 3: blackjack

- `src/games.ts` imports blackjack's solo renderer and player-info renderer,
  unmodified, and adds one row. The Table and Hand surfaces are not bundled.
- The harness reads a conformance document directly:
  `conformance/blackjack/scenario-bust-and-finish.json.gz` (three players,
  two rounds, to game over, 47 versions). It gunzips in the page and turns
  each viewer's view into protocol entries, with legality only at settled
  versions, as the live server reports it. No converter tool and no second
  copy of the recording.
- `tests/blackjack.spec.ts`, 7 tests, asserting against facts read from the
  recording itself:

| Claim | Evidence |
|---|---|
| The renderer is chosen by game name | `ready` lists `blackjack` and `memory`; after `init` the tree holds `boardgame-render-game-blackjack` and no Memory renderer; three player-info renderers |
| Hidden cards stay hidden per viewer | After the deal (version 10), for player 1, player 2 and the observer: each seat's hand shows as many faces as that viewer's view holds, 2-1-1, 1-2-1 and 1-1-1 |
| The whole game plays | 47 `animations-done` in order, version 0 `initial` and the rest `played`, no `error`, gate opens equal closes, no watchdog firing, no request or socket from the guest frame |
| Game over | The outcome element reads "Game over! You lost." for player 1; the recording's winner is player 3 |
| Moves | At version 11 the seat to move clicks Hit: one `propose-move`, `Current Player Hit`, `snapshotVersion` 11. As player 2 at the same version the button is disabled and nothing is proposed |

**Not checked for blackjack:** timers (the recording has none), a rejected
move, the reshuffle recording (`scenario-reshuffle`, 569 KB gzipped), the
Table and Hand surfaces, and how it looks beyond one screenshot I read (three
seats, own hand face-up, others' hidden cards face-down, Hit and Stand).

## Commands and counts

Run in `server/static` unless stated.

| Command | Merged tip `b3c609ce` | HEAD `beb4a6db` |
|---|---|---|
| `npm ci` | ok | ok |
| `npm run type-check` | pass | pass |
| `npm run type-check:authoring` | pass | pass |
| `npm run test:unit` | 1165 pass | 1165 pass |
| `npm run test:layers` | **7 pass, 1 fail** | 11 pass (three new tests) |
| `npm run test:check-client` | pass | pass |
| `npm run build` | pass | pass |
| `npm run test:facade-production` | pass | pass |
| `./scripts/go-local run ./boardgame-util --config config.json check-client` (repo root) | pass | pass |
| `npm run test:renderer` | 185 pass, 2 fail | 186 pass, 1 fail (`scenario-replay.spec.ts:3`) |
| `npm run test:e2e -- --reporter=line`, against `boardgame-util serve --storage memory --offline-dev-mode` | 278 pass, 14 fail | 275 pass, 17 fail: the same 14 and 3 others, see below |
| `npm ci`, `npm run type-check`, `npm test` in `fabric/guest` | 19 pass, 2 fail (in a temporary checkout of `f6ed1d39`) | 28 pass, twice, from a clean `npm ci` |
| `npm run build:fabric` in `fabric/guest` | not run | builds |
| `./scripts/go-local test ./boardgame-util/lib/build/... ./boardgame-util/lib/stub/...` | not run | 6 packages pass |
| `./scripts/go-local vet ./fabric/guest/tools/... ./boardgame-util/lib/build/static/` | not run | clean |
| `boardgame-util build static --prod <tmp>` | not run | completes; Vite build inside the assembled directory succeeds |

Intermediate runs:

| When | Renderer | e2e |
|---|---|---|
| After step 1 (`06077c74`) | 186 pass, 1 fail; 185 pass, 2 fail (two runs) | 277 pass, 15 fail |
| After step 2 (`099b964c`) | 186 pass, 1 fail | not run |

**e2e at HEAD against the merged tip.** HEAD fails every test the merged tip
failed, and three more: `die-shape.spec.ts:1372`, `waapi-attrs.spec.ts:37`
and `waapi-companion.spec.ts:81`. That run took 12.7 minutes against 11.1,
on a machine other sessions were using. Each was then run four times by
itself at HEAD:

| Test | In the full run | Alone at HEAD, 4 runs | Further check |
|---|---|---|---|
| `die-shape.spec.ts:1372` | `page.goto` timed out after 30 s | 4 pass | |
| `waapi-companion.spec.ts:81` | expected 600, received 599 | 4 pass | |
| `waapi-attrs.spec.ts:37` | waited 30 s for the first gate open | 3 pass, **1 fail** | 32 runs at `f6ed1d39` (before the packages): **6 fail**. 32 runs at HEAD: 3 fail |

So the one that reproduces alone fails at least as often before the package
extraction as after it. The run after step 1 (15 fail) had two other extras,
`die-roll.spec.ts:1904` and `waapi-buttons.spec.ts:30`; both passed five of
five alone. I conclude the e2e suite fails only where the merged tip does plus
load-dependent flakes, but **no single full e2e run at HEAD equals the
baseline's list exactly**, and none in this whole effort has matched another
exactly either.

The fourteen at the merged tip are the twelve recorded in client-layers.md
plus `die-shape.spec.ts:744` and `player-info-gate.spec.ts:236`, both already
identified there as load flakes.

One renderer run after step 1 was discarded: 177 pass, 3 fail, 7 not run,
with `server.spec.ts:8` lost to "Execution context was destroyed, most likely
because of a navigation". That is the dev-server reload already described in
client-layers.md; it was the first run after `vite.config.ts` changed, and
the next two runs were clean.

### Not run

- `go test ./...` and `npm run type-check:strict` (documented as not green
  before this work).
- The pattern's `cf test` and `tools/play-two-users.mjs`, and anything in a
  `cf-iframe`.
- Any browser but headless Chromium.
- `fabric/declarative` and `fabric/patterns-declarative`: not built or run. If
  they import client source by path under `server/static/src`, those paths
  moved (see "For sibling sessions").

## For sibling sessions

- **Paths moved.** Anything that imported
  `server/static/src/<components|motion|moves|solid|state|status|timers|utils|effects|legal|styles|renderer-host>/...`
  or `src/client.ts` by file path now finds layers 1 and 2 under
  `server/static/packages/elements/src/...` and
  `server/static/packages/renderer-kit/src/...`, same subpath. The intended
  way in is a `file:` dependency and a package specifier; `fabric/guest` is
  the worked example. If a module you need is not exported, add it to that
  package's `exports`.
- **`src/engine` is untouched** and still resolves nothing outside itself
  except `../../legal/renderLegalMessage.ts`.
- **Compiling the packages** needs `experimentalDecorators: true` and
  `useDefineForClassFields: false`.

## What remains

**Stream A**

| Item | State |
|---|---|
| Packages by `npm` link inside the webapp (the rest of A5) | Not done. The webapp resolves by alias because of the shared `node_modules` cache. Needs the Go cache to be per-checkout, or to install packages by copying |
| Specifiers in game clients and generator templates (`../../src/client.js` to `@boardgame/renderer-kit/client.js`) | Not done; the compatibility files carry them. Needs the generator, the checker's policy and the games repository changed together |
| S6: compiled JavaScript, an explicit `define()` entry point, no registration on import | Not done. The packages ship TypeScript source |
| S8 | Done as far as the packages need: no layer-1 type file imports store or API types |
| S5, the rest: shipped token art inside a package; art modules using `import.meta.url` | Not done. `src/assets` stayed in the webapp, so the guest draws no flat token art |
| The custom properties a host must set (FINDINGS F10, F11) | Not written down |
| `boardgame-render-game` calling `host.stateQueue` itself | Not done; hosts still forward two events, as recorded in client-layers.md |
| The 369 `'/src/...'` test strings | Still old paths, served by the plugin. Rewriting them is mechanical and unblocked |

**Stream B**

| Item | State |
|---|---|
| Step 2, the guest host with unmodified renderers | Done, on the real contract, for two games |
| Step 3, hand-written rules replaced by the engine | Not mine; the pattern status says the sandbox refuses classes at module scope |
| Step 4, blackjack on Fabric | The guest side is ready. The pattern, per-seat views and the list-release dependency are the pattern session's |
| One replay exporter | The harness now reads conformance documents. `tools/export-replay` and `replays/memory-full.json` (1.9 MB) are still used by the Memory tests and could be replaced by `conformance/memory/scenario-full-game` |

## Decisions for the owner

1. **Aliases, not `npm` links, inside the webapp.** This is the proposal's
   stated fallback. The cost: three resolution mechanisms derived from one
   `exports` map, and a TypeScript `paths` wildcard that only the layer test
   keeps honest. Making the build cache per-checkout would remove all of it.
2. **`check_client.go` was edited** though it is outside the paths assigned,
   because moving files out of `src/` breaks it otherwise.
3. **Virtual re-exports for old test URLs, not stub files.** A plugin serves
   them; nothing is on disk. The alternative was about 170 one-line files.
4. **`exports` lists what is imported today,** 73 and 11 subpaths, generated
   from actual use and then fixed. It is explicit but not yet curated: some
   entries are there only because a `.compile.ts` test reaches them. A
   deliberate public surface is a design task, best done with S6.
5. **`*.compile.ts` files stayed in `src/`** beside modules that moved away.
   They are consumer-side contract tests and may import both packages, which
   a file inside a package may not.
6. **The guest publishes the animation counters unconditionally.** They are
   test instrumentation on a sandboxed frame's own window. Under the trusted
   goal (S14) this should become a host option.
7. **The legal message renderer** is engine-layer code living at
   `src/legal/`. Its owner should decide where it goes.
8. **The theme stylesheet is in the kit.** It is the webapp's Material theme,
   and a guest needs some theme. Whether the kit should carry a neutral one
   and the webapp its own is open.
