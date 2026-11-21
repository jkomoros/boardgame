# Findings: a board driven by layer-1 elements alone, from data

Stream D, step 2 of [the proposal](../PROPOSAL.md). Date: 2026-11-21.

## Verdict

**The ultimate goal is workable for Memory and Blackjack.** Both boards run
with no renderer class: plain markup, trusted elements, and one JSON property
per state. Cards flip, fly to won-cards stacks, and deal into fanned hands
using the existing animator, unmodified.

Three facts carry that verdict:

- **Every callback and hook the two games' renderers use has a data form.**
  Of 27 rows in the table below, 18 are clean, 5 are expressible but awkward,
  4 were not built (they are already data-shaped, but untested), and none is
  inexpressible.
- **Animation survives element recreation**, in all three modes tried,
  including recreating the animation root itself. It does not survive with
  the animator as it is today: every transition becomes a jump cut.
- **Two tables animate independently on one page**, once the animator's
  module-global stack list is narrowed to a root. Without that, one table
  cuts the other's animation short, and two tables showing the same game do
  not animate at all.

What this does **not** establish:

- Step D3, the real test. Nothing here crossed a worker boundary or ran
  under Common Fabric's renderer. Page B simulates recreation in one
  synchronous task; a real renderer may remove and insert in separate tasks.
- Effects, cross-screen transfers, motion release, timers, dice, tokens, and
  the `board` and `spatial` layouts were not exercised.
- Games with multi-step input (selection and placement drafts). Memory and
  Blackjack only ever need "click one thing" or "press a button".

## What was built

All under `fabric/declarative/`. Nothing in `server/static/src` or
`examples/` was edited; client source is imported in place.

| File | What it is |
|---|---|
| `src/bg-table.ts` | The animation root. Owns one `boardgame-component-animator`, the state queue, and the ambient registry |
| `src/bg-stack.ts` | Wrapper over `boardgame-component-stack`: `face` and `pick` as data, `propose-move` out |
| `src/bg-extras.ts` | `bg-callout`, `bg-turn-status`, `bg-action`, `bg-seat`, `bg-value` |
| `src/spec.ts`, `src/expand.ts` | The vocabulary and wire-stack expansion. Pure, no DOM |
| `memory.html` | Page A |
| `memory-recreate.html` | Page B: `mode=all`, `mode=changed`, `mode=table`; `measure=live` turns the fix off |
| `two-tables.html` | Page C: `scope=document` turns scoping off; `same=1` gives both tables one replay |
| `blackjack.html` | Blackjack, animated, `viewer=0` or `viewer=1` |
| `replays/cmd/export` | Go command that writes per-version replays from the real engine |
| `tests/` | 34 Playwright tests on port 5199 |

The wrappers are 1,419 lines. The existing code they import is 48 files and
20.0k lines.

**To run:**

```bash
cd fabric/declarative && npm install && npm run dev
```

```bash
cd fabric/declarative && npx playwright test
```

Then open `http://localhost:5197/memory.html`.

**A note on the replay.** The checked-in
`examples/memory/client/scenario-replay.json` records four settled states and
never captures a pair, so it cannot show a card flying to a won-cards stack.
Page A plays it with `?replay=scenario`. By default the pages play
`replays/memory-a.json`, which the exporter here generates from the real
Memory engine in the same shape, with one frame for every version.

## Every callback, hook and imperative call, as data

Verdicts: **clean** (expressible as data with no loss), **awkward**
(expressible, at a cost named in the row), **not built** (already a plain
data shape in the source, but not wired or tested here).

Line numbers are in `examples/memory/client/boardgame-render-game-memory.ts`
(M), `examples/blackjack/client/boardgame-render-game-blackjack.ts` (B),
`-blackjack-table.ts` (BT) and `-blackjack-hand.ts` (BH).

| # | What the renderer does | Where | Declarative replacement | Verdict |
|---|---|---|---|---|
| 1 | Subclasses `GameRenderer` and registers with a decorator | M:9-10, B:9-10, BT:20, BH:15 | No class. Markup under `<bg-table>` | clean |
| 2 | `cardView({ render })` returning Lit `html` for a face | M:11-15 | `face="Type"`: a field name. The stack renders it as text | clean |
| 3 | `cardView({ properties })` mapping values to `suit`, `rank` | B:11-16, BT:22, BH:17 | `face='{"props":{"suit":"Suit","rank":{"field":"Rank","map":{"Ace":"A"}}}}'`, against an allow-list of host properties | clean |
| 4 | `.withProperties({ rotated: true })` | B:77, BT:101-119, BH:77-80 | A constant in the same `props`: `"rotated": true` | clean |
| 5 | Branching on `kind === 'visible'` | M:12, B:13-14 | Built in. A hidden slot has no fields, so no source resolves | clean |
| 6 | `static styles` reaching `boardgame-card > div` | M:24-30 | Inherited custom properties (`--bg-face-font-size`) and `::part(face-text)` | awkward: a pattern needs a way to set CSS variables, and `::part` lets it restyle a face |
| 7 | `this.move(name).targets(slots, slot => args)` bound to `.action` | M:98-100, 110 | `pick='{"move":"Reveal Card","arg":"CardIndex","slots":"hidden"}'`; the stack emits `propose-move` | clean |
| 8 | Per-slot legality inside the bound action (`canActivate`) | M:98 via `moves/target-action.ts` | The frame's `choices`, which is the JSON the server already projects: `{MoveName, FieldName, Candidates:[{Value, Available}]}` | clean |
| 9 | `this.move(name)` bound to `boardgame-action-button` | M:144, B:57-58, BH:84-85 | `<bg-action move="Hide Cards">`; enabled from the frame's `legality` | clean |
| 10 | `post-animation-delay` computed from state by `_revealHoldMs()` | M:59-68, 107 | `hold-before='{"Capture Cards": 600}'` on the table, keyed by the upcoming move | awkward: needs the move name in the frame, and only delays a version that is already known |
| 11 | `effectsForTransition()` returning `fx.pulse(...)` | M:70-93 | `transition.effects` in the frame. `EffectSpec` is plain data apart from `HTMLElement` anchors | not built |
| 12 | `diffVisibleComponents`, `isVisibleComponent` to decide the effect | M:62, 75-82 | Stays in the pattern's worker, which has both states and full JavaScript | clean |
| 13 | `boardgame-fading-text .trigger=${...}` | M:112-115, B:80-81 | `<bg-callout message="Match" path="Game.Cards.$count">` | awkward: see "callouts" below |
| 14 | `this.playerPresentation(i)` to a player badge | M:119, 131 | Not wrapped. The badge takes a plain object | not built |
| 15 | `this.turnStatus` to `boardgame-turn-status` | M:156, B:88 | `<bg-turn-status>`; the table supplies the turn fields | awkward: `animating` is a timing fact the element must own, so it is fixed to false here |
| 16 | `boardgame-timer .timer=${state...}` | M:148-153 | Not wrapped. Needs a clock inside the element | not built |
| 17 | `repeat(state.Players, ...)` for seats | B:69, BT:105 | The pattern writes one seat per player; the markup lists them | clean |
| 18 | `classMap({ eliminated })`, `.active=${i === current}` | B:71-73, BT:106 | `<bg-seat index="1" flags="Eliminated Stood">` reflects `[active]`, `[eliminated]` | clean |
| 19 | `boardgame-stat .value=${p.Score}` | BT:113, player-info | `<bg-value path="Computed.Players.0.HandValue">` | clean |
| 20 | `_calculateStatus()` choosing "Busted" or "Stood" | player-info:7-17 | Reflected attributes plus CSS, or computed in the worker | clean |
| 21 | `game-outcome` fed `gameFinished`, `gameWinners`, `animating` | B:41-47 | Not wrapped. All plain values except `animating` | awkward: same as row 15 |
| 22 | `p.Eliminated && !this.animating` | BT:123 | The table already withholds a state until motion settles, so the guard is unnecessary | clean |
| 23 | `animationLength(fromMove, toMove)` hook | base renderer:533 | `animation-length="400"` and `animation-lengths='{"Collect Cards":700}'` | clean |
| 24 | `motionCohortsForTransition()` | base renderer:312 | `transition.cohorts`, passed to the animator unchanged. Tested: a JSON cohort delays the second card by 150 ms | clean |
| 25 | `motionTransfersForTransition()`, `motionReleaseForTransition()`, `animationOverlap()` | base renderer:323, 334, 543 | The specs are plain data. Transfers name endpoints by DOM id, which a pattern cannot rely on | not built |
| 26 | `renderFakeDeckRow()`, `renderTopEdgeAnchor()`, `renderAvatarStrip()`, `renderHostControls()`, `renderRoomCodeBanner()` | BT:95-98, 127, BH:72-73 | Companion-surface chrome. Out of scope for a board; the pattern's runtime owns seats and rooms | clean (dropped) |
| 27 | `<boardgame-card spacer>` to hold a row's height | M:129 | `min-height` in page CSS | clean |

Totals: 18 clean, 5 awkward (rows 6, 10, 13, 15, 21), 4 not built (rows 11,
14, 16, 25), none inexpressible.

### What "awkward" turned out to mean

The awkward rows share one cause: **a value derived from state that a
renderer would compute and bind per state.** A pattern can compute anything
in its worker, so the value is never the problem. The problem is that
binding it changes a child's property, and a changed property means a
recreated element.

The way out, used throughout, is that **elements bind to the state by path
and the table holds the state.** Then a child's attributes are static
strings: `path="Players.0.WonCards"` never changes, so the child is never
recreated. That is why `resolvePath` has `.$count` and `.$length`: so a count
is not a reason to pass a changing number.

## Recreation: what survived and what did not

Page B's reconciler describes the board as plain objects and destroys and
recreates elements. Three modes:

| Mode | What is thrown away each state |
|---|---|
| `all` | Every `bg-stack` and `bg-callout`, so every card node (tested: zero card nodes survive) |
| `changed` | Only elements whose attributes or JSON properties changed |
| `table` | The whole `bg-table` and its subtree |

**With the animator as it is today (`measure=live`): nothing animates.** In
all three modes the captured pair ends in the right stack, but it never
occupies a point between its old and new position. The animator starts zero
animations.

**With the table's remembered geometry (default): everything animates.** The
flip plays, the pair flies, the "Match" callout plays, and no painted frame
shows a card at its destination before it starts moving.

### Why it broke, precisely

The surprise is that **the animator already pairs components by id, not by
DOM node.** `prepare()` builds `_infoById` keyed by `component.id`
(`boardgame-component-animator.ts:485-533`), and `_doAnimate` looks each
current node up by id (`:1333`). Node continuity was never the mechanism.

What breaks is **when "before" can be measured.** `prepare()` reads the live
DOM. In the webapp, `render-game` calls it immediately before it assigns the
new state, so the old layout is still there. A trusted element under a
pattern renderer is not in that path: by the time any of its code runs (a
`disconnectedCallback`), the old nodes are already detached and measure as
zero. There is no moment at which the old layout can still be read.

Two smaller things also depend on node continuity:

- **Stack ids are random per element instance**
  (`boardgame-component-stack.ts:534`, `:1699`). The animator keys
  collection geometry and history by them (`:469-476`, `:1460`), so a
  recreated stack is a different collection and "came from that stack"
  cannot be resolved.
- **`boardgame-fading-text` decides whether to play by comparing against a
  private previous trigger** (`boardgame-fading-text.ts:56`, `:122-123`). A
  recreated element has no previous value and never plays.

### What it took

One mechanism, in `bg-table.ts`:

1. **Remember where everything rests.** `_measure()` records, for every
   identified component, the same fields `prepare()` records. It runs once a
   cycle, in the microtask after the animator has measured the new layout
   and before playback starts, when everything is at rest. That snapshot is
   the next cycle's "before".
2. **Hand it to the animator.** When the first sign of change in a task is a
   structural one (a stack connecting or disconnecting), the table calls
   `prepare()` to open the cycle and then overwrites what `prepare()` just
   measured with the remembered snapshot (`_restore()`).
3. **Keep live measurement when it is possible.** If the first sign is a
   property being set, nothing has been removed yet, so the table calls
   `prepare()` then and there, exactly as the webapp does.

It also needed:

- **A declared stack identity.** `bg-stack` overrides its inner stack's `id`
  getter with `name` (or `path`).
- **A per-name memory on the table**, used by `bg-callout` to recall its
  previous trigger.
- **A module-level map keyed by the table's `name`**, so the snapshot
  outlives the table element in `mode=table`.

**The cost is that it reaches into the animator.** It writes four private
fields (`_infoById`, `_beforeSeenIds`, `_beforeCollectionOffsets`,
`_beforeCollectionViewportOffsets`), reads a fifth
(`_historicalPresentationById`), and repeats about 25 lines of `prepare()`'s
measurement. That is acceptable in an experiment and not in a trusted set.
It is the first of the three changes recommended at the end.

### Limits found

- **A state arriving mid-flight in recreate mode animates from the previous
  target, not from where the card visibly was.** The remembered snapshot is
  resting geometry. The in-place path has the same behaviour today, because
  `prepare()` force-finishes running animations before measuring.
- **Remembered geometry goes stale on reflow.** The table re-measures on its
  own resize. A reflow that does not resize the table would not be caught.
- **If the animation root itself is recreated, the queue is lost.** The
  snapshot survives by name; the queue of pending versions does not. In
  `mode=table` a state arriving mid-flight replaces the table and jump-cuts
  whatever was unfinished.

## Two tables on one page

### What breaks today (`scope=document`)

Each is pinned by a test in `tests/findings.spec.ts`.

| Symptom | Cause |
|---|---|
| One table installing a state **cuts the other table's animation short**. Two cards 200 ms into a 600 ms flight are finished instantly | `prepare()` calls `finishGatedAnimations()` on every component of every stack in the document (`boardgame-component-animator.ts:456-463`) |
| When both tables change together, **each animator also plays the other's cards** | Every loop in the animator iterates `sharedStackList`, a module-global array (`boardgame-component-stack.ts:48`, `:494`) |
| Two tables showing **the same game do not animate at all**: zero animated cycles on both | Each component id is sighted twice, and continuity treats that as unresolvable (`motion/continuity.ts:74-76`) |

A fourth problem did not appear only because of how the wrappers are built.
`BoardgameComponent` reflects the component id to the DOM `id` attribute
(`boardgame-component.ts:234-235`), so two light-DOM tables showing the same
game would put duplicate ids in the document. Here every card sits inside a
`bg-stack` shadow root, which scopes ids. `document.getElementById` in the
animator (`:552`) is used only by `fly()` and `animateBetween()`, which
these boards never call.

### What I had to do

**One property override.** The animator reaches the global list through its
inner stack's `_sharedStackList` getter. `bg-table` redefines that getter on
that one instance to return only stacks whose nearest `bg-table` ancestor is
itself (`bg-table.ts`, `_becomeReady` and `_stacks`). Every loop in the
animator then sees one table's stacks.

With that, the same replay on both tables animates correctly on both, one
table's install leaves the other's running animations untouched, and
`propose-move` is delivered only through the table that was clicked.

The ambient lookups needed no change. `animatableRegistry` and
`animationContext` are found by walking ancestors, so providing them on
`bg-table` is enough.

## Events that would leak layout or timing

Every custom event in the import closure bubbles and is composed, so every
one reaches the page. `bg-table` stops the first five at its boundary; a
test asserts none reaches `document`.

| Event | Source | Detail | What it leaks |
|---|---|---|---|
| `will-animate` | `boardgame-animatable-item.ts:389` | `{ ele, expectedSettleMs }` | A live element reference, and how long the animation will run |
| `animation-done` | `boardgame-animatable-item.ts:448` | `{ ele }` | A live element reference, and the exact moment an animation ended: a clock finer than the runtime's |
| `stack-slot-geometry-changed` | `boardgame-component-stack.ts:1404` | `{ geometry: { componentInlineSize, componentBlockSize, inlineSize, blockSize } }` | Measured pixel sizes, which vary with font, zoom and viewport |
| `component-tapped` | `boardgame-component.ts:570` | `{ index }` | Nothing sensitive, but it is synthesized, so it cannot tell a person's click from a scripted one |
| `motion-cycle-release` | `boardgame-component-animator.ts:1938` | `{ cycleId, key, reason }` | The moment an animation crossed a progress threshold |
| `roll-start`, `roll-end` | `boardgame-die.ts:1598`, `:1631` | none, then a roll detail | When a die tumble began and ended. Not reached by these boards |
| projected-choice consumption changed | `moves/projected-choice-consumption.ts:85` | `{ controller }` | A live object with methods. Not reached: the wrapper never sets `projectedChoices` |

Three more channels are not events:

- **`observeStructuralMotionEvents(fn)`** on the animator delivers every
  component's viewport rectangle and executed timing. It takes a function,
  so a pattern cannot call it, but anything on the main thread can.
- **`window.__bgAnimTestHooks`** is installed at import
  (`utils/anim-test-hooks.ts:53`). Its log holds `performance.now()`
  timestamps and element ids for every play and settle.
- **A native `click`** still bubbles out of a stack with coordinates,
  including clicks on disabled cards.

The stopping is done with `stopPropagation` in the bubble phase, which a
capture-phase listener on `window` sees first. A pattern cannot install one,
but the right fix is for these events not to be composed at all.

## Every global the elements touch

Over the 48 source files the wrappers import. No file in the closure uses
`fetch`, `XMLHttpRequest`, `WebSocket`, storage, `navigator`, `location`,
`history`, `crypto`, or `import.meta`.

| Global | Where (`server/static/src/...`) |
|---|---|
| `window.matchMedia('(prefers-reduced-motion: reduce)')` | `components/boardgame-animatable-item.ts:344`; `boardgame-component-animator.ts:947`, `:1180`; `boardgame-die.ts:1407`; `boardgame-token.ts:621` |
| `window.__bgAnimTestHooks` (written at import) | `utils/anim-test-hooks.ts:53` |
| `window.requestAnimationFrame`, `window.setTimeout` fallback | `components/boardgame-component-animator.ts:1666`, `:1668` |
| `requestAnimationFrame` | `components/boardgame-animatable-item.ts:411`, `:413`; `boardgame-component-animator.ts:615`, `:617`; `motion/activation.ts:61`; `motion/release.ts:142` |
| `cancelAnimationFrame` | `components/boardgame-animatable-item.ts:437`; `motion/activation.ts:48`, `:55`; `motion/release.ts:123` |
| `document.getElementById` | `components/boardgame-component-animator.ts:552` |
| `document.createElement` | `components/component-view.ts:126`, `:139`, `:158` |
| `getComputedStyle` | `components/boardgame-animatable-item.ts:166`; `boardgame-card.ts:737`, `:739`, `:742`, `:746`; `boardgame-component-animator.ts:272`, `:501`, `:591`, `:919`, `:1117`; `boardgame-component-stack.ts:1369`; `boardgame-component.ts:334`, `:363`; `boardgame-die.ts:1508`, `:1699`; `boardgame-token.ts:611`; `moves/projected-choice-consumption.ts:96` |
| `getBoundingClientRect` | `components/boardgame-component-animator.ts:278`, `:1037`, `:1038`; `motion/geometry.ts:34` |
| `offsetParent`, `offsetTop`, `offsetWidth` | `motion/geometry.ts:54-62`; `components/boardgame-component-stack.ts:1370` |
| `Element.animate` | `components/boardgame-animatable-item.ts:370`; `boardgame-component-animator.ts:1007`, `:1081`, `:1189` |
| `queueMicrotask` | `components/boardgame-animatable-item.ts:145`; `boardgame-component-stack.ts:499`, `:1361`; `boardgame-die.ts:1192`; `moves/projected-choice-consumption.ts:65` |
| `ResizeObserver` | `components/boardgame-component-stack.ts:1344`; `moves/projected-choice-consumption.ts:55` |
| `MutationObserver` | `moves/projected-choice-consumption.ts:59` |
| `Date.now` | `components/boardgame-component-animator.ts:1044`; `motion/timing.ts:42`, `:117` |
| `performance.now` | `utils/anim-test-hooks.ts:34` |
| `Math.random` | `components/boardgame-component-stack.ts:533`, `:1704` |
| `CSS.escape` | `components/boardgame-component-animator.ts:561`, `:576` |
| `console.error`, `console.warn` | 23 sites: `boardgame-animatable-item.ts:274`, `:426`; `boardgame-component-animator.ts` (14 sites from `:244` to `:1812`); `boardgame-die.ts` (5 sites); `moves/action.ts:932`; `moves/target-action.ts:367` |
| `customElements.define`, `.get` (registration at import) | every element file; `components/component-view.ts:295` |

**Network, indirectly.** Card and token art is a URL the element fetches.
`components/component-art.ts:60` builds `url("…")` for `art` and `backArt`,
and `components/boardgame-token.ts:503` falls back to a document-relative
`src/assets/token_<type>.svg` (seam S5). A trusted element fetching a URL a
pattern chose is an exfiltration channel. `spec.ts` therefore allows only
`data:image/` and relative paths in `art` and `backArt`.

The wrappers add `ResizeObserver`, `setTimeout` (the hold) and
`queueMicrotask` in `bg-table.ts`.

## Other findings

1. **Per-version input must be one atomic property.** A view, its legality
   and its hints belong to one version. A renderer may set separate
   properties in any order, and a view paired with the previous version's
   legality enables the wrong cards. `bg-table` takes one `frame`.

2. **Legality is already data.** The server's `projectedMoveChoices` is
   exactly "this move, this argument, these values". Passing it through the
   frame replaced all of `moves/target-action.ts` for these two games.

3. **A click cannot be tied to what was on screen unless the table holds
   the state.** In page B's per-stack mode the table has no version, and the
   first run of the click test got `version: null`. `bg-stack` gained a
   `version` property for that mode. The path-bound design does not need it.

4. **`component-tapped` discards whether a person clicked.** The component
   dispatches it from its own click handler. `bg-stack` records
   `event.isTrusted` from the native click in the capture phase and proposes
   nothing for a scripted one. Tested.

5. **The stack writes onto its input** (seam S10):
   `boardgame-component-stack.ts:1157-1158` and `:1221-1222` assign
   `item.ID` when a slot lacks one. `expand.ts` emits an opaque slot as
   `{ ID }` so the write never happens, and deep-freezes every expanded
   stack. No test failed on a frozen stack.

6. **The stack drags in code these boards never run.** `component-view.ts`
   imports the token and the die (`:9-11`), which brings the dice simulator.
   The stack imports four `moves/` files (`:5`, `:16-23`) that the wrapper
   never exercises, because `pick` replaces them. Together that is roughly
   half of the 20.0k-line closure.

7. **Memory's "Match" callout probably never plays in the webapp today.** Its
   trigger is `Cards.Components.length` (M:114). `Cards` is a fixed-size
   stack, and the exported replays show the length is 20 at every version,
   so the trigger never changes. This is inferred from the data; the webapp
   was not run to confirm it. The declarative version binds
   `Game.Cards.$count` and does play. Reported, not fixed: `examples/` is
   out of scope.

8. **Geometry held up where it was tested.** Flights start and end in the
   right place two shadow roots deep, under a `scale(0.7)` ancestor, and in a
   scrolled container. Rotated or 3D ancestors, and the viewport-based math
   in `fly()` and transfers, were not tested (seam S13).

9. **Measuring "was it a jump cut" needs care.** Sampling positions from a
   `requestAnimationFrame` callback reads the layout before the animator's
   own callback applies the inversion, and reported a false flash. The tests
   sample from a `ResizeObserver` callback, which runs after every rAF
   callback and before paint.

## Recommended declarative API

Names are placeholders, as in the proposal.

### `<bg-table>`: the animation root and state queue

| Input | Type | Meaning |
|---|---|---|
| `frame` | JSON property | `{ view, choices?, legality?, transition? }`. The only per-state input. Assigning it queues that version |
| `chest` | JSON property | The game's component catalogue. Set once |
| `game-name` | attribute | |
| `name` | attribute | Stable identity. Needed if the table itself may be recreated |
| `viewer` | attribute | Seat index, or -1 |
| `animation-length` | attribute | Milliseconds |
| `animation-lengths` | JSON attribute | Milliseconds by move name |
| `hold-before` | JSON attribute | Milliseconds to hold before a version produced by the named move |

`transition` is `{ move?, holdMs?, animationLengthMs?, cohorts?, effects?, release? }`.
The last two are recommended, not built.

Events out: **`propose-move` only.** No settle, layout or animation event.

The experiment switches `measure` and `scope` and the `whenSettled()` method
exist for tests and are not part of the API.

### `<bg-stack>`

| Input | Type | Meaning |
|---|---|---|
| `path` | attribute | Where this stack is in the state: `Players.0.WonCards` |
| `name` | attribute | Stable identity within the table. Defaults to `path` |
| `layout`, `messy`, `messiness`, `faux-components`, `stagger`, `no-default-spacer` | attributes | As on the existing stack |
| `component` | attribute | `card` or `token` |
| `face` | attribute or JSON | A field name, `prop:Field` pairs, or `{ text?, regions?, props? }` |
| `pick` | attribute or JSON | `Move Name:ArgName`, or `{ move, arg?, value?, args?, slots?, candidates?, enabled? }` |

A source inside `face` is a field name or `{ field?, value?, map?, default? }`.
Output is always bound as text or assigned to an allow-listed host property.
Nothing a caller supplies is parsed as markup.

Event out: `propose-move`, bubbling and composed, with detail
`{ move, args, stack, index, version }`. Every value is JSON.

### Companions

| Element | Inputs | Purpose |
|---|---|---|
| `<bg-action>` | `move`, `args`, `disabled` | A button that proposes a move. Emits `propose-move` |
| `<bg-callout>` | `message`, `path`, `auto-message`, `suppress`, `name` | Plays when the value at `path` changes |
| `<bg-seat>` | `index`, `flags` | Reflects `[active]`, `[viewer]` and named booleans for CSS |
| `<bg-value>` | `path` | A number or string, with a change callout |
| `<bg-turn-status>` | `player-labels` | Whose turn it is |

### Worked example: Memory

This is `memory.html`, complete. It replaces a 161-line renderer class.

```html
<bg-table name="memory" game-name="memory" animation-length="400"
          hold-before='{"Capture Cards": 600}'>
  <div class="board">
    <bg-stack path="Game.Cards" layout="grid" messy face="Type"
              pick='{"move":"Reveal Card","arg":"CardIndex","slots":"hidden"}'></bg-stack>
    <bg-callout message="Match" path="Game.Cards.$count"></bg-callout>
  </div>
  <div class="won">
    <bg-stack path="Players.0.WonCards" layout="stack" messy face="Type"></bg-stack>
    <bg-stack path="Players.1.WonCards" layout="stack" messy face="Type"></bg-stack>
  </div>
  <bg-action move="Hide Cards">Hide Cards</bg-action>
  <bg-turn-status player-labels='["Player 0","Player 1"]'></bg-turn-status>
</bg-table>
```

The script's whole job, per state:

```ts
table.frame = { view, choices, legality, transition: { move } };
```

and, once, `table.chest = chest`. A click on the face-down card in slot 1 at
version 1 produces:

```json
{ "move": "Reveal Card", "args": { "CardIndex": 1 },
  "stack": "Game.Cards", "index": 1, "version": 1 }
```

As pattern JSX it would be the same tree with `frame={frame}` and
`onProposeMove={proposeMove}`. No child has a property that changes.

### Worked example: Blackjack

From `blackjack.html`. One face description serves every stack:

```json
{ "props": {
    "suit": "Suit",
    "rank": { "field": "Rank", "map": { "Ace": "A", "Jack": "J", "Queen": "Q", "King": "K" } },
    "rotated": true } }
```

```html
<bg-table name="blackjack" game-name="blackjack" animation-length="350"
          animation-lengths='{"Collect Cards": 700}'>
  <bg-stack path="Game.DrawStack" layout="stack" messy face="…"></bg-stack>
  <bg-action move="Current Player Hit">Hit</bg-action>
  <bg-action move="Current Player Stand">Stand</bg-action>
  <bg-stack path="Game.DiscardStack" layout="stack" messy face="…"></bg-stack>

  <bg-seat index="0" flags="Eliminated Stood">
    Player 0: <bg-value path="Computed.Players.0.HandValue"></bg-value>
    <bg-stack path="Players.0.HiddenHand" layout="fan" messy no-default-spacer face="…"></bg-stack>
    <bg-stack path="Players.0.VisibleHand" layout="fan" messy face="…"></bg-stack>
    <bg-callout message="Busted!" path="Players.0.Eliminated"></bg-callout>
    <bg-callout message="Stand!" path="Players.0.Stood"></bg-callout>
  </bg-seat>
  <!-- one bg-seat per player -->
</bg-table>
```

What Blackjack tested that Memory did not, and the result:

- **Per-seat views.** With `viewer=1`, seat 1's hidden card shows its face;
  with `viewer=0` the same card is face down with empty `suit` and `rank`.
  No element code decides this. The view arrives already sanitized.
- **A card moving between two stacks** (hidden hand to visible hand) flies
  and flips.
- **Fan layout with a rotated card** needs only the constant `rotated`.
- **Argument-less moves** need only `bg-action` and the frame's `legality`.

What was missing, and was added for it: `bg-seat`, `bg-value`, the `map`
table, and per-move animation lengths. What is still missing: the game
outcome banner and the companion surface's cross-screen deal flights.

## The three largest changes `server/static/src` would need

**1. An animation root that owns scope and remembers "before" (S9, S12).**
The animator should take its stacks from the root it belongs to, not from
`sharedStackList` (`boardgame-component-stack.ts:48`). It should keep the
resting geometry of the last installed state by component id and use it
when the old nodes are gone, instead of requiring a caller to invoke
`prepare()` at a moment that under a pattern renderer does not exist
(`boardgame-component-animator.ts:424-538`). Stack identity should be a
declared key, not `_randomId` (`boardgame-component-stack.ts:534`). The
state queue belongs at the same root. This is the change everything else
depends on; here it is about 250 lines of wrapper reaching into five
private fields.

**2. A stack that takes data, not functions (S11, S10).** `face` and `pick`
in place of `componentView`, `action`, `componentActions`,
`projectedChoices` and `selection`, with a plain intent event out. That
removes the four `moves/` files from layer 1, lets `component-view.ts` stop
importing the die and the token for a card stack, and ends the write to
`item.ID`. The webapp's renderers would get thinner by the same change:
Memory's `cardView` and `targets` calls become two attributes.

**3. Events and ambient access made safe by construction (S14, S15).**
Internal events (`will-animate`, `animation-done`,
`stack-slot-geometry-changed`, `component-tapped`, `motion-cycle-release`)
should not be composed, and should not carry element references. The
gesture event should check `isTrusted`. The test hook should be opt-in
rather than written to `window` at import. Art URLs need a policy. And
`fading-text` needs its "previous value" to live somewhere that survives
the element.

## Open questions

1. **Does Common Fabric's renderer really recreate an element when one of
   its own properties changes?** If so, `<bg-table frame={…}>` is recreated
   every state and the queue must move to module scope, keyed by `name`.
   `mode=table` shows animation survives that; queuing does not.
2. **Are removal and insertion one task under the real renderer?** If not,
   there is a visible gap that no element can cover.
3. **Can a pattern set CSS custom properties or ship a stylesheet?** Sizing
   and face typography depend on one of the two.
4. **How should multi-step input work?** Selection and placement drafts keep
   local state between gestures. Nothing here tested them.
5. **Should the table accept the raw wire view or a pre-expanded one?**
   Expanding inside the trusted element keeps the chest out of every frame.
