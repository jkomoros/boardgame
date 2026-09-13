import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('component zones keep a useful flex width and shrink at a narrow viewport', async ({ page }) => {
  await page.setViewportSize({ width: 900, height: 600 });
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    await page.evaluate(async () => {
      const { cardView } = await import('/src/client.ts');
      const row = document.createElement('div');
      row.id = 'zone-row';
      Object.assign(row.style, {
        display: 'flex',
        gap: '16px',
        width: '100%',
        boxSizing: 'border-box',
      });
      const componentView = cardView({});
      for (const [index, label] of ['Draw pile', 'Discard pile'].entries()) {
        const id = `zone-card-${index}`;
        const zone = document.createElement('boardgame-component-zone');
        zone.label = label;
        zone.componentView = componentView;
        zone.layout = 'stack';
        zone.stack = {
          Deck: 'cards', Indexes: [index], IDs: [id], IDsLastSeen: {},
          ShuffleCount: 0, Size: 1, GameName: 'zone-layout',
          Components: [{
            ID: id, Index: index, Deck: 'cards', GameName: 'zone-layout', Values: {},
          }],
        } as any;
        row.append(zone);
      }
      document.body.append(row);
      const zones = [...row.querySelectorAll('boardgame-component-zone')];
      await Promise.all(zones.map(zone => zone.updateComplete));
      await Promise.all(zones.map(zone => (
        zone.shadowRoot!.querySelector('boardgame-component-stack') as HTMLElement & {
          updateComplete: Promise<unknown>;
        }
      ).updateComplete));
    });

    const measure = () => page.evaluate(() => {
      const row = document.querySelector<HTMLElement>('#zone-row')!;
      const zones = [...row.querySelectorAll<HTMLElement>('boardgame-component-zone')];
      return {
        rowWidth: row.getBoundingClientRect().width,
        rowScrollWidth: row.scrollWidth,
        widths: zones.map(zone => zone.getBoundingClientRect().width),
        cardCount: zones.map(zone => (
          zone.shadowRoot!.querySelector('boardgame-component-stack')!
            .querySelectorAll('boardgame-card').length
        )),
      };
    });

    const wide = await measure();
    await page.setViewportSize({ width: 360, height: 600 });
    const narrow = await measure();

    expect(wide.cardCount).toEqual([1, 1]);
    expect(wide.widths.every(width => width >= 190)).toBe(true);
    expect(narrow.widths.every(width => width >= 150)).toBe(true);
    expect(narrow.widths.every((width, index) => width < wide.widths[index])).toBe(true);
    expect(narrow.rowScrollWidth).toBeLessThanOrEqual(narrow.rowWidth + 1);
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
