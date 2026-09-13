import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

for (const reducedMotion of ['no-preference', 'reduce'] as const) {
  test(`custom component composes travel with its visual track (${reducedMotion})`, async ({ page }) => {
    await page.emulateMedia({ reducedMotion });
    const diagnostics = await prepareRendererFixturePage(page);
    try {
      const result = await page.evaluate(async () => {
        await import('/src/examples/custom-motion-piece.ts');
        const piece = document.createElement('example-compass-piece');
        piece.heading = 90;
        document.body.append(piece);
        await piece.updateComplete;
        const record = {
          before: { heading: 0 }, after: { heading: 90 },
          invertedTransform: 'translateX(120px)', finalTransform: '',
          beforeOpacity: '1', finalOpacity: '', needsHostTransition: true,
          durationMs: 1000,
        };
        const tracks = piece.planMotionTracks(record);
        const animations = piece.playAnimation({ ...record, tracks });
        const settled = Promise.allSettled(animations.map(animation => animation.finished));
        const inner = piece.shadowRoot!.querySelector<HTMLElement>('#inner')!;
        const frames: Array<{ x: number; angle: number }> = [];
        for (const progress of [0, 0.5, 1]) {
          for (const animation of animations) {
            animation.pause();
            const timing = animation.effect!.getComputedTiming();
            animation.currentTime = Number(timing.delay) + Number(timing.activeDuration) * progress;
          }
          await new Promise(requestAnimationFrame);
          const host = new DOMMatrixReadOnly(getComputedStyle(piece).transform);
          const visual = new DOMMatrixReadOnly(getComputedStyle(inner).transform);
          frames.push({ x: host.m41, angle: Math.atan2(visual.b, visual.a) * 180 / Math.PI });
        }
        for (const animation of animations) animation.cancel();
        await settled;
        await piece.settled();
        return {
          tracks: tracks.map(track => `${track.target}:${track.property}`), frames,
          active: piece.getAnimations({ subtree: true }).length,
          finalAngle: Math.atan2(new DOMMatrixReadOnly(getComputedStyle(inner).transform).b,
            new DOMMatrixReadOnly(getComputedStyle(inner).transform).a) * 180 / Math.PI,
        };
      });
      expect(result.tracks).toEqual(['host:transform', 'visual:transform']);
      expect(result.active).toBe(0);
      expect(result.finalAngle).toBeCloseTo(90);
      if (reducedMotion === 'no-preference') {
        expect(result.frames[0].x).toBeCloseTo(120);
        expect(result.frames[0].angle).toBeCloseTo(0);
        expect(result.frames[1].x).toBeGreaterThan(0);
        expect(result.frames[1].x).toBeLessThan(120);
        expect(result.frames[1].angle).toBeGreaterThan(0);
        expect(result.frames[1].angle).toBeLessThan(90);
      }
      diagnostics.assertEmpty();
    } finally { diagnostics.stop(); }
  });
}
