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
    await createOfflineGame(page, 'memory');
    const setup = await gateSnapshot(page);
    await expectCleanGate(page, setup, 60000, { allowAlreadySettled: true });
    const curves = await sampleMotionCurves(page, async () => {
      await page.locator('boardgame-card:not([disabled])').first().click();
    });
    await waitForAnimationCounterStability(page, { balance: 'plays' });
    expectCurvesMatchGolden(curves, 'geometry-memory-reveal');
  });
});
