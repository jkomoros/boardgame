# Animation and effects

This is the complete game-author guide to Boardgame's animation subsystem:
automatic structural motion, semantic effects, their extension points, and the
boundary between them. Game truth stays in state; renderers describe
presentation, and the framework owns measurement, timing, accessibility,
budgeting, cancellation, and cleanup.

For the short path through ordinary game authoring, start with **Client
animations** in [`TUTORIAL.md`](../TUTORIAL.md#client-animations). This guide is
the reference for games that need custom choreography. Framework maintainers
should also read [`server/static/src/ARCHITECTURE.md`](../server/static/src/ARCHITECTURE.md).

## What this system owns

There are two related but deliberately separate systems:

- **Structural motion** preserves continuity when authoritative state moves,
  resizes, reveals, or removes a component. The framework's component animator
  owns it automatically, and it holds the state queue until its real WAAPI
  animations settle.
- **Semantic effects** explain meaning—reward, warning, confirmation, magic.
  Games opt into these through the API below. They are disposable decoration
  and never hold the state queue.

They share finite geometry and timing compilers, but not ownership. An effect
cannot replace or delay a card's structural transform. This division is why a
particle-budget failure can never strand a game transition.

Internally, automatic component motion is planned as immutable, single-channel
tracks. Structural FLIP owns the component host's transform, fades own host
opacity, and card face/rotation changes own the card's inner visual transform.
The same frozen list decides whether work exists and drives WAAPI playback, so
a card flip is no longer a second imperative animation system. Structural plans
publish only the owned channel names for observation; effects still receive no
write access.

A track is a list of SAMPLES, not an endpoint pair. Most tracks are still
authored as `{from, to}` and compile to two samples at offsets 0 and 1 — a
no-op for WAAPI — but a component-owned `visual` track may instead supply a
CURVE, a pure function of progress that the compiler samples on a uniform grid
it owns. That is how motion which is not "interpolate A→B under one easing"
enters the system. A sampled track claims its channel's TIMELINE: it is pinned
to linear easing at the effect level and must be played with immediate rather
than version timing, because a version slot would clamp a bake of the wrong
length into its own and play the whole thing uniformly fast — a die's throw
takes as long as the physics says it takes (a few hundred ms for a d6, up to
2.8s for the longest shapes), which is not a number a 600ms slot can be told.
Curves are permitted only on
`visual`; a sampled host track would break the FLIP resting write, the
two-point structural motion path, and trail-echo synchronization.

The standalone die is the current sampled-track producer: it rolls a real 3D
solid through a seeded physics simulation baked to literal matrices, on the
same track-to-keyframe executor, without claiming FLIP provenance or
satisfying motion anchors. This framework contract is intentionally not a
game-author API.

## The three axes

Effects keep three independent choices separate:

1. **Recipe** — what motion happens: `burst`, `pulse`, `travel`, or `trail`.
2. **Tone** — what it means: `neutral`, `reward`, `confirm`, `attention`,
   `warning`, or `magic`.
3. **Intensity** — how strongly it speaks: `subtle`, `small`, `medium`, or
   `large`.

```ts
this.effects?.play(fx.burst({
  at: button,
  tone: 'confirm',
  intensity: 'small',
}));
```

The common API contains no particle counts, pixels, colors, or easing curves.
Those remain coherent framework policy. An `advanced` section provides an
explicit escape hatch when a game's visual language genuinely requires it.

## Local interaction feedback

Use the imperative service for feedback caused by a local click, drag, hover,
or selection—not for authoritative state changes.

```ts
private selected(event: Event) {
  const target = event.currentTarget as HTMLElement;
  const handle = this.effects?.play(fx.pulse({
    at: target,
    tone: 'attention',
    intensity: 'subtle',
  }));

  void handle?.finished.then(result => {
    // finished | cancelled | skipped (+ a reason)
    console.debug(result);
  });
}
```

Elements are the most robust anchors for immediate feedback. A viewport point
is also available through `fx.point(x, y)`.

## Authoritative transition effects

Override `effectsForTransition()` for scoring, transfers, placement, reveals,
or victory. The host invokes it exactly once for each installed snapshot, after
the renderer has settled. Do not diff state or start authoritative effects from
Lit lifecycle methods.

```ts
override effectsForTransition(
  context: EffectTransitionContext<State, MoveName>,
): readonly EffectSpec[] {
  if (
    context.kind === 'initial' ||
    context.move?.AnimationKey !== MoveNames.ClaimPoint
  ) {
    return [];
  }

  return [fx.parallel([
    fx.pulse({
      at: fx.anchor('score'),
      tone: 'reward',
      intensity: 'medium',
    }),
    fx.burst({
      at: fx.anchor('score'),
      tone: 'reward',
      intensity: 'small',
    }),
  ], {
    key: 'claim-point',
    timing: 'version',
  })];
}
```

The context supplies immutable `before`, `after`, animation-safe move metadata
(name and produced version), version, snapshot epoch, and a selector helper:

```ts
if (!context.changed(state => state.Game.Score)) return [];
```

Initial installation is explicit. Authors must opt in if an initial snapshot
should animate, preventing a refresh of an already-finished game from replaying
its victory celebration.

## Named anchors and disappearing sources

Declarative effects use renderer-scoped names:

```ts
html`
  <div data-effect-anchor="bank">...</div>
  <div data-effect-anchor="hand">...</div>
`
```

```ts
fx.travel({
  from: fx.anchor('bank'),
  to: fx.anchor('hand'),
  tone: 'reward',
  intensity: 'small',
})
```

The framework measures named anchors before and after each state installation.
Current geometry is preferred; the prior measurement is retained as a fallback
for an element that disappeared. Names never search another mounted renderer or
companion surface.

## Scheduling a structural motion cohort

When several components move during one authoritative transition, a renderer
can give their automatic structural animations an explicit start order:

```ts
override motionCohortsForTransition(context: EffectTransitionContext<State, MoveName>) {
  if (context.kind === 'initial' || context.move?.AnimationKey !== MoveNames.Deal) return [];
  return [motion.stagger({
    key: 'deal-cards',
    subjects: context.after.Game.Hand.IDs,
    intervalMs: 45,
  })];
}
```

Array order is the deterministic cadence: the first participating component
starts at zero, the second at `intervalMs`, and so on. IDs that did not animate
in this installation are ignored. For cohort members this timing replaces a
stack's legacy `stagger`; nonmembers retain their stack timing. Duplicate IDs,
overlapping cohorts, malformed declarations, or an exception in the hook reject
the complete explicit schedule and fall back atomically to stack timing.

This API schedules structural starts only. It does not select components from
private motion plans, retime effects, expose geometry, or create a group
completion event. Version-slot clipping, reduced motion, cancellation, and the
animation gate continue through the existing structural timing and settlement
primitives. `motion.stagger()` coordinates components inside one version.

For an already-buffered solo successor, `motionReleaseForTransition()` may
return `motion.release({ progress, subjects? })`. The release barrier observes
the real primary WAAPI animation of every selected FLIP or retained-transfer
segment. It admits a destructive generation cutover; it does not make two
structural generations coexist. Missing/ambiguous selections, cancellation,
skips, companion timing, or malformed declarations fall back to ordinary
cycle settlement. Lifecycle-bound decorations are terminalized at cutover.

## Decorating automatic component motion

Use a motion anchor when a pulse or burst should occur at the actual endpoint
of one component's automatic structural animation:

```ts
fx.burst({
  at: fx.motion(cardId), // arrival is the default
  tone: 'reward',
  intensity: 'small',
  timing: 'immediate',
})

fx.pulse({
  at: fx.motion(cardId, 'departure'),
  tone: 'attention',
  intensity: 'subtle',
  timing: 'immediate',
})
```

`departure` resolves only when the animator observes the card's primary WAAPI
channel entering its active interval;
`arrival` resolves only after it finishes successfully. A skipped, cancelled,
or missing structural animation produces an explicit skipped effect result.
Stationary face/property morphs still have endpoints, so an arrival burst can
decorate a card flip without claiming that the card traveled.

Motion anchors are intentionally point-only and currently work with `pulse`
and `burst`, not `travel`. They expose an ID and captured viewport center—never
a DOM reference, cloned card, hidden face, or transform ownership. They are for
authoritative transition descriptors; use ordinary element or point anchors
for local interaction feedback.

Because the structural event itself supplies synchronization, start a
departure/arrival decoration with `timing: 'immediate'`. Reusing
`timing: 'version'` after arrival may correctly skip because the version slot
has already been consumed by the card animation.

### Following the moving subject

Use `fx.trail()` when the decoration should follow a component's real automatic
movement rather than synthesize a second trip between named anchors:

```ts
fx.decorateMotion({
  subject: movedTokenId,
  trail: {
    tone: 'magic',
    intensity: 'small',
  },
  arrival: fx.burst({
    at: fx.motion(movedTokenId),
    tone: 'reward',
    intensity: 'small',
    timing: 'immediate',
  }),
})
```

The trail is prearmed with the structural motion's real `armed` event. It uses the
same captured viewport endpoints and derives its timing envelope (earliest
delay, latest visible end, and primary easing) from the actual compiled
animations. It is cancelled when that exact motion generation is interrupted.
It therefore has no independent `timing` option and never delays the state
queue. A component that did not move returns `no-motion-path`; a missing or
opted-out visual subject returns `missing-subject`.

`fx.decorateMotion()` is the durable composition boundary for lifecycle-bound
work. The effect layer discovers it recursively and subscribes its trail and
departure/arrival effects before structural playback, even when the descriptor
is nested inside an `fx.sequence()`. When the sequence later reaches that item,
it joins the already-running or completed decoration instead of replaying a
cached start event late. Use a bare top-level `fx.trail()` only when no grouped
departure or arrival decoration is needed.

The trail is a colored silhouette, not a component clone. The animator captures
only an explicit shape capability—`rectangle`, `rounded-rectangle`, or
`circle`—and obtains size and position from the separate geometry snapshot.
Cards publish a rounded rectangle; circular token types publish a circle; other
framework components have a conservative rectangle default. No DOM, text,
artwork, card face, computed style, or game property can cross this boundary.
Echo size is installed once and endpoint size changes are expressed as transform
scale, so active trails animate only compositor-friendly transform and opacity
rather than width or height.

Framework component implementations can make that capability explicit:

```ts
override motionSubjectSnapshot(): MotionSubjectSnapshot | null {
  return motionSilhouette('circle'); // return null to opt out
}
```

Under reduced motion, a trail becomes a stationary arrival pulse. Echoes share
the document-wide visual-node budget and degrade to available capacity.

### Running an effect after several motions

`fx.afterMotion()` is a success-only barrier over explicit automatic-FLIP
subjects:

```ts
fx.afterMotion({
  subjects: dealtCardIds,
  effect: fx.burst({
    at: fx.anchor('hand'),
    tone: 'reward',
    timing: 'immediate',
  }),
})
```

The barrier is recursively prepared before structural playback, including when
nested in a sequence. It binds each requested subject to exactly one segment in
that FLIP generation and runs its ordinary child effect only after every bound
segment finishes successfully. Missing, skipped, or ambiguous subjects skip the
barrier; cancellation or a replacement generation cancels it. The child cannot
itself contain a trail, motion decoration, or another motion barrier.

This is not a general cohort identity or a queue gate. It neither schedules nor
selects structural motion, does not observe explicit `fly()` flights,
and never changes structural settlement. Combine it with `motion.stagger()`
when the same typed local profile needs both an ordered start cadence and a
success flourish.

Hidden or sanitized components sometimes lack an exact visible endpoint. The
default preserves the historical ordered winner/runner-up collection inference,
including ties and same-collection fallbacks, so existing card travel does not
disappear. The pure continuity resolver also offers strict ambiguity rejection
for new integrations. Motion-bound effects must treat `motion-skipped` as an
ordinary deterministic outcome, not as an error.

### Declaring retained-carrier transfers

`motionTransfersForTransition()` is the pure, queue-critical companion to
decorative effects. Return `motion.transfer({ key, subjectId, source, carrier,
durationMs })` for each carrier that should arrive from
source geometry. The full batch is validated before playback, scoped to this
renderer's registered roots, published as one explicit generation, and settled
alongside automatic FLIP. Missing endpoints skip individual segments;
malformed or conflicting declarations reject the complete batch.

The hook consumes the real wire contract: `context.move.AnimationKey`, not
`Name`, plus only viewer-sanitized `Properties`. A declaration key is unique
within one transition. It is not cross-device identity: version timing aligns
surfaces, while a shared semantic transfer token requires an explicit,
privacy-reviewed server field. An exact after-only stack carrier is consumed by
automatic FLIP rather than animated twice: the declared anchor supplies only
spatial origin; source/destination stack defaults supply semantic pose; all host
and visual tracks retain one FLIP lifecycle. Retained, mismatched, or ambiguous
stack carriers fail closed. Viewport anchor translation also fails closed under
transformed ancestors until affine projection is a distinct primitive.

## Composition

Every recipe is an `EffectSpec`, so composition is ordinary immutable data:

```ts
fx.sequence([
  fx.travel({
    from: fx.anchor('bank'),
    to: fx.anchor('player'),
    tone: 'reward',
  }),
  fx.parallel([
    fx.pulse({ at: fx.anchor('player'), tone: 'reward' }),
    fx.burst({ at: fx.anchor('player'), tone: 'reward' }),
  ]),
], {
  key: 'resource-transfer',
  gapMs: 30,
  intensity: 'small',
  timing: 'version',
})
```

Parent tone, intensity, timing, and seed identity are inherited by children
unless a child overrides them. Composition and leaf recipes use the same
executor, cancellation, result, and budget semantics.

### Reusable game-local profiles

When a game repeats a visual phrase, keep the profile as an ordinary typed
factory beside its renderer:

```ts
const rewardPop = (at: EffectPointAnchor, key: string): EffectSpec => fx.parallel([
  fx.pulse({ at }),
  fx.burst({ at }),
], {
  key,
  tone: 'reward',
  intensity: 'small',
  timing: 'version',
});
```

This inherits the same validation, deterministic identity, theme, reduced
motion, and budgets as inline descriptors. Prefer this to a global string-based
preset registry: TypeScript keeps the inputs honest, the profile is local to
the game whose visual language it expresses, and changing it cannot silently
restyle unrelated games. Promote a profile into `fx` only after several games
demonstrate the same semantic contract—not merely similar keyframes.

## Determinism and companion timing

Particle geometry is seeded from game ID, snapshot epoch, version, descriptor
identity/path, and optional `seedKey`. It never depends on viewport position.

`timing: 'version'` schedules independently-derived effects into the existing
Table/Hand companion slot. It does not transmit an imperative click to another
device; synchronized effects must be returned by `effectsForTransition()` on
each participating surface.

Effects are decorative and never hold the game-state queue or disable moves.
Installing a newer transition cancels only effects owned by the older
transition; local interaction feedback is independently scoped.

## Themes

A game can override semantic palettes without reaching into the effect layer:

```ts
override effectTheme(): EffectTheme {
  return defineEffectTheme({
    tones: {
      reward: ['#ffd54f', '#ff8f00'],
      magic: ['#7c4dff', '#ff4081', '#80d8ff'],
    },
  });
}
```

The default theme falls back to Material system colors. Raw per-effect palettes
remain available only under `advanced.palette`.

## Accessibility, lifecycle, and budgets

- The overlay is pointer-transparent and `aria-hidden`; meaningful outcomes
  must also appear in semantic UI.
- Reduced motion substitutes a short stationary opacity emphasis for bursts and
  travel instead of moving particles across the screen.
- Effects return `finished`, `cancelled`, or `skipped` with a reason. Motion
  anchors use `motion-skipped` when their structural segment did not complete.
- Trails use `missing-subject` when no safe silhouette was published and
  `no-motion-path` when the subject changed without spatial motion.
- Renderer removal cancels all effects. A newer transition cancels stale
  transition-owned effects.
- One document-wide manager admits at most eight concurrent effects and 60
  visual nodes. Large bursts degrade to remaining capacity; existing effects
  are never killed to admit newer decoration.
- A burst is independently capped at 24 transform/opacity-only particles.

## Advanced customization

Advanced values are validated and clamped. Prefer semantic defaults first:

```ts
fx.burst({
  at: fx.anchor('portal'),
  tone: 'magic',
  intensity: 'large',
  advanced: {
    count: 18,
    spreadPx: 130,
    durationMs: 800,
    palette: ['#7c4dff', '#ff4081'],
  },
})
```

Trail tuning is deliberately limited to the overlay echo treatment:

```ts
fx.trail({
  subject: tokenId,
  tone: 'magic',
  intensity: 'medium',
  advanced: {
    echoes: 6,       // clamped to 1–10 and then to remaining budget
    lagMs: 18,       // clamped to 4–80
    opacity: 0.4,    // clamped to 0.05–0.7
    palette: ['#7c4dff', '#80d8ff'],
  },
})
```

The descriptor/executor boundary is the durable extension point. Future
recipes—trails, screen treatments, or visual text echoes—can join the same
contract without adding unrelated methods to the renderer service. Semantic
text remains normal accessible UI; any floating text effect is only its
`aria-hidden` visual echo.

## Configuration reference

| Choice | Values | Default | Scope |
| --- | --- | --- | --- |
| Recipe | `burst`, `pulse`, `travel`, `trail`, `decorateMotion`, `afterMotion`, `sequence`, `parallel` | none | Descriptor |
| Tone | `neutral`, `reward`, `confirm`, `attention`, `warning`, `magic` | `neutral` | Inherited through composition |
| Intensity | `subtle`, `small`, `medium`, `large` | `medium` | Inherited through composition |
| Timing | `immediate`, `version`, or `{ localStartAtMs }` | `immediate` | Inherited through composition |
| Identity | `key`, `seedKey` | descriptor path | Descriptor and deterministic seed identity |
| Structural point | `fx.motion(id, 'departure' \| 'arrival')` | `arrival` | `pulse` and `burst` in authoritative transitions |
| Structural trail | `fx.trail({ subject: id })` | none | Real automatic movement only; inherits its structural timing |
| Structural cohort | `motion.stagger({ subjects, intervalMs })` | none | Ordered starts within one authoritative transition |
| Structural transfer | `motion.transfer({ key, subjectId, source, carrier })` | none | Retained-carrier intent; stack arrivals arbitrate into FLIP |
| Buffered cutover | `motion.release({ progress, subjects? })` | settlement | Solo, already-buffered successor only |
| Theme | semantic tone palettes | Material-aware defaults | Renderer via `effectTheme()` |
| Escape hatch | recipe-specific `advanced` values | semantic policy | Validated and clamped |

Use `timing: 'version'` for effects returned from `effectsForTransition()` when
table and hand surfaces should independently align to the same version slot.
Use the default immediate timing for local input feedback. A local absolute
start is infrastructure-level scheduling, not a synchronization protocol.

## Working examples

- `examples/memory` follows the real revealed card's structural arrival, then
  pulses it and adds a reward burst for a match. The reveal is a stationary
  face morph, demonstrating that endpoint decoration does not imply travel.
- `examples/debuganimations` follows a real moved token with a silhouette trail,
  decorates its arrival, and uses an explicit cohort cadence for its visible
  shuffle, plus demonstrating imperative click celebration and theme/intensity
  controls.
- `examples/pig` exercises the standalone die's shared `visual:transform`
  track executor, now as a SAMPLED curve: one physics trajectory baked to
  literal matrices and played as a single track. Its roll takes its own
  duration rather than a version slot — a clamped bake is a die falling at
  five times gravity — but stays queue-gated, declaring that duration so the
  gate waits for it, without being misrepresented as structural travel. Its
  celebration is the worked example of an effect that CANNOT be planned in
  `effectsForTransition`: that hook runs at cycle start, which for a die that
  flies is the moment of the throw, not the moment of the result. Pig listens
  for the die's own `roll-end` instead and plays a pulse (plus a reward burst
  on a six) imperatively, at immediate timing, on the value the event carries.
- Companion Table/Hand bases preserve their established local choreography
  with `animateBetween()` compatibility flights derived from adjacent sanitized
  snapshots. Hand arrivals launch together from their final pose; Table stub
  flights remain decorative and do not gate structural settlement. Authored
  cross-stack motion uses `motion.transfer()` declarations. The two local
  observations are not a shared cross-device transfer identity.

What is not configurable yet is equally important: games cannot subscribe to
the private structural-motion plan, replace automatic FLIP, or request artwork
or arbitrary card content in a trail. Subject artwork is a separate future
capability, not an extension smuggled through the silhouette protocol.
