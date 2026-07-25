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
