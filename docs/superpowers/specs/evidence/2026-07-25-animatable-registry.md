# Evidence pack: ambient animatable-item registry (#714's non-component discovery gap)

**Task:** Task 9 of the animatable-item unification plan
(`.superpowers/sdd/task-9-brief.md`).

**Claim:** render-game's cycle-start reset (`_resetAnimating()`) now reaches
every `BoardgameAnimatableItem` in the tree, not just the shared component
animator's own stack-tracked components — closing #714's "non-component"
discovery gap for standalone items (a die, `boardgame-status-text`,
`boardgame-fading-text`, any game-authored token) — with no observable change
to any golden-covered scenario.

## What discovery covers now that it didn't before

Before this task, `_resetAnimating()` (`boardgame-render-game.ts`) only reset
the animator's own tracked stack components via
`_clearAllAnimatingComponents()` → `BoardgameComponentAnimator.
clearAnimatingComponents()`. Any `BoardgameAnimatableItem` living OUTSIDE that
stack bookkeeping — a `<boardgame-die>` sitting as a plain sibling in a game
renderer's template (see `examples/pig/client/boardgame-render-game-pig.ts`),
`boardgame-status-text`, `boardgame-fading-text`, or any future game-authored
animatable — was invisible to the reset. Nothing forced its leftover
animation to settle when a new cycle started, and nothing installed the new
cycle's `animationContext` on it directly (it only ever got the context
indirectly, via the ambient walk in `_ambientAnimationContext()`, exercised
lazily the next time it happened to call `play()`).

New pieces (`server/static/src/motion/animatable-registry.ts`):

- `AnimatableRegistry`: `register()`/`unregister()`/`items()`, backed by a
  `Set` for idempotent add/remove and a snapshot-returning `items()` so
  render-game's iteration is safe even if a caller's `finishAllAnimations()`
  side effect (via a synchronous DOM removal) mutates the registry mid-loop.
  Typed via a minimal structural interface
  (`{ finishAllAnimations(): void; animationContext: unknown }`) rather than
  importing `BoardgameAnimatableItem`, so `boardgame-animatable-item.ts` can
  import this module without forming a cycle.
- `BoardgameAnimatableItem.connectedCallback`/`disconnectedCallback`
  (`boardgame-animatable-item.ts`): walk up for the nearest ancestor exposing
  an `animatableRegistry` property, register on connect, unregister the
  SAME cached registry on disconnect (so unregistration still works after the
  element has already been detached and the ancestor chain is gone). The
  walk is the existing `_ambientAnimationContext()` walk, factored out into a
  shared `_ambientLookup(propName, isProvider)` helper — same DOM shape
  (crosses shadow roots and slots, per commit `7172dd24`), but a DIFFERENT
  provider predicate: `animationContext` needs the null-skip (every
  `BoardgameAnimatableItem` inherits that property defaulting to `null`, so
  presence alone would wrongly stop the walk at the nearest animatable
  ancestor rather than the real provider above it); `animatableRegistry` does
  NOT need it, because that property is declared only by real providers
  (render-game), never inherited — presence IS the provider signal.
- `boardgame-render-game.ts`: `readonly animatableRegistry = new
  AnimatableRegistry()`, and `_resetAnimating()` now iterates `items()`
  calling `finishAllAnimations()` and installing `this.animationContext`
  directly on each, mirroring what the animator already does for its own
  stack components — BEFORE opening the gate for the new cycle.

## Why goldens are unaffected

`finishAllAnimations()` (`boardgame-animatable-item.ts`) is a no-op for an
item with no live animations (`_liveAnimations` empty — nothing to iterate).
In every golden-covered parity scenario, by the time the NEXT cycle's
`_resetAnimating()` runs, every registered item from the PRIOR cycle has
already settled naturally (animations complete well within a normal turn
cadence). The registry sweep therefore touches nothing observable in steady
state: no extra `will-animate`/`animation-done`/`play`/`settle` hook is
recorded, no extra frame is spent, and directly assigning
`item.animationContext` is invisible to any hook (it is a plain property
write, not read until the item's own next `play()` call, which already
resolves the SAME value via the ambient walk regardless). Confirmed:
`git status --porcelain tests/animations/parity/goldens/` is empty before and
after this task's full changes, and the full parity + waapi-gate + waapi-play
+ waapi-companion sweep (below) is green.

## Unit tests (`src/motion/animatable-registry.test.ts`)

`node --test` coverage for `AnimatableRegistry` in isolation (jsdom-free —
the DOM-crossing walk itself is covered by the e2e layer below, per the task
brief's own note that shadow-DOM traversal can't be exercised from
`node:test`):

- register makes an item show up in `items()`
- register is idempotent (registering the same item twice does not
  duplicate it)
- unregister removes an item
- unregister is idempotent (an absent item, or a double-unregister, is a
  no-op — this matters because a real item unconditionally unregisters at
  disconnect even if no provider was ever found)
- `items()` with several registered items returns all of them
- `items()` returns a snapshot: mutating the registry after taking a
  snapshot does not affect the already-returned array
- iterating a snapshot while mutating the registry from inside the loop
  (simulating a `finishAllAnimations()` side effect that registers/
  unregisters something else) does not throw and visits every originally-
  snapshotted item exactly once

`npm run test:unit`: 254/254 passed (247 pre-existing + 7 new).

## E2E tests (`tests/animations/parity/registry.spec.ts`)

**Test 1 — discovery.** In pig, after game creation, `<boardgame-die>` sits
as a plain sibling in the renderer's template (not inside a
`boardgame-component-stack`) — exactly the "non-component" shape #714
describes. Reaches `<boardgame-render-game>` (deep-querying through shadow
roots) and asserts its `animatableRegistry.items()` contains an element with
`tagName === 'BOARDGAME-DIE'`.

**Test 2 — the interrupted-cycle race, driven deterministically.** The
brief's original framing ("click roll twice quickly") turned out to be
unreachable through the real UI: `MoveActionImplementation.reason`
(`src/moves/action.ts`) reports `'animation-running'` whenever
`this.animating` is true, so `canActivate` — and thus the die's own click
handler — stays blocked until render-game's gate has fully CLOSED. With no
`motionReleaseForTransition` configured for pig's renderer, gate-close
coincides with the die's animation having already settled naturally, so a
real second click can never land while the first roll's spin is still
running; there is no UI-observable window to click through (verified
empirically: real back-to-back clicks, waiting only for the button to
re-enable between them, never produced a leftover running animation, with or
without the registry fix — the race genuinely cannot be reached this way).

Directly synthesizing a second `state` install (`el.state = clonedState`)
was tried and rejected: `<boardgame-render-game>` is a managed child of
`boardgame-game-view`, which keeps reasserting its own Redux-derived `state`
value; a manually-injected differing value gets silently overwritten shortly
after, producing spurious extra animations unrelated to the fix under test
(reproduced and diagnosed during this task; not a viable test vehicle).

The test instead drives the exact mechanism directly, following the
established precedent already documented at `_resetAnimating()`'s call site
("animation tests deliberately reach it directly... to open the completion
gate in isolation") and demonstrated by `waapi-gate.spec.ts`'s "memory:
same-cycle state reinstall" test (which reaches into render-game internals
for the same reason: realistic timing can't reliably produce this shape):

1. Roll the die for real (clicking through the real UI) and wait for a
   genuinely running `Animation` on `#inner`. (Roughly 1 in 6 rolls lands on
   the SAME face already showing — `boardgame-die.ts`'s
   `_selectedFaceChanged` guards on `oldValue !== newValue` and plays no
   animation at all in that case — so the test retries the roll until one
   actually animates, rather than flaking on a no-op roll.)
2. Call `renderGame._resetAnimating()` directly — the same method
   `_stateChanged()` calls to open a new cycle, including on the
   interrupted-cycle path. This simulates "a new cycle's start lands while
   the die is still animating" precisely and deterministically, without
   contending with game-view's live pipeline.
3. Assert the captured animation's `playState` was `'running'` immediately
   before the call (proving the setup is real, not vacuous) and is no
   longer `'running'` immediately after.

**Verified as a genuine regression test**, not merely a passing assertion:
with the registry sweep temporarily removed from `_resetAnimating()`, this
test fails deterministically (3/3 runs) on
`expect(result.playStateAfterReset).not.toBe('running')` — `Expected: not
"running"`, i.e. the die's leftover animation was still running after the
simulated cycle start. Restoring the fix returns it to reliably passing
(6/6 runs, including 3 runs immediately preceding the regression check).

**Test 3 — registration lifecycle.** Fixture pattern from
`fading-text.spec.ts`: a plain `<div>` manually given an `animatableRegistry`
property, with a `<boardgame-fading-text>` mounted under it. Asserts the
fading-text is registered immediately after `appendChild` (connectedCallback
runs synchronously on insertion) and unregistered immediately after
`.remove()`.

## Full verification

```
$ npm run test:unit
  254/254 passed

$ npx tsc --noEmit -p .
  clean

$ npx playwright test tests/animations/parity/registry.spec.ts
  3/3 passed (repeated 6+ times for stability)

$ npx playwright test tests/animations/parity/ tests/animations/waapi-gate.spec.ts \
    tests/animations/waapi-play.spec.ts tests/animations/waapi-companion.spec.ts
  39 passed (3.4m) -- clean on the FIRST run, no reruns needed. None of the
  known flakes (waapi-gate same-cycle reinstall, trace pig-roll, geometry
  fixture page-load curves) triggered.

$ git status --porcelain tests/animations/parity/goldens/
  (empty -- no golden shifted)
```
