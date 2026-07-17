# Animation effects

Boardgame renderers can add small, coherent moments of delight through the
renderer's protected `effects` service. Effects run in a pointer-transparent
overlay, survive renderer layout and shadow boundaries, and use the framework's
existing Web Animations lifecycle.

```ts
private celebrateScore(event: Event) {
  this.effects?.burst(event.currentTarget as HTMLElement, {
    preset: 'score',
    seed: `${this.gameId}:${this.gameVersion}:score`,
  });
}
```

`burst()` accepts an element, an element id / `data-effect-anchor` value, or a
viewport point (`{ x, y }`). Built-in presets are `impact`, `score`, `success`,
and `celebrate`. Override `count`, `spread`, `duration`, or `colors` only when a
game's visual language needs it; the presets are intended to make the common
case one line.

The defaults are deliberately safe:

- decorative and non-blocking (effects never own game truth or hold the queue);
- immediate/local timing (`timing: 'immediate'`);
- deterministic when supplied a stable seed;
- capped at 24 particles per burst, 60 particles and eight effects globally;
- skipped when the player prefers reduced motion;
- cancellable, with a `finished` promise and automatic teardown cleanup;
- `aria-hidden` and never a substitute for meaningful status UI.

Use `{ timing: 'version' }` when a Table/Hand companion effect must share the
installed game's synchronized animation slot. This first API is intentionally
decorative-only: sparkles and celebrations never delay play or disable moves.

This first API is intentionally a small semantic vocabulary rather than a raw
particle engine. Future transition-declarative recipes can compile to the same
effect layer without changing its lifecycle or performance contract.
