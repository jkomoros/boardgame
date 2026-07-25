import { test, expect } from '@playwright/test';

// Task 4: boardgame-fading-text migrates from self-driven CSS keyframes onto
// the gated WAAPI play() kernel (BoardgameAnimatableItem). This pins GATE
// PARTICIPATION -- the thing CSS keyframes structurally cannot provide: a
// bubbling composed `will-animate`/`animation-done` pair on the element, and
// the shared animHooks play/settle counters incrementing. Mounts the
// component exactly like geometry.spec.ts's "fixture: fading-text fade
// curve" (import on '/', createElement, set .trigger twice) so this test and
// the geometry fixture stay in lockstep.
test('fading-text fade participates in the animation gate', async ({ page }) => {
  await page.goto('/');

  const result = await page.evaluate(async () => {
    // Importing the component module installs animHooks as a side effect
    // (boardgame-animatable-item.ts imports anim-test-hooks.js), so the
    // hooks singleton exists by the time the element is created below even
    // though this page ('/') never installs it on load.
    await import('/src/components/boardgame-fading-text.ts');
    const el = document.createElement('boardgame-fading-text') as any;
    el.style.cssText = 'position:fixed;top:200px;left:200px;width:120px;height:40px;';
    el.autoMessage = 'fixed';
    el.message = 'Parity!';
    document.body.appendChild(el);

    const events: string[] = [];
    el.addEventListener('will-animate', () => events.push('will-animate'));
    el.addEventListener('animation-done', () => events.push('animation-done'));

    const hooks = (window as any).__bgAnimTestHooks;
    const playsBefore = hooks ? hooks.plays : 0;
    const settlesBefore = hooks ? hooks.settles : 0;

    // First trigger change from a defined previous value fires the fade
    // (mirrors geometry.spec.ts's fixture: the very first assignment only
    // establishes _previousTriggerValue).
    el.trigger = 1;
    await el.updateComplete;
    el.trigger = 2;
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

// Code-review finding: animateFade()'s retrigger path calls
// finishAllAnimations() to force-settle the PRIOR fade before starting a new
// one. That prior play()'s own `.finished.catch().finally()` closure is
// still pending when finish() forces its settlement, and without a
// generation token that stale closure can clear `_visible` AFTER the new
// fade has already started (a genuine mid-flight retrigger, not a
// same-tick double-set) -- hiding the container while a fresh animation is
// legitimately running underneath. This drives that exact shape: start a
// fade, wait until its animation is OBSERVABLY running (currentTime > 0,
// not just started), retrigger, then sample the container's visibility
// repeatedly through a window that spans well past when the stale closure
// would have cleared it but well before the second fade's own real
// completion -- so a premature hide is caught even though the exact
// microtask interleaving isn't independently observable from outside.
test('retrigger mid-flight keeps the container visible through the second fade', async ({ page }) => {
  await page.goto('/');

  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-fading-text.ts');
    const el = document.createElement('boardgame-fading-text') as any;
    el.style.cssText = 'position:fixed;top:200px;left:200px;width:120px;height:40px;';
    el.autoMessage = 'fixed';
    el.message = 'Parity!';
    document.body.appendChild(el);

    const hooks = (window as any).__bgAnimTestHooks;
    const playsBefore = hooks.plays;
    const settlesBefore = hooks.settles;

    const message = () => el.shadowRoot.querySelector('#message') as HTMLElement;
    const container = () => el.shadowRoot.querySelector('#container') as HTMLElement;
    const isVisible = () => container().classList.contains('animating');
    const frame = () => new Promise<void>((r) => requestAnimationFrame(() => r()));

    el.trigger = 1;
    await el.updateComplete;
    el.trigger = 2; // starts fade #1
    await el.updateComplete;

    // Wait until fade #1's own animation is genuinely running (not just
    // created -- currentTime must have advanced past 0).
    const deadline1 = performance.now() + 5000;
    for (;;) {
      const anims = message().getAnimations();
      const running = anims.some((a) =>
        a.playState === 'running' && typeof a.currentTime === 'number' && a.currentTime > 0);
      if (running) break;
      if (performance.now() > deadline1) throw new Error('fade #1 never became observably running');
      await frame();
    }

    el.trigger = 3; // genuine mid-flight retrigger -> fade #2
    await el.updateComplete;

    // Sample visibility across a window that starts right after the
    // retrigger and runs well past where a stale closure would have
    // cleared it, but stays short of fade #2's own ~250ms completion.
    let everHiddenDuringWindow = false;
    const sampleDeadline = performance.now() + 150;
    while (performance.now() < sampleDeadline) {
      if (!isVisible()) { everHiddenDuringWindow = true; break; }
      await frame();
    }

    // Let fade #2 fully settle, then confirm the container hides for real.
    await new Promise<void>((resolve) => {
      el.addEventListener('animation-done', () => resolve(), { once: true });
      setTimeout(resolve, 5000);
    });
    await frame();

    return {
      everHiddenDuringWindow,
      hiddenAfterSettle: !isVisible(),
      playsDelta: hooks.plays - playsBefore,
      settlesDelta: hooks.settles - settlesBefore,
    };
  });

  expect(result.everHiddenDuringWindow,
    'the container must stay visible through fade #2; a stale generation-0 closure hid it early')
    .toBe(false);
  expect(result.hiddenAfterSettle).toBe(true);
  expect(result.playsDelta).toBe(2);
  expect(result.settlesDelta).toBe(2);
});

// Code-review finding (round 2): a superseded animateFade() continuation
// must not START a play either. Two synchronous animateFade() calls leave
// two pending updateComplete continuations; without the start guard the
// stale one launches a duplicate animation -- inflating the play count and
// holding the gate until BOTH settle.
test('overlapping animateFade calls start exactly one animation', async ({ page }) => {
  await page.goto('/');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-fading-text.ts');
    const el = document.createElement('boardgame-fading-text') as any;
    el.style.cssText = 'position:fixed;top:200px;left:200px;width:120px;height:40px;';
    el.autoMessage = 'fixed';
    el.message = 'Overlap!';
    document.body.appendChild(el);
    await el.updateComplete;
    const hooks = (window as any).__bgAnimTestHooks;
    const playsBefore = hooks.plays;
    // Two overlapping starts in the same tick: both continuations pend.
    el.animateFade();
    el.animateFade();
    await new Promise<void>((resolve) => {
      el.addEventListener('animation-done', () => resolve(), { once: true });
      setTimeout(resolve, 5000);
    });
    return { playsDelta: hooks.plays - playsBefore };
  });
  expect(result.playsDelta).toBe(1);
});

// Code-review finding (round 2, Critical): status-text now extends
// BoardgameAnimatableItem, so it carries an inherited `animationContext`
// property defaulting to null. The ambient-context walk must climb PAST a
// null context (an animatable wrapper is not a provider merely by
// existing) so the fading-text nested inside status-text still reaches the
// real provider above -- the render-game analog here is a plain container
// whose `animationContext` is populated. Observable: the hooks 'active'
// record carries the provider's version only when the walk resolved it.
test('fading-text nested in status-text resolves the ambient version context', async ({ page }) => {
  await page.goto('/');
  const result = await page.evaluate(async () => {
    await import('/src/components/boardgame-status-text.ts');
    const provider = document.createElement('div') as any;
    provider.animationContext = {
      version: 7,
      startAtMs: Date.now(),
      slotDurationMs: 1000,
      maxAnimationDurationMs: 800,
    };
    document.body.appendChild(provider);
    const status = document.createElement('boardgame-status-text') as any;
    provider.appendChild(status);
    status.value = 1;
    await status.updateComplete;
    const hooks = (window as any).__bgAnimTestHooks;
    const logStart = hooks.log.length;
    status.value = 2;
    await status.updateComplete;
    await new Promise<void>((resolve) => setTimeout(resolve, 1500));
    const entries = hooks.log.slice(logStart)
      .filter((e: any) => e.ev === 'active' && String(e.detail).startsWith('boardgame-fading-text'));
    return { versions: entries.map((e: any) => e.version ?? null) };
  });
  expect(result.versions).toContain(7);
});
