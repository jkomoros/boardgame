import { test } from '@playwright/test';
import { createOfflineGame, expectCleanGate, gateSnapshot, waitForAnimationCounterStability } from '../helpers.js';
import { sampleMotionCurves, expectCurvesMatchGolden } from './geometry-helpers.js';

// Motion-curve parity: pins the TIMING SHAPE (easing + duration + keyframe
// structure, displacement-normalized) of every animation a scenario drives.
// This is the harness half that catches "the card still arrives, but pops
// instead of gliding" — the trace suite pins event structure, this pins how
// the motion actually progresses. See geometry-helpers.ts for why raw rect
// goldens cannot work (per-game layout randomness).
//
// Fixed viewport: normalized curves are size-independent in principle, but
// pinning it removes responsive-layout branches from the sampled set.
test.use({ viewport: { width: 1280, height: 900 } });

const PARITY_TIMEOUT_MS = 180_000;

test.describe('animation motion-curve parity', () => {
  test('debuganimations: swap flight curves', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    await createOfflineGame(page, 'debuganimations');
    const setup = await gateSnapshot(page);
    await expectCleanGate(page, setup, 60000, { allowAlreadySettled: true });
    // Setup-drain stability: the clean-gate check is point-in-time and the
    // creation pipeline can start another wave right after it passes (the
    // per-player info renderers mounting again shifted creation timing and
    // exposed exactly that race). Hold until counters are stable+balanced.
    await waitForAnimationCounterStability(page, { balance: 'plays' });
    const curves = await sampleMotionCurves(page, async () => {
      await page.locator('#shortstacks').getByRole('button', { name: 'Swap' }).click();
    });
    // Let the finished animations' settlement drain before the test ends so
    // teardown never races the gate.
    await waitForAnimationCounterStability(page, { balance: 'plays' });
    expectCurvesMatchGolden(curves, 'geometry-debuganimations-swap');
  });

  test('memory: reveal flip curves', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    // adminMode:false pins the sanitized face-down grid deterministically
    // (see trace.spec.ts memory scenario for the full race explanation).
    await createOfflineGame(page, 'memory', { adminMode: false });
    const setup = await gateSnapshot(page);
    await expectCleanGate(page, setup, 60000, { allowAlreadySettled: true });
    // Setup-drain stability: the clean-gate check is point-in-time and the
    // creation pipeline can start another wave right after it passes (the
    // per-player info renderers mounting again shifted creation timing and
    // exposed exactly that race). Hold until counters are stable+balanced.
    await waitForAnimationCounterStability(page, { balance: 'plays' });
    const curves = await sampleMotionCurves(page, async () => {
      await page.locator('boardgame-card:not([disabled])').first().click();
    });
    await waitForAnimationCounterStability(page, { balance: 'plays' });
    expectCurvesMatchGolden(curves, 'geometry-memory-reveal');
  });

  test('debuganimations: interrupted swap retarget curves', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    // Harness-critic gap 2: no other scenario ever observes an interrupted
    // glide, yet Phase 3 changes exactly what an interruption does (CSS
    // transition retargeting vs finish-then-replay). Pin the CURRENT
    // retarget behavior: fire a second Swap while the first is mid-flight
    // and fingerprint the second cycle's curves.
    await createOfflineGame(page, 'debuganimations');
    const setup = await gateSnapshot(page);
    await expectCleanGate(page, setup, 60000, { allowAlreadySettled: true });
    // Setup-drain stability: the clean-gate check is point-in-time and the
    // creation pipeline can start another wave right after it passes (the
    // per-player info renderers mounting again shifted creation timing and
    // exposed exactly that race). Hold until counters are stable+balanced.
    await waitForAnimationCounterStability(page, { balance: 'plays' });
    const swap = page.locator('#shortstacks').getByRole('button', { name: 'Swap' });
    await swap.click();
    // Wait until the first cycle's animations are genuinely mid-flight
    // (deep-walk; document.getAnimations() sees no shadow animations).
    await page.waitForFunction(() => {
      const walk = (root: Document | ShadowRoot): boolean => {
        for (const el of Array.from(root.querySelectorAll('*'))) {
          const anims = (el as Element & { getAnimations?: (o?: object) => Animation[] })
            .getAnimations?.({ subtree: false }) ?? [];
          if (anims.some((a) => a.playState === 'running'
            && typeof a.currentTime === 'number' && a.currentTime > 50)) return true;
          const sr = (el as Element & { shadowRoot: ShadowRoot | null }).shadowRoot;
          if (sr && walk(sr)) return true;
        }
        return false;
      };
      return walk(document);
    }, undefined, { timeout: 20000 });
    const curves = await sampleMotionCurves(page, async () => {
      // The animator's own interruption path (prepare() finishes the prior
      // generation) runs before the second cycle's animations start; the
      // sampler's wave loop only sees post-interrupt animations.
      await swap.evaluate((el) => (el as HTMLButtonElement).click());
    });
    await waitForAnimationCounterStability(page, { balance: 'plays' });
    expectCurvesMatchGolden(curves, 'geometry-debuganimations-interrupted-swap');
  });

  // Component-fixture curves: Phase 1 reclasses these elements onto the
  // gated play() kernel, and these goldens are the before/after parity
  // anchors for their animations (harness-critic gap 4). Driving them
  // through full game flows is nondeterministic (scoring requires a lucky
  // match; outcomes require finishing a game), so mount the components
  // directly in the served app page and trigger their animations at the
  // component contract.
  test('fixture: fading-text fade curve', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    await page.goto('/');
    await page.waitForFunction(() => (window as any).__bgAnimTestHooks !== undefined,
      undefined, { timeout: 30000 }).catch(() => { /* hooks only exist on game pages; fixture works without them */ });
    const curves = await sampleMotionCurves(page, async () => {
      await page.evaluate(async () => {
        await import('/src/components/boardgame-fading-text.ts');
        const el = document.createElement('boardgame-fading-text') as any;
        el.style.cssText = 'position:fixed;top:200px;left:200px;width:120px;height:40px;';
        el.autoMessage = 'fixed';
        el.message = 'Parity!';
        document.body.appendChild(el);
        // First trigger change from a defined previous value fires the fade.
        el.trigger = 1;
        await el.updateComplete;
        el.trigger = 2;
        await el.updateComplete;
      });
    });
    expectCurvesMatchGolden(curves, 'geometry-fixture-fading-text');
  });

  test('fixture: game-outcome arrival curve', async ({ page }) => {
    test.setTimeout(PARITY_TIMEOUT_MS);
    await page.goto('/');
    const curves = await sampleMotionCurves(page, async () => {
      await page.evaluate(async () => {
        await import('/src/components/boardgame-game-outcome.ts');
        const el = document.createElement('boardgame-game-outcome') as any;
        el.style.cssText = 'position:fixed;top:100px;left:100px;width:400px;';
        el.finished = true;
        el.winners = [0];
        document.body.appendChild(el);
        await el.updateComplete;
      });
    });
    expectCurvesMatchGolden(curves, 'geometry-fixture-game-outcome');
  });
});
