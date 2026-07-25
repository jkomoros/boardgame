# Evidence pack: `boardgame-game-outcome` gains gate participation

**Task:** Task 6 of the animatable-item unification plan
(`docs/superpowers/specs/2026-07-24-animatable-item-unification-design.md`, Phase 1).

**Claim:** `boardgame-game-outcome` now extends `BoardgameAnimatableItem` and plays
its verdict arrival through the gated WAAPI `play()` kernel instead of a self-driven
CSS `@keyframes` animation. The visual motion itself (opacity 0->1, `scale(0.96)` ->
`scale(1)`, 220ms `ease-out`, `fill: backwards`) is byte-for-byte preserved. The one
observable behavior change is pre-declared and approved by the spec, same shape as
Task 4's fading-text migration.

## What changed observably

1. **Gate participation (the literal #714 ask).** Before this change, the arrival
   ran as `animation: outcome-arrive 220ms ease-out both;` directly in the component's
   CSS: it fired no DOM events and never touched the shared animation gate. After,
   the arrival fires a bubbling composed `will-animate` then `animation-done` pair on
   the element and increments the shared `animHooks.plays`/`settles` counters — proven
   by the new `tests/animations/parity/game-outcome.spec.ts`, which fails against the
   pre-migration CSS implementation (no such events exist; confirmed red before
   implementing) and passes after.

2. **Reveal gating is unchanged; it is now ALSO a gate participant.** The verdict's
   reveal was already gated by the `animating` property mirror: `render()` returns
   `null` while `!this.finished || this.animating`, so the outcome section never
   appears mid-cycle (#798) — the renderer flips `animating` to `false` only once the
   settling cycle has finished. That gate is untouched by this migration. What is new
   is that the arrival animation itself — which fires once the reveal gate opens —
   now ALSO holds the shared `will-animate`/`animation-done` completion gate via
   `BoardgameAnimatableItem.play()`, the same primitive every other Phase 1 migration
   target uses. Two independent gates, not one collapsing into the other: the
   presence gate (via `animating`) still controls whether `#outcome` exists at all;
   the new participation gate (via `play()`) controls whether downstream watchers of
   `animation-done` see the arrival as in-flight.

3. **Reduced-motion: duration-0 instant play, not a skip.** The old CSS had an
   explicit `@media (prefers-reduced-motion: reduce) { #outcome { animation: none; } }`
   block — no motion, no delay, but the DOM class list carried no signal either way.
   Per `src/motion/timing.ts:85-98`, `resolveMotionTiming`'s `reducedMotion` branch
   returns `kind: 'play'` (never `'skip'`) with `delay: 0, duration: 0` and endDelay
   preserved — `play()` still calls `element.animate(...)` and returns a real
   `Animation`, just an instantaneous one. So reduced motion goes from "no animation
   runs at all" to "an animation runs and settles in the same frame" — a declared,
   approved behavior change (same shape as fading-text's Task 4 correction), not a
   silent regression: the verdict still reaches its final opacity/scale state
   immediately either way. (This branch is checked before the timing policy
   described in point 4 below, so it applies regardless of which policy the call
   uses.)

4. **Timing policy is pinned to `'immediate'`, not the kernel's `'version'` default
   (review finding, fixed).** The initial implementation passed no `opts` to
   `play()`, so `timingPolicy` defaulted to `'version'`
   (`boardgame-animatable-item.ts:208`). `resolveMotionTiming`'s `'version'` branch
   (`src/motion/timing.ts:111-134`) treats a requested `duration` as a REQUEST, not
   a guarantee: when a still-usable ambient `VersionAnimationContext` is present
   (via `_ambientAnimationContext()`'s walk to a populated provider — the real
   embedding is game-outcome inside the renderer's shadow root inside the
   `boardgame-render-game` host, which carries a populated `animationContext`
   cleared only on the next bundle), the branch CLAMPS the active duration to
   whatever remains of `context.maxAnimationDurationMs` and can inject a nonzero
   `delay`. The old CSS `animation: outcome-arrive 220ms ease-out both;` had no
   such dependency: declared directly on the element, it always ran exactly 220ms
   regardless of any render-game cycle happening to be in flight at reveal time.
   Passing an explicit `duration: 220` alone does not restore this guarantee — the
   `'version'` policy can still shorten an explicit duration; only pinning the
   TIMING POLICY itself to `'immediate'` does. Per `src/motion/timing.ts:135`, the
   `'immediate'` policy's branch is a no-op (no clamp, no delay injection), so the
   requested timing passes through unchanged. The fix passes `{ timing: 'immediate'
   }` as the 4th (`opts`) argument to `play()`; `gated` still defaults to `true`
   independent of `timing`, so the arrival continues to hold the shared completion
   gate exactly as before. Proven by a new nested-provider regression test (below)
   that fails against the pre-fix `'version'`-default call — observed a clamped
   duration of 96ms against a provider whose `maxAnimationDurationMs` is 100 — and
   passes once `timing: 'immediate'` is added.

Everything else is unchanged: public API (`finished`, `animating`, `winners`,
`winnerLabels`, `viewer`, `title`), `_validateConfiguration()` logic, aria/role
wiring, and the arrival's exact keyframes (`{ opacity: 0, transform:
'scale(0.96)' } -> { opacity: 1, transform: 'scale(1)' }`, `duration: 220, easing:
'ease-out', fill: 'backwards'`), matching the old CSS's hardcoded `220ms ease-out`
exactly and — with `timing: 'immediate'` — matching its context-independence too.
(Correction to an earlier draft of this pack: the claim that an explicit `duration`
means "the curve does not vary with a game's `--animation-length`" was true but
incomplete. It explained why the duration isn't *defaulted* from
`animationLengthMs()`, not why it can't be *reshaped* by an ambient version
context — that required the separate `timing: 'immediate'` policy fix in point 4
above.)

## Implementation

`updated()` latches on `_arrivalPlayed` so the arrival plays exactly once per reveal
and resets when un-revealed:

```typescript
override updated(changed: Map<PropertyKey, unknown>) {
  super.updated(changed);
  const revealed = this.finished && !this.animating;
  if (revealed && !this._arrivalPlayed) {
    this._arrivalPlayed = true;
    const outcome = this.renderRoot.querySelector('#outcome') as HTMLElement | null;
    if (outcome) {
      this.play(outcome, [
        { opacity: 0, transform: 'scale(0.96)' },
        { opacity: 1, transform: 'scale(1)' },
      ], { duration: 220, easing: 'ease-out', fill: 'backwards' });
    }
  }
  if (!revealed) this._arrivalPlayed = false;
}
```

`updated()` fires synchronously after `render()` has applied the new DOM within the
same Lit update, so `#outcome` already exists by the time this queries for it — no
`updateComplete` wait is needed here (unlike, e.g., a case that must wait for a
slotted child to project).

**Why no generation-token guard, unlike fading-text's Task 4/5 fixes:** fading-text's
`animateFade()` defers its `play()` call through `updateComplete.then(...)`, and a
mid-flight retrigger's `finishAllAnimations()` force-settles the prior play's own
still-pending `.finished.finally()` closure — creating a real race between two
async continuations that a boolean alone can't disambiguate (hence the generation
counter added in 12912447/7172dd24). This reveal gate has no equivalent async gap:
`play()` is called synchronously inside `updated()`, and once `_arrivalPlayed` latches
true there is no code path that re-enters this branch until `revealed` goes false and
back to true again (a full un-reveal/re-reveal cycle, which itself resets the latch
correctly). There is no second concurrent continuation for a stale closure to belong
to, so a plain boolean latch is sufficient.

**Ambient-context walk-past-null (7172dd24):** unaffected by this task. That fix
lives entirely in `BoardgameAnimatableItem._ambientAnimationContext()` and this
migration adds no new wrapper layer between `boardgame-game-outcome` and its
render-game provider — it is a direct `BoardgameAnimatableItem` subclass, same
shape as `boardgame-fading-text`, so the existing walk-past-null behavior applies
unchanged with no code change required here. It is, however, exactly the mechanism
that makes point 4 above a live risk rather than a theoretical one: the walk finds
whatever populated context is nearest, and the render-game host is a realistic
source of one at reveal time.

## Regression coverage: gate participation and context-independence

`tests/animations/parity/game-outcome.spec.ts` carries two tests:

1. **Gate participation** (unchanged from the original submission): mounts the
   component standalone, sets `finished`/`winners`, and asserts the
   `will-animate`/`animation-done` event pair and matching `animHooks`
   plays/settles deltas.

2. **Context-independence under a populated ambient version context** (new, added
   for the review fix): mounts a plain provider `div` whose `animationContext` is
   populated with `maxAnimationDurationMs: 100` — deliberately less than 220, so
   the default `'version'` policy would clamp if consulted — appends
   `boardgame-game-outcome` as its direct child (so `_ambientAnimationContext()`'s
   walk finds the provider immediately via `parentNode`), reveals the verdict, and
   asserts the started `Animation`'s `effect.getComputedTiming().duration` (deep
   walk to `#outcome` inside the shadow root) is exactly `220`. Red-first proof:
   temporarily removing the `{ timing: 'immediate' }` opts and rerunning this test
   produced `Expected: 220, Received: 96` (the clamped value:
   `maxAnimationDurationMs: 100` minus the small elapsed-time slice between
   `startAtMs` and the actual reveal) — confirming the test genuinely exercises the
   clamp and that the fix genuinely defeats it.

```
$ npx playwright test tests/animations/parity/game-outcome.spec.ts
  2 passed (2.1s)
```

## Visual-parity proof: fixture golden passes unregenerated

`tests/animations/parity/geometry.spec.ts`'s `fixture: game-outcome arrival curve`
test mounts `boardgame-game-outcome` directly (outside any game), sets
`finished = true; winners = [0]`, and fingerprints the resulting motion curves
against the checked-in golden
(`tests/animations/parity/goldens/geometry-fixture-game-outcome.json`), which
carries three distinct curves sampled from the `/` fixture page (the outcome
element's own opacity+scale arrival plus two unrelated ambient curves that happen
to be running on the served app shell at fixture-load time — the golden's set
semantics tolerate that). The arrival curve of interest:

```json
{
  "transform": [0, 0.38, 0.68, 0.91, 1],
  "opacity": [0, 0.38, 0.68, 0.91, 1],
  "timing": [225, 0]
}
```

Run against the migrated implementation **without regenerating the golden**:

```
$ npx playwright test tests/animations/parity/geometry.spec.ts -g "game-outcome"
  ✓  1 [chromium] › fixture: game-outcome arrival curve
  1 passed (2.4s)
```

The `[225, 0]` timing (220ms rounded to the suite's 25ms grid, zero delay) and the
ease-out opacity/scale progression match exactly, confirming the WAAPI `play()` call
reproduces the CSS `@keyframes` curve bit-for-bit.

**Pre-existing flake, unrelated to this change:** this fixture's golden carries
three curves (the outcome arrival plus two unrelated ambient curves that happen to
be mid-flight on the `/` shell page at fixture-load time — see the golden comment
above). Rerunning this test in a loop surfaced roughly 1-in-4 failures where only
the outcome's own curve was observed and the two ambient curves were absent
(`sampleMotionCurves`'s quiet-window wave detection missing whichever page-load
animation those ambient curves come from). Confirmed present identically on the
already-committed pre-fix implementation via `git stash` + rerun (4 runs, 1
failure, same missing-curves signature) — this is a harness/page-load timing flake
predating and unrelated to both the original migration and this fix; the migrated
component's own curve (`timing: [225, 0]`) matched on every single run, fix or no
fix.

## Full regression run (goldens untouched)

```
$ npx playwright test tests/animations/parity/ tests/animations/waapi-gate.spec.ts tests/animations/waapi-play.spec.ts
  23 passed (2.6m)
```

All 23 tests passed on the first run, including `waapi-gate.spec.ts`'s known-flaky
"same-cycle reinstall" case — no rerun was needed.

`git status --porcelain tests/animations/parity/goldens/` shows no changes: every
golden, including `geometry-fixture-game-outcome.json`, is untouched.

`npm run type-check` is clean. `npm run test:unit` passes 234/234, unaffected by this
change (no unit-test surface touches this component).
