import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('animateBetween aligns differently-sized endpoints by viewport center', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const keyframes = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');

      const animator = document.createElement('boardgame-component-animator') as HTMLElement & {
        updateComplete: Promise<unknown>;
        animateBetween(
          real: HTMLElement,
          stub: HTMLElement,
          durationMs: number,
          options: { timing: 'immediate' },
        ): Promise<void>;
      };
      const real = document.createElement('div');
      const stub = document.createElement('div');
      Object.assign(real.style, {
        position: 'fixed', left: '200px', top: '100px', width: '20px', height: '10px',
      });
      Object.assign(stub.style, {
        position: 'fixed', left: '20px', top: '30px', width: '40px', height: '30px',
      });
      document.body.append(animator, real, stub);
      await animator.updateComplete;

      const finished = animator.animateBetween(real, stub, 10_000, { timing: 'immediate' });
      await new Promise(requestAnimationFrame);
      const animation = real.getAnimations()[0];
      if (!animation || !(animation.effect instanceof KeyframeEffect)) {
        throw new Error('animateBetween did not create a WAAPI keyframe animation');
      }
      const result = animation.effect.getKeyframes().map(frame => frame.transform);
      animation.finish();
      await finished;
      return result;
    });

    // real center = (210, 105), stub center = (40, 45).
    expect(keyframes[0]).toBe('translate(-170px, -60px)');
    expect(keyframes[1]).toBe('none');
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
