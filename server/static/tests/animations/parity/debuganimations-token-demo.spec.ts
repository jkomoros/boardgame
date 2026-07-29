import { test, expect } from '@playwright/test';
import { createOfflineGame, settleInitialLoad } from '../helpers';

/**
 * THE SIX-SHAPE DEMO TOKEN IN `debuganimations` HAS TO BE ON SCREEN.
 *
 * It is the only place in the repo where a person can switch a `boardgame-token`
 * through all six shapes and ten colours by hand, and it is the widget every
 * change to that component gets looked at in. It binds `color`, `type`,
 * `active` and `highlighted` -- and no `.item`. `BoardgameComponent._itemChanged`
 * turns a null `item` into `spacer = true`, and `.spacer` is
 * `visibility: hidden`. So the demo has been invisible.
 *
 * `token-throb.spec.ts` drives this exact element and cannot see it: it asserts
 * on `#outer`'s computed `filter` and on live WAAPI animations, and neither is
 * affected by `visibility` -- a hidden element still animates and still reports
 * its filter. The assertions here are the ones visibility DOES change:
 * `visibility` itself, and whether anything reached the screen.
 */
test.describe('the debuganimations demo token', () => {
  test('stands for something, and is therefore drawn', async ({ page }) => {
    await createOfflineGame(page, 'debuganimations');
    await settleInitialLoad(page);

    const probe = await page.evaluate(() => {
      const deep = (root: Document | ShadowRoot, selector: string): Element | null => {
        const direct = root.querySelector(selector);
        if (direct) return direct;
        for (const el of root.querySelectorAll('*')) {
          if ((el as any).shadowRoot) {
            const found = deep((el as any).shadowRoot, selector);
            if (found) return found;
          }
        }
        return null;
      };
      // The demo widget lives in the renderer's own `#token` row; the game's
      // token stacks are elsewhere, so scope the search to that row.
      const row = deep(document, '#token');
      const token = row?.querySelector('boardgame-token') as any;
      if (!token) return { error: 'no demo token' };
      const outer = token.renderRoot.querySelector('#outer');
      // The demo sits well down a long page; a clip outside the viewport is
      // not a screenshot.
      token.scrollIntoView({ block: 'center' });
      const rect = token.getBoundingClientRect();
      return {
        spacer: token.spacer,
        type: token.type,
        visibility: getComputedStyle(outer).visibility,
        rect: { x: rect.x, y: rect.y, width: rect.width, height: rect.height },
      };
    });

    expect(probe.error).toBeUndefined();
    // The bug, stated as the two facts that caused it.
    expect(probe.spacer, 'a demo token stands for itself; it is not a slot-holder')
      .toBe(false);
    expect(probe.visibility, 'and so it must not be visibility: hidden')
      .toBe('visible');

    // ...and the only assertion that would survive someone "fixing" visibility
    // without the token actually arriving: pixels. Screenshot the element's own
    // rect and require it to be painted with something other than the page.
    const rect = probe.rect!;
    expect(rect.width, 'the demo token must have a box to draw in').toBeGreaterThan(4);
    const shot = (await page.screenshot({
      clip: { x: rect.x, y: rect.y, width: rect.width, height: rect.height },
    })).toString('base64');
    const painted = await page.evaluate(async (base64: string) => {
      const image = new Image();
      image.src = 'data:image/png;base64,' + base64;
      await image.decode();
      const canvas = document.createElement('canvas');
      canvas.width = image.width;
      canvas.height = image.height;
      const context = canvas.getContext('2d', { willReadFrequently: true })!;
      context.drawImage(image, 0, 0);
      const data = context.getImageData(0, 0, canvas.width, canvas.height).data;
      // The demo defaults to a red cube. Count pixels that are decisively more
      // red than they are green or blue -- the page's own surface is a neutral
      // parchment, so nothing but the token can score here.
      let red = 0;
      for (let p = 0; p < data.length; p += 4) {
        if (data[p] > 90 && data[p] - data[p + 1] > 50 && data[p] - data[p + 2] > 50) red++;
      }
      return { red, total: data.length / 4 };
    }, shot);

    expect(painted.red, 'a red demo token has to actually reach the screen')
      .toBeGreaterThan(painted.total * 0.15);
  });
});
