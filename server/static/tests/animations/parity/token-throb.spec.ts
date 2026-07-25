import { test, expect } from '@playwright/test';
import { createOfflineGame, settleInitialLoad, gateSnapshot } from '../helpers';

// deepQueryFirst walks into every shadowRoot to find the first match for
// `selector`, since a per-game renderer's boardgame-token lives several
// shadow roots deep under boardgame-app.
function deepQueryFirstScript() {
  function deepQueryFirst(root: Document | ShadowRoot | Element, selector: string): Element | null {
    const direct = root.querySelector(selector);
    if (direct) return direct;
    for (const el of Array.from(root.querySelectorAll('*'))) {
      if ((el as any).shadowRoot) {
        const found = deepQueryFirst((el as any).shadowRoot, selector);
        if (found) return found;
      }
    }
    return null;
  }
  return deepQueryFirst;
}

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

  // Regression net for the ambient-animation-sweep bug (evidence pack
  // 2026-07-26-ambient-animation-sweep.md). The infinite throb is UNGATED
  // ambient decoration -- it was never a completion-cycle participant. But a
  // token mounted inside a live render-game is registered in that game's
  // animatableRegistry, whose cycle-start reset (_resetAnimating) force-
  // finishes every registered item. If that reset uses finishAllAnimations()
  // it CANCELS the infinite throb, and because the token's active/highlighted
  // did not change on that state install, updated()->_syncThrob never re-arms
  // it -- so a highlighted token stops glowing the moment ANY move is made.
  // The retired CSS @keyframes throb (base) was class-driven and never
  // affected by a state cycle. Every OTHER throb test in this file mounts the
  // token standalone (no render-game, no registry), so this cross-cycle
  // survival was the untested gap.
  test('highlight throb survives a real render-game cycle', async ({ page }) => {
    await createOfflineGame(page, 'debuganimations');
    await settleInitialLoad(page);

    const fnSrc = `(${deepQueryFirstScript.toString()})()`;

    // Highlight the standalone #token (a local-prop demo widget in the
    // debuganimations renderer -- not game state). This is the exact
    // active/highlighted affordance every game uses.
    const highlighted = await page.evaluate((src: string) => {
      // eslint-disable-next-line no-eval
      const dq = eval(`(${src})`);
      const token = dq(document, 'boardgame-token') as any;
      if (!token) return { error: 'no token found' };
      token.highlighted = true;
      return { ok: true };
    }, fnSrc);
    expect(highlighted).toEqual({ ok: true });

    // Let updated()->_syncThrob start the infinite play.
    await page.waitForTimeout(200);

    const readThrob = async () => page.evaluate((src: string) => {
      // eslint-disable-next-line no-eval
      const dq = eval(`(${src})`);
      const token = dq(document, 'boardgame-token') as any;
      const inner = token?.shadowRoot?.querySelector('#inner') as HTMLElement | null;
      const anims = inner ? inner.getAnimations({ subtree: false }) : [];
      const infiniteRunning = anims.filter((a: Animation) =>
        a.playState === 'running'
        && (a.effect as KeyframeEffect | null)?.getComputedTiming().iterations === Infinity).length;
      return { highlighted: token ? token.highlighted : null, infiniteRunning };
    }, fnSrc);

    expect((await readThrob()).infiniteRunning, 'throb must be live before the move').toBe(1);

    // Perform a real board move -> new state -> _stateChanged ->
    // _resetAnimating registry sweep. "Public Shuffle" (VisibleShuffle) is
    // used rather than "To Hidden": a shuffle is legal in every state, whereas
    // "To Hidden" is illegal (its button disabled) whenever components are
    // already hidden -- a state-dependent flake. The shuffle still drives a
    // full gated cycle (a #fan relayout).
    const snap = await gateSnapshot(page);
    const shuffle = page.getByRole('button', { name: 'Public Shuffle', exact: true });
    await expect(shuffle).toBeEnabled();
    await shuffle.click();
    await page.waitForFunction((s: number) => {
      const h = (window as any).__bgAnimTestHooks;
      return h.gateOpens > s;
    }, snap.gateOpens, { timeout: 20000 });
    await page.waitForFunction(() => {
      const h = (window as any).__bgAnimTestHooks;
      return h.gateCloses >= h.gateOpens;
    }, undefined, { timeout: 20000 });
    await page.waitForTimeout(300);

    const after = await readThrob();
    expect(after.highlighted, 'token stays highlighted across the move').toBe(true);
    expect(after.infiniteRunning, 'the ungated throb must survive the cycle sweep').toBe(1);
  });

  // Second family member: the retired CSS throb was class-driven and survived
  // DOM reparenting automatically; the WAAPI throb is cancelled in the token's
  // disconnectedCallback and only re-armed on an active/highlighted CHANGE.
  // Lit does not re-render on a reparent, so without a connectedCallback
  // re-arm a highlighted token dragged from one container to another loses
  // its glow forever.
  test('highlight throb survives a DOM reparent', async ({ page }) => {
    await page.goto('/');
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const a = document.createElement('div');
      const b = document.createElement('div');
      document.body.append(a, b);
      const el = document.createElement('boardgame-token') as any;
      a.appendChild(el);
      await el.updateComplete;

      const infiniteRunning = () => {
        const inner = el.shadowRoot?.querySelector('#inner') as HTMLElement | null;
        return (inner?.getAnimations({ subtree: false }) ?? []).filter((anim: Animation) =>
          anim.playState === 'running'
          && (anim.effect as KeyframeEffect | null)?.getComputedTiming().iterations === Infinity).length;
      };
      const frame = () => new Promise<void>((r) => requestAnimationFrame(() => r()));

      el.highlighted = true;
      await el.updateComplete;
      await frame();
      const beforeReparent = infiniteRunning();

      // Move the still-highlighted token to a different parent (fires
      // disconnectedCallback then connectedCallback synchronously; Lit does
      // NOT re-render, so no updated() and no active/highlighted change).
      b.appendChild(el);
      await frame();
      const afterReparent = infiniteRunning();

      el.remove();
      return { beforeReparent, afterReparent };
    });
    expect(result.beforeReparent, 'throb live before reparent').toBe(1);
    expect(result.afterReparent, 'throb must survive the reparent').toBe(1);
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
