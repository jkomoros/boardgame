import { test, expect } from '@playwright/test';

// Task 7: boardgame-token's infinite "throb" highlight (the drop-shadow
// pulse while active/highlighted) migrates from self-driven CSS @keyframes
// onto the shared WAAPI play() kernel (BoardgameAnimatableItem), but UNLIKE
// every other migrated component this is deliberately routed UNGATED
// ({ gated: false }): an ambient, infinite highlight must never hold the
// completion gate -- a gated infinite animation would never settle on its
// own and would wedge every cycle until the watchdog force-closed it. This
// spec pins the inverse of the fading-text/game-outcome parity tests: gate
// NON-participation is the point, not participation.
test.describe('boardgame-token throb', () => {
  test('highlighting a token starts a live, ungated, infinite throb', async ({ page }) => {
    await page.goto('/');

    const result = await page.evaluate(async () => {
      // Importing the component module installs animHooks as a side effect
      // (boardgame-animatable-item.ts imports anim-test-hooks.js), so the
      // hooks singleton exists by the time the element is created below even
      // though this page ('/') never installs it on load.
      await import('/src/components/boardgame-token.ts');
      const el = document.createElement('boardgame-token') as any;
      el.style.cssText = 'position:fixed;top:200px;left:200px;';
      document.body.appendChild(el);
      await el.updateComplete;

      const events: string[] = [];
      el.addEventListener('will-animate', () => events.push('will-animate'));
      el.addEventListener('animation-done', () => events.push('animation-done'));

      const hooks = (window as any).__bgAnimTestHooks;
      const gateOpensBefore = hooks.gateOpens;
      const playsBefore = hooks.plays;

      // Deep-walk getAnimations() through the shadow root: document
      // .getAnimations() does not see animations on elements inside a
      // shadow tree (confirmed by the existing geometry.spec.ts interrupt
      // scenario, which uses the identical walk for the same reason).
      const liveAnimations = (): Animation[] => {
        const inner = el.shadowRoot?.querySelector('#inner') as HTMLElement | null;
        return inner?.getAnimations({ subtree: false }) ?? [];
      };

      el.highlighted = true;
      await el.updateComplete;
      // Let the animation actually start running (one frame is enough --
      // element.animate() schedules synchronously, but give the engine a
      // tick before sampling playState).
      await new Promise<void>((r) => requestAnimationFrame(() => r()));

      const animsAfterStart = liveAnimations();
      const infiniteRunning = animsAfterStart.some((a) => {
        const effect = a.effect as KeyframeEffect | null;
        const timing = effect?.getComputedTiming();
        return a.playState === 'running' && timing?.iterations === Infinity;
      });

      // Ungated: no will-animate/animation-done ever fires for this play,
      // and the item's own gate-hold tracking (isAnimating) never turns on
      // -- an infinite gated animation would wedge isAnimating forever.
      const isAnimatingAfterStart = el.isAnimating;

      el.highlighted = false;
      await el.updateComplete;
      const animsAfterClear = liveAnimations();

      document.body.removeChild(el);
      const animsAfterDisconnect = liveAnimations();

      return {
        events,
        infiniteRunning,
        isAnimatingAfterStart,
        gateOpensDelta: hooks.gateOpens - gateOpensBefore,
        playsDelta: hooks.plays - playsBefore,
        animsAfterClearCount: animsAfterClear.length,
        animsAfterDisconnectCount: animsAfterDisconnect.length,
      };
    });

    expect(result.infiniteRunning, 'a live infinite Animation must be running on #inner').toBe(true);
    // play() records the 'play' hook unconditionally (it is not gate-scoped
    // instrumentation), so an ungated play still increments hooks.plays --
    // what it must NOT do is dispatch will-animate/animation-done or hold
    // the gate (isAnimating / gateOpens).
    expect(result.playsDelta).toBeGreaterThan(0);
    expect(result.events, 'ungated plays must never dispatch will-animate/animation-done').toEqual([]);
    expect(result.isAnimatingAfterStart, 'an ungated play must not hold the completion gate').toBe(false);
    expect(result.gateOpensDelta, 'a standalone-mounted token never touches the render-game gate').toBe(0);
    expect(result.animsAfterClearCount, 'clearing highlighted must cancel the throb').toBe(0);
    expect(result.animsAfterDisconnectCount, 'disconnecting the element must cancel the throb').toBe(0);
  });

  test('active-only and highlighted-only both throb; clearing both stops it', async ({ page }) => {
    await page.goto('/');

    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const el = document.createElement('boardgame-token') as any;
      el.style.cssText = 'position:fixed;top:200px;left:200px;';
      document.body.appendChild(el);
      await el.updateComplete;

      const liveCount = () => {
        const inner = el.shadowRoot?.querySelector('#inner') as HTMLElement | null;
        return inner?.getAnimations({ subtree: false }).length ?? 0;
      };
      const frame = () => new Promise<void>((r) => requestAnimationFrame(() => r()));

      el.active = true;
      await el.updateComplete;
      await frame();
      const afterActive = liveCount();

      el.active = false;
      el.highlighted = true;
      await el.updateComplete;
      await frame();
      const afterHighlighted = liveCount();

      el.highlighted = false;
      await el.updateComplete;
      await frame();
      const afterBothClear = liveCount();

      document.body.removeChild(el);
      return { afterActive, afterHighlighted, afterBothClear };
    });

    expect(result.afterActive).toBe(1);
    expect(result.afterHighlighted).toBe(1);
    expect(result.afterBothClear).toBe(0);
  });
});
