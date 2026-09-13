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

test('custom view participates in real stack snapshot capture and host reuse', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { compassView } = await import('/src/examples/custom-motion-piece.ts');
      await import('/src/components/boardgame-component-animator.ts');
      const animator = document.createElement('boardgame-component-animator');
      const stack = document.createElement('boardgame-component-stack');
      stack.componentView = compassView;
      stack.layout = 'spread';
      stack.style.cssText = 'display:block;width:500px;--animation-length:100ms';
      const item = (id: string, heading: number) => ({ ID: id, Index: id === 'a' ? 0 : 1,
        Deck: 'pieces', GameName: 'compass-fixture', Values: { Heading: heading } });
      const state = (items: ReturnType<typeof item>[]) => ({ Deck: 'pieces', GameName: 'compass-fixture',
        Components: items, Indexes: items.map(piece => piece.Index), IDs: items.map(piece => piece.ID),
        IDsLastSeen: {}, ShuffleCount: 0, Size: items.length });
      stack.stack = state([item('a', 0), item('b', 0)]);
      document.body.append(animator, stack);
      await Promise.all([animator.updateComplete, stack.updateComplete]);
      await Promise.all([...stack.querySelectorAll('example-compass-piece')].map(piece => piece.updateComplete));
      const original = [...stack.querySelectorAll('example-compass-piece')];
      const records: Array<{ before: unknown; after: unknown; channels: string[] }> = [];
      for (const piece of stack.querySelectorAll('example-compass-piece')) {
        const play = piece.playAnimation.bind(piece);
        piece.playAnimation = record => {
          records.push({ before: record.before['heading'], after: record.after['heading'],
            channels: (record.tracks ?? piece.planMotionTracks(record)).map(track => `${track.target}:${track.property}`) });
          return play(record);
        };
      }
      animator.prepare();
      stack.stack = state([item('b', 0), item('a', 90)]);
      await stack.updateComplete;
      await Promise.all([...stack.querySelectorAll('example-compass-piece')].map(piece => piece.updateComplete));
      await animator.animateFlip();
      return { retained: original.every(piece => [...stack.querySelectorAll('example-compass-piece')].includes(piece)),
        heading: stack.querySelector('example-compass-piece#a')!.heading,
        records, active: document.getAnimations().length };
    });
    expect(result.retained).toBe(true);
    expect(result.heading).toBe(90);
    expect(result.records).toContainEqual({ before: 0, after: 90, channels: ['host:transform', 'visual:transform'] });
    expect(result.active).toBe(0);
    diagnostics.assertEmpty();
  } finally { diagnostics.stop(); }
});
