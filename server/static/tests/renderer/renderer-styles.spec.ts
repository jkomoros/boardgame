import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

/**
 * The shared layout/state vocabulary, measured where it has to work: inside a
 * real game renderer's own shadow root, in a renderer that imports nothing.
 *
 * The audit that motivated this vocabulary found `.horizontal{display:flex;
 * flex-direction:row}` copied into five of five game renderers -- eight times
 * in debuganimations alone -- and four incompatible "this is selected"
 * languages. The vocabulary only stops that recurring if it arrives WITHOUT
 * being asked for, so the first assertion here is deliberately about delivery,
 * not about CSS: pig's renderer names no style module, and the classes must
 * still resolve inside its shadow root.
 */

/** Mount pig's fixture and read computed styles off probe nodes in its shadow root. */
async function probe(
  page: import('@playwright/test').Page,
  probes: readonly { readonly html: string; readonly read: readonly string[] }[],
): Promise<readonly Record<string, string>[]> {
  return page.evaluate(async (specs) => {
    await import('/game-src/pig/boardgame-render-game-pig.ts');
    const { pigRendererFixture } = await import(
      '/game-src/pig/boardgame-render-fixtures-pig.ts'
    );
    const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
    const handle = await mountRendererFixture(pigRendererFixture);
    const root = handle.renderer.shadowRoot;
    if (!root) throw new Error('pig renderer has no shadow root');
    return specs.map((spec) => {
      const holder = document.createElement('div');
      holder.innerHTML = spec.html;
      const node = holder.firstElementChild;
      if (!(node instanceof HTMLElement)) throw new Error(`probe produced no element: ${spec.html}`);
      root.append(node);
      const computed = getComputedStyle(node);
      const out: Record<string, string> = {};
      for (const property of spec.read) out[property] = computed.getPropertyValue(property);
      node.remove();
      return out;
    });
  }, probes);
}

test('the layout vocabulary reaches a renderer that imports no style module', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const [horizontal, vertical, center, justify, flexItem, gap, retunedGap, wrap] = await probe(page, [
      { html: '<div class="horizontal"></div>', read: ['display', 'flex-direction'] },
      { html: '<div class="vertical"></div>', read: ['display', 'flex-direction'] },
      // `.center` means align-items in every one of the copies the audit found.
      // Asserting justify-content stayed default is what makes this test able to
      // fail if the two are ever swapped.
      { html: '<div class="horizontal center"></div>', read: ['align-items', 'justify-content'] },
      { html: '<div class="horizontal justify-center"></div>', read: ['align-items', 'justify-content'] },
      { html: '<div class="flex"></div>', read: ['flex-grow'] },
      { html: '<div class="horizontal gap"></div>', read: ['column-gap'] },
      // The escape hatch: gap is the one value that genuinely varied between
      // games, so it reads a custom property rather than freezing a number.
      { html: '<div class="horizontal gap" style="--boardgame-gap: 5px"></div>', read: ['column-gap'] },
      { html: '<div class="horizontal wrap"></div>', read: ['flex-wrap'] },
    ]);

    expect(horizontal).toEqual({ display: 'flex', 'flex-direction': 'row' });
    expect(vertical).toEqual({ display: 'flex', 'flex-direction': 'column' });
    expect(center).toEqual({ 'align-items': 'center', 'justify-content': 'normal' });
    expect(justify).toEqual({ 'align-items': 'normal', 'justify-content': 'center' });
    expect(flexItem['flex-grow']).toBe('1');
    expect(gap['column-gap']).toBe('16px');
    expect(retunedGap['column-gap']).toBe('5px');
    expect(wrap['flex-wrap']).toBe('wrap');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('the state vocabulary expresses both kinds of active player, and every state is retintable', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const RING = ['outline-style', 'outline-width', 'outline-color', 'outline-offset'] as const;
    const [active, responding, selected, targetable, retinted, eliminated, gameEliminated, disabled, plain] =
      await probe(page, [
        { html: '<div class="active"></div>', read: [...RING] },
        { html: '<div class="responding"></div>', read: [...RING] },
        { html: '<div class="selected"></div>', read: [...RING] },
        { html: '<div class="targetable"></div>', read: [...RING] },
        {
          html: '<div class="active" style="--boardgame-state-active-ring: gold"></div>',
          read: ['outline-color'],
        },
        { html: '<div class="eliminated"></div>', read: ['filter'] },
        {
          html: '<div class="eliminated" style="--boardgame-state-eliminated-filter: saturate(0.5) blur(1px)"></div>',
          read: ['filter'],
        },
        { html: '<div class="disabled"></div>', read: ['pointer-events', 'cursor'] },
        { html: '<div></div>', read: [...RING, 'filter'] },
      ]);

    // Whose turn it is, and who is being asked to answer right now. These are
    // two DIFFERENT states in the wild (murdermrmonroe's `current` vs
    // `current-luck`), and a vocabulary that renders them identically
    // guarantees the second keeps being reinvented.
    expect(active['outline-style']).toBe('solid');
    expect(active['outline-width']).toBe('2px');
    expect(responding['outline-style']).toBe('solid');
    expect(responding['outline-color']).not.toBe(active['outline-color']);
    expect(selected['outline-color']).not.toBe(active['outline-color']);
    expect(selected['outline-color']).not.toBe(responding['outline-color']);

    // The ring is drawn INSIDE the element's own edge, so adopting a state
    // cannot reflow the element's neighbours -- the reason blackjack gave every
    // seat a permanent transparent border.
    expect(active['outline-offset']).toBe('-2px');

    // An invitation reads differently from a commitment.
    expect(targetable['outline-style']).toBe('dashed');

    expect(retinted['outline-color']).toBe('rgb(255, 215, 0)');

    expect(eliminated['filter']).toBe('saturate(0.35) opacity(0.55)');
    expect(gameEliminated['filter']).toBe('saturate(0.5) blur(1px)');

    expect(disabled).toEqual({ 'pointer-events': 'none', cursor: 'not-allowed' });

    // Nothing in the vocabulary is applied to an element that claims no state.
    expect(plain['outline-style']).toBe('none');
    expect(plain['filter']).toBe('none');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('pig actually adopted the vocabulary rather than merely being able to', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const row = await page.evaluate(async () => {
      await import('/game-src/pig/boardgame-render-game-pig.ts');
      const { pigRendererFixture } = await import(
        '/game-src/pig/boardgame-render-fixtures-pig.ts'
      );
      const { mountRendererFixture } = await import('/src/testing/renderer-fixture.ts');
      const handle = await mountRendererFixture(pigRendererFixture);
      const surface = handle.renderer.shadowRoot?.querySelector('boardgame-game-surface');
      const candidate = surface?.querySelector('div');
      if (!(candidate instanceof HTMLElement)) throw new Error('pig rendered no row');
      const computed = getComputedStyle(candidate);
      return {
        classes: candidate.className,
        display: computed.display,
        direction: computed.flexDirection,
        // The spacer that pushes Done to the far end of the row.
        spacerGrow: getComputedStyle(candidate.querySelector('.flex') as HTMLElement).flexGrow,
      };
    });

    // The row pig used to declare privately as `.container{display:flex;
    // flex-direction:row}` plus `.flex{flex:1}`.
    expect(row.classes).toBe('horizontal');
    expect(row.display).toBe('flex');
    expect(row.direction).toBe('row');
    expect(row.spacerGrow).toBe('1');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
