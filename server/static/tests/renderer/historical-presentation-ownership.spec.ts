import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('safe historical installations own their nodes and a superseded disposer cannot remove a newer face', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  try {
    const result = await page.evaluate(async () => {
      const { captureHistoricalPresentation, installHistoricalPresentation, historicalPresentationDisposer } =
        await import('/src/motion/historical-presentation.ts');
      const source = document.createElement('div') as HTMLDivElement & { historicalPresentationPolicy: 'clone-default-slot-safe' };
      source.historicalPresentationPolicy = 'clone-default-slot-safe';
      const target = document.createElement('div');
      target.innerHTML = '<span slot="fallback">authored fallback</span><span>authored live</span>';
      document.body.append(source, target);
      source.innerHTML = '<button id="source-identity">first face</button>';
      const first = captureHistoricalPresentation(source)!;
      installHistoricalPresentation(target, first);
      const disposeFirst = historicalPresentationDisposer(target);
      source.innerHTML = '<span>second face</span>';
      installHistoricalPresentation(target, captureHistoricalPresentation(source)!);
      const disposeSecond = historicalPresentationDisposer(target);
      disposeFirst();
      const current = target.querySelector('[slot="motion-history"]')!;
      const beforeClear = {
        text: current.textContent,
        inert: current.hasAttribute('inert'),
        ariaHidden: current.getAttribute('aria-hidden'),
        ids: target.querySelectorAll('[id]').length,
      };
      disposeSecond();
      disposeSecond();
      return { beforeClear, afterClear: target.textContent, historyCount: target.querySelectorAll('[slot="motion-history"]').length };
    });
    expect(result).toEqual({
      beforeClear: { text: 'second face', inert: true, ariaHidden: 'true', ids: 0 },
      afterClear: 'authored fallbackauthored live', historyCount: 0,
    });
    diagnostics.assertEmpty();
  } finally {
    diagnostics.stop();
  }
});
