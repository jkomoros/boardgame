# Ambient-animation sweep regression — evidence pack

**Change:** the cycle-start "force-finish stale animations" sweeps
(`render-game._resetAnimating` registry sweep and the shared animator's
`prepare()`) switch from `finishAllAnimations()` to a new
`finishGatedAnimations()`, so an UNGATED ambient loop (an infinite highlight
throb) survives a state cycle instead of being cancelled every move. The
`boardgame-token` also re-arms its throb on `connectedCallback` so a reparent
cannot leave it permanently glow-less.

Realizes the fix for the regression identified in the whole-branch prosecution
brief (Charge 1).

---

## 1. The regression (as convicted)

Prior to this branch, a token's highlight throb (the drop-shadow pulse while
`active`/`highlighted`) was a class-driven CSS `@keyframes throb`
(`animation-iteration-count: infinite`). CSS animations are declarative and
immune to any JS sweep, so the pulse ran continuously across any number of
moves.

Commit `9f80de0c` re-homed the throb onto the WAAPI `play()` kernel as a
tracked `Animation` (`{ gated: false }` — ambient, must never hold the
completion gate). Commit `49f398b4` added the ambient `AnimatableRegistry`,
whose cycle-start sweep (`_resetAnimating`) called `item.finishAllAnimations()`
on **every** registered item. `finishAllAnimations()` calls `anim.finish()`,
which throws `InvalidStateError` for an infinite animation and falls through to
`anim.cancel()` — so the throb was **cancelled at every state cycle**. Because
the move did not change `active`/`highlighted`, `updated()`→`_syncThrob()` never
re-armed it. Net: **a highlighted token stopped glowing the moment ANY move
was made** (and, since the throb's `play()` used `fill:'none'` and `_syncThrob`
had cleared `inner.style.filter`, the glow reverted to nothing — the highlight
affordance vanished, not merely stopped pulsing).

Framework-wide: `active`/`highlighted` are the general selection/cue states on
`boardgame-token` (tokens, pawns, meeples, cubes, discs). Every game that
highlights a piece lost that highlight on the first move.

### Mechanism (file:line, on the pre-fix branch HEAD)

- `boardgame-animatable-item.ts` `connectedCallback` — every
  `BoardgameAnimatableItem` (tokens included) registers with the ambient
  registry.
- `boardgame-token.ts` `_syncThrob()` — starts the infinite throb via `play(…,
  { iterations: Infinity }, { gated:false })`; invoked ONLY from `updated()`
  on an `active`/`highlighted` change.
- `boardgame-render-game.ts` `_resetAnimating()` — cycle-start sweep called
  `item.finishAllAnimations()` on every registered item, unguarded.
- `boardgame-animatable-item.ts` `finishAllAnimations()` — `finish()` throws
  for the infinite throb, `catch`→`cancel()` destroys it.

### Second family member — reparent

The throb is cancelled in `boardgame-token.disconnectedCallback` and otherwise
re-armed only on an `active`/`highlighted` change. Lit does **not** re-render on
a reparent, so a synchronous move of a still-highlighted token to a new
container fired `disconnectedCallback` (cancel) then `connectedCallback` (no
re-arm) and left the token throb-less forever. The retired CSS throb survived
reparenting automatically (class-driven). Verified empirically red-first (see
§3).

---

## 2. The remedy

### Gated-only sweep semantics

- `boardgame-animatable-item.ts`: `_liveAnimations` becomes
  `Map<Animation, boolean>` (the value is the per-animation `gated` flag, known
  at `play()` time). New `finishGatedAnimations()` force-settles only the gated
  entries; `finishAllAnimations()` keeps its everything semantics; both share a
  private `_forceSettle(anim)` helper.
- `boardgame-render-game.ts` `_resetAnimating` and
  `boardgame-component-animator.ts` `prepare()` (the two cycle-interruption
  sweeps) call `finishGatedAnimations()`. A stale cycle's job is to end its own
  GATED participants; an ungated ambient loop was never a cycle participant. In
  the animator's case the throb pulses `#inner`'s `filter`, not the host
  transform, so it is irrelevant to resting-position measurement.
- `AnimatableRegistry`'s `RegistrableAnimatableItem` interface now declares
  `finishGatedAnimations()` (the method the sweep actually calls).

### `finishAllAnimations` caller audit (deliberate, per call site)

| Caller | Decision | Why |
|---|---|---|
| `render-game._resetAnimating` (registry sweep) | → `finishGatedAnimations` | cycle interruption; ambient loops must survive |
| `component-animator.prepare()` (stack sweep) | → `finishGatedAnimations` | same cycle-interruption rationale; throb (filter) is irrelevant to resting-position measurement |
| `animatable-item.disconnectedCallback` (microtask) | keep `finishAllAnimations` | element left the tree — an ambient loop against the document timeline is pure waste; kill it |
| `animatable-item.beforeOrphaned()` | keep `finishAllAnimations` | about to be removed — same tree-departure rationale |
| `fading-text.animateFade()` (self-retrigger) | keep `finishAllAnimations` | self-scoped; fading-text owns no ungated ambient loop, so equivalent today; "finish everything before I restart" is the clearer intent |

### Reparent re-arm

`boardgame-token.connectedCallback` calls `super.connectedCallback()` then
`this._syncThrob()`. Safe on first connect (`innerElement` is null pre-render,
so `_syncThrob` no-ops; the first render's `updated()` starts it as before); on
a reparent it restarts the throb the persisted `active`/`highlighted` still
demand.

---

## 3. Red-first evidence

All three tests were watched failing on the unfixed branch HEAD before the fix,
each for the right reason:

```
token-throb.spec.ts "highlight throb survives a real render-game cycle"
  → Expected 1, Received 0  (throb idle after the move; token still highlighted)
token-throb.spec.ts "highlight throb survives a DOM reparent"
  → Expected 1, Received 0  (throb dead after reparent)   [empirically CONFIRMS the sibling bug]
finish-gated-animations.spec.ts "force-settles gated … leaves ungated ambient loops running"
  → TypeError: el.finishGatedAnimations is not a function  (method absent)
```

After the fix, all pass; and with `_resetAnimating` reverted to
`finishAllAnimations`, the survival test returns to red (Received 0) —
confirming the test genuinely pins the fix, not incidental behavior.

Note on the trigger: the survival test drives the cycle with **Public
Shuffle** (`VisibleShuffle`), not "To Hidden". A shuffle is legal in every
state; "To Hidden" is illegal (button disabled) whenever components are already
hidden, which made an earlier draft flake in-suite (30s timeout waiting for a
gate-open that a swallowed/illegal move never produced).

### Tests are hosted in Playwright, not `node --test`

The kernel-level `finishGatedAnimations` contract test lives in
`tests/animations/finish-gated-animations.spec.ts` (browser), not a colocated
`*.test.ts` node unit file, because the method operates on real WAAPI
`Animation` objects (`finish()`/`cancel()`/`playState`) on a `LitElement`
subclass — neither `Animation` nor `customElements` exists under this repo's
`node --test` runner (Node type-stripping, no DOM). A `boardgame-token` is the
concrete `BoardgameAnimatableItem` used to reach the inherited kernel.

---

## 4. Verification

- Unit (`npm run test:unit`): 254 passed.
- Parity (`tests/animations/parity/`): all passed, **goldens untouched**
  (`git status` clean under `tests/animations/parity/`); includes the two new
  token-throb tests.
- `waapi-gate` + `waapi-play`: 9 passed (the "same-cycle reinstall" gate test
  is a documented pre-existing flake; passes on rerun).
- `waapi-companion`: 8 passed.
- `finish-gated-animations.spec.ts`: 2 passed.
- `npx tsc --noEmit`: clean.
- No golden files regenerated; no committed test modified except the NEW
  registry unit test's fake item, renamed to the new interface method.
