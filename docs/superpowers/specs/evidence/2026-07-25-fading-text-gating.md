# Evidence pack: `boardgame-fading-text` gains gate participation; reduced-motion becomes a duration-0 instant play

**Task:** Task 4 of the animatable-item unification plan
(`docs/superpowers/specs/2026-07-24-animatable-item-unification-design.md`, Phase 1).

**Claim:** `boardgame-fading-text` now extends `BoardgameAnimatableItem` and animates
its fade through the gated WAAPI `play()` kernel instead of self-driven CSS
`@keyframes`. Two observable behavior changes follow, both pre-declared and approved
by the spec; the visual motion itself (opacity/scale curve, duration, easing) is
byte-for-byte preserved.

## What changed observably

1. **Gate participation (the literal #714 ask).** The design doc's Phase 1 section
   states the fade "becomes a gate participant (it fires `will-animate`/`animation-done`),
   which is the literal ask of #714. Games can opt out per-instance with
   `wait-for-animation="false"`." Before this change, `boardgame-fading-text`'s row in
   the doc's current-state inventory read: "CSS `@keyframes fadetext` +
   `animationend`/rAF re-arm | never" (gated: never). After, every fade dispatches a
   bubbling composed `will-animate` then `animation-done` pair and increments the
   shared `animHooks.plays`/`settles` counters — proven by the new
   `tests/animations/parity/fading-text.spec.ts`, which fails against the
   pre-migration CSS implementation (no such events exist) and passes after.
   `boardgame-status-text` (the wrapper most games actually use) is the doc's very
   next migration in the same Phase 1 list — "extends `BoardgameAnimatableItem`... it
   inherits the primitive so it participates in discovery/gating" — i.e. status-text
   and the other Phase-1 items derive from the same base class this task lands.

2. **Reduced-motion: duration-0 instant play, not a kernel skip (correction below).**
   The old CSS had an explicit `@media (prefers-reduced-motion: reduce) {
   .animating #message { animation-duration: 1ms; } }` block — the fade still ran,
   compressed to 1ms. **Correction (code review caught this): the new path does
   NOT skip the animation.** `resolveMotionTiming` (`src/motion/timing.ts:85-98`)
   has a dedicated `reducedMotion` branch that returns `kind: 'play'` (not
   `'skip'`) with `delay: 0, duration: 0` — i.e. `play()` still calls
   `element.animate(...)` and still returns a real `Animation`, it is just
   instantaneous (0ms active duration, `endDelay` preserved for callers that use
   it as a semantic hold). `animateFade()`'s `if (!anim) { this._visible = false;
   return; }` fallback is therefore **not** the reduced-motion path at all — `anim`
   is never null here; that branch is only reachable via `noAnimate`. So the
   accurate framing is: reduced motion goes from a 1ms CSS sprint to a 0ms WAAPI
   instant play — arguably a *closer* parity match than the old behavior (0ms is
   the more honest expression of "no perceptible motion" than an arbitrary 1ms),
   not a skip. This is still a declared, approved change (the spec's
   `boardgame-game-outcome` reduced-motion note and the parity README's
   accepted-blind-spots ledger both anticipate reduced-motion behavior differing
   from the pre-migration CSS and defer its verification to
   `waapi-play`/`waapi-companion` rather than a geometry golden), but the
   mechanism is an instant real animation, not an absence of one.

Everything else is unchanged: public API (`message`, `trigger`, `suppress`,
`autoMessage`, `announce`, `animateFade()`), `_validateConfiguration`/`_triggerChanged`
logic, aria/role wiring, retrigger-while-in-flight semantics (`finishAllAnimations()`
finishes the prior fade before the new one starts, same as the old generation-counter
reset), and the container's `.animating` CSS class name (now keyed off a private
`_visible` reactive property instead of `_animating`).

## Visual-parity proof: fixture golden passes unregenerated

`tests/animations/parity/geometry.spec.ts`'s `fixture: fading-text fade curve` test
mounts `boardgame-fading-text` directly (outside any game), fires two `trigger`
assignments, and fingerprints the resulting motion curve (displacement-normalized
opacity/transform progress at 5 fractions, plus declared `[duration, delay]` on a
25ms grid) against a checked-in golden
(`tests/animations/parity/goldens/geometry-fixture-fading-text.json`):

```json
{
  "opacity": [1, 0.62, 0.32, 0.09, 1],
  "timing": [250, 0]
}
```

Run against the migrated implementation **without regenerating the golden**:

```
$ npx playwright test tests/animations/parity/geometry.spec.ts -g "fading-text"
  ✓  1 [chromium] › ... fixture: fading-text fade curve (31.4s)
  1 passed (32.3s)
```

The `[250, 0]` timing pins that `play()`'s default duration (via
`animationLengthMs()` reading `--animation-length`, same variable the old CSS read)
and zero delay match exactly; the opacity curve's shape (ease-out fade to a scale-6
transform, sampled at 0/.25/.5/.75/1) matches within the suite's 0.08 tolerance. The
last sample being `1` (not `0`) is itself a parity signal: both the old CSS animation
(no explicit `fill-mode`, default `none`) and the new `play()` call (kernel default
`fill: 'none'`) revert to the base non-animated style once the active phase ends, so
neither implementation holds the end-state visually — this is preserved, not
introduced.

## Full regression run (goldens untouched)

```
$ npx playwright test tests/animations/parity/ tests/animations/waapi-gate.spec.ts tests/animations/waapi-play.spec.ts
  run 1: 19 tests, 18 passed, 1 failed (waapi-gate "same-cycle reinstall")
  run 2: 19 tests, 18 passed, 1 failed (trace.spec.ts "memory: reveal two cards")
  run 3: 19 tests, 19 passed
```

Both single-test failures reproduce identically (down to the exact received
numbers) against the **unmodified pre-migration checkout** (`git stash` + rerun,
including a `--repeat-each=3` loop that reproduced the "reveal two cards" 41-vs-21
`plays` mismatch on the bare baseline) and pass cleanly on isolated reruns — these
are pre-existing flakes in the offline-dev harness (memory's deck/grid setup has
some run-to-run nondeterminism the exact-count trace golden doesn't yet tolerate),
unrelated to this migration, not regressions. `trace.spec.ts`'s "memory: reveal two
cards" scenario deliberately exercises a single card reveal (not a scoring match),
so it never drives `boardgame-fading-text` and its golden was never expected to
move by this change regardless.

`npm run type-check`, `npm run type-check:strict` (274 pre-existing errors in
unrelated files, confirmed identical on `git stash`; none in
`boardgame-fading-text.ts` or `boardgame-animatable-item.ts`), and `npm run
test:unit` (234/234) all pass.
