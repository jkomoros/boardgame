# Integration: status

Branch `claude/fabric-a-client-layers-cdc3bc`, 2026-11-21, with
`fabric-factoring` at `a3af29f2` merged in (a fast-forward). Nothing pushed,
no PR, `master` untouched, nothing installed, no Go changed.

- **Toolchains:** Node v24.22.3, Go 1.27.4 with no `GOTOOLCHAIN`, Deno 2.9.8.
- **Labs:** `4815052651`, branch `fix/served-handler-write-identity`, in a
  detached worktree created for this work and removed afterwards. Nothing was
  committed there. Its local servers ran on port offset 470 and were stopped.
- **Final commit for every number below:** `1d7668c4`. The commit that adds
  this file changes nothing else.

## Result

- **Blackjack is playable on Common Fabric in a browser.** Three identities
  and a watcher played two whole games through the deployed pattern, each
  handed its own hole card and nobody else's in an unmodified client.
  **A move takes a median of 18.6 s.**
- Memory replayed on the rebuilt guest: **3.0 s** a move in a space of its
  own.
- One legal-message renderer remains, in the engine.
- Both element requests from the declarative session are applied.
- Every suite passes on the final commit except the two webapp browser
  suites, which fail where they failed before this work, plus timing flakes
  that change from run to run.

## Commits

| Commit | What |
|---|---|
| `be44b229` | The declarative experiment moved onto `@boardgame/elements` (item 1) |
| `cf074964` | One legal-message renderer, in the engine (item 2) |
| `e4ad9f04` | Blackjack on Fabric: a view a seat, a lobby from the table's info, the three-identity script, both games' screenshots (item 3) |
| `1d7668c4` | `collection-id` and `stackScope` on the stack (item 4) |
| (this commit) | This file |

## 1. What the merge broke

One thing, found by the baseline.

**`fabric/declarative` did not build or run.** It imported the client by file
path under `server/static/src` (eleven modules: the stack, the animator, the
component views, three status elements, four motion modules and the
animation counters). The package extraction
moved all of them. On the merged tip `tsc` reported eight errors and all 34
of its Playwright tests timed out.

Fixed in `be44b229`. It now depends on `@boardgame/elements` as a `file:`
dependency and imports by package subpath. Two things about a Vite **dev
server** consuming the package came out of that, and both would catch the
next consumer:

| Symptom | Cause | Fix in the consumer's `vite.config.ts` |
|---|---|---|
| Nothing renders; several copies of one module | The dev server pre-bundles each imported subpath of a linked package as a separate dependency | `optimizeDeps.exclude: ['@boardgame/elements']` |
| `Unsupported decorator location: field` | The dev server reads no tsconfig for source under `node_modules`, so Lit's legacy decorators are compiled as standard ones | `esbuild.tsconfigRaw` with `experimentalDecorators: true` and `useDefineForClassFields: false` |

The guest never hit either because it only runs `vite build`. Both go away
when the packages ship compiled JavaScript (seam S6).

**Nothing else broke.** The engine's rewrite without classes and the package
move did not collide: the webapp's type check, 1,168 unit tests, the layer
test and the build all passed on the merged tip before any change, as did the
guest (28 of 28), the pattern Node tests (36 of 36), `go build ./...` and the
conformance tests.

## 2. One legal-message renderer

`cf074964`. The engine's `engine/legal/message.ts` is the only one. The
client's `src/legal/renderLegalMessage.ts` had no importer anywhere in the
client, so it is deleted, with **no re-export**: nothing needed the old path.
Its eleven tests moved to `engine/legal/render-message.test.ts` and now run
against the engine's renderer. The test that compared the two copies became a
table of the ten texts they agreed on. The `TODO(client)` is gone. The layer
test's allow-list is still empty.

## 3. Blackjack on Fabric

### The protocol the two sides agreed

Written up in `fabric/guest/PROTOCOL.md`, "Inside a `cf-iframe`". The
pattern's proposal (`view` plus `seat0` to `seat6`) was kept, with one
addition so the guest does not have to probe for cells that may not exist.

| Change | Where |
|---|---|
| `TableInfo.seatViews`: how many seats have a view of their own. 7 for blackjack, 0 for Memory | `fabric/patterns/toolkit/view.ts`, `table.ts` (`infoOf`), `blackjack/game.ts` |
| `TableInfo.displayName` and `TableInfo.setupOptions`, from the engine's own variant description | same |
| The bridge reads `view` for who is seated, finds the viewer's seat from `me`, then reads `seat<that seat>` **and no other** | `fabric/guest/src/fabric-bridge.ts` |
| Until a seat's own view arrives for a deal, the viewer is drawn as the observer from `view`. Nothing is drawn for a viewer it was not computed for | same |
| The lobby's title and setup selects come from `info`. It is no longer Memory's | same |
| `init` carries `info.moveInputSchemaFingerprint` | same |

The message protocol itself (`fabric/guest/protocol.ts`) did not change, and
neither did either pattern's `table.tsx` or `main.tsx` beyond a comment. No
`TODO(cfc)` was added or removed; the two in `blackjack/table.tsx` stand.

### What was played

`fabric/patterns/blackjack/tools/play-three-users.mjs`. Four browser
contexts, each registering a fresh passphrase identity on the shell's login
screen: Alice, Bob and Carol sit, Dave watches. It deals and plays to the
end, hitting below 15 and standing otherwise.

| | Game 1 | Game 2 |
|---|---|---|
| Rounds | 1 | 2 |
| Moves | 3 (3 stands) | 9 (3 hits, 6 stands) |
| Result, on all four screens | Alice won | Carol won |
| Secrecy checks passed | 67 of 67 | 145 of 145 |
| Animations played, each guest | about 350 | about 1,035 (983 for the watcher) |
| Animation watchdog firings | 0 | 0 |
| Proposal to the mover's own view, median (min to max) | 17.7 s (16.9 to 23.6) | **18.6 s** (12.5 to 48.4) |
| Deal to every screen | 27.6 s | 27.4 s |

**The secrecy check, and its limit.** After the deal and before every move,
the script reads from each of the four unmodified clients (a) the state its
renderer host was handed and (b) whether each card on screen is face up. For
every hole card still hidden it checks that the card is given to, and shown
face up to, its own seat; given to nobody else; shown face down to everybody
else; and appears nowhere else in what anybody else was given. This is what
was asked: that each identity sees its own hole card and not another's **in
an unmodified client**. It does not show that a modified client could not
read another seat's cell. It can: every seat's view is released to the whole
space behind a `TODO(cfc)` (specs#59), as `blackjack/README.md` says.

**Screenshots:** `fabric/patterns/blackjack/screenshots/` (game 2): lobby,
dealt, first move and finished, for each of the four. I looked at two: Bob's
after the deal (his own Ace and 5 face up, Carol's hole card face down), and
Carol's after the first move.

### Memory, rerun on the rebuilt guest

`fabric/patterns/memory/tools/play-two-users.mjs`, changed in one place: it
picks the ten-card board from the new lobby, whose default is the engine's
(20 cards).

| | In a space of its own | In the space holding two blackjack tables |
|---|---|---|
| Reveals | 16 | 14 |
| Result on both screens | Bob won | Alice won |
| Proposal to the mover's view, median (min to max) | **3.0 s** (2.2 to 5.3) | 4.3 s (3.0 to 8.3) |
| To the other player's view, median | 3.0 s | 4.3 s |
| Animations played, each guest | 282 | 238 |
| Watchdog firings | 0 | 0 |

The engine session measured 3.7 s (2.6 to 6.1) before the guest was rebuilt.
I would not read 3.0 against 3.7 as an improvement: these are two games on a
machine with other work running, and the second column shows how far
neighbours in the same space move the number. The animation counters, which
read `{}` in the engine session's run, are real again.

**Screenshots:** `fabric/patterns/memory/screenshots/`, replaced with the
first of these two games.

## 4. The two element requests

Both applied, in `1d7668c4`, to
`server/static/packages/elements/src/components/boardgame-component-stack.ts`.

| Request | Applied | Note |
|---|---|---|
| `collection-id`, replacing the random per-instance id | As asked | Must be unique among the stacks one animator sees |
| `stackScope`, honored by the animator's stack list | As asked, **plus `connectedStacks()`** | See below |

The second diff as written would have hung the requester's own wrapper. The
wrapper lists every stack by calling the prototype's `_sharedStackList`
getter on the animator's inner stack and then filters to its table. With
`stackScope` set on that same stack, the getter calls the scope, which calls
the getter. `BoardgameComponentStack.connectedStacks()` returns every
connected stack whatever its scope, and the wrapper filters that.

`fabric/declarative` now uses both in place of its two
`Object.defineProperty` overrides, and its 34 tests pass, including the
two-tables tests that exist to exercise the scope. Neither change does
anything for a caller that does not opt in; the webapp sets neither.

## 5. Every suite, on `1d7668c4`

| Suite | Command (directory) | Count | Result |
|---|---|---|---|
| Webapp install | `npm ci` (`server/static`) | | ok |
| Webapp type check | `npm run type-check` | 2 projects | pass |
| Authoring type check | `npm run type-check:authoring` | | pass |
| Webapp and engine unit tests | `npm run test:unit` | 1,168 | 1,168 pass |
| Layers and package boundary | `npm run test:layers` | 11 | 11 pass |
| Client checker's own tests | `npm run test:check-client` | 13 | 13 pass |
| Webapp build | `npm run build` | | pass |
| Production assembly | `npm run test:facade-production` | | pass |
| Renderer fixtures | `npm run test:renderer` | 187 | **186 pass, 1 fail** |
| Webapp e2e | `npm run test:e2e -- --reporter=line`, against `boardgame-util serve --storage memory --offline-dev-mode` | 292 | **277 pass, 15 fail** |
| Go build | `./scripts/go-local build ./...` (root) | | pass |
| Conformance | `./scripts/go-local test ./boardgame-util/lib/conformance/...` | 2 packages | pass |
| Go build tooling | `./scripts/go-local test ./boardgame-util/lib/build/...` | 5 packages | pass |
| Client check of every game | `./scripts/go-local run ./boardgame-util --config config.json check-client` | | pass |
| Guest type check | `npm run type-check` (`fabric/guest`) | | pass |
| Guest | `npm test` (build, then Playwright) | 28 | 28 pass |
| Guest, Fabric entry | `npm run build:fabric` | | builds; `guest.js` 1,093,043 bytes, `dist/guest.html` 595,388 |
| Declarative type check | `npm run type-check` (`fabric/declarative`) | | pass |
| Declarative | `npx playwright test` | 34 | 34 pass |
| Declarative bundle | `npm run build:bundle` | | builds; 409,111 bytes |
| Pattern Node tests | `node --test` on the three `*.test.ts` under `fabric/patterns` | 36 | 36 pass |
| Memory parity in the sandbox | `cf test fabric/patterns/memory/parity.test.tsx --root "$PWD"` | 1 | pass |
| Blackjack parity in the sandbox | `cf test fabric/patterns/blackjack/parity.test.tsx --root "$PWD"` | 1 | pass |
| Memory, multi-user | `cf test fabric/patterns/memory/multi-user.test.tsx --root "$PWD"` | 202 | 202 pass, 153 s |
| Blackjack, multi-user | `cf test fabric/patterns/blackjack/multi-user.test.tsx --root "$PWD"` | 92 | 92 pass, 226 s |

The browser games in section 3 were played at `e4ad9f04`, one commit before
the final one. The commit between them adds two opt-in members to the stack
that the guest does not set. The Fabric guest was rebuilt at `1d7668c4`; the
games were not replayed on that build.

### The same suites on the merged tip, before any change

| Suite | Merged tip `a3af29f2` | Final `1d7668c4` |
|---|---|---|
| Webapp type check, unit (1,168), layers (11), build | pass | pass |
| Renderer | 185 pass, 2 fail | 186 pass, 1 fail |
| e2e | 277 pass, 15 fail | 277 pass, 15 fail |
| Guest | 28 pass | 28 pass |
| Declarative | **`tsc` fails; 0 of 34 pass** | 34 pass |
| Pattern Node tests | 36 pass | 36 pass |
| `go build ./...`, conformance | pass | pass |

The renderer and e2e baseline was taken at `be44b229`, which differs from the
merged tip only under `fabric/declarative` and by one added line of
`exports`.

## Known failures, by test name

"Predates" means the test failed at `d5093955`, the commit before any source
change of this whole effort, in the baseline recorded in
[client-layers.md](client-layers.md) and rerun there on Node 24.

### Renderer (`npm run test:renderer`)

| Test | In the final run | Predates this work? | Evidence |
|---|---|---|---|
| `scenario-replay.spec.ts:3` real Memory replay renders reveal, mismatch, and timer cleanup for each viewer | fails | Yes | Failed at `d5093955` in three runs. It is load-dependent, not deterministic: the spec file alone, four times over, passes 12 of 12 ([client-packages.md](client-packages.md)) |
| `scenario-replay.spec.ts:30` timed Werewolf replay exposes native votes and the partial-vote deadline | passes; failed on the merged tip | Not seen at `d5093955` | First seen at the earlier merged tip `b3c609ce`; passes in the same 12 of 12. Fails in some full parallel runs and not others |

### e2e (`npm run test:e2e`)

Twelve of the final run's fifteen are in the list recorded at `d5093955`:

| Test | Predates? |
|---|---|
| `parity/die-roll.spec.ts:921` reduced motion: runs no tumble and still presents the right face | Yes |
| `parity/die-roll.spec.ts:1726` a roll whose playback never starts still lands where it should | Yes |
| `parity/geometry.spec.ts:22`, `:56`, `:69`, `:106` motion-curve parity (four tests) | Yes |
| `parity/token-3d.spec.ts:105` builds no scene at all for a spacer | Yes |
| `parity/trace.spec.ts:12`, `:53`, `:127` animation parity traces | Yes |
| `parity/trace.spec.ts:159` animation parity traces, pig: die roll | Yes, intermittently: it passed in some baseline runs |
| `tests/games/memory.spec.ts:73` game renderer module is loaded | Yes |

The other three in the final run:

| Test | Predates? | Evidence |
|---|---|---|
| `parity/die-shape.spec.ts:744` the solid survives mounted in the real app (pig) | Not in the `d5093955` list. A load flake that predates the host injection | Failed in a run at `cfa18672`, before the change it could be blamed on; passes 8 of 8 alone ([client-layers.md](client-layers.md)). Also failed on this merged tip |
| `waapi-gate.spec.ts:120` animation completion gate, memory: interrupted cycles | Not in the `d5093955` list | Failed on the merged tip before any change of this task, and again in the final run. **Not run alone**, so I cannot say whether it is deterministic there |
| `parity/die-shape.spec.ts:2280` a d16 reserves exactly `--die-size` | No, and not seen before the final run | A 30 s test timeout. Run alone five times over (55 tests, the whole parametrized group): 55 pass |

Two tests failed on the merged tip and passed in the final run:
`player-info-gate.spec.ts:294` and `waapi-attrs.spec.ts:86`. No full e2e run
in this whole effort has had the same failure list as another. The shared
symptom of the gate tests (no gate open recorded after game creation, so a
helper waits out its timeout) is described in client-layers.md; one of them
fails 1 of 12 at `d5093955`.

## Not run

- `go test ./...` beyond the conformance and build packages.
- `npm run type-check:strict` (documented as not green before this work).
- `fabric/patterns-declarative`: its `cf check`, its two-user script and its
  probes. They need the labs spike branch
  (`spike/boardgame-trusted-components`) and its patched shell, not the fix
  branch this task used. Its dependency `fabric/declarative` was fixed and
  its bundle builds, but the pattern was not run against it.
- The standalone guest's Playwright suite does not cover the Fabric bridge.
  The view-a-seat logic is exercised only by the browser games.
- A modified client, for blackjack's secrecy. Out of scope by the task's own
  wording.
- More than three players and a watcher; a second deal on one table; leaving
  a seat; a reload mid-game; blackjack's reshuffle; any browser but headless
  Chromium.
- The whole README sequence from a clean checkout in one go.

## Still inconsistent between the pieces

1. **Blackjack takes 18 s a move; Memory takes 3 s.** Each input rewrites
   eight views (the watcher's and seven seats'), each holding every version
   of the last input, and the deal alone is eleven versions. Seven seat views
   are computed whatever the number of people seated. This is the pattern's
   shape, not the guest's, and it is the first thing a person playing will
   notice.
2. **What "each seat sees only its own" rests on.** The guest asks only for
   its own seat's cell. The release rule lets any reader of the space read
   all of them. The guest's document and the pattern's README both say so;
   nothing enforces it.
3. **A deal is not animated for the person who deals.** The view carries
   every version of one input, but the guest shows only the newest entry of
   the first window under a key (PROTOCOL.md, rule 4), and a seated player's
   own view arrives as a new key. So the eleven versions of blackjack's deal
   are cut to, not played. Game 1's guests opened the animation gate 12 times
   for about 20 versions.
4. **Two ways into the client.** `fabric/guest` and `fabric/declarative`
   depend on the packages. Game clients, and so anything that bundles them,
   still reach the kit through a `src/client.ts` compatibility file
   (`../../src/client.js`), which the guest reproduces under its own
   `game-src/`. The generator's templates are unchanged.
5. **Three resolution mechanisms for one `exports` map** inside the webapp
   (Vite aliases, a Node hook, TypeScript `paths`), because of the Go build's
   shared `node_modules` cache. Consumers outside it use ordinary `file:`
   dependencies and need the two Vite settings in section 1 for a dev server.
6. **The lobby's default changed for Memory.** It now offers the engine's
   variants with the engine's defaults, so a Memory table deals 20 cards
   unless told otherwise, where the old lobby's first option was 10. It also
   offers "Card Set", which the old lobby did not.
7. **Two recordings of Memory for the guest's tests.** The blackjack tests
   read a conformance document. The Memory tests still read
   `replays/memory-full.json` (1.9 MB) from `tools/export-replay`.
8. **The vendored Fabric client reports `script-src blocked wasm-eval`** on
   every load in a `cf-iframe`. Memory's FINDINGS already records it; it is
   the labs client, not the boardgame guest, and nothing depends on it.
9. **`fabric/patterns-declarative/elements.ts`** types the declarative
   elements for a pattern and was not rechecked after the two element
   changes. They are additive, so I expect no effect; I did not run it.
