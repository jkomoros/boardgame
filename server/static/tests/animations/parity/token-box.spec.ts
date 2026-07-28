import { test, expect } from '@playwright/test';

/**
 * A TOKEN'S BOX IS SQUARE, AND ITS ART IS DRAWN IN TRUE PROPORTION INSIDE IT.
 *
 * `boardgame-token` used to carry `#outer.pawn { --component-aspect-ratio: 2.0 }`
 * and `#outer.meeple { …: 1.25 }`. Neither ever applied. Custom-property
 * substitution happens where the property that USES the reference is declared,
 * and `--component-effective-height` is declared at `:host`
 * (`boardgame-component.ts`), above `#outer` -- so it was always substituted
 * with the `:host` ratio of 1.0 no matter what `#outer` said. Measured: a pawn
 * computed `--component-aspect-ratio: 2.0` on `#outer` while
 * `--component-effective-height` computed to `calc(calc(1.0 * 30px) * 1.0)` and
 * every shape drew in a 30x30 box.
 *
 * The rules were deleted rather than made to work, and the two assertions here
 * are why:
 *
 * 1. THE BOX IS THE LAYOUT CONTRACT AND EVERY CONSUMER ASSUMES IT IS SQUARE.
 *    The board layout puts `aspect-ratio: 1` on every component host,
 *    `boardgame-spatial-board.tokenPosition` centres a piece with
 *    `coords - tokenSize / 2` in BOTH axes, and the stack's spread/fan margins
 *    and the FLIP scale ratio all key off that one box. A stack-hosted
 *    component may not reserve extra space -- that is the same rule that makes
 *    a 3D token size by drawn extent instead of by circumsphere.
 *
 * 2. THE ART IS ALREADY DRAWN IN TRUE PROPORTION. An SVG's default
 *    `preserveAspectRatio` is `xMidYMid meet`, so `token_pawn.svg` (89.536 by
 *    207.215) draws at its own 0.432 inside whatever box it is given. The dead
 *    rules would not have removed that letterbox, only changed its size -- and
 *    they did not even match the assets: 2.0 against the pawn's 2.31, and 1.25
 *    against the meeple's 1.11. Honouring 2.0 was rendered and looked at: at a
 *    120px component width it draws a pawn 240px tall beside a 120px cube.
 *
 * The invariant asserted is deliberately the one BOTH resolutions satisfy: a
 * shape may not declare an aspect ratio its box does not have. A future
 * non-square token is free to make the mechanism work; it is not free to leave
 * a declaration that lies.
 */

/** width / height of each asset, from its own `viewBox`. */
const ASSET_ASPECT: Record<string, number> = {
  meeple: 146.083 / 161.979,
  pawn: 89.536 / 207.215,
};

test.describe('a token\'s box', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('is square, and says so', async ({ page }) => {
    const boxes = await page.evaluate(async () => {
      await import('/src/components/boardgame-token.ts');
      document.body.innerHTML = '';
      const out: Record<string, any> = {};
      for (const type of ['cube', 'token', 'chip', 'disc', 'meeple', 'pawn']) {
        const el = document.createElement('boardgame-token') as any;
        el.type = type;
        el.item = { ID: type };
        el.style.cssText = 'display:block';
        el.style.setProperty('--component-width', '200px');
        document.body.appendChild(el);
        await el.updateComplete;
        await el.updateComplete;
        const outer = el.renderRoot.querySelector('#outer');
        const inner = el.renderRoot.querySelector('#inner');
        const rect = inner.getBoundingClientRect();
        out[type] = {
          // What the shape DECLARES, read where the throb and the box would
          // read it: on #outer, and inherited down to #inner.
          declared: Number(getComputedStyle(outer).getPropertyValue('--component-aspect-ratio')),
          drawn: rect.height / rect.width,
          width: rect.width,
        };
        el.remove();
      }
      return out;
    });

    for (const [type, box] of Object.entries(boxes) as [string, any][]) {
      expect(box.width, `${type} sizes off --component-width`).toBeCloseTo(200, 3);
      expect(box.drawn, `${type}'s box is square`).toBeCloseTo(1, 6);
      // THE ASSERTION. A shape may not claim a ratio its box does not have --
      // the two deleted rules claimed 2.0 and 1.25 and drew 1.0.
      expect(box.declared, `${type} declares the ratio it actually draws at`)
        .toBeCloseTo(box.drawn, 6);
    }
  });

  test('lets the authored art keep its own proportions inside that square', async ({ page }) => {
    for (const [type, aspect] of Object.entries(ASSET_ASPECT)) {
      const drawn = await page.evaluate(async (shape: string) => {
        await import('/src/components/boardgame-token.ts');
        document.body.innerHTML = '';
        document.body.style.cssText = 'margin:0;background:#fff';
        const el = document.createElement('boardgame-token') as any;
        el.type = shape;
        el.item = { ID: shape };
        el.noShadow = true;
        el.style.cssText = 'position:fixed;left:0;top:0;display:block';
        el.style.setProperty('--component-width', '200px');
        document.body.appendChild(el);
        await el.updateComplete;
        await el.updateComplete;
        await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
        return true;
      }, type);
      expect(drawn).toBe(true);

      const shot = (await page.screenshot({
        clip: { x: 0, y: 0, width: 210, height: 210 },
      })).toString('base64');
      const extent = await page.evaluate(async (base64: string) => {
        const image = new Image();
        image.src = 'data:image/png;base64,' + base64;
        await image.decode();
        const canvas = document.createElement('canvas');
        canvas.width = image.width;
        canvas.height = image.height;
        const context = canvas.getContext('2d', { willReadFrequently: true })!;
        context.drawImage(image, 0, 0);
        const data = context.getImageData(0, 0, canvas.width, canvas.height).data;
        let left = Infinity; let right = -Infinity; let top = Infinity; let bottom = -Infinity;
        for (let y = 0; y < canvas.height; y++) {
          for (let x = 0; x < canvas.width; x++) {
            const p = (y * canvas.width + x) * 4;
            if (Math.hypot(255 - data[p], 255 - data[p + 1], 255 - data[p + 2]) < 120) continue;
            left = Math.min(left, x); right = Math.max(right, x);
            top = Math.min(top, y); bottom = Math.max(bottom, y);
          }
        }
        return { width: right - left + 1, height: bottom - top + 1 };
      }, shot);

      // The drawn silhouette's own proportions, against the asset's viewBox.
      // A stretched-to-fill token would land at 1.0 here; a letterboxed one
      // lands on the asset's own number.
      expect(extent.width / extent.height,
        `${type} is drawn at its asset's proportions, not stretched to the box`)
        .toBeCloseTo(aspect, 1);
      // ...and it fills the box in its long axis, so the square box is spent
      // on the piece rather than on padding.
      expect(Math.max(extent.width, extent.height),
        `${type} fills its box in the long axis`).toBeGreaterThan(190);
    }
  });
});
