import { test, expect } from '@playwright/test';

// Task 6: boardgame-game-outcome migrates from a self-driven CSS keyframe
// arrival onto the gated WAAPI play() kernel (BoardgameAnimatableItem). This
// pins GATE PARTICIPATION -- the thing CSS keyframes structurally cannot
// provide: a bubbling composed `will-animate`/`animation-done` pair on the
// element, and the shared animHooks play/settle counters incrementing.
// Mounts the component exactly like geometry.spec.ts's "fixture:
// game-outcome arrival curve" (import on '/', createElement, set finished +
// winners) so this test and the geometry fixture stay in lockstep.
test('game-outcome arrival participates in the animation gate', async ({ page }) => {
  await page.goto('/');

  const result = await page.evaluate(async () => {
    // Importing the component module installs animHooks as a side effect
    // (boardgame-animatable-item.ts imports anim-test-hooks.js), so the
    // hooks singleton exists by the time the element is created below even
    // though this page ('/') never installs it on load.
    await import('/src/components/boardgame-game-outcome.ts');
    const el = document.createElement('boardgame-game-outcome') as any;
    el.style.cssText = 'position:fixed;top:100px;left:100px;width:400px;';
    el.finished = true;
    el.winners = [0];

    const events: string[] = [];
    el.addEventListener('will-animate', () => events.push('will-animate'));
    el.addEventListener('animation-done', () => events.push('animation-done'));

    const hooks = (window as any).__bgAnimTestHooks;
    const playsBefore = hooks ? hooks.plays : 0;
    const settlesBefore = hooks ? hooks.settles : 0;

    document.body.appendChild(el);
    await el.updateComplete;

    await new Promise<void>((resolve) => {
      if (events.includes('animation-done')) { resolve(); return; }
      el.addEventListener('animation-done', () => resolve(), { once: true });
      setTimeout(resolve, 5000);
    });

    return {
      events,
      playsDelta: hooks ? hooks.plays - playsBefore : 0,
      settlesDelta: hooks ? hooks.settles - settlesBefore : 0,
    };
  });

  expect(result.events).toEqual(['will-animate', 'animation-done']);
  expect(result.playsDelta).toBeGreaterThan(0);
  expect(result.settlesDelta).toBe(result.playsDelta);
});

// Code-review finding: play()'s default timing policy is 'version'
// (boardgame-animatable-item.ts), and resolveMotionTiming's 'version' branch
// (src/motion/timing.ts) CLAMPS an explicit duration to whatever remains of
// a populated ambient VersionAnimationContext's maxAnimationDurationMs. The
// old CSS `animation: outcome-arrive 220ms ease-out both;` could never be
// reshaped this way -- it always ran exactly 220ms regardless of any
// surrounding render-game cycle. Mirrors fading-text.spec.ts's nested
// ambient-context test shape: a plain provider div carries a populated
// animationContext (maxAnimationDurationMs deliberately 100, well under
// 220, so the version policy would clamp if consulted), game-outcome mounts
// as its direct child (its own _ambientAnimationContext() walk starts at
// parentNode and finds the provider immediately), and the reveal's actual
// started Animation is inspected directly via getComputedTiming().duration.
test('game-outcome arrival keeps its declared 220ms under a populated ambient version context', async ({ page }) => {
  await page.goto('/');

  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-game-outcome.ts');
    const provider = document.createElement('div') as any;
    provider.animationContext = {
      version: 7,
      startAtMs: Date.now(),
      slotDurationMs: 1000,
      maxAnimationDurationMs: 100,
    };
    document.body.appendChild(provider);

    const el = document.createElement('boardgame-game-outcome') as any;
    el.style.cssText = 'position:fixed;top:100px;left:100px;width:400px;';
    provider.appendChild(el);
    el.finished = true;
    el.winners = [0];
    await el.updateComplete;

    const outcome = el.shadowRoot.querySelector('#outcome') as HTMLElement | null;
    const anim = outcome?.getAnimations()[0];
    const timing = anim?.effect?.getComputedTiming();
    return { duration: timing ? Number(timing.duration) : null };
  });

  expect(result.duration).toBe(220);
});
