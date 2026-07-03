# Animation Timing Reliability: WAAPI Rewrite — Design

**Date:** 2026-07-03
**Issues:** #720, #726, #713, #714, #715, #716, #721, #728, #798 (capstone)
**Branch:** `animation-waapi-timing`

## Problem

The client animation system infers when animations complete instead of knowing.
`boardgame-component-animator` *predicts* which CSS transitions will fire (diffing
offsets/transforms/opacity against 0.5px / 0.01 thresholds), registers expected
`transitionend` events per element/property in `boardgame-animatable-item`, and
counts them down as events arrive. The state-version queue is pumped by that
countdown reaching zero. Any wrong guess — a transition that never fires, an
interrupted transition (`transitioncancel` is not handled), a property the browser
coalesces — stalls the queue until a 15-second watchdog force-fires
`all-animations-done` (#720). Every feature that needs to know when animation ends
(button disabling #721, staggering #728, wait-for-animation #716, companion
cross-screen sync #798) inherits this fragility.

## Decision

Replace CSS-transition-driven animation with the Web Animations API
(`element.animate()`). Every animation becomes a real `Animation` object with a
native `finished` promise. Completion is observed, never inferred. The expectation
machinery (`_expectedTransitionEnds`, `_outstandingTransitonEnds`,
`_transitionEnded`, `_expectTransitionEnd`) is deleted entirely.

Decisions locked in during brainstorming:

- **Scope:** full refactor including the declarative features (#715, #716, #721,
  #728), with #798 (companion sync consumption) as the final phase.
- **Done bar:** an automated Playwright regression gate — known repro scenarios run
  repeatedly with zero watchdog firings and no premature state installs.
- **API compatibility:** break the renderer-facing API freely; migrate all in-repo
  example games **and** the external clients in `../games`
  (murdermrmonroe, pass, valentine) as part of this campaign.

## Architecture & ownership

Four layers, each with one job (this lands the #713/#714 ownership refactor):

### 1. `BoardgameAnimatableItem` (base class, rewritten) — owns ground truth

Holds a set of live WAAPI `Animation` objects. Core method:

```ts
play(element: HTMLElement, keyframes: Keyframe[], timing: OptionalEffectTiming,
     opts?: { gated?: boolean }): Animation
```

- Calls `element.animate(keyframes, timing)`.
- Adds the Animation to the live set; fires `will-animate` on 0→1.
- When the animation settles (its `finished` promise resolves **or** it is
  cancelled), removes it; fires `animation-done` on 1→0.
- Exposes `isAnimating: boolean` (derived from the live set, always correct) and
  `settled(): Promise<void>` (resolves when the live *gated* set empties; never
  rejects — cancellation counts as settled).
- `opts.gated` (default `true`) controls whether this animation participates in the
  global completion gate (this is #716's `wait-for-animation`, see Features).

No transition event listeners, no expectation maps, anywhere.

### 2. `BoardgameComponent` / card / token / die — own what animating looks like

Each component type translates `(beforeProps, afterProps, positionDelta,
opacityDelta)` into keyframes:

- FLIP transform on the host element (inverted position → identity, `fill: 'none'`
  so the natural after-layout is the end state).
- Opacity fades.
- Intra-component effects: the card's 3D `rotateY` flip runs as keyframes on the
  inner element; final Lit properties (`faceUp`, `rotated`) are set immediately so
  the DOM's resting state is the after state, with the animation as a visual
  overlay on top.
- Duration comes from the `--animation-length` CSS variable, read via
  `getComputedStyle` at play time — games keep controlling speed from CSS.
- `noAnimate` and spacer components simply never call `play()`: styles apply
  instantly and there are no phantom expectations to suppress (`willNotAnimate()`
  bookkeeping goes away).
- `prefers-reduced-motion` is respected by forcing duration to 0 (animations
  settle immediately but all lifecycle events still fire in order).

### 3. `boardgame-component-animator` — owns orchestration

The measure/compute FLIP phases (recently fixed: `getBoundingClientRect`, center
alignment, id reflection) are kept **unchanged**. The INVERT/PLAY phases change:
instead of writing inline transforms + expectations, the animator asks each item to
`play()` with computed keyframes and per-item timing (`delay` for staggering,
`endDelay` for post-animation holds, absolute scheduling for companion sync).
`animateFlip()` returns `Promise.all` of the items' `settled()` promises.

A wrong `needsAnimation` guess is now harmless in both directions: a false positive
is a zero-distance animation that still settles; a false negative just means no
animation plays. Neither can stall the queue — the thresholds become purely a
performance optimization.

Faux/cross-stack animating components (PolicyNonEmpty flights,
`newAnimatingComponent()`) use the same `play()` path; their cleanup
(`clearAnimatingComponents()`) cancels any still-live animations, which resolves
their settled promises rather than orphaning expectations.

`animateBetween()` (cross-screen flights) moves to the same mechanism: one WAAPI
animation on the flying clone, whose `finished` promise replaces the current
`durationMs + 100` cleanup timer.

### 4. `boardgame-render-game` — owns the gate

`_activeAnimations` map + event counting is replaced by awaiting the animator's
promise plus any renderer-initiated gated animations. `all-animations-done` fires
when that combined promise resolves. The flow `all-animations-done` →
`ready-for-next-state` → install next version stays structurally as-is.

The 15s watchdog shrinks to a **~4s last resort** that logs component diagnostics
and force-releases the gate. Invariant: the watchdog firing is a bug, and the
Playwright gate asserts it never fires.

## Interruption semantics

When a new state version must install while animations are live (admin fast-
forward, rapid fixup chains, version skipping via #717's SkipAnimation):

- `animator.prepare()` for the new cycle first calls `finish()` on every live
  animation (jump to end state, resolve `finished`), so measurement sees resting
  positions. Infinite/scheduled-but-unstarted animations are `cancel()`ed instead.
- `settled()` treats both `finish()` and `cancel()` as settlement; promise
  consumers never hang and never see rejections (cancel rejections are caught
  internally).

## Renderer-facing API

Event names and meanings survive (`will-animate`, `animation-done`,
`all-animations-done`) — their semantics simply become trustworthy. What changes
for game renderers:

- **`delayAnimation(fromMove, toMove)`** (imperative, move-name-keyed) is
  **removed**, replaced by declarative attributes (below). All in-repo renderers
  and `../games` renderers are migrated.
- **`post-animation-delay="<ms>"`** attribute on any animatable item (#715): maps
  to WAAPI `endDelay`; `animation-done` fires after the hold. (If `endDelay`
  proves quirky cross-browser, implementation may chain an explicit delay promise —
  behavior, not mechanism, is the contract.)
- **`wait-for-animation`** boolean attribute (#716): sets the default `gated` value
  for that item's animations. Defaults: components `true`, `status-text` /
  `fading-text` `false`.
- **`stagger="<fraction>"`** attribute on `boardgame-component-stack` (#728):
  children animating in the same cycle get `delay = index * fraction *
  animationLength`. Overlap works because animations are independent WAAPI objects.
- **`isAnimating`** exposed on `boardgame-render-game` (#721):
  `boardgame-move-form` buttons auto-disable while the gate is open (opt out with a
  `no-animation-disable` attribute); `propose-move` events arriving while animating
  are rejected with a console warning (belt and suspenders).
- `computeAnimationProps()` / component subclass hooks survive with the same
  intent but now return keyframe contributions instead of setting transition
  styles; card/token/die in-repo subclasses are the reference implementations.

## Companion sync consumption (#798, final phase)

The `companion-sync.ts` estimator and server `serverPlayAt` stamps exist; this
phase makes them load-bearing:

- `boardgame-game-state-manager` schedules installation of queued animated
  versions at `localEquivalent(serverPlayAt)` (clamped: never earlier than "now",
  never later than now + 2s) instead of installing immediately on
  `ready-for-next-state`.
- Auto-fly cross-screen animations in `hand-view-base` / `table-view-base`
  schedule their `play()` at the same timestamp, so phone and table surfaces start
  flights within estimator error (~1 frame on LAN) instead of drifting up to ~1s.
- Verdict/status text that reveals outcomes is gated on `all-animations-done`
  (now trustworthy) so no surface announces results before the cards land.
- The FLIP↔animateBetween handoff blip (≤2 frames) is addressed by having the
  handoff share one Animation timeline where feasible; if not fully fixable it is
  documented and tracked, not silently shipped.

murdermrmonroe (companion-mode game in `../games`) is the end-to-end validation
bed for this phase.

## Migration scope

- **In-repo example game clients** (blackjack, memory, debuganimations,
  tictactoe, pig, and any other `examples/*/client` with renderers).
- **`../games` clients:** murdermrmonroe, pass, valentine (`<game>/client/*.ts`).
  Changes committed in that repo alongside this branch's changes.
- Anything using `delayAnimation`, transition-based animation hooks, or listening
  for raw `transitionend` on components gets migrated to the new attributes/API.

## Error handling

- Watchdog (4s): logs pending item tags/ids + their live Animation states, then
  force-releases. Firing = bug.
- `settled()` never rejects; `Animation.finished` rejections from `cancel()` are
  caught at the item layer.
- Scheduled (companion) installs falling >2s behind are installed immediately with
  a console warning — the game must never hang on a bad latency estimate.
- Unknown/missing animation targets (id lookups) keep the recent console.warn
  behavior; a missed flight never blocks the gate.

## Testing

**Playwright regression gate** (extends `server/static/tests/animations/`),
running against `boardgame-util serve --offline-dev-mode`:

1. **debuganimations:** the #720 repro (MoveAllComponents + undo, repeated ×10) —
   assert zero watchdog console.errors, `all-animations-done` within budget each
   cycle, versions install in order.
2. **blackjack:** fresh game deal ×10 — the historical ~20% wedge scenario.
3. **memory:** flip pair + hide (timer-driven fixup chain) — rapid sequential
   fixups.
4. **Feature assertions:** stagger produces monotonically increasing start times;
   `post-animation-delay` defers `animation-done`; buttons disabled while
   animating; `wait-for-animation=false` items don't hold the gate.
5. **Companion phase:** murdermrmonroe table+hand two-page test — deal flight
   start-time skew below threshold; verdict text never precedes card landing.
6. Type-check (`tsc`) and `vite build` stay green; Go tests unaffected.

Tests assert on instrumentation counters exposed via a `window.__bgAnimTestHooks`
debug object (gate open/close timestamps, watchdog count, per-item play/settle
log) rather than screenshots.

## Phasing (implementation order)

1. **Phase A — timing core:** rewrite `BoardgameAnimatableItem` around WAAPI
   `play()`/`settled()`; add test hooks; watchdog to 4s.
2. **Phase B — cutover:** component/card/token/die keyframes; animator
   INVERT/PLAY → `play()`; render-game gate on promises; delete expectation
   machinery; Playwright gate for scenarios 1–3 green.
3. **Phase C — features + migration:** #715/#716/#721/#728 attributes; remove
   `delayAnimation`; migrate in-repo + `../games` renderers; feature tests.
4. **Phase D — companion sync (#798):** scheduled installs, synced auto-fly,
   gated verdict; murdermrmonroe two-page test.

Each phase lands as one or more commits on `animation-waapi-timing`; the branch
stays green (type-check + existing tests) at every commit.

## Issues resolved / affected

| Issue | Outcome |
|---|---|
| #726 | Superseded: observation via WAAPI promises instead of transitionstart |
| #720 | Root cause removed; watchdog kept as 4s last resort + test invariant |
| #713/#714 | Ownership refactor landed as the four-layer split |
| #715 | `post-animation-delay` attribute (endDelay) |
| #716 | `wait-for-animation` attribute (gated opt) |
| #721 | Auto-disable via reliable `isAnimating` |
| #728 | `stagger` attribute; overlap native to WAAPI |
| #798 | Companion phase: scheduled installs + synced flights + gated verdict |
