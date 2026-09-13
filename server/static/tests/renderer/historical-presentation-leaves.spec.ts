import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('stat and pips copy property-only rendered scalars without retaining state', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      await import('/src/components/boardgame-stat.ts');
      await import('/src/components/boardgame-pips.ts');

      const stat = document.createElement('boardgame-stat');
      stat.label = 'Supply';
      stat.stack = {
        Deck: 'cards', Indexes: [0, 1, 2, -1, -1, -1],
        IDs: ['a', 'b', 'c', '', '', ''], Size: 6,
      } as any;
      const pips = document.createElement('boardgame-pips');
      pips.glyph = '☀';
      pips.emptyGlyph = '○';
      pips.count = 2;
      pips.max = 5;
      pips.label = 'Suns';
      document.body.append(stat, pips);
      await Promise.all([stat.updateComplete, pips.updateComplete]);

      const statClone = stat.cloneNode(true) as typeof stat;
      const pipsClone = pips.cloneNode(true) as typeof pips;
      const beforeCopy = {
        statValue: statClone.value,
        pipsCount: pipsClone.count,
      };
      stat.copyHistoricalPresentationTo(statClone);
      pips.copyHistoricalPresentationTo(pipsClone);
      // Historical presentation clones once into the cache and once again
      // for installation. Property-only values must survive both generations.
      const installedStat = statClone.cloneNode(true) as typeof stat;
      const installedPips = pipsClone.cloneNode(true) as typeof pips;
      statClone.copyHistoricalPresentationTo(installedStat);
      pipsClone.copyHistoricalPresentationTo(installedPips);
      stat.remove();
      pips.remove();
      document.body.append(installedStat, installedPips);
      await Promise.all([installedStat.updateComplete, installedPips.updateComplete]);

      const status = installedStat.shadowRoot!.querySelector('boardgame-status-text')!;
      await status.updateComplete;
      return {
        beforeCopy,
        stat: {
          label: installedStat.shadowRoot!.querySelector('.label')?.textContent?.trim(),
          value: status.shadowRoot?.textContent?.trim(),
          capacity: installedStat.shadowRoot!.querySelector('.capacity')?.textContent?.trim(),
          retainedStack: installedStat.stack,
          announce: installedStat.announce,
        },
        pips: {
          count: installedPips.shadowRoot!.querySelectorAll('.pip').length,
          empty: installedPips.shadowRoot!.querySelectorAll('.empty').length,
          spoken: installedPips.shadowRoot!.querySelector('.sr-only')?.textContent,
        },
      };
    });

    expect(result).toEqual({
      beforeCopy: { statValue: undefined, pipsCount: 0 },
      stat: {
        label: 'Supply', value: '3', capacity: '/6', retainedStack: null, announce: false,
      },
      pips: { count: 2, empty: 3, spoken: '2 Suns' },
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
