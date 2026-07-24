# Animatable Item Unification — Design

Date: 2026-07-24
Issues: #714 (primary), #713 (architecture realized in Phase 3), touches #709's spirit via the parity harness.
Status: Approved (sub-project 1 of 2; sub-project 2 = new animations, separate spec).

## Goals

1. Every element that animates game-semantic state derives from `BoardgameAnimatableItem`
   and animates through its gated WAAPI `play()` kernel.
2. The legacy `boardgame-component-stack` CSS `transition` path is retired, along with the
   `noAnimate` class-suppression dance.
3. The two structural gaps in #714 are closed: the framework can discover/reset
   non-component animatable items, and player-info-renderer animations participate in the
   completion gate.
4. **Zero observable regression of stack animations**, verified mechanically by a parity
   harness — not by inspection. The only permitted behavior changes are the explicitly
   declared fixes in this spec, each requiring an evidence pack before landing.

## Non-goals

- UI chrome transitions (chat panel, drawer, player-chip hover, game-item hover) stay
  plain CSS. They are affordance feedback, not game-state motion, and must never hold the
  state queue.
- `boardgame-effect-layer` stays ungated raw WAAPI by design (see
  docs/animation-effects.md: effects are disposable and never gate).
- New animations (score motion, roster handoff, outcome ceremony, new fx recipes) are
  sub-project 2, designed separately after this lands. Taste bar for that work: quiet and
  purposeful — short durations, motion only where it carries meaning, no bounce; warmth
  via the existing tone/intensity axes.

## Current state (verified inventory)

Routes through the gated primitive today: `boardgame-component` (→ `boardgame-card`,
`boardgame-token`) and `boardgame-die` only.

Animates outside it:

| Element | Mechanism | Gate today |
|---|---|---|
| `boardgame-fading-text` | CSS `@keyframes fadetext` + `animationend`/rAF re-arm | never |
| `boardgame-status-text` | bare `LitElement` wrapping fading-text | never |
| `boardgame-game-outcome` | CSS `@keyframes outcome-arrive` | never |
| `boardgame-token` throb | CSS `@keyframes throb` (infinite) | never (correctly) |
| `boardgame-component-stack` | ambient CSS `transition` on slotted components, suppressed via `noAnimate` during FLIP | indirectly, legacy path |

Structural gaps: the animator only enumerates `_sharedStackList` collections'
`Components` (never standalone animatable items), and `boardgame-player-roster` is a DOM
sibling of `boardgame-render-game`, so player-info animations bubble past the gate
listeners installed at `boardgame-render-game.ts:347-348` and are silently un-gated.

## Phase 0 — Parity harness

New Playwright suite `server/static/tests/animations/parity/`:

- Drives scripted state transitions in **debuganimations** (richest structural motion),
  **memory** (fx + card flips), **blackjack** (companion Table/Hand `animateBetween`
  flights), and **pig** (standalone die visual track).
- Records per scenario:
  - the ordered `animHooks` event log — `play` / `active` / `settle` with element
    identity and declared timing (version, targetAtMs);
  - gate open/close counts and watchdog firings (**watchdog count must be 0**);
  - **mid-flight geometry samples**: computed transform matrices of moving components at
    fixed progress fractions, compared within a small numeric tolerance.
- Golden traces are recorded from the merge base (`d7bbda2a`) and checked in under the
  suite. A recording mode regenerates them; CI mode compares.
- Before any migration lands, a **harness critic sub-agent** answers: "name a regression
  this harness would not catch." Holes get closed first; residual accepted blind spots
  are documented in the suite README.

## Phase 1 — Classing migrations (behavior-preserving)

- **`boardgame-fading-text`** extends `BoardgameAnimatableItem`. The `fadetext`
  keyframes, `.animating` class machinery, `animationend`/`animationcancel` handlers, and
  rAF re-arm are replaced by a single gated `play()` on the `#message` element with
  identical keyframes (opacity 1→0, scale 1→6), `ease-out`, `--animation-length`
  duration (the base class reads the same variable, default 250ms). Visibility toggling
  keys off animation settlement instead of the `.animating` class. Interruption
  (retrigger while in flight) preserves current semantics: the prior fade is finished and
  a fresh one starts.
  - *Declared change (evidence pack recorded; approved via this spec):* reduced-motion currently
    plays a 1ms sprint; the timing kernel skips instead. Also: the fade becomes a gate
    participant (it fires `will-animate`/`animation-done`), which is the literal ask of
    #714. Games can opt out per-instance with `wait-for-animation="false"`.
- **`boardgame-status-text`** extends `BoardgameAnimatableItem`. No visual change; it
  inherits the primitive so it participates in discovery/gating and is ready for
  sub-project 2.
- **`boardgame-game-outcome`** extends `BoardgameAnimatableItem`. `outcome-arrive`
  becomes a gated `play()` on verdict reveal — identical 220ms `ease-out`
  opacity/scale keyframes. Reduced-motion parity (currently `animation: none`; kernel
  skip is equivalent here).
- **`boardgame-token` throb** routes through `play()` with `gated: false`,
  `timing: 'immediate'`, infinite iterations, identical keyframes (drop-shadow filter,
  1s alternate ease-in-out). Ambient highlights must never hold the queue.
  `finishAllAnimations()` already cancels infinite animations safely.

## Phase 2 — Discovery and gate topology

Two intended behavior changes; each is a #714 checklist item, and each lands only after
an evidence pack demonstrating the current misbehavior:

1. **Ambient registry for non-component items.** `BoardgameAnimatableItem` self-registers
   on connect (and unregisters on disconnect) into a registry provided ambiently by
   `boardgame-render-game`, discovered by the same walk-up used by
   `_ambientAnimationContext()`. `render-game` uses the registry at cycle start to run
   interruption (`finishAllAnimations`) and install the version animation context on
   every live animatable item — not just stack-registered components.
2. **Gate extraction + player-info participation.** The gate accounting currently inlined
   in `render-game` (will-animate counting, expected-settle bookkeeping, watchdog) is
   extracted into `src/motion/animation-gate.ts` — a pure, unit-testable kernel.
   `render-game` owns the instance and its external behavior (`all-animations-done`
   timing, watchdog semantics) is unchanged for the render-game subtree.
   `boardgame-game-view` additionally pipes `will-animate`/`animation-done` bubbling out
   of `boardgame-player-roster` into the same gate instance, so player-info animations
   hold state advancement exactly like board animations.

## Phase 3 — Retire legacy stack transitions (realizes #713)

- `boardgame-component-stack` stops relying on ambient CSS `transition`. It hands each
  slotted component its layout transform through an explicit setter (the "external tweak
  transform" of #713). `boardgame-component` reacts to a setter change by playing
  old→new through its own gated `play()` with the same `var(--animation-length)`
  duration and `ease-in-out` easing the CSS used.
- The component animator coordinates through the same setter, eliminating the need to
  toggle `noAnimate`/`.no-animate` to suppress the CSS path; the suppression CSS and flag
  choreography are deleted.
- Strictest parity bar of the project: harness traces before/after must match on event
  ordering, timing, and sampled geometry within tolerance. Any deviation is a stop-line.

## Verification protocol (every phase)

1. Implementer works TDD: failing test → code → green.
2. In parallel after implementation:
   - **Regression critic** sub-agent reads the full diff hunting observable behavior
     changes not declared in this spec.
   - **Harness critic** sub-agent hunts parity-suite blind spots opened by the change.
   - **Fresh verifier** sub-agent runs `npm run type-check:strict`, unit tests, and the
     Playwright parity + existing animation suites, reporting raw output.
3. Critic findings are adversarially confirmed before acting on them.
4. Parity deviations triage to exactly one of: implementation bug (fix it) or
   pre-existing bug (evidence pack: reproduction, spec/doc citation, trace comparison).
   Execution is autonomous: behavior changes **declared in this approved spec** proceed,
   with their evidence packs recorded in `docs/superpowers/specs/` for review.
   **Undeclared** deviations default to preserving current behavior; their evidence packs
   are recorded and surfaced in the final report rather than acted on.
5. A phase merges only with a green harness and all critics resolved.

## Testing

- Unit: `src/motion/animation-gate.test.ts` (new kernel), plus tests for the registry
  and the stack-transform setter path, colocated per existing convention.
- Browser: the Phase 0 parity suite plus existing `tests/animations/` and
  `tests/renderer/` suites stay green throughout.
- Types: `npm run type-check` and `type-check:strict` clean at every phase boundary.

## Risks

- **Companion flights (Table/Hand `animateBetween`)** ride the legacy transition path
  indirectly; blackjack scenario in the harness exists specifically to pin them.
- **Gate participation of fading-text** could lengthen perceived state advancement in
  games with long fades; mitigated by per-instance opt-out and by the fact that fades
  share `--animation-length` with the motion they accompany.
- **Registry lifecycle** must not leak: unregister on disconnect is mandatory and tested.
