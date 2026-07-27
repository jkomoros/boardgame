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

## Structural motion pipeline

Card motion is not a list of special animation modes. One transition is
compiled through orthogonal layers:

1. **Continuity and presence.** Exact stable identity wins. When one endpoint
   is not rendered, unique privacy-safe stack history may establish an
   appearing or departing subject; ties and contradictions fail closed.
2. **Provenance and presentation.** The plan records why an endpoint is known.
   A missing DOM endpoint uses a fresh inert carrier with a surface-scoped
   historical presentation or the destination stack's safe defaults. Sanitized history
   never reconstructs private card content or an exact hidden position.
3. **Geometry and path.** Before/after measurements are immutable and branded
   by coordinate space. The solver produces numeric translation/scale and an
   explicit stationary-or-travel path. A declared transfer anchor contributes
   geometry only; it does not overwrite the card's semantic pose.
4. **Owned tracks.** Host transform/opacity and component-owned visual tracks
   such as card face and orientation are compiled separately, preflighted, and
   played together. Ownership arbitration ensures a stack arrival is either
   automatic FLIP or an explicit retained-carrier transfer, never both.
5. **Timing and cohorts.** One timing compiler handles CSS/default duration,
   reduced motion, companion version slots, clipping, delay, repeats, fill,
   and settlement budget. `motion.stagger()` replaces start delays for an
   explicit subject cohort without changing geometry or effects.
6. **Execution and lifecycle.** The immutable generation-bound structural plan
   is published before playback, then records armed, active, finished, skipped,
   or cancelled outcomes from real WAAPI animations. `Animation.finished` is
   settlement truth. Effects may observe exact structural points but cannot
   own or retime them.
7. **Queue policy.** Normal state admission waits for exact-cycle settlement.
   A solo renderer may declare `motion.release()` for an already-buffered
   successor; its barrier samples actual primary-animation progress. Cutover
   terminalizes the old generation and is not concurrent overlap.

The concrete DOM cases are consequences of those layers. A card rendered at
both endpoints uses exact continuity and its retained host. A virtualized or
departing endpoint uses an inert carrier. `IDsLastSeen` supplies only ambiguous
or unique provenance, never content. Table/Hand defaults preserve their local
compatibility choreography through `animateBetween()`; authored cross-stack
motion uses transfer declarations. A face flip, rotation, resize,
fade, spatial move, trail, and arrival effect can therefore coexist without a
new card-specific execution path.

`--animation-length` remains the component default, and
`renderer.animationLength()` may temporarily override it or skip an
intermediate buffered bundle with a negative value. That requested duration is
only an input: executed timing may be reduced or clipped by the shared timing
policy. Declarative transfer, cohort, release, and effect hooks are evaluated
once from the authoritative before/after transition rather than Lit lifecycle
callbacks.

## Motion tracks: endpoint pairs and sampled curves

`src/motion/component-track.ts` is the one place a component says what it wants
animated. `componentMotionTracks()` takes a list of *requests* and returns
frozen, normalized `ComponentMotionTrack`s; `playMotionTracks()` on
`BoardgameAnimatableItem` resolves each track's `target` to an element and
plays it. Requests come in two forms, and they compile to ONE shape:

- **Endpoint pair** — `{ target, property, from, to }`, on either the `host` or
  the `visual` channel. This is what structural FLIP, fades, and card
  face/orientation changes use.
- **Sampled curve** — `{ target: 'visual', property, curve, resolution?,
  resting? }`, where `curve` is a pure function of progress in `[0,1]`
  returning a CSS value. This is how motion that is not "interpolate A→B under
  one easing" — a baked physics trajectory, a path follow, a multi-bounce —
  gets into the system.

Both compile to `samples: readonly {offset, value}[]` plus a `timeline` of
`'eased'` or `'sampled'`. An endpoint pair becomes two samples at offsets 0 and
1, which is a WAAPI no-op for a two-keyframe list, so nothing about existing
motion changes. A curve is evaluated by the COMPILER on a uniform grid it owns:
the producer never authors offsets, and the closure is never stored (it can
capture DOM, cannot be frozen, and cannot be diffed by a golden). `resolution`
is clamped to `[2, 256]` (default 64) rather than rejected, matching the
clamp-and-degrade precedent in `effect-budget.ts` — an over-budget bake
decimates, because a 90-sample trajectory resampled to 64 is the same
trajectory.

Two things a curve track may not do, enforced in the type and at compile time:

- **A sampled `host` track is forbidden.** The host channel has three
  contracts that a sampled timeline would break — the FLIP resting write, the
  two-point viewport contract in `StructuralMotionPath`, and trail-echo
  synchronization, which animates echoes along a straight `fromCenter →
  toCenter` line under the effect easing. Sampled motion belongs on `visual`.
- **A constant curve throws.** Silently vacating a claimed channel is worse
  than a loud failure at the producer boundary; validation lives in the bake,
  which is pure and unit-tested, not in a runtime throw that the animator's
  try/catch would swallow into a `console.error`.

### Per-track timing, and the effect-level easing pin

`playMotionTracks` used to hand ONE `timing` object to every track in a batch.
It now derives timing per track: `componentMotionTrackEasing(track)` returns
`'linear'` for a sampled track and `undefined` for an eased one, and the linear
value is merged into that track's own timing.

The pin is at the **effect** level, not per keyframe. Per-keyframe
`easing: 'linear'` would be a no-op anyway (WAAPI keyframe easing already
defaults to linear); what actually warps a sampled trajectory is the
`ease-in-out` the kernel injects at the effect level. And the effect level is
also what `BoardgameComponentAnimator` reads into
`StructuralExecutedTiming` via `effect.getTiming().easing` — encoding
character in keyframes would make that published field report `linear` for
everything and quietly stop meaning anything.

Supplying `timing.easing` yourself alongside a sampled track **throws**: two
things claiming one channel's timeline is the same class of error as two
owners claiming one channel.

A sampled track must also be played with `{ timing: 'immediate' }`. Under the
default `'version'` policy a multi-second bake is clamped into the companion
cycle's slot and plays uniformly fast — geometrically faithful, physically
absurd, and silent — and the same policy can resolve to skip outright, which
makes `playMotionTracks` report `not-started` and takes the batch's sibling
tracks down with it. Both are invisible in solo play, where the ambient
animation context is null.

### The resting write

Animations run with `fill: 'none'`, so both natural completion and a forced
`finish()` render the element's **resting style**, not the animation's last
keyframe. `playAnimation` writes a resting style for the `host` channel only;
the `visual` channel has none, which is why `boardgame-card` maintains its
inner transform by hand. A track may therefore carry `resting` (defaulting to
`curve(1)` for a curve track), and `playMotionTracks` writes it to the resolved
target immediately after starting the animation. A producer whose `resting` and
final sample disagree by one rounding digit gets a visible twitch as the
animation is removed, so the two are generated by the same formatter.

## The die: a solid rolled by physics

`boardgame-die` renders a real 3D solid and rolls it by simulation. The
pipeline is five pure modules and one component, each testable without a DOM:

1. **Geometry** (`motion/die-geometry.ts`). `dieGeometry(faceCount)` returns a
   solid: outward face normals, plane offsets, vertices, face polygons, an
   inertia tensor, and a `ReadingRule`. Platonic counts (4, 6, 8, 12, 20) are
   closed forms; 10 is a trapezohedron; everything else is a barrel (an N-sided
   band with two cone caps). Solids are NOT size-normalized — circumradius runs
   from 1.00 to 1.90 — so every consumer divides through by circumradius.
   `finishSolid` rejects an open or non-manifold surface rather than returning
   a silently wrong inertia tensor.
2. **Relabeling** (`motion/die-faces.ts`). The outcome is the SERVER's and is
   known before any pixel moves, so the simulation is never asked to produce
   it. `presentedFaceIndex(geometry, orientation)` reads which face a resting
   orientation turned up, and `assignFaceValues` paints the server's value onto
   exactly that face while permuting the rest into a still-legitimate die
   (opposite faces still pair, and a d6 still winds 1-2-3 right-handed, so the
   result is not a mirrored die). Re-simulating until the physics agreed would
   cost `sides^dice` throws in expectation; this costs one, always. The visible
   price is that a die's OTHER faces carry different numbers after each roll.
   A d4 and an odd-sided barrel are read from the face they rest ON, which is
   what `ReadingRule` records.
3. **Simulation** (`motion/dice-sim.ts`). A seeded rigid-body throw: vertex
   contacts against an invisible tray, an impulse solver with restitution
   applied once after convergence (Poisson's hypothesis, capped per contact by
   Newton's rule — applying it inside the iteration loop pumps energy at high
   restitution), and rest detection after a continuous still-hold. It is
   bitwise reproducible from its seed, verified across fresh processes.
   `RollTrajectory` reports `cocked` and `restAlignment` so a caller can tell a
   throw the simulator could not settle flat.
4. **Bake** (`motion/dice-bake.ts`). `trajectoryCurve(die, durationMs, opts)`
   returns the pure function of progress that the curve track wants, emitting
   **literal `matrix3d(...)`** — never `var()`/`calc()`, which forces
   main-thread animation — by slerping orientation and lerping position between
   samples. `restingTransform(die, opts)` is byte-identical to `curve(1)`. The
   physics world is +Y up and CSS screen-Y points DOWN, so the bake applies the
   reflection `S = diag(1, −1, 1)` to the WHOLE pose (`p → Sp`, `R → S R S`) in
   one place. It must be a reflection: the obvious-looking `(x, −y, −z)` is a
   proper rotation into CSS's left-handed frame and renders the solid's mirror
   image. `opts.radiusPx` converts trajectory positions, which are in die
   circumradii, into pixels — a NUMBER read from the DOM, because interpolating
   a CSS variable into the matrix would forfeit compositing.
5. **The component** (`components/boardgame-die.ts`). It renders one element
   per face, placed by the face's own normal and centroid and cut to the face
   polygon with `clip-path`, on a `transform-style: preserve-3d` carrier.
   Face content is computed, not hard-coded: an author symbol map first, then
   generated pips on a 3×3 lattice, then numerals past nine, all laid out
   inside the largest square INSCRIBED in each facet's polygon (not its
   bounding box, which is what smears content across a barrel's long faces).

   A roll is triggered by `DynamicValues.RollCount`, the server-side counter
   that `Roll()` increments and nothing else touches — a throw landing on the
   face already showing leaves `SelectedFace` and `Value` byte-identical, so
   the face index cannot be the trigger. That count, with the component ID, is
   also the seed: it identifies the THROW, and unlike the state version it does
   not move while one throw is on screen. A die whose game does not use
   `components/dice` falls back to the face change and the state version.

   The bake describes the simulator's world, not the player's, so the component
   composes a constant scene prefix in front of every keyframe: a landing
   square-up (the minimal turn putting the read face exactly on the reading
   axis, so a cocked throw still lands readable), a fixed camera elevation
   above the tray, and a recentring translation back to the middle of the die's
   own box. The simulator's trailing rest-hold is trimmed before playback —
   it is ~300ms of a ~1s roll spent holding the whole game's animation gate
   open on a die that has already stopped.

   The whole thing plays as ONE sampled curve track on the `visual` channel
   with `timing: 'immediate'`, so it is an ordinary gate participant: it
   declares its own settle budget through `will-animate`, extends the watchdog,
   and honors reduced motion by resolving to duration 0 and landing on its
   resting pose.

## Animation timing: play() / settlement / the gate

The timing logic described above (computing before/after transforms) is
unchanged, but *knowing when an animation is done* uses the Web Animations
API (WAAPI) directly instead of counting `transitionend` events. There is no
more expectation counting, no `_expectTransitionEnd`, no `willNotAnimate`,
and no `transitionend` listening anywhere in the animation path —
`Animation.finished` is ground truth.

`BoardgameAnimatableItem` (`src/components/boardgame-animatable-item.ts`) is
the common base for every game-semantic animated element, not just the
components a stack lays out: `boardgame-component` (and its
`boardgame-card`/`boardgame-token` subtypes) and
`boardgame-component-stack`'s faux animating elements extend it, and so do
the standalone primitives that live outside a stack — `boardgame-die`,
`boardgame-fading-text`, `boardgame-status-text`, `boardgame-game-outcome`.
Whichever element it is, it animates through the exact same kernel, entered
through a single method, `play(element, keyframes, timing, opts)`:

- It calls `element.animate(keyframes, timing)` and gets back a real WAAPI
  `Animation`, always with `composite: 'replace'`. That is pinned
  deliberately, not incidentally: when two transform animations run on the
  same host in one cycle (e.g. a stack's `layoutTransform` self-play racing
  the same cycle's FLIP host track — see below), `'replace'` semantics mean
  the higher animation wins outright each frame, instead of the two adding
  together and visibly doubling the motion. No parity golden can catch a
  violation here — curves are displacement-normalized, so a uniform doubling
  divides back out of the comparison. Do not remove or parameterize this
  without a test that pins the same-host composite case.
- Unless `noAnimate` is set (a barrier used while the animator is measuring
  before/after layout) or `opts.gated === false`, the animation counts toward
  the item's *gated* set: **every** gated animation that starts fires
  `will-animate`; when the gated count returns to zero it fires
  `animation-done`. `will-animate` is a *declaration*, not a 0→1 transition —
  each firing declares that play's own settle budget, so a long animation
  starting alongside a short one still extends the watchdog to cover itself.
  Listeners must therefore be idempotent; `AnimationGate.willAnimate` is, by
  construction (a keyed write plus a strictly monotone deadline).
- Timing resolves against the ambient `animationContext` discovered by
  climbing ancestors — crossing shadow roots and slots, and past any
  intermediate `null` context — so a wrapper element (e.g. status-text
  sitting above its own nested fading-text) never shadows the real
  `boardgame-render-game` provider above it. Standalone dice, tokens, and
  game-authored animatable items resolve the same installed version-slot
  context as stack-managed components, with no game-renderer plumbing
  required.
- `settled(): Promise<void>` resolves once every gated animation on that item
  has finished (or was cancelled — `anim.finished` rejects on cancel, and
  both paths count as settlement).
- `finishAllAnimations()` force-finishes (or cancels) every live animation on
  the item synchronously, resolving settlement immediately. This backs
  interruption semantics (a new animation cycle starting while a previous one
  is still in flight), `beforeOrphaned()` (an animating faux component is
  about to be removed from the DOM — settle first so the gate never waits on
  a detached element), and a disconnect safety net: `disconnectedCallback`
  force-settles any still-gated item a microtask after a genuine removal
  (deferred so a same-tick reparent never snaps a live animation), because a
  detached element's `animation-done` has no parent left to bubble to and
  would otherwise leave the gate waiting on it forever.
- `animationLengthMs()` reads the effective `--animation-length` CSS custom
  property (set by `boardgame-render-game` from the renderer's
  `animationLength()` hook) and is what `play()` uses as the default duration
  unless the caller overrides `timing`.
- Reduced motion (`prefers-reduced-motion: reduce`) is a complete scheduling
  policy inside `play()`, not a default an explicit duration can override: it
  collapses duration and delay to 0 while preserving `endDelay` (so a
  semantic hold, like a matched pair lingering before capture, still works).
  An ambient infinite-iteration effect can't take that shortcut (a 0-duration
  infinite play renders nothing) — the token throb instead holds the strong
  end of its glow statically under reduced motion, keeping the affordance
  without the pulse; the old shadow-scoped CSS animation ignored the
  preference entirely, so neither of those two extremes is what it does now.

`BoardgameComponentAnimator._startAnimations` calls `play()` (via each
component's `playAnimation()`) for every component that needs to animate this
cycle, collects each one's `settled()` promise, and resolves the promise it
handed back from `animateFlip()` with `Promise.all(settledPromises)`. That
promise means "everything has visually SETTLED", not "everything has
started" — callers await real completion, not just animation kickoff.

### The ambient registry

Not every animatable item is tracked by the component animator's own
stack-component bookkeeping — a standalone die, a `status-text`/
`fading-text`, or any other game-authored `BoardgameAnimatableItem` living
outside a `boardgame-component-stack` is invisible to it. `AnimatableRegistry`
(`src/motion/animatable-registry.ts`) closes that gap: a single instance
lives on `boardgame-render-game` as `animatableRegistry`. Every
`BoardgameAnimatableItem` climbs ancestors at `connectedCallback` (the same
walk shape as the ambient `animationContext` lookup above) to find it and
self-register, unregistering symmetrically at `disconnectedCallback` from
the same registry it found at connect time (disconnect can't re-walk — by
then the element may already be unparented). At the start of each cycle,
render-game iterates a snapshot of the registry's items and force-settles +
re-contexts each one, the same treatment the animator already gives its own
tracked components, so a same-cycle interruption never leaves an untracked
item's stale animation running into the next cycle.

### The gate

The completion gate itself is a pure, DOM-free kernel, `AnimationGate`
(`src/motion/animation-gate.ts`), extracted verbatim (behavior-for-behavior)
from what used to be inlined in `boardgame-render-game.ts`, so its
timing/bookkeeping invariants — in particular the watchdog-deadline
extension, which has no e2e coverage long enough to exercise it — can be
unit-tested with fake timers instead of only indirectly through the DOM.
`boardgame-render-game` owns one `AnimationGate` instance plus a boolean
`isAnimating` (reflected as the `is-animating` attribute so tests/CSS can
observe it, and broadcast via an `animating-changed` event since it isn't
reachable as a reactive property from ancestors) that mirrors the kernel's
open/closed state through its callbacks:

- `_resetAnimating()` calls `gate.open(cycleId)` at the start of an animation
  cycle. Before opening, it force-settles every item in the ambient registry
  (see above) so an untracked item never carries a stale animation into the
  new cycle. The kernel's `onOpen` callback flips `isAnimating = true` and
  mirrors the value onto the active renderer's `animating` property
  (`_applyAnimatingToRenderer`, called at both gate flips and at renderer
  instantiation so a renderer created mid-cycle starts with the right value).
- Each component's `will-animate` event is forwarded into
  `gate.willAnimate(ele, label, expectedSettleMs)`, and its later
  `animation-done` into `gate.animationDone(ele)`; when the gate's last
  participant clears, its `onAllDone` callback flips `isAnimating = false`
  and `boardgame-render-game` fires `all-animations-done`, which
  `boardgame-game-view` uses to ask the state manager to install the next
  pending state bundle.
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

#### Roster / player-info gating

`boardgame-player-roster` (and the per-game player-info renderers it mounts)
is a DOM **sibling** of `boardgame-render-game`, not a descendant — so a
roster-hosted animatable's (e.g. `boardgame-status-text`'s nested
`boardgame-fading-text`) `will-animate`/`animation-done` events bubble past
render-game entirely and would otherwise never reach its gate.
`boardgame-game-view` pipes them in through two thin public delegates on
render-game, `gateWillAnimate`/`gateAnimationDone`, wired to
`@will-animate`/`@animation-done` listeners on the roster element in its
template:

- `will-animate` is forwarded **only** while a board cycle is already open
  (`renderEle.isAnimating`) — a roster animation outside any cycle (e.g. a
  hover-triggered fade) can never open or wedge a new one.
- `animation-done` is **always** forwarded: a participant admitted at open
  must always be able to settle, and the gate's `animationDone()` is a safe
  no-op for an unregistered element.
- Because the ambient registry lives on `boardgame-render-game` and the
  roster is its sibling, roster-hosted items are gated but **not**
  registry-swept: on a cycle handoff, a mid-flight board animatable is
  force-finished (snaps) while a roster one completes smoothly. This
  asymmetry is deliberate and accepted (see
  `tests/animations/parity/README.md`) — a late roster settle is a harmless
  kernel no-op, and roster items intentionally resolve a null ambient
  animation context, which is correct: their animations are local effects,
  not version-slot participants.
- A roster item removed from the DOM mid-animation is force-settled by
  `BoardgameAnimatableItem.disconnectedCallback` (see above), but that
  settlement's `animation-done` bubble has no parent left to reach once
  detached. `game-view`'s `_rosterWillAnimate` additionally subscribes to the
  item's `settled()` promise as a second, DOM-independent done channel at
  will-animate time, closing that orphan-settle gap without disturbing the
  normal bubbled path — in the ordinary attached case both fire (the bubbled
  path closes the gate first; the kernel's idempotent `animationDone()` makes
  the later `settled()`-driven call a harmless no-op), and only the orphaned
  case actually depends on the second path. See
  `docs/superpowers/specs/evidence/2026-07-25-roster-orphan-settle.md`.

### The watchdog

The gate kernel arms a watchdog timer every time `open()` is called, and
clears it whenever the cycle closes normally or a fresh cycle opens. If the
deadline passes without every registered participant clearing — a hung
`Animation`, a component that never fired `animation-done`, a bug — the
kernel fires `onWatchdog(pending, budgetMs)` immediately followed by
`onAllDone()`, force-closing the cycle regardless. `boardgame-render-game`'s
watchdog callback logs an error naming the still-pending components and
records a `watchdog` event via the `animHooks` test-hook singleton
(consulted by the Playwright suite's `watchdogFirings` assertions: a passing
run must see zero watchdog firings, since a firing means some animation path
didn't settle on its own). This is the invariant that guarantees the gate —
and therefore move entry and state installation — can never wedge
permanently, regardless of what bugs exist upstream in timing/settlement
logic.

The deadline is **not** a flat timeout: a legitimate cycle can run much
longer than a few seconds (e.g. 15 cards staggered at 0.2 with a 2s
`--animation-length` — the last card doesn't even start until ~5.6s), and a
flat 4s watchdog would force-close that cycle mid-flight, violating its own
"firing = bug" invariant. Instead each gated `play()` reports its declared
settle time (`delay + duration + endDelay`, covering stagger delay,
animation length, and `post-animation-delay`) in the `will-animate` event's
`expectedSettleMs`. The gate kernel's `willAnimate()` tracks the largest such
value declared in the current cycle and extends the watchdog to *that
declared settle instant + a 1.5s margin* whenever a play would outlast the
current deadline. The deadline never drops below a 4s floor, so trivially
short cycles still get a prompt backstop; it only ever grows to cover a
cycle's own declared animation budget. A firing therefore still unambiguously
means an animation overran the time it *itself declared* it would take — a
real bug, never just a long-but-honest cycle.

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

### Stack layout: layoutTransform

`boardgame-component-stack` no longer positions its children with a CSS
`transition: transform …, opacity …` rule plus a `.no-animate` container
class to suppress it during measurement — that mechanism, and the stack's
container-level `noAnimate` toggle that drove it, are gone entirely. Instead,
for every component it lays out, the stack assembles the messy/pile/fan
transform pieces into one string and assigns it through
`BoardgameComponent.layoutTransform` — a self-animating setter on the common
component base. Setting a new value snaps `this.style.transform` to it
immediately (so layout/hit-testing always sees the true final value, exactly
like an authored CSS property does under a transition), then — unless
suppressed — plays a gated host animation from the pre-snap *computed*
transform to the new computed transform through the same `play()` kernel.
Capturing the computed value (not the previous setter argument) is what
reproduces CSS-transition-style retargeting: an interrupted retarget
continues from wherever the box actually is on screen, never from the stale
target of the animation it interrupts. Setting the same value is a no-op;
setting while `noAnimate` is up or the element is disconnected only snaps.
The **component-level** `noAnimate` barrier used during FLIP measurement is
unaffected and still gates `play()` exactly as before.

The stack's relayout write lands from Lit's slotchange/updated pass, which
runs microtasks *before* the animator raises that component-level barrier
for the cycle — so the setter's self-play and the same cycle's FLIP host
track can both be live on one host at once. This is not a double-animation
regression: it reproduces what the retired CSS transition already did (it
fired at the same pre-barrier slotchange moment, with the same
easing/duration, co-existing with FLIP), and `play()`'s pinned
`composite: 'replace'` (see above) is exactly what makes two same-host
transform animations resolve to one correct motion instead of doubling it.
See `docs/superpowers/specs/evidence/2026-07-26-stack-transition-cutover.md`
and the `geometry-debuganimations-fan-draw` parity golden, which was
recorded from the old CSS path and passes unregenerated against the setter.

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
