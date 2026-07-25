# Animation parity harness

Regression net for the animatable-item unification project
(`docs/superpowers/specs/2026-07-24-animatable-item-unification-design.md`).
Two complementary suites pin the CURRENT animation behavior so migrations can
prove zero regression mechanically.

## Running

Needs the offline-dev server serving this checkout (`boardgame-util serve
--offline-dev-mode`; in a worktree, launch with a GOPATH shim so the import
path resolves to the worktree). Then from `server/static/`:

```bash
npx playwright test tests/animations/parity/            # compare vs goldens
PARITY_RECORD=1 npx playwright test tests/animations/parity/  # re-record
```

Re-record ONLY when a behavior change is intended and declared; every golden
diff must be explainable hunk by hunk.

## What each suite pins

**Trace suite** (`trace.spec.ts`): animation-hook event streams. Per-scenario
contracts, matched to what each game can promise deterministically:

| Golden | Contract | Why not stricter |
|---|---|---|
| `memory-reveal-one` | exact counters + per-element event sequences (ids canonicalized by first appearance) | deck shuffle renames ids; everything else is deterministic |
| `debuganimations-card-move` | exact gated-cycle counts + required kinds | FLIP skips no-op transforms and messy rotations hash per-game random ids, so play counts vary (observed 92–138 for the same action) |
| `blackjack-deal` | structural invariants + required kinds | deal length depends on the shuffled deck |
| `pig-roll` | structural invariants + required kinds | post-roll cycles depend on the rolled value |

All modes assert: watchdog 0, every play settles in-window; exact modes also
assert open/close balance.

**Geometry suite** (`geometry.spec.ts`): motion-curve fingerprints. Every
animation in a scenario (deep shadow-root walk — `document.getAnimations()`
returns NOTHING for shadow-tree animations in this Chromium) is paused and
seeked to fractions 0/.25/.5/.75/1 of its own delay+duration; per curve:
displacement-normalized progress, transform-matrix (Frobenius) progress,
opacity, declared `[duration, delay]` on a 25ms grid, and z-index. Curves
compare as a SET under 0.08 tolerance (counts are per-game random; count
regressions are the trace suite's job). Wave-union sampling captures chained
cohorts. Scenarios: swap flight, reveal flip, interrupted-swap retarget, plus
component fixtures for `fading-text` and `game-outcome` (the Phase 1
before/after anchors — full-game flows can't drive them deterministically).

## Teeth (verified failure detection)

- Suppressing card `play` hooks → trace suite fails debuganimations, memory,
  blackjack (pig legitimately unaffected).
- `ease-in-out` → `linear` sabotage → geometry fails the swap scenario.
- Flip-shape sabotage (`rotateY(180)` → `rotateY(90)`) → geometry fails the
  memory reveal.

## Accepted residual blind spots (harness-critic ledger)

Reviewed adversarially at Phase 0 close; these are ACCEPTED, with owners:

- **Absolute endpoints / raw positions** — per-game layout randomness makes
  raw-rect goldens unreproducible. Wrong-final-position bugs that preserve
  curve shape are not caught here; the existing behavioral suites
  (`waapi-gate`, `waapi-companion`) and geometry's z/timing channels bound
  the exposure.
- **Blackjack companion flight SHAPE** — only cross-surface skew sync is
  pinned (`waapi-companion.spec.ts`); the flight's own curve has no golden
  (deal randomness). Phase 3 must rely on the debuganimations curves plus
  companion suite.
- **Reduced-motion goldens** — Phase 1 *declares* a behavior change here
  (kernel skip vs CSS 1ms sprint), so pinning the current behavior would
  contradict the approved spec. Kernel-level reduced-motion behavior is
  covered by `waapi-play`/`waapi-companion`.
- **Mobile viewport** — geometry runs at 1280×900 only; curves are
  size-normalized in principle. Low value vs cost.
- **Roster / player-info animations** — currently un-gated (the #714 gap);
  the Phase 2 change lands with its own gate-witness test (plan Task 10).
- **`expectedSettleMs` watchdog extension end-to-end** — no scenario is long
  enough to need it; owned by the gate-kernel unit tests (plan Task 8).
- **0.08 tolerance** — validated against large-effect teeth (0.17–0.21
  midpoint deltas); a sub-tolerance easing tweak (<0.08 at every fraction)
  would pass. The midpoint sample carries most discriminative power.
- **Short-lived animations** — a sub-two-frame animation could finish before
  the wave pause catches it; would surface as a loud existential-match
  failure (flake), not a silent pass.

### Phase 2 topology asymmetry (accepted, documented)

The registry/context providers live on `boardgame-render-game`; the roster
is its DOM SIBLING. Roster-hosted animatables are therefore gated (via the
game-view event pipe) but NOT registry-swept: on a cycle handoff a
mid-flight board animatable is force-finished (snaps) while a roster one
completes smoothly. Benign — a late roster settle is a kernel no-op. Roster
items also resolve a null ambient animation context (default timing), which
is correct: their animations are local effects, not version-slot
participants.

Orphan-settle (a roster animatable removed from the DOM mid-animation) IS
now covered: `BoardgameAnimatableItem.disconnectedCallback` force-settles
any still-gated item a microtask after disconnect (deferred so a
same-tick reparent never snaps a live animation), and
`boardgame-game-view`'s `_rosterWillAnimate` additionally subscribes to the
item's `settled()` promise as a done channel that does not depend on DOM
presence — the bubbled `animation-done` event this suite otherwise relies
on cannot reach render-game from a detached node. See
`tests/animations/parity/player-info-gate.spec.ts`'s third test and
`docs/superpowers/specs/evidence/2026-07-25-roster-orphan-settle.md`.
