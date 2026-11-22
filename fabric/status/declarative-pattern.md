# Memory in pattern JSX with boardgame elements: status

Stream D, step 3 of [`../PROPOSAL.md`](../PROPOSAL.md). Branch
`claude/fabric-d-declarative-99437f`, with `fabric-factoring` (`b3c609ce`)
merged in; rerun later the same day on `dced5490`, see "Rerun on the
integrated branch". Nothing pushed, no PR, `master` untouched. Written 2026-11-21.

Start with [`../patterns-declarative/FINDINGS.md`](../patterns-declarative/FINDINGS.md).
How to run it is in [`../patterns-declarative/README.md`](../patterns-declarative/README.md).

## Outcome

- **A two-identity game of Memory was played, twice, through a pattern whose
  own JSX is boardgame elements** in the shell's document: no `cf-iframe`.
  The sealed table, `move` and `seatView` are imported from
  `fabric/patterns/memory` unchanged.
- **The real reconciler does not defeat the approach**, but it behaves
  differently from the first experiment's simulation: it never recreated an
  element, and it does deliver props one at a time. One atomic input owned
  by the table element is what makes it work.
- **A pattern may not reshape the released view.** A `computed` or `.map`
  over it is withheld by the table's policy. So the element takes the view
  whole, and a variable number of seats cannot be written in JSX.
- **Labs has no admission mechanism to extend.** A tag is usable when the
  page defines it; nothing else gates it. The spike loads a script.

## Where things are

| What | Where |
|---|---|
| The pattern, JSX types, experiments, scripts, screenshots | `fabric/patterns-declarative/` |
| Element additions: `tableView`, `principal`, `timer-elapsed`, the one-file bundle | `fabric/declarative/src/table-view.ts`, `bg-table.ts`, `bundle.ts`, `vite.bundle.config.ts` |
| The labs change, as a patch | `fabric/patterns-declarative/labs-spike/0001-spike-shell-trusted-components.patch` |
| The labs change, as a branch | labs worktree `.claude/worktrees/boardgame-trusted-components`, local branch `spike/boardgame-trusted-components` at `1ef3d7ce99`, on `fix/served-handler-write-identity` (`4815052651`). Never pushed |

Not touched: `server/static/**`, `fabric/guest/**`, `fabric/patterns/**`,
`examples/**`, any Go.

## What ran, with what result

Node v24.22.3, Deno 2.9.8, headless Chromium (Playwright 1.58.2), labs'
local servers on port offset 460, started and stopped by this session.

| Claim | How | Result |
|---|---|---|
| A pattern can type an unlisted tag itself | `cf check` on a probe, then on every pattern here | An unlisted tag is a type error; with `elements.ts` imported it compiles, and a wrong prop is still an error |
| The shell loads the bundle and the elements upgrade | `tools/look.mjs` | `bg-table` defined and upgraded inside the piece; `/scripts/trusted-components.js` served, 409 kB |
| A complete two-identity game | `tools/play-two-users.mjs`, two fresh pieces | Both to the end. 16 reveals each; 3 mismatches each hidden by the timer; same winner on both screens. Output in `memory/screenshots/run.log` (second game) |
| Animations ran | The animator's `plays` counter and the table's own stats, read at the end | 58 plays, 0 watchdog firings, a player; 24 of 35 motion cycles started an animation |
| The reconciler recreated nothing | Connections counted by wrapping each element class as it is defined | 1 connection for the table and for each stack, both players, whole game |
| Versions are lost without a window | Assignments and installs counted | 24 `tableView` assignments carried 35 versions |
| Derived values are withheld | `tools/probe-release.mjs` | Path read shown; `computed` and `.map` replaced by "Content hidden by policy"; a stack read by path drew 10 cards |
| The per-stack shape draws nothing against the sealed table | `tools/probe-per-stack.mjs` | Two placeholders and an empty board; the script times out waiting for cards |
| Per-stack props do not animate even with public data | `tools/probe-reconciler.mjs`, two fresh pieces | 0 of 9 cycles animated; each stack's assignment arrived alone, about 3 ms from its partner |
| The first experiment still passes with the element changes | `npx playwright test` in `fabric/declarative` | 34 passed; `tsc --noEmit` clean |
| The labs change type-checks | `deno check` on the three files | Clean |

## Only reasoned, or not done

- **Citations in FINDINGS section 1.** The applicator, the type interface,
  keying, event serialization, provenance, the shell's entry and felt config
  were read directly. The absence of a CSP and of any document describing
  third-party components is from a search, not a proof.
- **Why derived values are withheld** was not traced into the runtime. The
  result is observed; the explanation ("another computation is not what
  `seatView` computed") is the reading consistent with the Memory pattern's
  F3 and F4.
- **Animation was not seen** in the first runs, only counted. The rerun
  below shows it: see "Rerun on the integrated branch". The script's sampler
  of running Web Animations still reads 0 and is still not diagnosed.
- Remembered geometry and table recreation were never exercised: nothing was
  recreated.
- Two tables on one labs page; more than two players; Blackjack; a watcher;
  a reload mid-game; a second deal; any browser but headless Chromium.
- The labs test suites were not run against the spike. It changes the shell
  only when an environment variable is set.
- A modified client. Legality is enforced by `move`; the elements only
  disable what the view says is not legal.

## Element changes needed from the client session

Two, both optional and both already asked for after the first experiment;
exact diffs are in FINDINGS section 4.

1. `boardgame-component-stack`: a `collection-id` that replaces the random
   id when given.
2. `boardgame-component-stack`: a `stackScope` the animator's list honors.

The third request of the first experiment (hand the animator a remembered
"before") is withdrawn for this path.

## How far this is from a trusted component set a user's policy admits

The elements are close: a pattern uses them directly, typed, and a sealed
game plays through them. Admission is not started. Missing in labs, none of
it attempted here: a record of what a user admits; loading by that record
with a digest; a namespace an admitted set is confined to; the runtime
refusing an unadmitted tag; a way for a set to declare which props fetch or
show text; and isolation, since an admitted element runs with the shell's
authority over the whole page. FINDINGS section 1 lists each with where it
would go.

## Rerun on the integrated branch (2026-11-21)

`fabric-factoring` at `dced5490` merged in (a fast-forward: it already held
this branch). Labs unchanged: worktree at `1ef3d7ce99` on
`spike/boardgame-trusted-components`, servers on offset 460, stopped
afterwards. The labs patch is byte-identical to the commit, so it was not
rebuilt. The shell reads the bundle by path from this repository, so there
is nothing to copy into labs: rebuilding the bundle and restarting the
servers is the whole re-export.

### What the merge broke, and the fix

All inside the two directories this stream owns. Nothing outside needed a
change.

| Broke | Cause | Fix |
|---|---|---|
| `cf check` on `memory/main.tsx`, `per-stack.tsx`, `probe-release.tsx` | They imported `memoryInfo` and `CardValues` from `fabric/patterns/memory/rules.ts`, which the engine work deleted | Import `memoryInfo` from `game.ts`; `info` is a plain `TableInfo` |
| The two probes' reads into the state | The table now releases its state as plain JSON, so `state.Game.Cards` has no type | `memory/state.ts` names the two paths they read. In `probe-release.tsx` the view is narrowed with one `as`; a cast inside the path expression stopped the transformer rewriting it, and `as unknown as` is refused by labs |
| The board size | A deal with no options is now the engine's default, 20 cards | The three patterns deal `{ numcards: "small" }`, the ten-card board every earlier number is for |

`fabric/declarative` needed nothing: the integration session had already
moved it onto `@boardgame/elements` and onto `collectionId` and
`stackScope`. It needs `npm ci` again after the merge, to link the package.

**`elements.ts` against the changed stack: no change.** It types the `bg-*`
wrappers, not `boardgame-component-stack`. `collection-id` and `stackScope`
are set by `bg-stack` and `bg-table` on the stack inside them and add no
prop a pattern writes. `cf check` passes on all four patterns with it.

### Commands

```bash
(cd fabric/declarative && npm ci && npm run type-check && npm run build:bundle && npx playwright test)
```

```bash
"$LABS/bin/cf" check fabric/patterns-declarative/memory/main.tsx fabric/patterns-declarative/memory/per-stack.tsx fabric/patterns-declarative/memory/probe-release.tsx fabric/patterns-declarative/reconciler/main.tsx --root "$PWD"
```

Then the README's steps 3 to 6, with a new space, the game deployed twice
(`memory-r1`, `memory-r2`) and each probe once (the reconciler probe twice).

### Counts

The "before" figures for game 1 are those of the game logged before the
merge (the second of that pair). The rerun's game 1 is in
`memory/screenshots/run.log`; its game 2 and the probes' output were read
and not kept, except the reconciler probe's, in `reconciler/screenshots/run.log`.

| | Before the merge | Rerun |
|---|---|---|
| First experiment's Playwright tests | 34 passed | 34 passed |
| Bundle | 409 kB | 409.11 kB, 83 modules |
| `cf check`, four patterns | clean | clean after the fixes above |
| Game 1: reveals, result on both screens | 16, Alice 6 to 4 | 16, Alice 6 to 4 |
| Game 1: animator plays, watchdog firings, each player | 58, 0 | 58, 0 |
| Game 1: motion cycles that started an animation | 24 of 35 | 24 of 35 |
| Game 1: `tableView` assignments, versions carried | 24, 35 | 24, 35 |
| Game 1: connections of the table and each stack | 1 each | 1 each |
| Game 1: mismatches hidden by the timer | 3 | 3 |
| Game 2 | 16 reveals | 14 reveals, Alice 6 to 4 on both screens; 54 plays, 0 watchdog; 21 of 30 cycles; 21 assignments; 1 connection each; 2 timer hides |
| Click to face, median (first move) | 2.2 s (15 s) | 3.4 s and 3.9 s (15 s) |
| Placeholders "hidden by policy" in the game | 0 | 0 |
| Release probe | path read shown; `computed` and `.map` hidden; 10 cards by path | the same |
| Per-stack probe against the sealed table | 2 placeholders, empty board, times out | the same |
| Reconciler probe, two pieces | 0 of 9 cycles animated; stacks set alone | 0 of 9 and 0 of 8; 12 of 13 batches held one stack's assignment, partners 2 to 3 ms apart |

The slower click-to-face is the one number that moved. The table now runs
the real engine in `move`, where it ran a hand-written Memory before. I did
not measure where the extra second goes.

### Animation, seen

`play-two-users.mjs` now records the first capture on the mover's screen two
ways at once: the browser's own screencast frames (over the DevTools
protocol), and the box of every card on every animation frame. A card counts
as in flight while it is between its first and last position.

| | Game 1 | Game 2 |
|---|---|---|
| Cards that flew (the pair) | 2 | 2 |
| Distance | 357 px and 616 px | 510 px and 616 px |
| Animation frames with the card strictly between its ends | 22 each | 23 each |
| Time across those frames | 317 ms | 306 ms |
| Screencast frames taken while a card was between a quarter and three quarters of the way | 7 | 7 |

The configured flight is 400 ms; the 300-odd ms is the part measured
strictly between the ends. Three of game 1's seven frames are saved as
`fabric/patterns-declarative/memory/screenshots/05-mid-flight-{1,2,3}-alice.png`.
I looked at the second: both tree cards are face up between the grid and
Alice's pile. Each card's whole path is in `screenshots/run.log` under
`flight`. They replace the eight timed `05-capturing-*` screenshots, none of
which had caught a card moving.

This is one capture a game on one screen, in headless Chromium. Flips and
the hide after a mismatch are still counted only: a flip does not move a
card's box.

## On the viewer-bound release (2026-11-21)

`fabric-factoring` at `ff70c37c` merged in (a fast-forward; it holds
`a0a66980`, the merge of the port). The pattern is on the per-viewer view
and two games were played on it.

- **Labs:** a new local branch `spike/boardgame-trusted-components-viewer-bound`
  at `87788f82c8`: the spike patch applied with `git am` on `70f5c390ec`. It
  applied unchanged, so `labs-spike/0001-…patch` is still the one patch. The
  worktree is `labs/.claude/worktrees/boardgame-trusted-components-vb`.
  Neither existing labs branch was committed to. Nothing pushed. Servers on
  offset 460 with `EXPERIMENTAL_VIEWER_BOUND_RELEASE=true`, stopped.

### What changed here

| File | Change |
|---|---|
| `fabric/declarative/src/table-view.ts` | A view has `viewer` and its seats have only a `label`. `seatOf(view)` reads `view.viewer` |
| `fabric/declarative/src/bg-table.ts` | The `principal` property is gone; the seat comes from the view |
| `fabric/patterns-declarative/elements.ts` | `principal` removed from `bg-table`'s props |
| `memory/main.tsx`, `per-stack.tsx`, `probe-release.tsx` | `viewFor({ table })`; the table defaults to `null`; no `me` cell, and `sit` no longer writes one |
| `memory/probe-release.tsx` | Also shows `view.viewer` by path, and the view through a module-scope `lift` |
| `fabric/guest/README.md` | "The Common Fabric build" no longer names a `me` cell |

The pattern lost its only per-user identity cell. It now holds a name field
and nothing else of its own.

### Is a `computed` over the view shown to its viewer? No.

**Ran:** `memory/probe-release.tsx` with `tools/probe-release.mjs`, two
identities seated at one table, read from each one's screen after the deal.

| Read of the view | Seat 0's screen | Seat 1's screen |
|---|---|---|
| `{view?.revision}` | `3` | `3` |
| `{view?.viewer}` | `0` | `1` |
| `computed(() => view?.revision ?? -1)` | Content hidden by policy | Content hidden by policy |
| a module-scope `lift` of the same | Content hidden by policy | Content hidden by policy |
| `view?.seats.map(seat => …)` | Content hidden by policy | Content hidden by policy |
| `<bg-stack stack={…state.Game.Cards}>`, by path | 10 cards | 10 cards |

So the limit stands exactly as before the port: a fixed-path read of the
released view is shown, each viewer's own; anything a pattern computes from
it is withheld from the very viewer it was released to. `per-stack.tsx`,
which picks the newest entry in a `computed`, still draws two placeholders
and an empty board. The viewer-intrinsic rule changed who a view is for,
not what may be done with it. I did not trace why in the runtime; the port
session reports the same from the guest's side.

### What ran

| | Result |
|---|---|
| First experiment's Playwright tests | 34 passed |
| Bundle | 408.91 kB |
| `cf check`, four patterns, with the flag | clean |
| Game 1 | 16 reveals; Bob 6 to 4 on both screens; 3 timer hides |
| Game 2 | 16 reveals; Alice 6 to 4 on both screens; 3 timer hides |
| Animator plays, watchdog firings, each player | 58, 0 (both games) |
| Motion cycles that started an animation | 24 of 35 (both games) |
| `tableView` assignments | 22 (game 1), 24 (game 2), for 35 versions |
| Connections of the table and each stack | 1 each |
| Placeholders "hidden by policy" in the game | 0 |
| Each player's own seat, from `view.viewer` | found by both in both games: the script waits for it and each then moves only on their own turn |
| Capture seen in flight | Game 1: 18 animation frames, 307 ms, 510 px and 616 px. Game 2: 21 frames, 302 ms |
| Click to face, median (first move) | 1.6 s and 1.5 s (15 s) |

Click to face was 3.4 s and 3.9 s on the earlier labs. The port session's
status attributes most of that to labs itself having become faster; I did
not separate the causes. Game 1's output and screenshots are in
`memory/screenshots/`; I looked at none of this run's screenshots, only the
counts.

The pattern has no `cf test` file; its tests are these browser scripts.

### Not run

- The reconciler probe. It reads nothing from a table and nothing in it
  changed.
- A watcher (someone with no seat), a third player, or the table's owner
  looking on.
- Whether a view can still skip a revision. A view now carries the versions
  of the last input only, where it carried a window of recent ones. 22
  assignments delivered 35 versions with none missed in these two games,
  but a reader that missed a whole revision would now miss its versions.

## Decisions for the owner

1. **Is "trusted" meant to be full trust?** An admitted element can read
   every other piece on the page and call the runtime. If that is acceptable
   for a set a user chose, admission is mostly bookkeeping. If not, the set
   needs isolation, and that is a different project from this one.
2. **Should a table's policy let a pattern reshape its view?** Today the
   pattern can only pass it through, which pushes version choice, seat
   repetition and every derived label into elements. The alternative is for
   the table to bless more lifts. This decides how much the element set must
   grow (a repeater, text from state, a lobby).
3. **Should `bg-table` know the toolkit's `TableView`?** It does now, in
   `table-view.ts`, restated structurally. That couples the element set to
   one wire shape. The other choice is a second element that adapts.
4. **Where do a set's types and its policy declarations live?** Types
   travel with the set today, with no labs change. Nothing ties them to what
   a viewer actually loaded.
5. **Should labs enforce the tag list at all?** Today the closed type is the
   only check, and a pattern can widen it itself.

## Blockers

None. No account, login or system setting was needed.
