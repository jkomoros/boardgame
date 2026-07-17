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

test('shared timing keeps raw and gated card flights in the same version window', async ({ page }) => {
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-component-animator.ts');
      await import('/src/components/boardgame-animatable-item.ts');

      type FlightAnimator = HTMLElement & {
        updateComplete: Promise<unknown>;
        animationContext: {
          version: number;
          startAtMs: number;
          slotDurationMs: number;
          maxAnimationDurationMs: number;
        };
        animateBetween(
          real: HTMLElement,
          stub: HTMLElement,
          durationMs: number,
        ): Promise<void>;
      };
      type Animatable = HTMLElement & {
        updateComplete: Promise<unknown>;
        isAnimating: boolean;
        settled(): Promise<void>;
      };

      const animator = document.createElement('boardgame-component-animator') as FlightAnimator;
      const item = document.createElement('boardgame-animatable-item') as Animatable;
      const raw = document.createElement('div');
      const stub = document.createElement('div');
      for (const target of [item, raw]) {
        Object.assign(target.style, {
          position: 'fixed', left: '200px', top: '100px', width: '20px', height: '10px',
        });
      }
      Object.assign(stub.style, {
        position: 'fixed', left: '20px', top: '30px', width: '40px', height: '30px',
      });
      document.body.append(animator, item, raw, stub);
      await Promise.all([animator.updateComplete, item.updateComplete]);
      animator.animationContext = {
        version: 9,
        startAtMs: Date.now() + 300,
        slotDurationMs: 1_000,
        maxAnimationDurationMs: 600,
      };

      let willAnimate = 0;
      let animationDone = 0;
      document.addEventListener('will-animate', () => { willAnimate += 1; });
      document.addEventListener('animation-done', () => { animationDone += 1; });
      const itemFinished = animator.animateBetween(item, stub, 400);
      const rawFinished = animator.animateBetween(raw, stub, 400);
      await new Promise(requestAnimationFrame);

      const itemAnimation = item.getAnimations()[0];
      const rawAnimation = raw.getAnimations()[0];
      if (!(itemAnimation?.effect instanceof KeyframeEffect)
        || !(rawAnimation?.effect instanceof KeyframeEffect)) {
        throw new Error('both flight paths must create WAAPI animations');
      }
      const itemTiming = itemAnimation.effect.getTiming();
      const rawTiming = rawAnimation.effect.getTiming();
      const gatedDuring = item.isAnimating;
      itemAnimation.cancel();
      rawAnimation.cancel();
      await Promise.all([itemFinished, rawFinished, item.settled()]);

      return {
        itemTiming: {
          delay: itemTiming.delay,
          duration: itemTiming.duration,
          fill: itemTiming.fill,
        },
        rawTiming: {
          delay: rawTiming.delay,
          duration: rawTiming.duration,
          fill: rawTiming.fill,
        },
        gatedDuring,
        gatedAfter: item.isAnimating,
        willAnimate,
        animationDone,
      };
    });

    expect(result.itemTiming.duration).toBe(400);
    expect(result.rawTiming.duration).toBe(400);
    expect(result.itemTiming.fill).toBe('backwards');
    expect(result.rawTiming.fill).toBe('backwards');
    expect(Math.abs(result.itemTiming.delay - result.rawTiming.delay)).toBeLessThan(30);
    expect(result.gatedDuring).toBe(true);
    expect(result.gatedAfter).toBe(false);
    expect(result.willAnimate).toBe(1);
    expect(result.animationDone).toBe(1);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
