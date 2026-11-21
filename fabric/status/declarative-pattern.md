# Memory in pattern JSX with boardgame elements: status

Stream D, step 3 of [`../PROPOSAL.md`](../PROPOSAL.md). Branch
`claude/fabric-d-declarative-99437f`, with `fabric-factoring` (`b3c609ce`)
merged in. Nothing pushed, no PR, `master` untouched. Written 2026-11-21.

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
- **Animation was not seen**, only counted. No screenshot caught a card
  mid-flight, and the script's sampler of running Web Animations read 0 and
  was not diagnosed.
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
