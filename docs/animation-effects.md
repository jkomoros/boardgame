# Animation effects

Boardgame effects are immutable descriptions of presentation. Game truth stays
in state; renderers describe a visual cue, and the framework owns measurement,
timing, accessibility, budgeting, cancellation, and cleanup.

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
write access. Standalone die spins use the same track-to-keyframe executor and
ambient version timing, but do not claim FLIP provenance or satisfy motion
anchors. This framework contract is intentionally not a game-author API.

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
    context.move?.Name !== MoveNames.ClaimPoint
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

`departure` resolves only when the card's real WAAPI playback starts;
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
fx.parallel([
  fx.trail({
    subject: movedTokenId,
    tone: 'magic',
    intensity: 'small',
  }),
  fx.burst({
    at: fx.motion(movedTokenId),
    tone: 'reward',
    intensity: 'small',
    timing: 'immediate',
  }),
])
```

The trail begins from the structural motion's real `started` event. It uses the
same captured viewport endpoints and derives its timing envelope (earliest
delay, latest visible end, and primary easing) from the actual compiled
animations. It is cancelled when that exact motion generation is interrupted.
It therefore has no independent `timing` option and never delays the state
queue. A component that did not move returns `no-motion-path`; a missing or
opted-out visual subject returns `missing-subject`.

The trail is a colored silhouette, not a component clone. The animator captures
only an explicit shape capability—`rectangle`, `rounded-rectangle`, or
`circle`—and obtains size and position from the separate geometry snapshot.
Cards publish a rounded rectangle; circular token types publish a circle; other
framework components have a conservative rectangle default. No DOM, text,
artwork, card face, computed style, or game property can cross this boundary.

Framework component implementations can make that capability explicit:

```ts
override motionSubjectSnapshot(): MotionSubjectSnapshot | null {
  return motionSilhouette('circle'); // return null to opt out
}
```

Under reduced motion, a trail becomes a stationary arrival pulse. Echoes share
the document-wide visual-node budget and degrade to available capacity.

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
| Recipe | `burst`, `pulse`, `travel`, `trail`, `sequence`, `parallel` | none | Descriptor |
| Tone | `neutral`, `reward`, `confirm`, `attention`, `warning`, `magic` | `neutral` | Inherited through composition |
| Intensity | `subtle`, `small`, `medium`, `large` | `medium` | Inherited through composition |
| Timing | `immediate`, `version`, or `{ localStartAtMs }` | `immediate` | Inherited through composition |
| Identity | `key`, `seedKey` | descriptor path | Descriptor and deterministic seed identity |
| Structural point | `fx.motion(id, 'departure' \| 'arrival')` | `arrival` | `pulse` and `burst` in authoritative transitions |
| Structural trail | `fx.trail({ subject: id })` | none | Real automatic movement only; inherits its structural timing |
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
- `examples/debuganimations` follows a real moved token with a silhouette trail
  and decorates its arrival, plus demonstrating imperative click celebration
  and theme/intensity controls.
- `examples/pig` exercises the standalone die's shared `visual:transform`
  track executor. Its roll stays version-timed and queue-gated without being
  misrepresented as structural travel.
- Companion table/hand renderers use `animateBetween()` for real card flights;
  that structural API shares timing and geometry foundations with effects but
  remains queue-critical.

What is not configurable yet is equally important: games cannot subscribe to
the private structural-motion plan, replace automatic FLIP, or request artwork
or arbitrary card content in a trail. Subject artwork is a separate future
capability, not an extension smuggled through the silhouette protocol.
