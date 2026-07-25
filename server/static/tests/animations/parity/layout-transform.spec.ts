import { test, expect } from '@playwright/test';

// Task 11: BoardgameComponent grows a self-animating `layoutTransform`
// setter -- the mechanism Phase 3 (Task 12) will point the stack's
// per-layout writes through, retiring the CSS
// `transition: transform var(--animation-length, 0.25s) ease-in-out` in
// boardgame-component-stack.ts. NOTHING calls the setter in production yet
// (Task 12's job); this suite pins the setter's own standalone contract so
// that cutover is a pure "swap the caller" change with zero new behavior to
// invent.
//
// Fixture pattern mirrors token-throb.spec.ts / fading-text.spec.ts: a
// standalone element mounted directly on the harness's '/' page (no game
// needed -- this is a component-level mechanism, not a stack-driven flip),
// imported with a guarded `customElements.get` check (see
// player-info-gate.spec.ts's mountRosterFadingText comment: an unconditional
// import can re-execute a module the harness already loaded under a
// different resolved specifier and throw "name already used").
test.describe('boardgame-component layoutTransform', () => {
  test('setting a new value plays a gated host animation from the mid-flight computed transform', async ({ page }) => {
    await page.goto('/');

    const result = await page.evaluate(async () => {
      if (!customElements.get('boardgame-component')) {
        await import('/src/components/boardgame-component.ts');
      }
      const el = document.createElement('boardgame-component') as any;
      el.style.cssText = 'position:fixed;top:100px;left:100px;width:40px;height:40px;';
      el.item = { ID: 'probe-a' };
      document.body.appendChild(el);
      await el.updateComplete;

      const events: string[] = [];
      el.addEventListener('will-animate', () => events.push('will-animate'));
      el.addEventListener('animation-done', () => events.push('animation-done'));

      const hooks = (window as any).__bgAnimTestHooks;
      const playsBefore = hooks ? hooks.plays : 0;
      const settlesBefore = hooks ? hooks.settles : 0;

      el.layoutTransform = 'translateX(120px)';

      const liveAnimations = () => el.getAnimations({ subtree: true })
        .filter((a: Animation) => (a.effect as KeyframeEffect | null)?.target === el);
      const anims = liveAnimations();
      const anim = anims[0] as Animation | undefined;
      const timing = anim?.effect?.getComputedTiming();

      await new Promise<void>((resolve) => {
        if (events.includes('animation-done')) { resolve(); return; }
        el.addEventListener('animation-done', () => resolve(), { once: true });
        setTimeout(resolve, 3000);
      });

      // Independently compute what the browser resolves 'translateX(120px)'
      // to as a matrix, so the final-computed-style assertion below does not
      // just restate the setter's own input string.
      const probe = document.createElement('div');
      probe.style.transform = 'translateX(120px)';
      document.body.appendChild(probe);
      const expectedTransform = getComputedStyle(probe).transform;
      probe.remove();

      return {
        animCount: anims.length,
        duration: timing ? Number(timing.duration) : null,
        easing: timing?.easing ?? null,
        events,
        finalTransform: getComputedStyle(el).transform,
        expectedTransform,
        layoutTransformGetter: el.layoutTransform,
        playsDelta: hooks ? hooks.plays - playsBefore : 0,
        settlesDelta: hooks ? hooks.settles - settlesBefore : 0,
      };
    });

    expect(result.animCount, 'exactly one host animation must start').toBe(1);
    expect(result.duration, 'duration must match animationLengthMs() (no --animation-length set -> 250ms fallback)').toBe(250);
    expect(result.easing).toBe('ease-in-out');
    // Gate participation: Task 12's cutover DECLARES layout tweaks join the
    // completion gate (the CSS transition it replaces was gate-invisible).
    // Pinning it here so that declaration is accurate once something calls
    // this setter for real.
    expect(result.events).toEqual(['will-animate', 'animation-done']);
    expect(result.playsDelta).toBeGreaterThan(0);
    expect(result.settlesDelta).toBe(result.playsDelta);
    expect(result.finalTransform, 'the host must land on the new value once WAAPI settles').toBe(result.expectedTransform);
    expect(result.layoutTransformGetter).toBe('translateX(120px)');
  });

  test('setting the same value is a no-op: no style write side effect, no play', async ({ page }) => {
    await page.goto('/');

    const result = await page.evaluate(async () => {
      if (!customElements.get('boardgame-component')) {
        await import('/src/components/boardgame-component.ts');
      }
      const el = document.createElement('boardgame-component') as any;
      el.style.cssText = 'position:fixed;top:100px;left:100px;width:40px;height:40px;';
      el.item = { ID: 'probe-b' };
      document.body.appendChild(el);
      await el.updateComplete;

      el.layoutTransform = 'translateX(10px)';
      // Drain the first play so the baseline below is clean.
      el.getAnimations({ subtree: true }).forEach((a: Animation) => a.finish());
      await new Promise<void>((r) => requestAnimationFrame(() => r()));

      const hooks = (window as any).__bgAnimTestHooks;
      const playsBefore = hooks.plays;
      const styleBefore = el.style.transform;

      el.layoutTransform = 'translateX(10px)'; // unchanged value

      const liveCount = el.getAnimations({ subtree: true })
        .filter((a: Animation) => (a.effect as KeyframeEffect | null)?.target === el
          && a.playState === 'running').length;

      return {
        liveCount,
        playsDelta: hooks.plays - playsBefore,
        styleBefore,
        styleAfter: el.style.transform,
      };
    });

    expect(result.liveCount, 'an unchanged value must not start an animation').toBe(0);
    expect(result.playsDelta, 'an unchanged value must not call play() at all').toBe(0);
    expect(result.styleAfter).toBe(result.styleBefore);
  });

  test('noAnimate snaps the transform without animating', async ({ page }) => {
    await page.goto('/');

    const result = await page.evaluate(async () => {
      if (!customElements.get('boardgame-component')) {
        await import('/src/components/boardgame-component.ts');
      }
      const el = document.createElement('boardgame-component') as any;
      el.style.cssText = 'position:fixed;top:100px;left:100px;width:40px;height:40px;';
      el.item = { ID: 'probe-c' };
      el.noAnimate = true;
      document.body.appendChild(el);
      await el.updateComplete;

      const events: string[] = [];
      el.addEventListener('will-animate', () => events.push('will-animate'));
      el.addEventListener('animation-done', () => events.push('animation-done'));

      el.layoutTransform = 'translateY(30px)';

      const liveCount = el.getAnimations({ subtree: true })
        .filter((a: Animation) => (a.effect as KeyframeEffect | null)?.target === el).length;

      return {
        liveCount,
        events,
        styleTransform: el.style.transform,
        layoutTransformGetter: el.layoutTransform,
      };
    });

    expect(result.liveCount, 'noAnimate must snap with no animation at all').toBe(0);
    expect(result.events).toEqual([]);
    expect(result.styleTransform).toBe('translateY(30px)');
    expect(result.layoutTransformGetter).toBe('translateY(30px)');
  });

  // CSS-transition retargeting parity: the retired
  // `transition: transform var(--animation-length) ease-in-out` always
  // continued from wherever the box actually WAS on screen (the mid-flight
  // computed value) when interrupted, never from the stale authored target
  // of the animation it interrupted. The setter must reproduce that by
  // capturing `getComputedStyle(this).transform` fresh on every set, not by
  // reusing the previous set's target string.
  test('retargeting mid-flight continues from the mid-flight computed transform, and only one animation survives', async ({ page }) => {
    await page.goto('/');

    const result = await page.evaluate(async () => {
      if (!customElements.get('boardgame-component')) {
        await import('/src/components/boardgame-component.ts');
      }
      const el = document.createElement('boardgame-component') as any;
      el.style.cssText = 'position:fixed;top:100px;left:100px;width:40px;height:40px;';
      el.item = { ID: 'probe-d' };
      document.body.appendChild(el);
      await el.updateComplete;

      const liveAnimations = () => el.getAnimations({ subtree: true })
        .filter((a: Animation) => (a.effect as KeyframeEffect | null)?.target === el);

      el.layoutTransform = 'translateX(200px)';
      // Let the first animation run partway (250ms total) before retargeting.
      await new Promise<void>((r) => setTimeout(r, 60));

      // Sample the mid-flight computed transform and issue the retarget in
      // the SAME evaluate call -- a Node-side round trip between the sample
      // and the set would let the animation advance further and skew the
      // comparison.
      const midFlight = getComputedStyle(el).transform;
      el.layoutTransform = 'translateX(400px)';

      const anims = liveAnimations();
      const newEffect = anims[0]?.effect as KeyframeEffect | null;
      const fromTransform = (newEffect?.getKeyframes()[0] as any)?.transform;

      return {
        liveCount: anims.length,
        midFlight,
        fromTransform,
      };
    });

    expect(result.liveCount, 'the setter must cancel its own prior layout animation so a retarget leaves exactly one live animation').toBe(1);
    expect(result.fromTransform, 'the new animation must start from the mid-flight computed transform (CSS-retargeting parity)').toBe(result.midFlight);
  });
});
