import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

/**
 * `<boardgame-pips>`, and the `hide-when-zero` half of `<boardgame-stat>`.
 *
 * They are tested together because they replace one helper each from the same
 * 87-line hand-rolled card face: a loop that concatenated one clover per luck
 * point, and a `_valueOrNothing` that returned `''` so a `Room:` label would not
 * print with nothing after it.
 */

async function setup(page: import('@playwright/test').Page) {
  const diagnostics = await prepareRendererFixturePage(page);
  await page.evaluate(async () => {
    await import('/src/client.ts');
    await customElements.whenDefined('boardgame-pips');
    await customElements.whenDefined('boardgame-stat');
  });
  return diagnostics;
}

test('a count is drawn as that many symbols', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const read = async (apply: (pips: HTMLElement & Record<string, unknown>) => void) => {
        const pips = document.createElement('boardgame-pips');
        apply(pips as never);
        document.body.append(pips);
        await pips.updateComplete;
        const root = pips.shadowRoot!;
        return {
          filled: [...root.querySelectorAll('[part="pip"]')].map(node => node.textContent),
          empty: [...root.querySelectorAll('[part="empty"]')].map(node => node.textContent),
          spoken: root.querySelector('.sr-only')!.textContent,
          decorativeHidden: !!root.querySelector('[aria-hidden="true"]'),
        };
      };
      return {
        three: await read(pips => { pips.glyph = '☘'; pips.count = 3; pips.label = 'Luck'; }),
        zero: await read(pips => { pips.glyph = '☘'; pips.count = 0; }),
        capped: await read(pips => { pips.glyph = '●'; pips.count = 2; pips.max = 5; }),
        overfull: await read(pips => { pips.glyph = '●'; pips.count = 9; pips.max = 3; }),
        fractional: await read(pips => { pips.glyph = '●'; pips.count = 2.7; }),
      };
    });

    expect(result.three.filled).toEqual(['☘', '☘', '☘']);
    expect(result.three.empty).toEqual([]);
    // The number, not "shamrock shamrock shamrock".
    expect(result.three.spoken).toBe('3 Luck');
    expect(result.three.decorativeHidden).toBe(true);

    expect(result.zero.filled).toEqual([]);
    expect(result.zero.spoken).toBe('0');

    // A capacity draws the whole row, quiet remainder included.
    expect(result.capped.filled).toHaveLength(2);
    expect(result.capped.empty).toHaveLength(3);

    // A count above the cap cannot draw more symbols than the cap has slots.
    expect(result.overfull.filled).toHaveLength(3);
    expect(result.overfull.empty).toHaveLength(0);

    expect(result.fractional.filled).toHaveLength(2);

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('hide-when-zero removes the row from layout, not merely its contents', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      // A flex row with a gap is where the difference shows: an emptied but
      // present inline-flex host still costs one gap.
      const row = document.createElement('div');
      row.style.display = 'flex';
      row.style.gap = '20px';
      row.innerHTML = '<span id="a">A</span>'
        + '<boardgame-pips id="p" glyph="☘" hide-when-zero></boardgame-pips>'
        + '<span id="b">B</span>';
      document.body.append(row);
      const pips = row.querySelector('boardgame-pips')!;
      await pips.updateComplete;
      const gapWithZero = row.querySelector('#b')!.getBoundingClientRect().left
        - row.querySelector('#a')!.getBoundingClientRect().right;

      pips.count = 2;
      await pips.updateComplete;
      const gapWithTwo = row.querySelector('#b')!.getBoundingClientRect().left
        - row.querySelector('#a')!.getBoundingClientRect().right;

      return {
        gapWithZero: Math.round(gapWithZero),
        gapWithTwo: Math.round(gapWithTwo),
        suppressedAttr: pips.hasAttribute('empty'),
        display: getComputedStyle(pips).display,
      };
    });

    // One gap, because the pips are not there at all.
    expect(result.gapWithZero).toBe(20);
    // Two gaps plus the symbols once there is something to show.
    expect(result.gapWithTwo).toBeGreaterThan(40);
    expect(result.suppressedAttr).toBe(false);
    expect(result.display).not.toBe('none');
  } finally {
    diagnostics.stop();
  }
});

test('a stat with hide-when-zero disappears at zero and stays for a capacity', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const read = async (apply: (stat: HTMLElement & Record<string, unknown>) => void) => {
        const stat = document.createElement('boardgame-stat');
        stat.hideWhenZero = true;
        stat.label = 'Room';
        apply(stat as never);
        document.body.append(stat);
        await stat.updateComplete;
        const status = stat.shadowRoot!.querySelector('boardgame-status-text');
        return {
          display: getComputedStyle(stat).display,
          width: Math.round(stat.getBoundingClientRect().width),
          // The value lives inside status-text's own shadow root, so the stat's
          // textContent would never see it; ask the composed element instead.
          value: (status as { value?: unknown } | null)?.value ?? null,
          rendered: (stat.shadowRoot!.textContent ?? '').trim(),
        };
      };
      const zero = await read(stat => { stat.value = 0; });
      const blank = await read(stat => { stat.value = ''; });
      const absent = await read(() => { /* nothing bound at all */ });
      const present = await read(stat => { stat.value = 'Parlor'; });
      const withCapacity = await read(stat => {
        stat.value = 0;
        stat.stack = {
          Deck: 'food', Indexes: [-1, -1], IDs: ['', ''], IDsLastSeen: {},
          ShuffleCount: 0, Size: 2, GameName: 'g', Components: [null, null],
        };
      });
      // Opting out is the default, and it must stay the default.
      const optedOut = document.createElement('boardgame-stat');
      optedOut.label = 'Score';
      optedOut.value = 0;
      document.body.append(optedOut);
      await optedOut.updateComplete;
      return {
        zero, blank, absent, present, withCapacity,
        optedOutDisplay: getComputedStyle(optedOut).display,
      };
    });

    expect(result.zero.display).toBe('none');
    expect(result.zero.width).toBe(0);
    expect(result.blank.display).toBe('none');
    expect(result.absent.display).toBe('none');

    expect(result.present.display).not.toBe('none');
    expect(result.present.value).toBe('Parlor');
    expect(result.present.rendered).toContain('Room');

    // `Food 0/2` is a fact worth printing; the capacity is the something to say.
    expect(result.withCapacity.display).not.toBe('none');
    expect(result.withCapacity.rendered).toContain('/2');

    // A scoreboard whose zeroes silently vanish would be worse than one with
    // zeroes in it, so nothing happens unless a game asks for it.
    expect(result.optedOutDisplay).not.toBe('none');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('a pip row is restyleable from outside without abandoning the element', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const result = await page.evaluate(async () => {
      const sheet = document.createElement('style');
      sheet.textContent = 'boardgame-pips::part(pip) { color: rgb(4, 5, 6); }'
        + 'boardgame-pips::part(empty) { color: rgb(7, 8, 9); }';
      document.head.append(sheet);
      const pips = document.createElement('boardgame-pips');
      pips.glyph = '●';
      pips.count = 1;
      pips.max = 2;
      document.body.append(pips);
      await pips.updateComplete;
      const root = pips.shadowRoot!;
      return {
        pip: getComputedStyle(root.querySelector('[part="pip"]')!).color,
        empty: getComputedStyle(root.querySelector('[part="empty"]')!).color,
      };
    });

    expect(result.pip).toBe('rgb(4, 5, 6)');
    expect(result.empty).toBe('rgb(7, 8, 9)');

    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});

test('an empty glyph is refused by name', async ({ page }) => {
  const diagnostics = await setup(page);
  try {
    const message = await page.evaluate(async () => {
      const pips = document.createElement('boardgame-pips');
      pips.glyph = '';
      pips.count = 2;
      document.body.append(pips);
      try {
        await pips.updateComplete;
      } catch (error) {
        return (error as Error).message;
      }
      return `NO THROW: ${pips.shadowRoot?.textContent}`;
    });

    expect(message).toContain('glyph must be a non-empty symbol');

    diagnostics.stop();
  } finally {
    diagnostics.stop();
  }
});
