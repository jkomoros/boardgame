# Status: the patterns on the viewer-bound release

Branch `claude/fabric-b-memory-pattern-70154b`, 2026-11-21, with
`fabric-factoring` at `f76bde9b` merged in first (a fast-forward). Nothing
pushed, no PR, `master` untouched, nothing installed, no Go changed.

- **Toolchains:** Node v24.22.3, Deno 2.9.8, Go 1.27.4 with no `GOTOOLCHAIN`
  (Go was not run for this work).
- **Labs:** `70f5c390ec`, the head of the local branch
  `worktree-agent-a53d3b2fae9b79d6b`, in a detached worktree of my own
  (`.claude/worktrees/boardgame-viewer-port`), removed afterwards. Nothing
  was committed there; I made no labs change and needed none. Its local
  servers ran on port offset 340 with
  `EXPERIMENTAL_VIEWER_BOUND_RELEASE=true` and were stopped.
- One comparison run used labs `4815052651`
  (`fix/served-handler-write-identity`), and a bisect used three commits
  between the two, each in a detached worktree of its own on port offset
  350, all stopped and removed.

## Result

- **The patterns are on the viewer-bound release, in its idiomatic shape.**
  One sealed record, one `move`, one `viewFor` that asks who it is computed
  for, and the card table's rule with its input witness. Both interim seams
  are gone. Memory and blackjack are now the same pattern over different
  games.
- **At the runtime's display boundary the table's owner is shown no hand.**
  In labs' own harness against our blackjack table, and in sixteen browser
  games with an owner watching (2,255 secrecy checks, none failed). The
  owner binding `move` from another module, a second run of the module over
  the live table, and a write by other code all get nothing.
- **A person with the `cf` CLI still reads the sealed table.** A key with
  only `READ` printed the game's secret salt. And a space's owner has a
  documented command to release a clause by intent, which lists ours. The
  first is the open delivery non-conformance (specs#95); the second is by
  design. See "What is not held".
- **The 18.6 s move is gone, and the runtime did that, not the pattern.**
  The unchanged pattern takes a median 2.6 s a move and 5.5 s to deal on
  today's labs branch, where it took 18.6 s and 27 s on the older commit,
  which still does today. A bisect puts it on one labs commit, `2033f78fae`,
  the fix that stamps a computation's result whole.
- **The port by itself made moves slower:** median 3.6 s, deal 8.3 s. Each
  viewer's view is computed after the commit, about 0.7 s behind it.
- **The steps after it got that back and more.** Final: **deal 1.2 s, a
  move a median 2.6 s, a round's end 9.6 s** at blackjack (5.5 s, 2.6 s and
  13.0 s at baseline); **Memory 1.5 s a move** (2.2 s at baseline, 2.9 s
  after the port alone).
- **What is left is the versions.** A move costs about 0.6 s for each
  version it produces, because a fix-up move is a whole state in every
  viewer's view. Holding only the head makes every blackjack move 1.2 to
  2.1 s, a round's end included, and gives up a card at a time. That is the
  owner's call; it is measured and not shipped.

## Commits

| Commit | What |
|---|---|
| `9a6ac223` | The port: one `viewFor` lift a game, the witnessed viewer rule, the table a record that starts as `null`, the bridge reading one view (steps 1 to 3) |
| `1c848c95` | Step 4(a): the stored window becomes the last input's states and the state before it |
| `c8db1ab2` | Step 4(c): the view says which seat is its viewer's; `identify`, `me` and the seats' principals are gone |
| `c52d9224` | Step 4(a), second part: a table keeps only the head of a deal |
| `1803989f` | A view is computed from one plain copy of the table; the display test's timing case; the browser script's timing split; the `tick` marker |
| the commit after it | This file, the READMEs, the screenshots |

Steps 4(b) and 4(d) changed no code. Why is under "What is still in
userland".

## What changed

**One view per viewer.** Each game's `table.tsx` exports one released lift:

```ts
export const viewFor = lift(({ table }: Viewed): TableView =>
  viewForPrincipal(blackjackGame(), table, currentPrincipal())
);
```

`viewForPrincipal` (`toolkit/table.ts`) finds the principal's seat in the
sealed table and returns the engine's sanitized view for it, or a watcher's
if they hold none. `seat0` to `seat6`, `seatView`, Memory's shared view,
`TableInfo.seatViews` and `MAX_SEATS` are deleted. Nothing in the types says
how many seats a table has: a seat exists when someone sits, and the deal
creates the game with that many players.

**The rule** is labs' card-table rule: `TransformedBy{viewFor, inputWitness:
move, viewer: v("viewer")}` releasing to `User(v("viewer"))`. The input
witness is back. Both interim seams are gone: the dropped witness (Memory
and blackjack) and every seat's view released to the whole space
(blackjack).

**The state is a record**, `TableState | null`, written whole by `move`. It
is not a string, and nothing about it is shaped around a gap but its
default, `null`, which is the one `TODO(cfc)` left in each `table.tsx`. The
handler steps a plain copy of what it read (`tableOf`), as the card table's
`gameOf` does.

**The two games' `table.tsx` are now the same file** but for the game they
name. Memory used to have one view for everyone; it now has one a viewer,
like blackjack. That costs Memory something: see the measurements.

**The guest bridge** reads `view` and `info` and sends `move`. It has no
per-seat cells, no `me`, no `identify`. `PROTOCOL.md`, "Inside a
`cf-iframe`", is rewritten to match.

## Measurements

### Method

The same two scripts as before, against the deployed piece on local labs
servers, with server execution on (the toolshed's default):

- `blackjack/tools/play-three-users.mjs`, two rounds: three browser
  identities sit and play, hitting below 15, and a fourth watches. The
  watcher is now granted `OWNER` of the table's space (it was `WRITE`).
- `memory/tools/play-two-users.mjs`, the ten-card board.

Each number is from a fresh space. "Click to the mover's screen" is from the
click to the mover's guest showing a newer version. A proposal's cost
follows how many versions it produces (the move and its fix-ups), so the
blackjack moves are also given by that count. Every run of every step
passed all of its secrecy checks.

Each step was deployed from a snapshot worktree of its commit, so that the
steps could be interleaved on the same servers in one sitting. The baseline
is `f76bde9b`, the pattern as it was.

Bytes are from `fabric/patterns/tools/measure-bytes.mjs`, which plays the
same games through the toolkit under Node and sizes the JSON a pattern
would store and release, for the viewers who are looking (three players and
a watcher at blackjack; two players and a watcher at Memory). They do not
move with load.

### The baseline is not what it was: labs got five to seven times faster

The first thing measured was the unchanged pattern on today's labs branch.

| Unchanged pattern (`f76bde9b`) | labs `4815052651`, measured earlier | labs `4815052651`, today | labs `70f5c390ec`, today |
|---|---|---|---|
| Deal to every screen | 27.4 s, 27.6 s | 27.4 s | 5.5 s (5.2 to 6.8), 7 games |
| A move, median | 18.6 s (12.5 to 48.4) | 9.9, 12.4, 14.3, 18.2 and 48.5 s for the first five moves | 2.6 s (0.6 to 15.2), 60 moves |

The older commit reproduces the earlier figures today, so the difference is
not load. `70f5c390ec` is twelve commits on top of `4815052651`. I bisected
them on the deal alone, one run a commit, each on its own servers:

| Labs commit | Deal to every screen |
|---|---|
| `4815052651` (the base) | 27.4 s |
| `16122f1e4e` "Keep the writer's stamp on a record its writer reads and sets" | 21.2 s |
| **`2033f78fae` "Stamp a computation's result as the run that wrote it"** | **5.1 s** |
| `e7cdd1270c` (two commits later) | 5.2 s |
| `70f5c390ec` (the head) | 5.5 s |

`2033f78fae` is the commit directly after `16122f1e4e`, so it is the one.
Its message says a computation's result was stamped only at the paths whose
value changed and is now recorded as one whole-value destination.
**Inferred:** that this made a large view's label one entry where it had
been one for each path, and that the server's label work per delivery fell
with it. I did not look further. It is a correctness fix that happens to be
worth a factor of four to any pattern that releases a large computed value.

So "18.6 s a move" is no longer the number to beat. On this branch the
unchanged pattern takes a median 2.6 s a move.

### Blackjack, by step

Click to the mover's screen: median (least to most), and how many.

| Step | Games | Deal to every screen | A move, all | 1 version | 3 versions | 6 versions | 16 versions |
|---|---|---|---|---|---|---|---|
| Baseline, `f76bde9b` | 7 | 5.5 s (5.2 to 6.8) | 2.6 s (0.6 to 15.2), n=60 | 1.4 s (1.1 to 2.7), n=10 | 2.7 s (1.6 to 4.6), n=16 | 5.4 s (4.8 to 6.0), n=4 | 13.0 s (12.8 to 15.2), n=4 |
| The port (steps 1 to 3), `9a6ac223` | 5 | 8.3 s (7.8 to 8.8) | 3.6 s (1.4 to 13.8), n=41 | 2.0 s (1.6 to 4.9), n=8 | 3.5 s (2.7 to 7.8), n=16 | 6.0 s (5.5 to 6.4), n=4 | 12.8 s (11.9 to 13.8), n=4 |
| 4(a) the last input's states, `1c848c95` | 2 | 8.6 s (8.1 to 9.0) | 2.9 s (1.2 to 12.5), n=21 | 2.2 s (1.2 to 3.9), n=9 | 3.0 s (2.6 to 3.3), n=8 | 5.5 s (5.1 to 5.9), n=2 | 12.3 s (12.1 to 12.5), n=2 |
| 4(c) no `identify`, `c8db1ab2` | 1 | 7.8 s | 3.1 s (1.6 to 11.7), n=8 | 2.4 s (1.6 to 3.2), n=2 | 2.9 s (2.6 to 4.0), n=4 | 5.0 s, n=1 | 11.7 s, n=1 |
| 4(a) the deal's head only, `c52d9224` | 2 | 1.2 s (1.2 to 1.3) | 2.5 s (1.2 to 11.9), n=17 | 1.3 s (1.2 to 1.6), n=5 | 2.8 s (2.5 to 4.3), n=8 | 5.4 s (5.2 to 5.6), n=2 | 11.8 s (11.8 to 11.9), n=2 |
| **Final**, `1803989f` (a view from one copy) | 3 | **1.2 s** (1.2 to 1.4) | **2.6 s** (1.1 to 12.1), n=27 | 2.1 s (1.1 to 2.8), n=9 | 2.6 s (1.9 to 4.3), n=12 | 4.8 s (4.5 to 5.3), n=3 | **9.6 s** (9.3 to 12.1), n=3 |
| *Experiment, not shipped: the head only, for every input* | 2 | 1.2 s (1.1 to 1.4) | 1.6 s (1.2 to 2.1), n=16 | 1.6 s (1.2 to 1.7), n=4 | 1.6 s (1.2 to 2.1), n=8 | 2.1 s, n=2 | 1.5 s, n=2 |

Twenty-two games, each with every secrecy check passing: 1,009 checks in
the seven baseline games and 2,121 in the fifteen after. (A sixteenth game
on the ported pattern, an earlier form of the experiment with the head only
in the view and every state still stored, passed 134 more; its moves took a
median 1.5 s and its round's end 3.5 s.) One more baseline game ran while
the `cf test` suites were running; its deal took 9.6 s, and it is left out
of the table. A game is eight to eleven moves, so the
columns by version count are small samples, and the one-version column of
the last two rows should be read as the same.

### Memory, by step

| Step | Games | Click to the face showing, mover | Proposal to the mover's view, each game | Until `send()` settles, each game |
|---|---|---|---|---|
| Baseline | 2 | 2.2 s (1.4 to 5.8), n=30 | 2.2 s (1.4 to 3.2); 2.4 s (1.5 to 5.7) | 1.8 s (1.0 to 2.9); 1.9 s (1.0 to 4.1) |
| The port | 1 | 2.9 s (1.5 to 3.4), n=16 | 3.0 s (1.4 to 3.4) | 1.9 s (0.6 to 3.4) |
| **Final** | 2 | **1.5 s** (1.0 to 2.5), n=34 | 1.4 s (1.0 to 2.4); 1.5 s (0.9 to 2.4) | 0.8 s (0.5 to 2.1); 0.7 s (0.4 to 2.0) |
| *Experiment: the head only* | 1 | 1.1 s (0.9 to 1.4), n=14 | 1.2 s (0.9 to 1.3) | 0.7 s (0.3 to 1.3) |

Memory was not run at the steps in between. In the final games the
hide-cards timer fired seven times, between 4.0 and 5.2 s after the
mismatch showed, by the client-driven `tick`.

### Bytes, by step

JSON sizes of what a pattern stores and releases, from `measure-bytes.mjs`,
for the same games. "Written" models the store's diff: every path whose
value differs from what was stored before. "Released" is every view computed
for a move, added up.

Blackjack, three players and a watcher:

| Step | Views | Deal: stored | Deal: released | A move, median: stored | written | released | Largest move: stored | released |
|---|---|---|---|---|---|---|---|---|
| Baseline | 8 | 113 KB | 596 KB | 48 KB | 33 KB | 170 KB | 180 KB | 868 KB |
| Port | 4 | 113 KB | 596 KB | 48 KB | 33 KB | 170 KB | 180 KB | 868 KB |
| 4(a) | 4 | 112 KB | 568 KB | 48 KB | 22 KB | 165 KB | 179 KB | 829 KB |
| Final | 4 | 21 KB | 52 KB | 48 KB | 22 KB | 165 KB | 179 KB | 829 KB |
| *Head only (experiment)* | 4 | 21 KB | 52 KB | 26 KB | 3 KB | 57 KB | 28 KB | 59 KB |

Memory, two players and a watcher:

| Step | Views | Deal: stored | Deal: released | A move, median: stored | written | released | Largest move: stored | released |
|---|---|---|---|---|---|---|---|---|
| Baseline | 1 | 20 KB | 22 KB | 60 KB | 39 KB | 44 KB | 84 KB | 67 KB |
| Port | 3 | 20 KB | 63 KB | 60 KB | 39 KB | 127 KB | 84 KB | 194 KB |
| 4(a) | 3 | 19 KB | 63 KB | 60 KB | 1 KB | 125 KB | 84 KB | 191 KB |
| Final | 3 | 19 KB | 62 KB | 60 KB | 1 KB | 125 KB | 84 KB | 191 KB |
| *Head only (experiment)* | 3 | 19 KB | 62 KB | 42 KB | 0 KB | 64 KB | 45 KB | 65 KB |

Three things these show that the clock does not:

- **The port by itself releases the same bytes at blackjack.** The baseline
  computed eight views, but the four for empty seats were `undefined`. What
  the port changes is who is sent them: at baseline every client's
  `cf-iframe` was handed all eight cells, and now each is handed its own.
- **The port triples what Memory releases.** Memory had one view for
  everyone; it now has one for each principal looking, each holding the same
  board.
- **Step 4(a) barely moves the released bytes** (170 to 165 KB). What it
  removed was bookkeeping. The versions themselves are what an animator
  plays, and they are still there.

### What each step bought

| Step | What it changed for a player | Why |
|---|---|---|
| **The port** | Slower: a move 2.6 s to 3.6 s, the deal 5.5 s to 8.3 s, Memory 2.2 s to 2.9 s | Each viewer's view is its own instance, computed after the commit. At baseline a one-version move's view arrived 0.1 s after `send()` settled; after the port, 0.7 s. The server's own delivery work fell about five times (see below), which did not show in what a player waits for |
| **4(a), the last input's states** | A move 3.6 s to 2.9 s. Memory's commit 1.9 s after the port to 0.8 s at the end; Memory was not run in between, and the bytes say this is the step | Not the bookkeeping. The stored states were a record keyed by version, which I had chosen earlier so that a held version would "keep its place"; a new version was then always a new key and a whole state written. As a plain list, a state sits where the last input's sat and the store's diff writes only what differs: Memory writes 1 KB a move where it wrote 39 KB, blackjack 22 KB where it wrote 33 KB |
| **4(c), no `identify`** | Nothing measurable, as expected | It removed an event each guest sent once at load, a per-user cell and the seats' principals from every view. One game; within the noise of 4(a) |
| **4(a), the deal's head only** | The deal 8.6 s to 1.2 s | A deal was twelve versions, stored and released to everyone, and no guest ever played them: a new deal cuts to its newest version. 568 KB released became 52 KB |
| **A view from one copy** | A round's end 11.8 s to 9.6 s (two and three games) | The lift read the stored table through the runtime many times over; it now reads it once. Measured directly in labs' harness: about 110 ms a view to about 80 ms |

### Where the time goes now

A move's time is two parts, and the browser script now reports them apart:
until `send()` resolves, which is the table's commit settling, and from
there until the mover's own view arrives.

Blackjack in the browser, by how many versions the proposal produced:
the median time until `send()` settles, then the median time from there to
the mover's view. "Then 0.0 s" means the view arrived first.

| | Proposals | 1 version: settles, then view | 3 versions | 6 versions | 16 versions | `sit` round trip |
|---|---|---|---|---|---|---|
| Baseline (one game) | 8 | 1.1 s, then 0.1 s | 1.8 s, then 1.1 s | 3.0 s, then 2.5 s | 7.4 s, then 7.7 s | 0.7 s (0.4 to 0.8) |
| The port (one game) | 10 | 1.4 s, then 0.7 s | 2.2 s, then 0.9 s | 1.9 s, then 3.7 s | 3.2 s, then 8.7 s | 0.4 s (0.2 to 0.8) |
| Final (three games) | 27 | 1.1 s, then 0.8 s | 2.0 s, then 0.0 s | 4.4 s, then 0.1 s | 3.0 s, then 6.6 s | 0.4 s (0.2 to 1.5) |
| *Experiment: the head only (two games)* | 16 | 0.7 s, then 0.8 s | 0.9 s, then 0.4 s | 0.8 s, then 1.3 s | 0.6 s, then 0.9 s | 0.4 s (0.2 to 0.7) |

The two parts are not the same two things before and after the port. At
baseline the views were shared values, so much of their computation fell
before `send()` settled; since the port they are per-viewer and fall after
it. The sum is what compares.

**Measured.** What the split says:

- **The commit follows the size of what is stored.** `sit`, which writes a
  few hundred bytes, round-trips in about 0.4 s. A move takes one to three
  seconds to settle with every version of the input stored, and 0.6 to
  0.9 s with only the head stored, however many versions it produced.
- **The views follow the number of versions.** A sixteen-version move's view
  arrives 6.6 s after its commit settles (about 9 s before the view was
  computed from one copy), and with only the head released, 0.9 s after. End to end,
  a move costs about 0.6 s for each version it produces, for four viewers.
- **The server's main thread is not where it goes.** The toolshed's
  `/api/health/stats`, read before and after a game, times the memory
  server and the executor's port. At baseline they added up to about 14 s a
  game (`memory/flush/refresh` 9.9 s, `send/encode` 5.6 s,
  `prepareLabels` 3.7 s, `commitWave` 3.7 s; spans nest, so these overlap).
  After the port, about 2.5 s (`flush/refresh` 1.5 s, `encode` 1.0 s,
  `commitWave` 0.9 s). One game each; every later game read like the
  second. The port cut the server's delivery work by about
  five times and the moves did not get faster, so the time is somewhere
  that endpoint does not see: the serving worker, or the browser.
- **It is not the display check.** In labs' harness the check of a 107 KB
  view of ten versions took about 60 ms, and of a 12 KB view 5 ms.
- **It is not the engine.** Blackjack's sandbox parity test replays 16
  versions and computes 64 viewer outputs in 139 ms inside a lift. Compiling
  the game, which every run repeats, is 0.6 ms under Node.
- **It is in the runtime, with no browser involved.** The display test's
  timing case (`TIMING=1`) runs one deal and one three-version move for two
  players and a watcher in one process, and prints the runtime's own timing
  statistics. On the final code:

  | | Clients compute | A serving runtime computes |
  |---|---|---|
  | The move: `send()` | 464 ms | 382 ms |
  | The move: from there to all three views | 591 ms | 1,207 ms |
  | The deal (one version held): `send()` | 290 ms | 257 ms |
  | The deal: from there to all three views | 248 ms | 808 ms |

  In the served move's 1.2 s, by the runtime's spans (they nest and
  overlap, so they do not add up): the view lift's own JavaScript 381 ms in
  five runs for three principals; `prepareCfc` 359 ms in 15 calls;
  `scheduler/run/commit` 237 ms; `traverse` 167 ms; `deriveFlowJoin`
  142 ms; and one client's `watchAdd/request` waiting 709 ms on the server.
  In the `send()`: the handler's own JavaScript is 23 ms, and `prepareCfc`
  about 200 ms in three calls.
- **Most of a lift's own time is reading the table, not computing.** A view
  computed from a plain copy of the stored table took about 80 ms where one
  computed from what the lift was handed took about 110 ms (three runs each
  way), and the engine's share of either is a few milliseconds. The copy is
  one walk of a 50 KB record through the runtime's view of the cell.
- **The port adds about two thirds of a second between the commit and the
  view.** At Memory, `send()` settled at a median 1.8 s before the port and
  1.9 s after, and the mover's view arrived 0.4 s later before and 1.0 s
  later after (14 and 16 moves). Blackjack's one-version moves show the
  same: 0.1 s behind the commit at baseline, 0.7 to 0.8 s after the port.

**Inferred, not measured:**

- That the extra time is the per-viewer instance being computed on each
  viewer's demand, in a later wave than the commit (labs
  `docs/specs/server-side-execution/scopes.md`: a node that discovers a
  `user` scope "runs once per demanding principal"). The five lift runs for
  three principals in the served breakdown fit that, and I did not trace
  the waves.
- That the serving worker's time is what the toolshed's stats endpoint is
  missing. The harness breakdown above is one process holding all four
  runtimes, not the toolshed's worker.
- How the same costs divide in the browser's own runtime worker. Nothing
  here instrumented the browser.

**What a pattern cannot remove,** then, on this branch: about half a second
of `prepareCfc` and commit for a handler that rewrites a 50 KB record; a
view computation per viewer that is mostly the cost of reading that record;
and a second, per-viewer round of the same when a serving runtime computes
the views. Each scales with the versions held. Holding fewer is the only
lever a pattern has, and the two places it could be pulled without losing
anything (the bookkeeping, and the deal) are pulled.

## Tests

All on the final code (`1803989f`), with
`EXPERIMENTAL_VIEWER_BOUND_RELEASE=true`. "Off" and "on" are
`EXPERIMENTAL_SERVER_EXECUTION=false` and `=true`.

| What | Command | Result |
|---|---|---|
| Pattern Node tests | `node --test fabric/patterns/memory/table.test.ts fabric/patterns/memory/parity.test.ts fabric/patterns/blackjack/parity.test.ts` | 39 of 39 |
| Memory parity in the sandbox | `cf test fabric/patterns/memory/parity.test.tsx` | 1 of 1, off and on |
| Blackjack parity in the sandbox | `cf test fabric/patterns/blackjack/parity.test.tsx` | 1 of 1, off and on |
| Blackjack, multi-user | `cf test fabric/patterns/blackjack/multi-user.test.tsx` | 92 of 92 off (35 s); 92 of 92 on (43 s) |
| Memory, multi-user | `cf test fabric/patterns/memory/multi-user.test.tsx` | 202 of 202 off (49 s); 202 of 202 on (69 s) |
| The display boundary and the attacks | `deno test -A test/boardgame-blackjack-display-fit.test.ts` in labs `runtime/html` | 4 tests, 13 steps: passed without a serving runtime (2 runs) and with one (2 runs). The timing case is skipped unless asked for |
| Guest type check | `npm run type-check` in `fabric/guest` | clean |
| Guest tests | `npm test` in `fabric/guest` | 28 of 28 |
| `cf check` | both `main.tsx`, both multi-user tests | clean |
| Blackjack in a browser | `play-three-users.mjs`, three players and an owner who watches | Three whole games on the final code: 10, 9 and 8 moves, the same result on all four screens, 424 of 424 secrecy checks, 745 to 1,002 animations a guest, no watchdog firing |
| Memory in a browser | `play-two-users.mjs` | Two whole games on the final code: 16 and 18 reveals, the same winner on both screens, seven mismatches hidden by the timer, 282 and 326 animations a guest, no watchdog firing |

The same four `cf test` suites passed on the port commit (`9a6ac223`), off
and on, after one fix to Memory's test: it worked out the other player's
next move from that player's legality, which a view no longer holds, and
now uses `legalForAnyone` from its own. Steps 4(a) and 4(c) were checked at
their own commits by the Node tests and `cf check` only.

What the multi-user suites can no longer check: one participant comparing
two principals' views. Each reads their own instance and has no way to read
another's. That a hole card one seat is shown appears in nobody else's view
of the same version is checked under Node over whole games, at the display
boundary in labs' harness, and in the browser script.

## The attacks

`blackjack/attacks/display-fit.test.ts` is labs' display-boundary test for
its card table, pointed at `blackjack/table.tsx` with the engine and toolkit
loaded as the program. Alice owns the space and holds no seat; Bob and Carol
play. Each is a runtime of their own on one memory server. It ran in my labs
worktree, without and with a serving runtime.

| Case | Outcome |
|---|---|
| A player reads their own instance of the view | Shown; it holds their own hole card and no other |
| The owner, with no seat, reads their own instance | Shown; it holds no hole card |
| Any principal's ceiling against another principal's instance (owner against each player, player against player, player against owner) | Refused |
| The sealed table, and `peek` (the secret salt), for all three | Refused |
| **The owner binds `move` from another module over the live table** | Makes the owner's own proposals and no others: a `sit` naming Bob as proposer seats the owner; the owner's `start` and `play` with no seat are refused. The witness holds: the players are still shown their views. The owner is shown no hole card |
| **The owner runs the module a second time over the live table** | Computes the owner's own watcher view again. It is shown to nobody, the owner included, and `peek` of it is refused. The live table is unchanged |
| Another module's handler writes the table, putting the owner in every seat | Refused at commit: "writeAuthorizedBy requires a trusted verified binding identity at /table". Seats and revision unchanged |

On the final code: 4 tests, 13 steps, all passing, twice without a serving
runtime and twice with one. With a serving runtime each run logs three
`StorageTransactionAborted` ("seal-after-park-refused") lines for a commit
after the serving tenure parked, and passes; I did not look into it.

In a browser, the watcher in every game above was an `OWNER` of the space.
It was handed no hole card and shown none face up, at every check of every
game.

## What is not held

These are not gaps in the pattern, and no pattern can close them. They
decide whether "the owner cannot read hands" is true.

1. **The official CLI hands the sealed table to any reader of the space.**
   Observed on `70f5c390ec`, against a finished game of the ported pattern:

   ```
   cf cell get --cell blackjack --input --no-run table/record/secretSalt
   ```

   printed the game's secret salt, from which the deck's order derives, for
   the key that owns the space **and for a fresh key granted only `READ`**.
   The viewer rule governs what a runtime shows at its display boundary;
   `cf cell get` has no such boundary, and the memory server delivers the
   document. This is the delivery non-conformance already ruled in specs#95,
   still open on this branch. Until it is closed, a seat's hole card is kept
   from a browser showing the piece and not from a person with `cf`.
2. **A space's owner may release a clause by intent.** `cf cell
   release-by-intent` (labs `docs/specs/cfc-persisted-declassification.md`,
   3.4; specification 8.12.7 route 2d) removes named confidentiality
   clauses at a stored path "whoever they stand for", on a request the
   space's owner signs. Run without `--clause` against the sealed table, it
   listed this module's policy clause as one it would release. **I did not
   run the release**: it needs a typed confirmation at a terminal. Read
   from the document, not observed: an owner who releases the clause on the
   table could then be shown it. This is a designed power, with a signed
   record. Whether it should reach a clause that stands for nobody the
   owner is, is a question for a ruling.
3. **A client runtime is trusted for its own user.** With server execution
   off, each client computes its own instance from a replica of the sealed
   table. The proposal says so itself.

Taken together: against an unmodified browser, at the display boundary, in
`cf test` and in labs' own harness, the owner is shown no hand, and the
three ways around the table that a pattern could be blamed for are closed.
Against the owner as a person with the CLI, nothing is hidden today.

## What is still in userland

| Machinery | Where | Why it is still here | What would let it go |
|---|---|---|---|
| **The versions of the last input, in every view** | `TableState.states`, `TableView.entries` | A guest plays one animation a version, as the Go client does, and a fix-up move is a version. A cell delivers its latest value, not a log, so a view carries every version its reader has not played | **Runtime:** a log a reader consumes in order, released per viewer, where an entry is written once and never rewritten. Then a view is the head and the log's new entries. **Or a decision:** play only the head (measured above as "head only"), and lose a card at a time |
| **The state before the last input** | `TableState.before` | The engine sanitizes a move against the state it was applied to | Nothing in the runtime; it is one state |
| **The move history** (step 4b) | `TableState.history` | Left. The engine takes the whole history: it reads its length as the version, its last record's name, and every record back to where the current phase began, which in a one-phase game such as Memory is all of it. It is 100 bytes a version: 4.7 KB of an 85 KB table at the end of the two-round blackjack game | **Not the runtime: our engine.** An engine API that takes a version number and the tail it needs, and for one-phase games a progression position kept in the state instead of derived from history. Not done here: the brief's scope was the patterns, and the measured share is 5% |
| **`tick`** (step 4d) | `Proposal.kind`, the bridge's `onTick` | Left. Nothing wakes a pattern at a time. A guest shown a timer sends `tick` when its own clock runs out, and the table fires the timer only if its own clock agrees | **Runtime:** a wake at an instant a pattern's state names. Labs `docs/plans/scheduled-work-in-the-server.md` ("Nothing wakes a piece on a timer") proposes a wake on a literal interval of minutes, which cannot serve a four-second timer. One `TODO(cfc)`, in `toolkit/view.ts` |
| **`revision`, and reading the view again after `send()`** | `TableView.revision`, the bridge's `propose` | A refused proposal writes nothing and says nothing. The bridge learns of a refusal by seeing that the table did not move on | **Runtime:** a handler's reply to its sender, or a refusal the sender's `send()` rejects with |
| **`epoch`** | `TableView.epoch` | A guest keys what it has shown by it, to tell a new deal from a newer version | Nothing; it is the game's identity |
| **Copies** | `tableOf` before a step; `toJson` over every view | The writer's stamp needs a final value with no reference in it, and a value returned as part of its input reaches the store as a reference to the sealed table (Memory FINDINGS F4). The card table copies too. **Ran:** with `tableOf` returning what it was given, the display test fails at its first step: no player is shown their view | **Runtime:** a `set` of a value read from the same cell storing it as a value. The view's copies (F4) were not rerun on this branch |
| **A JSON type for the engine's state** | `toolkit/json.ts` | A cell position typed `unknown` reads back as `{}` (F7) | **Runtime:** F7 was not rerun on this branch |
| **Compiling the game in every run** | `blackjackGame()`, `memoryGame()` | A pattern module holds only functions and plain data at module scope | Nothing needed: 0.6 ms a call under Node |
| **The lobby** | the bridge | Seating is a move; the guest draws the seats | The pattern's own JSX, once a value derived from a released view is not sealed again |

## The `TODO(cfc)` marks

| Where | Gap |
|---|---|
| `memory/table.tsx`, `blackjack/table.tsx` | The table starts as `null`: a record default is nobody's, and would keep every view sealed |
| `toolkit/view.ts` | `tick`: nothing wakes a pattern at a time |

Removed: the dropped input witness (both games) and the seat's audience
(blackjack).

Open on the labs branch and not marked, because nothing here works around
it: a value a pattern derives from the released view is sealed again, so
`main.tsx` hands the view whole to the `cf-iframe` and its `[NAME]` is a
constant.

## Not verified

- **The owner's release by intent**, with `--clause`. Listed, not run.
- **A modified client.** Nothing here ran one. The CLI read above is the
  nearest thing: an official client with no display boundary.
- **Why `2033f78fae` is so much faster.** The bisect names it; the reason
  above is read from its message.
- **The serving worker's and the browser's share of a move.** The toolshed
  endpoint does not report the worker, and nothing instrumented the browser.
  The breakdown is from labs' harness, one process, two players and a
  watcher.
- **Why a served run logs `seal-after-park-refused`** in the second-binding
  case.
- **F4, F6 and F7** of Memory's FINDINGS on this branch, and whether the
  views' `toJson` copies are still needed.
- **The lobby before the first move.** The table is `null` then, and a view
  of it carries no witness. I did not look at what a guest shows a person
  before anyone has sat; the bridge has a line for a withheld view and
  sitting does not depend on one.
- **A timer firing in a browser on the final code**, beyond the Memory game
  below.
- **Step 4(a) and 4(c) under `cf test`** at their own commits.
- **Go.** Not run; nothing in Go changed.
- **The declarative pattern** (`fabric/patterns-declarative`), which this
  breaks. See "For other sessions".

## For other sessions

- **`fabric/patterns-declarative/memory/*.tsx` no longer compiles.** It
  imports `seatView` and `EMPTY_TABLE` from `fabric/patterns/memory/table.tsx`
  and reads every seat's legality from one view. I did not edit it: it is
  another session's. It needs `viewFor({ table })`, a table that defaults to
  `null`, and the viewer's own legality from the head entry's
  `viewers[view.viewer]`.
- **`fabric/declarative/src/table-view.ts`** finds the viewer's seat by
  matching a principal against `seats[].principal`, which a view no longer
  holds. It should read `view.viewer`. Also not edited.
- **`fabric/guest/README.md`**, "The Common Fabric build", still names the
  `me` cell. Only `fabric-bridge.ts` and `PROTOCOL.md` were mine to edit.
- **`fabric/HANDOFF.md`** still describes the earlier shape and the 18.6 s
  figure. It is the orchestrator's file.

## Decisions for the owner

1. **Is a space's owner who can read the table with `cf` acceptable?** The
   viewer-bound release closes what a pattern could close. What is left is
   delivery (specs#95) and the owner's release by intent, both the
   runtime's, and the second by design. A table hosted in a space none of
   its players owns would sidestep the second and not the first.
2. **Per-step animation, or speed.** Every version a fix-up produces is a
   whole state in every viewer's view, and a round's end at blackjack is
   sixteen of them. Holding only the head makes that move 1.5 s where it
   takes 9.6 s, and the median move 1.6 s where it takes 2.6 s; Memory 1.1 s
   where it takes 1.5 s. The guest then animates straight from the old state to
   the new: cards that the Go client moves one at a time move together. I
   left it out, because it changes what a player sees. It is this, in
   `withVersions` in `toolkit/table.ts` (the parity tests compare every
   version, so they would keep the untrimmed step, as they do for the deal):

   ```ts
   states: [toJsonObject(head.state)],
   before: result.versions.length > 1
     ? toJsonObject(result.versions[result.versions.length - 2].state)
     : previous ?? null,
   ```
3. **Memory pays for the uniform shape.** Its seats all see the same board,
   so one shared view was enough; one a viewer triples what it releases and
   took a move from 2.2 s to 2.9 s before the later steps brought it to 1.5 s. A second rule could
   release a game's public view to the space and keep per-viewer views for
   games with hands. I did not add it: it is two shapes where there is one.
4. **The move history** wants an engine API to shrink. It is small today.
5. **Filing.** Four things here are worth a note to the Fabric team, and
   nothing was sent: that `2033f78fae` is a fourfold speed-up as well as a
   fix; the two thirds of a second a per-viewer view arrives behind its
   commit; what reading a 50 KB record costs a lift; and that `cf cell get`
   prints a sealed record for a `READ` key.
