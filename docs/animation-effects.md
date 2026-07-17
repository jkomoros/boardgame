# Animation effects

Boardgame effects are immutable descriptions of presentation. Game truth stays
in state; renderers describe a visual cue, and the framework owns measurement,
timing, accessibility, budgeting, cancellation, and cleanup.

## The three axes

Effects keep three independent choices separate:

1. **Recipe** — what motion happens: `burst`, `pulse`, or `travel`.
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
- Effects return `finished`, `cancelled`, or `skipped` with a reason.
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

The descriptor/executor boundary is the durable extension point. Future
recipes—trails, screen treatments, or visual text echoes—can join the same
contract without adding unrelated methods to the renderer service. Semantic
text remains normal accessible UI; any floating text effect is only its
`aria-hidden` visual echo.
