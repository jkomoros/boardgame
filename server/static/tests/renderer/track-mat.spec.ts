import { expect, test } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test('track and mat retain labelled state and fit a narrow containing surface', async ({ page }) => {
  const diagnostics = await prepareRendererFixturePage(page);
  await page.setViewportSize({ width: 1280, height: 800 });
  try {
    await page.evaluate(async () => {
      await import('/src/client.ts');
      const root = document.createElement('main');
      root.id = 'arrangements';
      root.style.cssText = 'max-width:1000px;margin:auto';
      const track = document.createElement('boardgame-track');
      track.label = 'Climate';
      track.steps = Array.from({ length: 9 }, (_, key) => ({ key, label: `Climate ${key + 1}`, color: `hsl(${220 - key * 20} 45% 85%)` }));
      track.value = 4;
      const mat = document.createElement('boardgame-mat');
      mat.label = 'Species 1';
      mat.art = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="40" height="30"%3E%3Crect width="40" height="30" fill="teal"/%3E%3C/svg%3E';
      const action = document.createElement('button');
      action.slot = 'actions'; action.textContent = 'Select';
      const stats = document.createElement('div');
      stats.textContent = 'Population 3 · Body 2 · Food 1';
      mat.append(action, stats);
      root.append(track, mat); document.body.append(root);
      await Promise.all([track.updateComplete, mat.updateComplete]);
    });
    const track = page.locator('boardgame-track');
    await expect(track.getByRole('listitem')).toHaveCount(9);
    await expect(track.locator('[aria-current="step"]')).toHaveText(/▼\s*Climate 5/);
    await expect(page.getByRole('heading', { name: 'Species 1' })).toBeVisible();
    await page.evaluate(async () => {
      const track = document.querySelector('boardgame-track')!;
      track.value = 8; await track.updateComplete;
    });
    await expect(track.locator('[aria-current="step"]')).toHaveText(/▼\s*Climate 9/);
    for (const width of [320, 200]) {
      await page.setViewportSize({ width, height: 800 });
      const geometry = await page.evaluate(() => {
        const root = document.querySelector<HTMLElement>('#arrangements')!;
        const mat = document.querySelector('boardgame-mat')!;
        return { width: root.clientWidth, scroll: root.scrollWidth,
          content: mat.shadowRoot!.querySelector('#content')!.getBoundingClientRect().width };
      });
      expect(geometry.scroll).toBeLessThanOrEqual(geometry.width + 1);
      expect(geometry.content).toBeGreaterThan(80);
    }
    await page.evaluate(async () => {
      const mat = document.querySelector('boardgame-mat')!;
      mat.art = ''; await mat.updateComplete;
      const track = document.querySelector('boardgame-track')!;
      track.value = 'unavailable'; await track.updateComplete;
    });
    await expect(page.locator('boardgame-mat img')).toHaveCount(0);
    await expect(track.locator('[aria-current]')).toHaveCount(0);
    diagnostics.assertEmpty();
  } finally { diagnostics.stop(); }
});
