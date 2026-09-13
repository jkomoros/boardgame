import { test, expect } from '@playwright/test';
import { prepareRendererFixturePage } from './renderer-fixture-helpers.js';

test.use({ reducedMotion: 'no-preference' });

for (const viewport of [{ width: 1280, height: 800 }, { width: 320, height: 640 }]) test(`five-die keep/reroll example uses stable-ID selection and real zone transfers at ${viewport.width}px`, async ({ page }) => {
  await page.setViewportSize(viewport);
  const diagnostics = await prepareRendererFixturePage(page);
  await page.evaluate(async () => {
    await import('/src/examples/dice-keep-demo.ts');
    document.body.append(document.createElement('dice-keep-demo'));
  });
  const demo = page.locator('dice-keep-demo');
  const tray = demo.locator('boardgame-component-zone').filter({ has: page.getByRole('heading', { name: 'Tray', exact: true }) });
  const kept = demo.locator('boardgame-component-zone').filter({ has: page.getByRole('heading', { name: 'Kept', exact: true }) });
  await expect(tray.locator('boardgame-die:not([spacer])')).toHaveCount(5);
  await tray.locator('boardgame-die[id="die-1"]').click();
  await expect(demo.locator('p[role="status"]')).toHaveText('1 selected');
  await tray.locator('boardgame-die[id="die-3"]').focus();
  await expect(tray.locator('boardgame-die[id="die-3"]')).toBeFocused();
  await page.keyboard.press('Space');
  await expect(demo.locator('p[role="status"]')).toHaveText('2 selected');
  await demo.getByRole('button', { name: 'Keep selected', exact: true }).click();
  await expect(kept.locator('boardgame-die:not([spacer])')).toHaveCount(2);
  await expect(tray.locator('boardgame-die:not([spacer])')).toHaveCount(3);
  await expect(demo.getByRole('button', { name: 'Repeat result', exact: true })).toBeEnabled();
  const transfers = await demo.evaluate((element: any) => element.animator._solvedMotionPlan.segments.filter((s: any) => ['die-1', 'die-3'].includes(s.subjectId)).map((s: any) => ({ id: s.subjectId, path: s.path.kind, status: s.execution.status })));
  expect(transfers).toEqual([{ id:'die-1', path:'travel', status:'finished' }, { id:'die-3', path:'travel', status:'finished' }]);
  await page.screenshot({ path: test.info().outputPath('dice-demo.png') });
  await tray.locator('boardgame-die[id="die-2"]').click();
  await expect(tray.locator('boardgame-die[id="die-2"]')).toHaveAttribute('aria-pressed', 'true');
  const before = await kept.locator('boardgame-die').evaluateAll(dice => dice.map((d: any) => d.value));
  await page.evaluate(() => {
    (window as any).demoStarts = [];
    document.querySelector('dice-keep-demo')!.addEventListener('roll-start', (e: Event) => {
      (window as any).demoStarts.push(e.composedPath().find(node => node instanceof HTMLElement && node.localName === 'boardgame-die')!.id);
    });
  });
  await demo.getByRole('button', { name: 'Repeat result', exact: true }).click();
  await expect(demo.getByRole('button', { name: 'Repeat result', exact: true })).toBeEnabled();
  expect(await page.evaluate(() => (window as any).demoStarts.sort())).toEqual(['die-0', 'die-2', 'die-4']);
  expect(await kept.locator('boardgame-die').evaluateAll(dice => dice.map((d: any) => d.value))).toEqual(before);
  await expect(demo.getByRole('button', { name: 'Keep selected', exact: true })).toBeEnabled();
  const state = await demo.evaluate((element: any) => ({ selected: element.selection.draft({ candidates: element.state.tray.map((d: any) => d.ID), rebase:'keep-valid', maxSelected:20 }).selected,
    width: document.documentElement.scrollWidth }));
  expect(state.selected).toEqual(['die-2']);
  expect(state.width).toBeLessThanOrEqual(viewport.width);
  await demo.getByRole('button', { name: 'Switch 5 / 20 dice', exact: true }).click();
  await expect(tray.locator('boardgame-die:not([spacer])')).toHaveCount(20);
  await expect(demo.getByRole('button', { name: 'Repeat result', exact: true })).toBeEnabled();
  expect(await tray.locator('boardgame-die').evaluateAll(dice => dice.filter(d => d.shadowRoot!.querySelector('#inner.solid')).length)).toBe(5);
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBeLessThanOrEqual(viewport.width);
  diagnostics.assertEmpty();
});

