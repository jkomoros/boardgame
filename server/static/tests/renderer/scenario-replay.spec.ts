import { expect, test } from '@playwright/test';

test('real Memory replay renders reveal, mismatch, and timer cleanup for each viewer', async ({ page }) => {
  const failures: string[] = [];
  page.on('pageerror', error => failures.push(error.message));
  await page.goto('/scenario-review.html');
  await expect(page.locator('body')).toHaveAttribute('data-frame', '0');
  for (const viewer of ['-1', '0', '1']) {
    await page.locator('#viewer').selectOption(viewer);
    for (const [frame, visible] of [0, 1, 2, 0].entries()) {
      await expect(page.locator('body')).toHaveAttribute('data-frame', String(frame));
      const cards = page.locator('boardgame-render-game-memory boardgame-component-stack').first().locator(':scope > boardgame-card[boardgame-component]');
      await expect(cards).toHaveCount(20);
      await expect.poll(() => cards.evaluateAll(elements => elements.filter(card => card.querySelector('div') !== null).length)).toBe(visible);
      await expect(page.locator('#error')).toBeEmpty();
      if (viewer === '0' && frame === 1) {
        await expect(cards.nth(0)).toHaveAttribute('aria-disabled', 'true');
        await expect(cards.nth(1)).toHaveAttribute('aria-disabled', 'false');
      }
      if (frame < 3) await page.getByRole('button', { name: 'Next', exact: true }).click();
    }
    await expect(page.getByText(viewer === "1" ? "Your turn" : "Player 2's turn", { exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Next', exact: true })).toBeDisabled();
    for (let i = 0; i < 3; i++) await page.getByRole('button', { name: 'Previous', exact: true }).click();
  }
  expect(failures).toEqual([]);
});
