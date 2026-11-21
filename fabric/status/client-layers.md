# Stream A status: client layers

Branch `claude/fabric-a-client-layers-cdc3bc`, cut from `34bfd655`.
Written 2026-11-21. Scope: stream A of [PROPOSAL.md](../PROPOSAL.md), steps 1
to 3 of that stream (the layer table, S3, the test hook, S8, S4, S2, S7, S5).
No file moved into a package, and no import specifier used by
`examples/*/client`, the code generator, or an existing Playwright string
changed.

## Result

- `server/static/src` has a checked-in layer table and a test that fails on
  any upward import. The allow-list is **empty**: no runtime or type-only
  import goes from a lower layer to a higher one.
- `boardgame-render-game` imports nothing from the host layer, at runtime or
  in types. Everything it needs from outside its subtree arrives through a
  typed `RendererHost`.
- The webapp behaves as before, as far as the suites can show. The renderer
  suite matches baseline at every commit. The e2e suite fails only where
  baseline fails, plus one or two tests per run that vary from run to run and
  that also fail before these changes (see "Verification per commit").

## Commits

| # | Commit | Step | What changed |
|---|---|---|---|
| 1 | `d5093955` | Layer table and import test | `scripts/layers.json`, `scripts/check-layers.mjs`, `scripts/check-layers.test.mjs`, `npm run test:layers`, one line in the `strict-client` CI job |
| 2 | `1d968dc2` | S3 | `expandGameStateSnapshot` and helpers to `src/state/expand.ts`; `reselect` declared in `package.json` |
| 3 | `68115d30` | Test hook opt-in | `installAnimTestHooks()`; the webapp opts in from a `localStorage` key that `playwright.config.ts` seeds |
| 4 | `2bb39498` | S8 | `src/types/wire.d.ts`; `api.d.ts` and `store.d.ts` re-export from it; `src/status/gathering-start.ts` |
| 5 | `cfa18672` | S4 | `TableActions` and `BoardPreviewTransport` in `src/renderer-host/renderer-host.ts`; webapp implementation |
| 6 | `fe983f72` | S2 | Surface resolver and module loader injected; full `RendererHost`, `RendererMoveInputs`, `RendererMount`; `src/webapp/renderer-host.ts` |
| 7 | `07355e5a` | S7 | `RendererHost.loadPlayerInfoModule`, `mountPlayerInfo()`, `ensurePlayerInfoRenderer()`; the roster loads through the host |
| 8 | `5dcc5f2b` | S5 (small) | `setTokenArtResolver()`; the webapp's path in `src/webapp/token-art.ts` |
| 9 | `0d95e985` | Not in the brief | Draft binding interfaces to `src/moves/draft-bindings.ts`, which empties the allow-list |
| 10 | this commit | Docs | This file and a "Layers and the renderer host" section in `src/ARCHITECTURE.md` |

## Environment problems found before any change

Both are recorded for sibling sessions in the shared notes folder
(`stream-a-environment-notes.md`).

1. **`boardgame-util serve` and `check-client` do not run under the installed
   Go 1.27.4.** `go.mod` pins `golang.org/x/tools v0.48.0`, which cannot read
   Go 1.27 export data ("export data version 5 is greater than maximum
   supported version 4"). `npm run test:renderer` starts its server with
   `go run ./boardgame-util serve`, so as checked in, **neither browser suite
   can start on this machine**.
   Workaround, touching no tracked file: a copy of `go.mod`/`go.sum` under the
   git-ignored `.cache/gomod/` with `x/tools` at `v0.53.0`, a `boardgame-util`
   binary built with `-modfile`, and a `go` shim first on `PATH` that redirects
   `go run ./boardgame-util …` to that binary. Every browser result below was
   produced this way. The real fix is a `go.mod` bump, which is not stream A's.
2. **`npm run test:unit` does not run on the installed Node (18, 20, 21).** It
   needs type stripping (CI uses Node 24). Under Node 20 it reports "Could not
   find 'src/**/*.test.ts'". The same test files were run on Node 20 through a
   scratch `--import` loader that transpiles `.ts` with the repo's own
   `typescript`, with the per-file timeout raised to 120s because Node 20
   applies `--test-timeout` to a whole file. "Unit" below means that run.
   `scripts/check-layers.test.mjs` is plain `.mjs` for the same reason: it runs
   on any Node, under `npm run test:layers`.

## Baseline

Taken at `d5093955` (commit 1, which changes no file under `src/` or `tests/`)
in a separate checkout, so nothing was being edited underneath it.

| Suite | Command | Baseline |
|---|---|---|
| Install | `npm ci` | ok |
| Type check | `npm run type-check` | pass |
| Unit | `npm run test:unit` | cannot start on Node 20; via the loader, 787 of 787 pass |
| Build | `npm run build` | pass |
| Renderer | `npm run test:renderer` | 184 pass, **1 fail** |
| e2e | `npx playwright test` against `boardgame-util serve --storage memory --offline-dev-mode` | 280 pass, **12 fail** |

Baseline renderer failure (fails in every run at every commit, including
three runs at baseline):

- `tests/renderer/scenario-replay.spec.ts:3` real Memory replay renders
  reveal, mismatch, and timer cleanup for each viewer

Baseline e2e failures:

- `parity/die-roll.spec.ts:921` reduced motion: runs no tumble and still presents the right face
- `parity/die-roll.spec.ts:1726` a roll whose playback never starts still lands where it should
- `parity/geometry.spec.ts:22`, `:56`, `:69`, `:106` motion-curve parity (four tests)
- `parity/token-3d.spec.ts:105` builds no scene at all for a spacer
- `parity/trace.spec.ts:12`, `:53`, `:127`, `:159` animation parity traces (four tests; `:159` is intermittent)
- `tests/games/memory.spec.ts:73` game renderer module is loaded

## Verification per commit

Each commit was checked out by itself in the separate checkout.

| Commit | Type check | Layer test | Unit | Renderer | e2e |
|---|---|---|---|---|---|
| 1 `d5093955` | pass | pass | 787/787 | 184 pass, 1 baseline fail | 280 pass, 12 baseline fail |
| 2 `1d968dc2` | pass | pass | 787/787 | 184 pass, 1 baseline fail | not run |
| 3 `68115d30` | pass | pass | 787/787 | 184 pass, 1 baseline fail (second run; see below) | not run alone; covered by 5 |
| 4 `2bb39498` | pass | pass | 787/787 | 184 pass, 1 baseline fail | not run |
| 5 `cfa18672` | pass | pass | 787/787 | 184 pass, 1 baseline fail | 281 pass, 11 fail, all baseline |
| 6 `fe983f72` | pass | pass | 787/787 | 184 pass, 1 baseline fail | 278 pass, 14 fail: 12 baseline and 2 others, see below |
| 7 `07355e5a` | pass | pass | 787/787 | 186 pass, 1 baseline fail | not run |
| 8 `5dcc5f2b` | pass | pass | 788/788 | 186 pass, 1 baseline fail | not run |
| 9 `0d95e985` | pass | pass | 788/788 | 186 pass, 1 baseline fail | run 1: 279 pass, 13 fail (11 baseline, 2 others); run 2: 279 pass, 13 fail (12 baseline, 1 other). See below |

The renderer suite has 186 tests from commit 7 on: commit 7 adds
`tests/renderer/player-info-mount.spec.ts` (two tests).

Also at HEAD: `npm run build`, `npm run type-check:authoring`,
`npm run test:check-client` (13 of 13), `npm run test:facade-production`,
`boardgame-util check-client` ("Client checks passed"), and
`go test ./boardgame-util/lib/build/static/... ./boardgame-util/lib/stub/...`
all pass (the two Go commands with the `x/tools` workaround).

**The two extra e2e failures at commit 6** were
`parity/die-shape.spec.ts:744` (the solid survives mounted in the real app)
and `parity/player-info-gate.spec.ts:236` (a roster animation with no board
cycle open leaves the gate untouched). That run took 11.1 minutes against a
usual 9, on a machine other sessions were also using. Both tests were then run
four times each by themselves at commit 6 and four times each at baseline:
8 of 8 pass at both.

**The e2e suite is flaky under load on this machine, at baseline as well.**
Five full runs were made (baseline, commit 5, commit 6, HEAD twice). The 12
baseline failures recur in every run, except `trace.spec.ts:159`, which passed
in two of them. Beyond those, a different test or two failed in three of the
four later runs, never the same set:

| Run | Extra failures |
|---|---|
| Commit 5 | none |
| Commit 6 | `die-shape.spec.ts:744`, `player-info-gate.spec.ts:236` |
| HEAD, run 1 | `die-shape.spec.ts:744`, `player-info-gate.spec.ts:114` |
| HEAD, run 2 | `waapi-attrs.spec.ts:37` |

Each was chased down rather than assumed:

- `die-shape.spec.ts` and `player-info-gate.spec.ts`, whole files, twice each
  at HEAD: 226 of 226 pass. The same at commit 5, which predates the host
  injection: 223 pass and 3 fail, **including `die-shape.spec.ts:744`**, with
  the machine busier (6.2 minutes against 4.7).
- `player-info-gate.spec.ts:114` twelve times at baseline: **1 fails**.
  Twelve times at HEAD: 0 fail.
- `waapi-attrs.spec.ts`, whole file, four times at HEAD and four at baseline:
  20 of 20 pass at both.

So every extra failure either reproduces before the change it could be blamed
on, or does not reproduce at all. The shared symptom of the gate tests (the
animation gate records no open after game creation, so the helper waits out
its timeout) is an existing race between the first state and the renderer
module load. It is worth its own fix, and it is not stream A's.

**One renderer run was discarded.** The first renderer run at commit 3 lost
`tests/renderer/server.spec.ts:31` to "Execution context was destroyed, most
likely because of a navigation" (a dev-server reload right after the checkout
changed files; `renderer-fixture-helpers.ts` documents the same hazard), and
six serial tests after it did not run. A second run of the same commit gave 184 pass and the one baseline failure, and that is the row above.

## The layer table

`server/static/scripts/layers.json`. Order, lowest first: `engine`,
`elements`, `kit`, `host`. `test-support` is unranked: it may import anything
and nothing ranked may import it.

- **Rules:** `src/engine/**` is `engine` (stream C's directory; the test
  accepts new files there without a table entry and fails if they import
  upward). `*.compile.ts`, `src/testing/**` and `src/examples/**` are
  `test-support`. `*.test.ts` files are not source and are skipped.
- **Every other file is listed by name.** A new file that is not in the table
  fails the test.
- **What the test sees:** static imports, `import type`, imports whose every
  binding is `type`, re-exports, dynamic `import('…')` with a literal, and
  `import('…')` type references. Type-only edges are reported by a separate
  test from runtime edges. A test against a throwaway tree checks each of
  those forms, so a parser regression cannot make the gate pass by seeing
  nothing.
- **Not seen:** a dynamic import of a computed path. The only ones are the
  `game-src` loads, which now live in the host layer.

## Allow-list

Empty. It started at 30 entries (6 runtime, 24 type-only), briefly 31 after
commit 2, and went to 16, 11, 10, 10, 10 and 0 over commits 4 to 9.

## Decisions for the orchestrator to review

1. **Layer assignments the proposal did not spell out.**
   - `moves/input.ts` and `moves/target-list.ts` are `elements`, beside the
     four move-intent files the proposal names: `action.ts` imports `input.ts`,
     and the `boardgame-target-list` element imports `target-list.ts` at
     runtime.
   - `timers/*`, `status/*`, `state/*`, `utils/retry-policy.ts`,
     `utils/seeded-random.ts`, `utils/anim-test-hooks.ts` and
     `companion-avatar-catalog.ts` are `elements` (pure, imported by elements).
   - `effects/*`, `legal/*`, `styles/renderer-styles.ts` and
     `boardgame-render-player-info` are `kit`.
   - `components/companion-sync.ts` is `host`. It is pure logic plus a
     `window.__bgCompanionSync` write at import. Nothing below the host imports
     it any more; if the guest needs the estimator it should move down and its
     global should become opt-in like the animation hook.
   - `boardgame-player-chip` and `boardgame-game-path` are `host` because only
     host files use them.
2. **Commit 9 goes beyond the brief.** The brief expected deferred allow-list
   entries. The last ten were type-only imports of draft binding interfaces by
   elements; moving the interfaces (not the controllers) to a layer-1 file was
   mechanical and leaves the kit's public names unchanged through re-exports.
   It can be reverted alone if the orchestrator prefers the entries deferred.
3. **How much of the state queue moved.** `RendererHost.stateQueue` is real
   and the webapp releases through it, but `boardgame-render-game` still
   *announces* settlement with `all-animations-done` and `motion-cycle-release`
   events, and the host forwards them to its queue. The forwarding stayed in
   `boardgame-game-view` because it holds two webapp decisions: drop a release
   for a stale cycle, and never let local progress advance a server-scheduled
   companion slot. A guest host must listen for the same two events. Moving
   the call into `render-game` is small but changes timing-sensitive code, so
   it was left for a step with its own e2e run.
4. **Reactive inputs are properties, not host members.** The move, preview and
   target-preview transports and the submission gate were already injected as
   properties and change with the connection. They are typed as
   `RendererMoveInputs`, which `boardgame-render-game` implements, rather than
   folded into `RendererHost`.
5. **Behavior with no host.** `boardgame-render-game` with `host` unset mounts
   the solo surface, shows a renderer error instead of loading a module, grays
   no board target by preview, and a Table surface reports that table actions
   are unavailable and hides the join address and QR. The webapp always passes
   a host, so none of this is reachable there.
6. **One existing Playwright test was edited** (`renderer-fixture.spec.ts`,
   "dynamic host ignores a module load superseded by a newer game route"). It
   stubbed the private `_loadRendererModule` and relied on tab storage being
   read with no host. It now mounts with `createWebappRendererHost({
   importModule })`. Its assertions, including the three expected
   `../../game-src/…` module paths, are unchanged. This adds one new
   `'/src/webapp/renderer-host.ts'` string for S1's alias to keep resolving;
   `player-info-mount.spec.ts` adds three more (`boardgame-render-player-info`,
   `boardgame-base-player-info-renderer`, `client`).
7. **The animation hook's opt-in is a `localStorage` key**
   (`boardgame:anim-test-hooks`), seeded by `storageState` in
   `playwright.config.ts`, so no animation spec changed. A person can set it by
   hand for a bug report. Under `fabric`'s trust goal (S14) the hook is now
   reachable only when a host calls `installAnimTestHooks()`.
8. **Token art uses a module-level resolver** (`setTokenArtResolver`), not an
   ancestor lookup, because tokens are drawn outside `boardgame-render-game`
   too (the roster's player-info renderers). That is one more module global
   for S9 to scope. With no resolver a token draws a transparent square in
   place of shipped art; the generated solids do not use the image.
9. **`src/webapp/` is a new directory** for the webapp's host files. It sits
   two levels below `server/static`, like `src/components/`, so the relative
   `../../game-src/…` specifiers are byte-for-byte what they were.
10. **`reselect` is declared as `^4.0.0`**, the version the lockfile already
    resolved through `reselect-tools`. Only the root dependency line of
    `package-lock.json` was added, by hand, because the local npm rewrites
    unrelated `libc` fields.

## Not done, by design

- S1 (packages, path aliases, the 371 `'/src/…'` Playwright strings), S6
  (compiled output, an explicit `define()` entry point) and the build-cache and
  games-repository work are later stream A steps.
- Art modules that use `import.meta.url` (the other half of S5) are untouched.
- `boardgame-render-game` still reads `navigator.onLine` and listens for
  `online`/`offline` on `window`, and uses `requestAnimationFrame`. Those are
  ambient access for S14's audit, not host-layer imports.
