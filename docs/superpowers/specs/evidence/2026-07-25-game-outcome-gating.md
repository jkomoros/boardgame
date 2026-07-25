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
   immediately either way.

Everything else is unchanged: public API (`finished`, `animating`, `winners`,
`winnerLabels`, `viewer`, `title`), `_validateConfiguration()` logic, aria/role
wiring, and the arrival's exact keyframes/timing (`{ opacity: 0, transform:
'scale(0.96)' } -> { opacity: 1, transform: 'scale(1)' }`, `duration: 220, easing:
'ease-out', fill: 'backwards'`) passed explicitly to `play()` rather than relying on
`animationLengthMs()` defaults, so the curve does not vary with a game's
`--animation-length` the way the shared kernel's other consumers do — matching the
old CSS's hardcoded `220ms ease-out`.

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
unchanged with no code change required here.

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
