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

      // Re-arm BEFORE disconnecting: the disconnect assertion below must
      // prove disconnectedCallback's own cleanup path, not merely observe
      // an already-cancelled throb left over from the clear above (which
      // would make the assertion pass regardless of whether disconnect
      // cleanup exists at all).
      el.highlighted = true;
      await el.updateComplete;
      await new Promise<void>((r) => requestAnimationFrame(() => r()));
      const animsBeforeDisconnect = liveAnimations();
      // Capture the Animation object itself (runtime access to the
      // TypeScript-`private` field is unaffected -- privacy is compile-time
      // only) rather than relying on getAnimations() after removal: Chrome
      // returns an EMPTY list from getAnimations() for a disconnected
      // element's own animations regardless of whether .cancel() was ever
      // called (confirmed empirically -- an animate() call left running,
      // never cancelled, on an element that is then removed also reports
      // getAnimations().length === 0 post-removal). So the only way to
      // prove disconnectedCallback's cancel() actually fired is to inspect
      // the specific Animation instance's own `playState`, which .cancel()
      // synchronously flips to 'idle' -- independent of the target
      // element's connectedness.
      const throbBeforeDisconnect = (el as any)._throb as Animation | null;
      const throbStateBeforeDisconnect = throbBeforeDisconnect?.playState ?? null;

      document.body.removeChild(el);
      const throbStateAfterDisconnect = throbBeforeDisconnect?.playState ?? null;

      return {
        events,
        infiniteRunning,
        isAnimatingAfterStart,
        gateOpensDelta: hooks.gateOpens - gateOpensBefore,
        playsDelta: hooks.plays - playsBefore,
        animsAfterClearCount: animsAfterClear.length,
        animsBeforeDisconnectCount: animsBeforeDisconnect.length,
        throbStateBeforeDisconnect,
        throbStateAfterDisconnect,
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
    expect(result.animsBeforeDisconnectCount, 'the re-armed throb must be live immediately before disconnect').toBe(1);
    expect(result.throbStateBeforeDisconnect, 'the re-armed throb must genuinely be running before disconnect').toBe('running');
    expect(result.throbStateAfterDisconnect, 'disconnecting while throbbing must cancel() the throb (playState -> idle)').toBe('idle');
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

// Phase 1 gate regression-critic finding: under prefers-reduced-motion the
// legacy shadow-scoped CSS kept pulsing (the global reduced-motion sheet
// does not pierce the shadow root), and the raw kernel path would suppress
// the glow entirely (duration-0 infinite play renders nothing with fill
// 'none'). The declared behavior is a STATIC strong glow: no motion, but
// the highlight affordance survives.
test('reduced motion holds a static glow instead of pulsing', async ({ browser }) => {
  const context = await browser.newContext({ reducedMotion: 'reduce' });
  try {
    const page = await context.newPage();
    await page.goto('/');
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const el = document.createElement('boardgame-token') as any;
      el.style.cssText = 'position:fixed;top:200px;left:200px;';
      document.body.appendChild(el);
      await el.updateComplete;
      el.highlighted = true;
      await el.updateComplete;
      await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
      const inner = el.renderRoot.querySelector('#inner') as HTMLElement;
      // Setting the static filter starts the pre-existing 280ms CSS filter
      // TRANSITION on #inner (component base styles) — transient and fine.
      // The property under test is that no INFINITE pulse runs.
      const infiniteAnimations = inner.getAnimations()
        .filter((a) => a.effect?.getComputedTiming().iterations === Infinity).length;
      const filter = inner.style.filter;
      el.highlighted = false;
      await el.updateComplete;
      const filterCleared = inner.style.filter;
      return { infiniteAnimations, filter, filterCleared };
    });
    expect(result.infiniteAnimations).toBe(0);
    expect(result.filter).toContain('drop-shadow');
    expect(result.filterCleared).toBe('');
  } finally {
    await context.close();
  }
});
