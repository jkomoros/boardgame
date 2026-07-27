import { test, expect } from '@playwright/test';

// Kernel-level contract for BoardgameAnimatableItem.finishGatedAnimations()
// (evidence pack 2026-07-26-ambient-animation-sweep.md). This method backs
// render-game's cycle-start registry sweep: interrupting a stale cycle must
// force-settle the cycle's GATED participants without touching UNGATED
// ambient loops (an infinite highlight throb). finishAllAnimations() keeps
// its everything semantics for the orphan/disconnect paths (an element
// leaving the tree SHOULD kill even ambient loops).
//
// This lives in the Playwright suite rather than a node --test unit file
// because it exercises real WAAPI Animation objects (finish()/cancel()/
// playState) and a LitElement subclass -- neither exists under node --test.
// A boardgame-token is the concrete BoardgameAnimatableItem instance used to
// reach the inherited play()/finishGatedAnimations()/finishAllAnimations().

test.describe('finishGatedAnimations', () => {
  test('force-settles gated animations but leaves ungated ambient loops running', async ({ page }) => {
    await page.goto('/');
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const el = document.createElement('boardgame-token') as any;
      el.style.cssText = 'position:fixed;top:200px;left:200px;';
      document.body.appendChild(el);
      await el.updateComplete;
      const inner = el.shadowRoot.querySelector('#inner') as HTMLElement;
      // The throb pulses #outer's filter (a filter on #inner would flatten a
      // component-owned 3D scene), so the ambient loop is counted there.
      const outer = el.shadowRoot.querySelector('#outer') as HTMLElement;
      const frame = () => new Promise<void>((r) => requestAnimationFrame(() => r()));

      // Ungated infinite ambient loop (the throb).
      el.highlighted = true;
      await el.updateComplete;

      // A separate GATED finite animation, started through the same play()
      // kernel with the default gated:true.
      const gatedAnim = el.play(
        inner,
        [{ opacity: '1' }, { opacity: '0.5' }],
        { duration: 5000 },
        { timing: 'immediate' },
      ) as Animation;
      await frame();

      const runningInfinite = () => outer.getAnimations({ subtree: false }).filter((a: Animation) =>
        a.playState === 'running'
        && (a.effect as KeyframeEffect | null)?.getComputedTiming().iterations === Infinity).length;

      const gatedStateBefore = gatedAnim.playState;
      const infiniteBefore = runningInfinite();
      const isAnimatingBefore = el.isAnimating; // gated play holds the item's gate

      el.finishGatedAnimations();
      await frame();

      return {
        gatedStateBefore,
        infiniteBefore,
        isAnimatingBefore,
        gatedStateAfter: gatedAnim.playState, // finish() -> 'finished'
        infiniteAfter: runningInfinite(),      // untouched -> still 1
        isAnimatingAfter: el.isAnimating,      // gated participant settled -> false
      };
    });

    expect(result.gatedStateBefore).toBe('running');
    expect(result.infiniteBefore).toBe(1);
    expect(result.isAnimatingBefore, 'a gated play holds the completion gate').toBe(true);
    expect(result.gatedStateAfter, 'gated animation must be force-finished').toBe('finished');
    expect(result.infiniteAfter, 'ungated ambient loop must be left running').toBe(1);
    expect(result.isAnimatingAfter, 'the gated participant must have settled').toBe(false);
  });

  test('finishAllAnimations still cancels ungated ambient loops (orphan/disconnect path)', async ({ page }) => {
    await page.goto('/');
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      const el = document.createElement('boardgame-token') as any;
      el.style.cssText = 'position:fixed;top:200px;left:200px;';
      document.body.appendChild(el);
      await el.updateComplete;
      const outer = el.shadowRoot.querySelector('#outer') as HTMLElement;
      const frame = () => new Promise<void>((r) => requestAnimationFrame(() => r()));

      el.highlighted = true;
      await el.updateComplete;
      await frame();
      const infiniteBefore = outer.getAnimations({ subtree: false }).filter((a: Animation) =>
        a.playState === 'running'
        && (a.effect as KeyframeEffect | null)?.getComputedTiming().iterations === Infinity).length;

      el.finishAllAnimations();
      await frame();
      const infiniteAfter = outer.getAnimations({ subtree: false }).filter((a: Animation) =>
        a.playState === 'running'
        && (a.effect as KeyframeEffect | null)?.getComputedTiming().iterations === Infinity).length;

      el.remove();
      return { infiniteBefore, infiniteAfter };
    });
    expect(result.infiniteBefore).toBe(1);
    expect(result.infiniteAfter, 'finishAllAnimations must still kill ambient loops').toBe(0);
  });
});
