import { test, expect } from '@playwright/test';
import { createOfflineGame } from './helpers';

test('play() fires will-animate/animation-done and settles', async ({ page }) => {
  await createOfflineGame(page, 'blackjack'); // any page with components registered
  const result = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    document.body.appendChild(ele);
    await ele.updateComplete;
    const events: string[] = [];
    ele.addEventListener('will-animate', () => events.push('will-animate'));
    ele.addEventListener('animation-done', () => events.push('animation-done'));
    const anim = ele.play(ele, [{ transform: 'translateX(100px)' }, { transform: 'none' }],
      { duration: 50 });
    const animatingDuring = ele.isAnimating;
    await ele.settled();
    return { events, animatingDuring, animatingAfter: ele.isAnimating, gotAnim: !!anim };
  });
  expect(result.gotAnim).toBe(true);
  expect(result.animatingDuring).toBe(true);
  expect(result.animatingAfter).toBe(false);
  expect(result.events).toEqual(['will-animate', 'animation-done']);
});

test('cancel counts as settlement; finishAllAnimations unblocks settled()', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const ok = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    document.body.appendChild(ele);
    await ele.updateComplete;
    ele.play(ele, [{ opacity: 0 }, { opacity: 1 }], { duration: 60000 });
    const p = ele.settled();
    ele.finishAllAnimations();
    await p; // must resolve, not hang or reject
    return true;
  });
  expect(ok).toBe(true);
});

test('noAnimate suppresses play; ungated play does not hold settled()', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const r = await page.evaluate(async () => {
    const ele = document.createElement('boardgame-component') as any;
    document.body.appendChild(ele);
    await ele.updateComplete;
    ele.noAnimate = true;
    const a1 = ele.play(ele, [{ opacity: 0 }, { opacity: 1 }], { duration: 50 });
    ele.noAnimate = false;
    const a2 = ele.play(ele, [{ opacity: 0 }, { opacity: 1 }],
      { duration: 60000 }, { gated: false });
    await ele.settled(); // must resolve immediately despite a2 running
    return { a1IsNull: a1 === null, a2Running: a2.playState === 'running' };
  });
  expect(r.a1IsNull).toBe(true);
  expect(r.a2Running).toBe(true);
});

test('animateBetween flies an element and resolves on settlement', async ({ page }) => {
  await createOfflineGame(page, 'blackjack');
  const r = await page.evaluate(async () => {
    const animator = document.createElement('boardgame-component-animator') as any;
    document.body.appendChild(animator);
    await animator.updateComplete;
    const a = document.createElement('div');
    const b = document.createElement('div');
    a.style.cssText = 'position:fixed;top:10px;left:10px;width:20px;height:20px';
    b.style.cssText = 'position:fixed;top:300px;left:300px;width:20px;height:20px';
    document.body.append(a, b);
    const p = animator.animateBetween(a, b, 150);
    const liveDuring = a.getAnimations().length > 0;
    const inlineTransitionDuring = a.style.transition;
    await p;
    const liveAfter = a.getAnimations().length > 0;
    return {
      liveDuring,
      liveAfter,
      inlineTransformUntouched: a.style.transform === '',
      inlineTransitionUntouched: inlineTransitionDuring === '' && a.style.transition === '',
    };
  });
  expect(r.liveDuring).toBe(true);
  expect(r.liveAfter).toBe(false);
  expect(r.inlineTransformUntouched).toBe(true);
  expect(r.inlineTransitionUntouched).toBe(true);
});
