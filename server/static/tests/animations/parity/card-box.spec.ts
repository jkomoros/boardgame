import { test, expect } from '@playwright/test';

/**
 * A CARD'S BOX IS SQUARE, AND THE CARD IS DRAWN IN TRUE PROPORTION INSIDE IT.
 *
 * Same substitution trap `token-box.spec.ts` documents, one component over.
 * `boardgame-card` wrote `--component-aspect-ratio: <aspectRatio>` inline on
 * `#outer`, and `#outer`'s own `height: var(--component-effective-height)`
 * never saw it: a custom property that REFERENCES another is substituted where
 * it is DECLARED, and `--component-effective-height` is declared at `:host`,
 * above `#outer`. Measured before the change, at `--component-width: 200px`:
 * `#outer` computed `--component-aspect-ratio: 0.6666666` and drew a 200x200
 * box, with `--component-effective-height` still resolving to
 * `calc(calc(1.0 * 200px) * 1.0)`. Real blackjack cards measured 105x105 hosts
 * around a 103x71 card.
 *
 * Unlike the token's, the rule was NOT dead -- it was doing a different job
 * under a colliding name. `#inner` reads `--component-aspect-ratio` DIRECTLY
 * rather than through `--component-effective-height`, so the inline write on
 * `#outer` inherited down and shaped the drawn CARD. It just never shaped the
 * BOX, and it silently shadowed any value an author set at `:host` -- the one
 * place `--component-aspect-ratio` is meant to be set and the one place it
 * works. Measured: a card given `--component-aspect-ratio: 1.5` on its host
 * drew a 200x300 box around a 100x66.7 card, box and art disagreeing.
 *
 * So the two meanings were split rather than either being deleted:
 * `--component-aspect-ratio` now means the BOX's ratio and only that, and the
 * card's own `aspectRatio` property publishes `--card-aspect-ratio` for the
 * art. Nothing about what is drawn changed -- the box stays square, because
 * every consumer assumes it is (the board layout's `aspect-ratio: 1` on each
 * component host, `boardgame-spatial-board.tokenPosition` centring at
 * `coords - tokenSize / 2` in both axes, the stack's spread/fan margins and
 * the FLIP scale ratio) and the card already draws in true proportion inside
 * it, exactly as the token's SVG does.
 *
 * The invariant pinned is `token-box.spec.ts`'s, extended to cards: a
 * component may not declare an aspect ratio its box does not have.
 */

const CARD_ART_RATIO = 0.6666666;

test.describe('a card\'s box', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('is square, and says so, while the card keeps its own proportions', async ({ page }) => {
    const boxes = await page.evaluate(async () => {
      await import('/src/components/boardgame-card.ts');
      document.body.innerHTML = '';

      const measure = async (configure: (el: any) => void) => {
        const el = document.createElement('boardgame-card') as any;
        el.item = { ID: 'probe' };
        el.style.cssText = 'display:block;width:200px';
        el.style.setProperty('--component-width', '200px');
        configure(el);
        document.body.appendChild(el);
        await el.updateComplete;
        await el.updateComplete;
        await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
        const outer = el.renderRoot.querySelector('#outer');
        const inner = el.renderRoot.querySelector('#inner');
        const o = outer.getBoundingClientRect();
        const i = inner.getBoundingClientRect();
        const out = {
          // What the box DECLARES, read where a reader inside the shadow tree
          // would read it.
          declared: Number(getComputedStyle(outer).getPropertyValue('--component-aspect-ratio')),
          box: o.height / o.width,
          boxWidth: o.width,
          art: i.height / i.width,
          artWidth: i.width,
        };
        el.remove();
        return out;
      };

      return {
        plain: await measure(() => {}),
        // The public API: `aspect-ratio` as an ATTRIBUTE, the spelling
        // component-view reads off a slotted front.
        steeper: await measure((el) => el.setAttribute('aspect-ratio', '1.4')),
      };
    });

    for (const [label, box] of Object.entries(boxes) as [string, any][]) {
      expect(box.boxWidth, `${label} sizes its box off --component-width`).toBeCloseTo(200, 3);
      expect(box.box, `${label}'s box is square`).toBeCloseTo(1, 6);
      // THE ASSERTION, borrowed from token-box.spec.ts: a component may not
      // claim a ratio its box does not have. Before the split this read
      // 0.6666666 against a box of 1.0.
      expect(box.declared, `${label} declares the ratio its box actually has`)
        .toBeCloseTo(box.box, 6);
    }

    // ...and the art is still shaped by the card's own ratio, which is the job
    // the colliding declaration was really doing. Not vacuous: the two cases
    // differ only in `aspect-ratio`, and only the art moves.
    expect(boxes.plain.art, 'a plain card draws at its default proportions')
      .toBeCloseTo(0.6666666, 3);
    expect(boxes.steeper.art, 'and aspect-ratio still reshapes the card')
      .toBeCloseTo(1.4, 3);
    expect(boxes.steeper.artWidth, 'without touching how wide it is drawn')
      .toBeCloseTo(boxes.plain.artWidth, 3);
  });

  /**
   * `--component-aspect-ratio` set where it is MEANT to be set -- at `:host` or
   * above -- still shapes the box, and the card no longer overrides it out
   * from under itself. This is the half that proves the split did not simply
   * disconnect the mechanism.
   */
  test('honours a host-level ratio, and still declares what it has', async ({ page }) => {
    const box = await page.evaluate(async () => {
      await import('/src/components/boardgame-card.ts');
      document.body.innerHTML = '';
      const el = document.createElement('boardgame-card') as any;
      el.item = { ID: 'probe' };
      el.style.cssText = 'display:block;width:400px';
      el.style.setProperty('--component-width', '200px');
      el.style.setProperty('--component-aspect-ratio', '1.5');
      document.body.appendChild(el);
      await el.updateComplete;
      await el.updateComplete;
      await new Promise<void>((r) => requestAnimationFrame(() => requestAnimationFrame(() => r())));
      const outer = el.renderRoot.querySelector('#outer');
      const inner = el.renderRoot.querySelector('#inner');
      const o = outer.getBoundingClientRect();
      const i = inner.getBoundingClientRect();
      const out = {
        declared: Number(getComputedStyle(outer).getPropertyValue('--component-aspect-ratio')),
        box: o.height / o.width,
        art: i.height / i.width,
      };
      el.remove();
      return out;
    });

    expect(box.box, 'a host-level ratio still shapes the box').toBeCloseTo(1.5, 3);
    // The point of the split: `#outer` used to report 0.6666666 here, having
    // overwritten the author's 1.5 for everything inside it.
    expect(box.declared, 'and the box declares the ratio it has').toBeCloseTo(1.5, 6);
    expect(box.art, 'while the card is still drawn at its own ratio')
      .toBeCloseTo(0.6666666, 3);
  });
});
