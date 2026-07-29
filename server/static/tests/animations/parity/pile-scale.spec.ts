import { test, expect } from '@playwright/test';

/**
 * A PILE'S PADDING FOLLOWS ITS OWN SCATTER RADIUS.
 *
 * `#container.pile` is padded
 * `calc(calc(var(--pile-scale, 1.0) * 3em) + 3em)`, and `--pile-scale` is
 * supposed to be the same normalized component count that sets how far the
 * pile scatters: `_pileOffsetsForId` clamps the count to 5..25 and returns
 * `multiplier = (clamped - 5) / 20`, then spreads each component over
 * `20 + 30 * multiplier` px. Padding of `3em * (1 + multiplier)` is that
 * scatter radius expressed as room to scatter into.
 *
 * It never once fired. `_pileScaleFactor` was a plain field rather than
 * `@state`, so the `changedProperties.has('_pileScaleFactor')` test that
 * recomputed the style could never be satisfied -- Lit only puts REACTIVE
 * properties in `changedProperties`. `_style` stayed `''` for the component's
 * whole life, `--pile-scale` was never set, and every pile in every game fell
 * back to `1.0`: a flat 6em, the value meant for a 25-card pile, no matter how
 * few components it held. A whole compute chain
 * (`_pileOffsetsForId` -> `lastPileScaleFactor` -> `_pileScaleFactor` ->
 * `_computeStyle`) wrote a value nothing read.
 *
 * Making it live only ever SHRINKS padding: the dead fallback of 1.0 is the
 * maximum the live value can reach.
 */

/** The multiplier `_pileOffsetsForId` computes, restated. */
function expectedScale(numComponents: number): number {
  const clamped = Math.min(Math.max(numComponents, 5), 25);
  return (clamped - 5) / 20;
}

async function measurePile(page: any, numComponents: number) {
  return page.evaluate(async (count: number) => {
    await import('/src/components/boardgame-component-stack.ts');
    document.body.innerHTML = '';

    const stack = document.createElement('boardgame-component-stack') as any;
    stack.layout = 'pile';
    for (let i = 0; i < count; i++) {
      const component = document.createElement('boardgame-component') as any;
      component.id = `c${i}`;
      component.item = { ID: `c${i}` };
      component.setAttribute('boardgame-component', '');
      stack.appendChild(component);
    }
    document.body.appendChild(stack);
    await stack.updateComplete;
    // One more frame so the slotchange-driven relayout and the render it
    // schedules have both settled.
    await new Promise((resolve) => requestAnimationFrame(() => resolve(null)));
    await stack.updateComplete;

    const container = stack.shadowRoot.querySelector('#container') as HTMLElement;
    const style = getComputedStyle(container);
    return {
      pileScale: style.getPropertyValue('--pile-scale').trim(),
      paddingTop: parseFloat(style.paddingTop),
      fontSize: parseFloat(style.fontSize),
    };
  }, numComponents);
}

test.describe('a pile\'s --pile-scale', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('is actually set, and matches the scatter multiplier', async ({ page }) => {
    for (const count of [5, 15, 25]) {
      const measured = await measurePile(page, count);
      expect(measured.pileScale, `a ${count}-component pile must set --pile-scale`).not.toBe('');
      expect(parseFloat(measured.pileScale)).toBeCloseTo(expectedScale(count), 3);
    }
  });

  test('makes a small pile pad less than a large one', async ({ page }) => {
    const small = await measurePile(page, 5);
    const large = await measurePile(page, 25);

    // 3em * (1 + multiplier): 3em at the floor, 6em at the ceiling.
    expect(small.paddingTop).toBeCloseTo(3 * small.fontSize, 1);
    expect(large.paddingTop).toBeCloseTo(6 * large.fontSize, 1);
    expect(small.paddingTop).toBeLessThan(large.paddingTop);
  });
});
