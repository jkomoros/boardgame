# Evidence pack: `boardgame-token` throb routes through ungated `play()`

**Task:** Task 7 of the animatable-item unification plan
(`docs/superpowers/specs/2026-07-24-animatable-item-unification-design.md`, Phase 1).

**Claim:** `boardgame-token`'s infinite highlight ("throb": a pulsing
`drop-shadow` on `#inner` while `active` and/or `highlighted`) now starts through
`BoardgameAnimatableItem.play()` instead of a self-driven CSS `animation-name:
throb` / `@keyframes throb` rule, deliberately routed **ungated**
(`{ gated: false, timing: 'immediate' }`). This is the mechanism inverse of
Tasks 4-6 (fading-text, game-outcome): those pin GATE PARTICIPATION as the
value CSS keyframes structurally cannot provide; this pins GATE
NON-PARTICIPATION as the value that must be preserved. `BoardgameToken` already
extended `BoardgameComponent` -> `BoardgameAnimatableItem`, so no class-hierarchy
change was needed — only the throb mechanism itself moves.

## Why ungated + immediate is correct for an ambient infinite highlight

The throb is not a state-arrival cue with a natural endpoint (unlike a fade-in,
a flip, or a verdict arrival) — it is an ambient, indefinitely-repeating
decoration that lasts as long as `active || highlighted` stays true, which can
be the entire remaining lifetime of the game. If it were gated
(`{ gated: true }`, the `play()` default), starting it would increment
`_liveGatedCount` and never decrement it on its own — `anim.finished` only
settles on `cancel()`/`finish()`, and an `iterations: Infinity` animation never
reaches its own natural finish. The completion gate (`animation-done`,
`isAnimating`, the render-game watchdog) would then wedge open for as long as
the throb kept running, and every *other* unrelated animation cycle that also
happens to hold the same gate would be blocked from ever reporting done until
the watchdog eventually force-closed it — turning an ambient highlight into a
latent hang for completely unrelated gameplay animations. `timing: 'immediate'`
is the companion choice: the throb has no reason to synchronize against a
render-game version slot (`policy: 'version'`'s stagger/clipping logic is for
coordinated multi-component cycles), and a standalone-mounted token (this
task's test fixture) has no `animationContext` to synchronize against anyway.

`play()`'s own accounting confirms this is safe by construction, not merely by
convention (`server/static/src/components/boardgame-animatable-item.ts:203-262`):

- `gated = (opts?.gated ?? true) && this.waitForAnimation;` — `gated: false`
  makes this `false` regardless of `waitForAnimation`.
- The `will-animate` event and `_liveGatedCount++` are inside `if (gated) {...}`
  (lines 225-238) — never reached for the throb.
- `animHooks.record('play', ...)` (line 240) is **unconditional** — it fires
  for every `play()` call regardless of `gated`. This is a load-bearing
  distinction the test asserts precisely: `hooks.plays` DOES increment for the
  throb even though no `will-animate` fires and the gate is never held.
- `_animationSettled()` (lines 264-282) also records `'settle'` unconditionally,
  but only decrements `_liveGatedCount` / dispatches `animation-done` `if
  (gated)` — for the throb, `gated` is `false`, so cancelling it later still
  records a `'settle'` hook entry but never touches the gate or dispatches
  `animation-done`.

So "ungated" concretely means: no `will-animate`, no `animation-done`, no
`_liveGatedCount` change — but `hooks.plays`/`hooks.settles` still move, because
those are unscoped instrumentation, not gate state. The test suite asserts
exactly this shape, not the naively-simpler "no hooks fire at all."

## A latent bug this task surfaced and fixed: `iterations: Infinity` was silently zeroed

`resolveMotionTiming()` (`server/static/src/motion/timing.ts`) is the shared
pure timing resolver every `play()` call funnels through. Before this task, it
contained:

```typescript
for (const field of ['delay', 'duration', 'endDelay', 'iterations'] as const) {
  const value = timing[field];
  if (typeof value === 'number' && !Number.isFinite(value)) timing[field] = 0;
}
```

This loop exists to protect the three genuinely-must-be-finite WAAPI fields
(`delay`, `duration`, `endDelay`) from NaN/Infinity leaking in from malformed
runtime input (the `timing.test.ts` case "never emits non-finite WAAPI timing
from malformed runtime input" pins exactly this for `delay`/`duration`). But it
lumped `iterations` into the same blanket guard — and `Number.isFinite(Infinity)
=== false`, so any caller requesting `iterations: Infinity` (WAAPI's own
"repeat forever" sentinel, and exactly what Task 7's sketch calls for) had it
silently rewritten to `iterations: 0` before ever reaching
`element.animate()`. An `Animation` created with `iterations: 0` has zero
active duration — it does not visibly run at all. No prior task exercised this
path (Tasks 4-6 all use small finite durations), so nothing had hit it before.

Fix (same file): only NaN, not Infinity, is malformed input for `iterations`
specifically — `delay`/`duration`/`endDelay` still get the original
finite-or-zero guard:

```typescript
for (const field of ['delay', 'duration', 'endDelay'] as const) {
  const value = timing[field];
  if (typeof value === 'number' && !Number.isFinite(value)) timing[field] = 0;
}
if (typeof timing.iterations === 'number' && Number.isNaN(timing.iterations)) {
  timing.iterations = 0;
}
```

Verified this is a targeted fix, not a behavior change for existing callers:
`timing.test.ts`'s malformed-input case never sets `iterations` (stays
`undefined`, untouched by either guard), and its finite-`iterations: 3` cases
(`compiles version wait...`, `preserves forward fill and clips repeated active
duration...`) were never at risk from the old guard either (`Number.isFinite(3)`
was already `true`). `npm run test:unit` (234/234) confirms no regression.

**Proof the fix is load-bearing, not incidental:** stashing just the
`timing.ts` change and re-running `token-throb.spec.ts` reproduces exactly the
predicted failure — `infiniteRunning` false (no `Animation` with
`getComputedTiming().iterations === Infinity` is ever actually running) and the
active/highlighted live-animation counts are `0` instead of `1`. Restoring the
fix turns both tests green with no other change. This is the TDD red/green
pair for the fix itself, on top of the red/green pair for the throb migration.

One consequence of the fix that is deliberately left as-is:
`effectiveIterations()` (used only to estimate `expectedSettleMs` for the
`will-animate` event's `detail`, and only reachable when `gated` is true) still
treats `Infinity` iterations as `0` via `finiteTimingMs()` — this would
under-report an *expected settle time* for a hypothetical **gated** infinite
animation. That path is intentionally not exercised: gating an infinite
animation is nonsensical for exactly the reason this evidence pack opens with
(it would never settle), so no caller should ever request `gated: true` with
`iterations: Infinity`, and none does.

## Keyframe/color-resolution parity

The old CSS:

```css
#outer.active #inner, #outer.highlighted #inner {
  animation-name: throb;
  animation-duration: 1s;
  animation-timing-function: ease-in-out;
  animation-direction: alternate;
  animation-iteration-count: infinite;
}
@keyframes throb {
  from { filter: drop-shadow(0 0 0.25em var(--throb-color-to)) drop-shadow(0 0 0.25em var(--throb-color-to)); }
  to   { filter: drop-shadow(0 0 0.25em var(--throb-color-from)) drop-shadow(0 0 0.25em var(--throb-color-from)); }
}
```

Only the `animation-*` declarations and the `@keyframes throb` block were
deleted. The `--throb-color-from`/`--throb-color-to` custom properties and
their per-state selectors (`#outer.active #inner`, `#outer.highlighted #inner`,
`#outer.active.highlighted #inner`) are untouched — they still carry the three
theme colors (olive for active-only, black for highlighted-only, yellow for
both) and are still selected purely by the existing `classMap` output, no new
plumbing.

WAAPI keyframes cannot resolve `var()` portably (`element.animate()` keyframes
are resolved at construction time against the CSSOM, not live-recomputed like
a CSS animation's declaration would be), so `_syncThrob()` resolves both colors
once via `getComputedStyle(inner).getPropertyValue(...)` at (re)start time and
bakes the literal color strings into the two `filter` keyframes — otherwise
identical to the CSS `@keyframes` (same `drop-shadow` doubling, same "to" is
the darker/`-from` color per the original code comment). Timing is passed
through `play()` explicitly (`duration: 1000, easing: 'ease-in-out', direction:
'alternate', iterations: Infinity`) rather than through `animationLengthMs()`
defaults, matching the CSS's hardcoded `1s` (the old rule never referenced
`--animation-length`).

**Restart-on-change semantics (same edge case as before, now explicit):**
`updated()` calls `_syncThrob()` whenever `active` or `highlighted` changes.
`_syncThrob()` unconditionally cancels any prior throb `Animation` before
deciding whether to start a new one. This matters even when the throb-state
boolean (`active || highlighted`) stays `true` across a change — e.g.
`active: true, highlighted: false` -> `active: false, highlighted: true` keeps
throbbing the whole time in both old and new implementations, but the *color*
changes (olive -> black), so the animation must restart to pick up the new
`getComputedStyle` values. The legacy CSS had the identical restart-like
discontinuity implicitly (the browser recomputes `var()` inside a running CSS
animation on a best-effort, engine-dependent basis at keyframe boundaries) —
this migration makes the same tradeoff explicit and deterministic instead of
engine-dependent.

`disconnectedCallback()` cancels the throb before calling `super.disconnectedCallback()`,
so removing a still-throbbing token from the DOM does not leak a running
infinite `Animation` (`beforeOrphaned()` on the shared base class is a
gate-settlement hook for gated animations and would not touch this ungated
one — this class needed its own disconnect cleanup).

## Test evidence

New `server/static/tests/animations/parity/token-throb.spec.ts`, mounted
standalone on `/` (same fixture pattern as `fading-text.spec.ts` /
`game-outcome.spec.ts`, lighter than a full debuganimations e2e):

1. **Live infinite Animation.** Deep-walks `#inner`'s own
   `getAnimations({ subtree: false })` (not `document.getAnimations()`, which
   does not see into shadow roots) and confirms a `running` `Animation` whose
   `effect.getComputedTiming().iterations === Infinity` exists after setting
   `highlighted = true`.
2. **Ungated.** No `will-animate`/`animation-done` ever fires; `el.isAnimating`
   (`_liveGatedCount > 0`) stays `false`; a standalone mount never touches
   `hooks.gateOpens` (that counter is a render-game watchdog concept, entirely
   separate from per-item gate-holding — confirmed by reading
   `boardgame-render-game.ts:631/727`, the only two call sites). `hooks.plays`
   DOES increase (unconditional instrumentation, per the mechanism section
   above) — asserted as `toBeGreaterThan(0)`, not asserted to stay flat, so the
   test cannot be satisfied by an implementation that skips `play()` entirely.
3. **Clearing state cancels it.** Setting `highlighted = false` (with `active`
   already false) leaves zero live animations on `#inner`.
4. **Disconnect cancels it.** Removing the element from the document after
   re-arming the throb also leaves zero live animations.
5. A second test independently exercises `active`-only and `highlighted`-only
   (not just `highlighted`), confirming both state sources reach the same
   `_syncThrob()` path and that clearing back to neither state stops it.

```
$ npx playwright test tests/animations/parity/token-throb.spec.ts
  ✓ highlighting a token starts a live, ungated, infinite throb
  ✓ active-only and highlighted-only both throb; clearing both stops it
  2 passed
```

Confirmed red before the `timing.ts` fix (see the "latent bug" section above)
and red before the `boardgame-token.ts` migration (manually verified by
stashing each file independently) — genuine TDD red/green for both changes
this task made.

## Full regression run (goldens untouched)

```
$ npx playwright test tests/animations/parity/ tests/animations/waapi-gate.spec.ts tests/animations/waapi-play.spec.ts
  24 passed, 1 failed (first run)
```

The sole failure was `waapi-gate.spec.ts`'s `memory: interrupted cycles at
game creation close every gate they open` (a `waitForAnimationCounterStability`
timeout) — not the specific "same-cycle reinstall" case the task brief
flagged as known-flaky, but the same file/category of pre-existing timing
flake, unrelated to tokens or throbs (the scenario is memory-game creation
cycles; nothing in this task touches memory or game-creation timing). Re-run
in isolation:

```
$ npx playwright test tests/animations/waapi-gate.spec.ts -g "memory: interrupted cycles at game creation"
  ✓ 1 passed
```

Passed clean on rerun, confirming it is the pre-existing flake, not a
regression from this task's changes.

`git status --porcelain tests/animations/parity/goldens/` shows no changes —
the throb is ungated and unhooked, and no golden-covered scenario mounts a
highlighted/active `boardgame-token`, so no golden needed regeneration.

`npm run type-check` is clean. `npm run test:unit` passes 234/234 (no
unit-test surface touches `boardgame-token.ts` or the throb; the `timing.ts`
fix is covered by its own existing `timing.test.ts` suite, all of which still
passes unchanged).
