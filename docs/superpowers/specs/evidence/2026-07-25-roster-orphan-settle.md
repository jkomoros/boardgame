# Evidence pack: roster-hosted animatables have no orphan-settle path

**Task:** Phase 2 gate-finding follow-up (deferred at Phase 2 close, see
`.superpowers/sdd/progress.md`'s "PHASE 2 GATE CLOSED" entry): "roster
orphan-settle gap (UNRESOLVED, deferred: my inline fix attempt was reverted
after breaking specs -- MUST re-land as a proper red-first mini-task before/
with Phase 3)." This mini-task re-lands that fix with red-first discipline.

**Claim:** a board/stack component gets `beforeOrphaned()` (force-settle) from
the animator before removal — see `BoardgameAnimatableItem.beforeOrphaned()`
in `server/static/src/components/boardgame-animatable-item.ts`. A
roster-hosted animatable (gated via `boardgame-game-view.ts`'s
`_rosterWillAnimate`/`_rosterAnimationDone` pipe, Task 10) has **no**
equivalent caller. If such an element is removed from the DOM mid-animation
while admitted to the gate, two gaps compound:

1. Nothing force-settles it, so its WAAPI animation keeps running against the
   document timeline after detach (`Element.animate()`'s returned `Animation`
   is not implicitly cancelled by removing its target from the document).
2. Even once it does finish naturally, its `animation-done` `CustomEvent`
   (`bubbles: true, composed: true`) dispatches from a node with **no
   parent** — there is nothing to bubble to — so `render-game`'s gate
   listener (installed on itself, reached only via the game-view forwarding
   pipe) never hears it.

The gate is stuck open until the kernel's watchdog force-closes it
(`server/static/src/motion/animation-gate.ts`'s `DEFAULT_FLOOR_MS = 4000`,
extended further here by the probe's own declared `expectedSettleMs`), firing
a watchdog and logging a `console.error` in production.

## Post-mortem of the reverted attempt (verified, not just trusted)

A prior inline attempt (mentioned in the Phase 2 gate-close ledger) tried:
(1) in `disconnectedCallback`, a microtask-deferred
`if (!this.isConnected) this.finishAllAnimations()` guarded by
`_liveGatedCount > 0`; (2) in `_rosterWillAnimate`, after forwarding,
`ele.settled().then(() => renderGame.gateAnimationDone(...))` as a
detachment-proof done channel. It was reverted when a sweep went red, but
per the ledger the failures were later found to be **unrelated pre-existing
issues** (import collision + pig same-face rolls, since fixed
independently). This mini-task re-verified the approach from scratch under
strict red-first TDD rather than assuming the post-mortem's account — the
approach turned out to be sound (see Implementation below) and needed no
changes beyond what's implemented here.

One subtlety the post-mortem explicitly flagged as never checked: **does
`settled()`'s resolution ordering vs. the bubbled `animation-done` event
create any observable difference in gate-close timing for the normal
(attached) path?** Verified explicitly below — no, and the reasoning is
mechanical, not empirical-only.

## Red-first proof

Added a third test to `tests/animations/parity/player-info-gate.spec.ts`
("a roster animatable removed from the DOM mid-animation does not stall the
gate to the watchdog"): while a real `pig` board cycle is open (a real
Roll-die move genuinely animating, confirmed via the `animHooks` log — same
technique as the suite's first test), the roster-mounted
`boardgame-fading-text` probe is admitted to the gate with a real
`postAnimationDelay = 2500`, then removed from the DOM
(`(el as HTMLElement).remove()`) the instant its own `play()` genuinely
started (confirmed via the `animHooks` log observing a `play` record for
`boardgame-fading-text#task10-roster-probe`) — i.e. demonstrably mid-animation,
not before start (which the fading-text's own `isConnected` guard in
`animateFade()`'s continuation would just no-op) and not after natural
settle (which would prove nothing about the orphan path).

Run against current HEAD (`ffe1419a`, before this task's implementation),
the test failed exactly as the finding predicts — the gate is stuck open
until the watchdog fires, not merely late:

```
Error: gate closed 3754.399999856949ms after the die's own settle
(dieSettleAt=2562.800000190735, gateCloseAt=6317.200000047684); expected
the orphaned roster probe to be force-settled promptly rather than
stalling the gate to the watchdog

expect(received).toBeLessThan(expected)
Expected: < 1500
Received:   3754.399999856949
```

The ~3.75s gap matches the predicted mechanism precisely: the probe's
declared `expectedSettleMs` is `duration(~250ms, animationLengthMs()) +
endDelay(postAnimationDelay=2500ms) = ~2750ms` (see
`server/static/src/motion/timing.ts`'s `resolveMotionTiming`, `timing:
'immediate'` policy branch), which the gate kernel's `willAnimate()` uses to
extend the watchdog deadline to `now + expectedSettleMs + marginMs(1500) =
~4250ms` past the point the probe's `will-animate` was forwarded (see
`animation-gate.ts`). That is comfortably past both the 1500ms tolerance and
the die's own near-instant settle, and matches the observed gap. This is not
a typo or setup error — the failure is the exact watchdog-stall symptom the
finding describes, reproduced deterministically (confirmed across the run
captured above and consistent in shape run to run).

## Implementation

**Part 1 — reparent-safe disconnect settle**
(`boardgame-animatable-item.ts`'s `disconnectedCallback`): when
`_liveGatedCount > 0` at disconnect, defer a check to a microtask; if the
element is still disconnected then (`!this.isConnected`), call the existing
`finishAllAnimations()`. Deferring is load-bearing: Lit can disconnect then
immediately reconnect an element within the same synchronous span (a
same-tick reparent, e.g. moving an item to a new parent), which also fires
`disconnectedCallback`. Checking `isConnected` synchronously would
needlessly snap an in-flight animation on every such reparent; checking one
microtask later means only a genuine removal (still disconnected once the
microtask queue drains) triggers the force-settle.

**Part 2 — detachment-proof done channel**
(`boardgame-game-view.ts`'s `_rosterWillAnimate`): after the existing
`gateWillAnimate` forward, additionally subscribe to the admitted element's
own `settled()` promise (a public method already on
`BoardgameAnimatableItem`, resolving when `_liveGatedCount` returns to 0 —
the same bookkeeping that drives the bubbled `animation-done` event) and
forward a synthesized `animation-done`-shaped event through
`gateAnimationDone` when it resolves. Unlike the bubbled `CustomEvent`, a
promise resolution does not depend on the element still being attached to
any DOM tree, so it is exactly the channel the orphaned case needs. Both
parts are necessary and neither alone suffices: Part 1 without Part 2 would
finish the animation but still have no way to tell the gate (still detached,
still can't bubble); Part 2 without Part 1 would have a working channel but
nothing driving `_liveGatedCount` to 0 promptly (the animation would keep
running for its full real-time duration, only eventually resolving
`settled()` once its own timer naturally elapses — better than the
watchdog, but not "promptly").

### Why double-delivery is safe (verified against the kernel, not assumed)

In the **normal (attached, never removed)** case, both delivery paths now
fire for every roster gate participant: the bubbled `animation-done` (via
`_rosterAnimationDone`, unconditionally forwarded) and the new `settled()`
channel (via `_rosterWillAnimate`'s subscription). Both ultimately call
`AnimationGate.animationDone(ele)`
(`server/static/src/motion/animation-gate.ts`):

```ts
animationDone(ele: object): void {
  if (this.activeAnimations.size === 0) return;
  this.activeAnimations.delete(ele);
  if (this.activeAnimations.size === 0) {
    this.close(this.cycleId);
  }
}
```

A second call for the same `ele` is always a safe no-op: either
`activeAnimations` is already empty (the early return fires), or `ele` was
already deleted so `.delete(ele)` is a harmless no-op that does not change
`.size`, so the `size === 0` branch (and thus `close()`) is reached at most
once as a result of that participant. `close()` itself is additionally
guarded by `allDoneFired` (`if (this.allDoneFired) return;`), so even a
close-then-close sequence for the same cycle is idempotent. This holds
**regardless of delivery order** — confirmed by construction from the
kernel source above, not just by observing green tests.

### Verifying the ordering question explicitly (the post-mortem's unchecked subtlety)

Does `settled()`'s resolution order relative to the bubbled event change
observable gate-close timing in the **normal, attached** path? No — and the
reason is mechanical, traced through
`BoardgameAnimatableItem._animationSettled`:

```ts
private _animationSettled(anim: Animation, gated: boolean) {
  ...
  if (this._liveGatedCount <= 0) {
    this._liveGatedCount = 0;
    const resolvers = this._settledResolvers;
    this._settledResolvers = [];
    for (const r of resolvers) r();                 // (A) resolves settled() promises
    this.dispatchEvent(new CustomEvent('animation-done', ...)); // (B) bubbles synchronously
  }
}
```

Calling a promise's resolve function (A) does not run that promise's
`.then()` continuations synchronously — those are always scheduled as
microtasks, queued after the resolve call returns. `dispatchEvent` (B),
immediately following on the next line, invokes any bubbling listeners
**synchronously**, within the same call stack — including
`boardgame-game-view`'s `_rosterAnimationDone` listener, which calls
`gateAnimationDone` → `AnimationGate.animationDone()` right then. So for the
attached path, the bubbled event closes the gate (or removes the
participant) *before* the settled()-channel's `.then()` continuation even
gets a turn — that continuation only runs once the current synchronous
execution unwinds and the microtask queue is drained, i.e. strictly later
(by one or more microtask ticks, not a macrotask or animation frame — no
observable wall-clock difference). By the time it does run, the kernel call
it makes is a guaranteed no-op per the double-delivery argument above. This
was verified structurally (reading the exact resolve-then-dispatch ordering
in `_animationSettled`), not merely inferred from tests passing — though
the existing player-info-gate tests (both directions, unchanged, still
green — see Verification below) are consistent with it: the first test's
"gate must not close strictly before the roster participant's settle"
assertion continues to hold with the same tolerance as before this change.

## Verification

- **Red-first**: the new test failed against pre-fix HEAD with the exact
  watchdog-stall shape shown above (captured before any implementation
  code was written).
- **Green**: after implementing both parts, all three
  `player-info-gate.spec.ts` tests pass, including the two pre-existing
  ones (both directions) unchanged.
- **New test run 4x for stability**: 4/4 green, ~2.7-2.9s each (well under
  the suite's own generous ceiling and the assertion's 1500ms tolerance).
- **Full parity sweep**: `tests/animations/parity/` — 26/26 green. Goldens
  untouched (`git status` shows no golden-file diffs — only the two
  implementation files and the extended spec).
- **waapi-gate + waapi-play**: 9/9 green, including the known-flaky
  "memory: same-cycle state reinstall mid-gate" test (passed clean this
  run).
- **Unit suite** (`npm run test:unit`): 254/254 green.
- `tsc --noEmit` clean (no `any` introduced beyond the existing
  `e.detail.ele as HTMLElement`-style cast idiom already used by
  `_componentWillAnimate`; the new code casts to the concrete
  `BoardgameAnimatableItem` type instead, since `.settled()` is needed).

### Known flakes encountered (not caused by this change)

- `waapi-companion.spec.ts`'s "common play policy covers composed-tree
  providers and the full remaining budget" failed once with a 1ms rounding
  mismatch (`Expected: 600, Received: 599`) on an incidental extra sweep of
  that file (not part of this task's required parity/waapi-gate/waapi-play
  scope); it passed clean on immediate rerun as the sole failure. This test
  exercises a companion timing-budget calculation unrelated to roster
  gating, disconnect, or the settled() channel touched here — a pre-existing,
  unrelated timing-jitter flake, not a regression from this change.
- Documented pre-existing flakes from earlier phases (unaffected by this
  change, not re-triggered in this sweep): waapi-gate same-cycle reinstall;
  geometry fixture page-load curves.

## Conclusion

The reproduction confirms the Phase 2 gate finding: roster-hosted
animatables had no orphan-settle path, stalling the completion gate to the
4s+ watchdog when such an element is removed from the DOM mid-animation.
The two-part fix — a reparent-safe disconnect settle on the shared
`BoardgameAnimatableItem` base class, plus a detachment-proof `settled()`
done channel in `boardgame-game-view`'s roster forwarding — closes the gap
without disturbing the normal (attached) path's existing behavior or timing,
verified both by full-sweep regression and by tracing the exact
resolve-then-dispatch ordering that makes double-delivery observably
harmless.
