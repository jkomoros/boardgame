This document describes at high level how the client side game views are
architected.

Its primary audience is people who are working on the architecure themselves,
for example to add new features to the frameweork. People who are simply using
the framework shouldn't need to understand this document.

`boardgame-game-view` is the top-level view that renders the page for games.
Its responsibility is to identify the game type and game ID to render (based
on URL). It then passes that information to the `boardgame-render-player-
info`, and `boardgame-render-game`, which are responsible for dynamically
loading and instantiating the renderers for that game type.

Within `boardgame-game-view` is an instance of `boardgame-game-state-manager`.
Its job is to fetch state from the server and pass it up to the game-view to
render when a new one should show. It fetches informationa about the game in
general (number of players, their names, etc) and then also opens a socket to
the server which will receive a message whenever a new state is available on
the server. It will then fetch that state from the server (and any states
between what it last fetched and that new version). It will then pass bundles
of state up to game-view to render, one at a time. It waits until the game-
view asks for another one to render (because the previous one is done
animating). It also modifies the state objects as received from the server to
include additional information that is useful for databinding. 

Finally, when `game-state-manager` is told by `game-view` to render another state, it checks with the current game renderer to see if it has an `animationLength` method, which can (temporarily) overridde the `--animation-length` css property. If you return 0 then the default CSS values will be used, but any value above that will be set on the renderer object as `--animation-length` until the next one is received. If the length is negative, then that bundle will simply be skipped (unless it's the last bundle in the queue, which is always installed).

To hold an item on screen for a beat after its animation completes (for example so players can see a matched pair before it's captured), renderers bind the declarative `post-animation-delay` attribute (milliseconds) on the relevant `boardgame-component-stack` or item, computed from current render-time state. This replaced an earlier imperative per-move delay hook on the renderer that the state manager used to consult directly before installing the next bundle.

For buffered catch-up, `motionReleaseForTransition()` can declare an early
cutover after selected structural primaries cross an actual active-progress
fraction. The animator derives this from executed WAAPI timing, not
`animationLength`. Both progress release and final settlement carry one opaque
install-cycle ID; stale events cannot advance a newer cycle. Cutover
terminalizes the old generation, so this is intentionally not described as
concurrent overlap.

`boardgame-game-view` also listens for `propose-move` events emanating from
within the rendererd game, and then forwards them to the `boardgame-admin-
controls`, where the logic to actually serialize them and pass to the server
resides. If the move is successfully applied, the state manager will hear
about it via the socket, and that will kick off more states being downloaded.
(The same mechanism applies no matter if other players or the game itself
applied moves).

`boardgame-render-game`'s primary job is to a) instatiate the specific type of
renderer for this game type (and pass through any updated state it receives
from `boardgame-game-view`) and b) to coordinate animations of state. We will
come back to animations.

Most specific games' renderers inherit from `boardgame-base-game-renderer`.
Its job is entirely to listen for `tap` events on components within the
renderered game view that have attributes about move proposing, and when that
happens, to fire a new `propose-move` event upward for the `game-view` to
capture and process.

The primary goal of a given game renderer is to take a state object and data-
bind it into a well laid-out view. It typically does this by data-binding into
normal layout elements, as well as buttons, `boardgame-component` (and its
sub-types, `boardgame-card` and `boardgame-token`), and `boardgame-component-
stack`. Again, any item that has attributes that talk about propose-move will,
when they are tapped, emit an event that the base game renderer will catch and
then re-throw as a propose-move.

`boardgame-component` and sub-types are (almost?) always rendered as children
of `boardgame-component-stack`. `boardgame-component-stack` has a few
responsibilities. It creates new `boardgame-components` from the renderer's
typed, renderer-scoped `componentView`. It also can
do advanced beahvior where for large stacks it only data-binds a few
components for real, and does faux components for others. Its primary job
though is to layout the children components according to its own layout
attributes. They can fan out cards, stack them, arranage in a grid, etc. They
also may perturb the exact position of the children to give a messier layout.
`boardgame-component-stack` also helps out with animations, generating faux
animating elements when necesary. More on that later.

The actual `boardgame-component` are generally either `boardgame-token` or
`boardgame-card`. The former is way simpler; it is just a simple object whose
appearance is defined by the attributes it has. Cards are way more
complicated; they can be tall or wide, rotated or not, and flipped or not. All
of those attributes, when changed, animate. If you select one in the DOM and
change one of those properties, you'll see them animate smoothly. Of course,
that doesn't happen literally in normal practice, because it's all one-way
databound statically from the state. These animations are referred to as
"internal" animations. They affect the layout properties of the component, but
they are based on information set internally.

Note that all animations of all types have a default length set by the CSS var
`--animation-length`. If you want to change the animation, you can target a
different CSS var at the item. You can also override renderer.animationLength to set a different animation value temporarily. `motionReleaseForTransition()` is the separate, declarative policy for admitting an already-buffered successor from actual structural progress.

Components have three types of transforms that can apply. The first is
*internal*. These are transformations on the inner element. For cards this
includes whether the card is faceUp and whether it's rotated. The next is
*external*. These are transform tweaks applied by the `component-stack` to
perturb the final layout, for example to make messy cards or fanned cards be
in their final layout. (Normal layout is used for gross position; these are
just small tweaks). The final are *inverse* transforms, which are applied by
component animator during animations in order to position a component where it
was in the last state, so it can animate to its new location. *External* and
*inverse* transforms are in practice applied the same way currently, which
means that animator has to figure out how to munge them toegether, setting
what is properly an *external plus internal* transform.

This is all pretty straightforward. However, the real benefit of the engine is
that it handles animations as components move between states well. At a high
level, the game logic on the server has decided how granularly to break up
moves. Correct animations can only happen between versions; the server game
logic thus decides where full animations MAY happen. It's up to the client to
actually calculate the animations to occur, set them in motion, and figure out
when they're done. *In the future it will also be possible for the client to
decide to skip certain states because it doesn't want to animate each state
change individually, by looking at a before and after state and choosing to
not databind the former.*

At a high level, what we do is bind the first state, then bind the second
state as a totally separate item. Items that just so happen to be in the same
place might be re-used by Polymer's data-binding engine, but components that
have logically moved to a different location from state to state (for example,
a card that moved from the draw stack to the discard stack) are almost
certainly represented by different physical DOM nodes before and after.

Most of the magic is organized by `boardgame-component-animator`. Before a new
state is bound, it goes through and collects the current location and state of
all of the components, keeping track of which is which by comparing the "id".
Then it allows the new state to be bound. It then goes through each element
and sees where its new location is. (It does that in one pass before going on
to the next step to avoid layout thrashing).

Now for the hard part. It goes through and generates inverse transforms to
move each component, visually, back to where it was in the previous state, and
then applies a CSS transform to bring it back to the location it literally is
in the DOM, by reducing those transforms to 0 in an animation. This
transformation is referred to as the *inverse* transform. This is a very
challenging calculation to do, especially because components have internal
animations that could change their layout.

Components who are represented by a literal DOM element before and after are
(relatively) easy. Just calculate the inverse transform and apply it.

Slightly more difficult are cases where either before or after had a literal
DOM element, but the other end of the transition doesn't; perhaps it's going
to a `boardgame-component-stack` with so many elements that we print only a
handful of faux components instead of one per actual item. In those cases, we
ask the stack that contains the component to generate a fake component to
animate (the stack gives it a default position in the middle), that will act
like the literal element. When the animation is over, the faux animating
component is removed.

The hardest case is when there is a component who either before or after is
not known to be in a specific location in a stack. This happens, for exmaple,
when a component moves from a normal stack to one that's sanitized with
PolicyLen. That means that the actual list of component IDs is elided, and all
that's left is stack.IDsLastSeen. This captures that the last time the given
ID was seen was in this particular stack, but not _where_ in the stack it was
seen. In this case `boardgame-component-animator` does a behavior like the one
immediatley above. It creates a faux animating element. It positions the
component in the middle of the stack, and styles the element to be very small
and transparent, so as the component animates back to 0 state it's visually
clear which stack the component went to in general, but not where in the
component it went.

## Animation timing: play() / settlement / the gate

The timing logic described above (computing before/after transforms) is
unchanged, but *knowing when an animation is done* was rewritten to use the
Web Animations API (WAAPI) directly instead of counting `transitionend`
events. There is no more expectation counting, no `_expectTransitionEnd`, no
`willNotAnimate`, and no `transitionend` listening anywhere in the animation
path — `Animation.finished` is ground truth.

`BoardgameAnimatableItem` (the mixin/base that both `boardgame-component` and
`boardgame-component-stack`'s faux animating elements extend) exposes a single
entry point, `play(element, keyframes, timing, opts)`, that every animating
transform goes through:

- It calls `element.animate(keyframes, timing)` and gets back a real WAAPI
  `Animation`.
- Unless `noAnimate` is set (a barrier used while the animator is measuring
  before/after layout) or `opts.gated === false`, the animation counts toward
  the item's *gated* set: on the first gated animation to start, the item
  fires `will-animate`; when the gated count returns to zero it fires
  `animation-done`.
- `settled(): Promise<void>` resolves once every gated animation on that item
  has finished (or was cancelled — `anim.finished` rejects on cancel, and
  both paths count as settlement).
- `finishAllAnimations()` force-finishes (or cancels) every live animation on
  the item synchronously, resolving settlement immediately. This backs
  interruption semantics (a new animation cycle starting while a previous one
  is still in flight) and `beforeOrphaned()` (an animating faux component is
  about to be removed from the DOM — settle first so the gate never waits on
  a detached element).
- `animationLengthMs()` reads the effective `--animation-length` CSS custom
  property (set by `boardgame-render-game` from the renderer's
  `animationLength()` hook) and is what `play()` uses as the default duration
  unless the caller overrides `timing`.

`BoardgameComponentAnimator._startAnimations` calls `play()` (via each
component's `playAnimation()`) for every component that needs to animate this
cycle, collects each one's `settled()` promise, and resolves the promise it
handed back from `animateFlip()` with `Promise.all(settledPromises)`. That
promise means "everything has visually SETTLED", not "everything has
started" — callers await real completion, not just animation kickoff.

### The gate

`boardgame-render-game` owns a boolean `isAnimating` (reflected as the
`is-animating` attribute so tests/CSS can observe it, and broadcast via an
`animating-changed` event since it isn't reachable as a reactive property
from ancestors). This is "the gate":

- `_resetAnimating()` flips it open (`isAnimating = true`) at the start of an
  animation cycle, resets the set of components it's waiting on, and mirrors
  the value onto the active renderer's `animating` property
  (`_applyAnimatingToRenderer`, called at both gate flips and at renderer
  instantiation so a renderer created mid-cycle starts with the right value).
- Each component fires `will-animate` (tracked in `_activeAnimations`) and
  later `animation-done` (removed from the map); when the map empties,
  `_notifyAnimationsDone()` flips the gate closed (`isAnimating = false`) and
  fires `all-animations-done`, which `boardgame-game-view` uses to ask the
  state manager to install the next pending state bundle.
- `boardgame-game-view` also swallows `propose-move` events while
  `isAnimating` is true (#721): the on-screen state is mid-transition to what
  the server already considers current, so a move proposed now would judge
  stale-looking state, and it also guards the classic
  double-click-a-move-button case. The gate re-opening (normally or via the
  watchdog) is what lets move entry resume — it never permanently wedges.
- Table/Hand view renderer bases (`boardgame-table-view-base.ts`,
  `boardgame-hand-view-base.ts`, and the shared `animating` property they
  inherit from `boardgame-base-game-renderer.ts`) gate verdict/outcome UI
  (`renderGameOverBanner()`, `renderHandHeader()`'s status text) on this same
  mirrored `animating` flag, so a game-over banner or "You won!" string can
  never render while the winning move's animation is still in flight (#798).

### The watchdog

`_resetAnimating()` also arms a watchdog every time the gate opens, and
clears it whenever the gate closes normally (`_notifyAnimationsDone`) or a
fresh cycle starts. If the deadline passes without every awaited animation
settling — a hung `Animation`, a component that never fired `animation-done`,
a bug — the watchdog force-fires `_notifyAnimationsDone()` anyway, logs an
error naming the still-pending components, and records a `watchdog` event via
the `animHooks` test-hook singleton (consulted by the Playwright suite's
`watchdogFirings` assertions: a passing run must see zero watchdog firings,
since a firing means some animation path didn't settle on its own). This is
the invariant that guarantees the gate — and therefore move entry and state
installation — can never wedge permanently, regardless of what bugs exist
upstream in timing/settlement logic.

The deadline is **not** a flat timeout: a legitimate cycle can run much
longer than a few seconds (e.g. 15 cards staggered at 0.2 with a 2s
`--animation-length` — the last card doesn't even start until ~5.6s), and a
flat 4s watchdog would force-close that cycle mid-flight, violating its own
"firing = bug" invariant. Instead each gated `play()` reports its declared
settle time (`delay + duration + endDelay`, covering stagger delay,
animation length, and `post-animation-delay`) in the `will-animate` event's
`expectedSettleMs`. `_componentWillAnimate` tracks the largest such value and
extends the watchdog to *that declared settle instant + a 1.5s margin*
whenever a play would outlast the current deadline. The deadline never drops
below a 4s floor, so trivially short cycles still get a prompt backstop; it
only ever grows to cover a cycle's own declared animation budget. A firing
therefore still unambiguously means an animation overran the time it *itself
declared* it would take — a real bug, never just a long-but-honest cycle.

### Attributes

Three DOM attributes give game renderers declarative control over animation
behavior without touching the state manager or the gate directly:

- **`post-animation-delay`** (milliseconds, on an item or a
  `boardgame-component-stack`, forwarded to stamped children): sets
  WAAPI's `endDelay` on `play()`'s timing, holding the element in its final
  state for a beat after the animation visually completes before
  `animation-done`/settlement fires. Replaced an earlier imperative
  per-move delay hook that the state manager used to consult directly
  before installing the next bundle — this is the same effect (e.g. let a
  matched pair linger before it's captured) expressed declaratively at the
  render site instead.
- **`wait-for-animation`** (boolean-ish, default true; the literal string
  `"false"` is the only way to opt out, since the property already defaults
  true): controls whether this item's animations hold the gate open at all
  (`opts.gated` in `play()`). An item with `wait-for-animation="false"`
  still animates, but the gate doesn't wait for it to settle.
- **`stagger`** (fraction, on `boardgame-component-stack`): when > 0, each
  animating child in the stack's cycle is delayed by
  `index * stagger * animationLengthMs()` before its `play()` call, producing
  a cascading start (e.g. cards dealing one after another) instead of every
  card starting simultaneously.

### Companion scheduling (#798)

Companion mode (Table + Hand views on separate physical screens) wants the
same logical event — e.g. a card dealt from the table to a phone — to start
animating on both screens at roughly the same wall-clock instant, even
though each screen receives its own WebSocket push and renders independently.

- The server notifier owns a monotonic animation lane per observed game. An idle lane
  starts at `serverSentAt + 500ms`; rapid versions reserve 800ms-spaced slots
  (600ms synchronized motion + 200ms preparation),
  so visible fix-up versions cannot all target the same instant. With no
  listeners, notifications coalesce to the latest version at `now + 500ms`
  instead of accumulating an invisible future backlog. Every socket receives
  the same frame, and a reconnect receives the retained slot for its current
  version rather than inventing a new target.
- Socket setup performs three clock-sync request/reply rounds. The estimator
  uses the offset from the lowest-RTT midpoint sample, avoiding direct
  one-way-route bias; complete timing-policy frames provide a one-way fallback
  when clock-sync replies are unavailable. Older frames without the declared
  slot policy are rejected rather than partially interpreted.
  `companionSync.localEquivalent(serverEpochMs)`
  converts a server timestamp into the local-clock instant of the same
  wall-clock moment; with fewer than 3 samples it falls back to the raw
  server timestamp (animations play immediately — safe degradation).
- `CompanionAnimationTimeline` stores timing by `(gameID, version)`, not as a
  mutable latest value. A version waits at most 200ms for its sibling timing
  frame; missing timing or a cold estimator degrades to immediate playback.
- `_scheduleNextStateBundle()` resolves the pending bundle's exact version,
  installs it inside the 200ms preparation window, and carries its scoped policy through
  `boardgame-game-view` into the shared animator. Delayed WAAPI uses backwards
  fill so the source pose remains visible until compositor-owned launch.
- The common `play()` primitive resolves the installed context through the
  composed render tree, so stack FLIP, property effects, standalone dice, and
  game-authored animatable items share the same target automatically.
  `fly()` uses that policy too. It explicitly names a geometry-only source and
  the retained carrier at its natural destination. Both APIs accept
  `{ timing: 'immediate' }` for a local effect or
  `{ timing: { localStartAtMs } }` for an explicit local timeline.
  Stagger, visible duration, and post-animation hold share one remaining-slot
  budget; effects that cannot begin before its end are omitted. Synchronized
  cycles do not use renderer overlap.
- The game-over verdict (`renderGameOverBanner()` / the Hand header's
  outcome text) is gated on the mirrored `animating` flag described above,
  so the outcome banner can't race ahead of a still-in-flight companion
  flight. For that guarantee to hold, the flight itself must keep the gate
  open: when `fly()`'s resolved carrier is an animatable
  item, it routes the flight through that item's `play()` (gated) rather
  than a raw `real.animate()`, so the departing/arriving card registers a
  will-animate/animation-done pair and the gate stays open — and the
  verdict stays suppressed — until the flight (sync delay included) settles.
  Only plain, non-item elements fall back to the ungated raw path. The
  watchdog is the backstop that guarantees the gate can't wedge if a
  flight's `Animation` never settles; because the flight reports its full
  `delay + duration` budget in `will-animate`, the watchdog extends to
  cover it rather than force-closing a legitimately delayed synced flight.
